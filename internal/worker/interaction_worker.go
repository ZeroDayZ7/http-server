package worker

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sony/gobreaker"
	"github.com/zerodayz7/http-server/internal/service"
	"github.com/zerodayz7/http-server/internal/shared/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

const (
	streamName     = "interaction_events"
	dlqStream      = "interaction_events_dlq"
	groupName      = "interaction_workers"
	idemKey        = "interaction_events_idem"
	idemTTL        = 24 * time.Hour
	readBlockTime  = 2 * time.Second
	readBatchSize  = 500
	batchSizeLimit = 2000

	reclaimInterval   = 10000 * time.Millisecond
	minIdleTime       = 5000 * time.Millisecond
	maxReclaimPerLoop = 500

	flushTimeout   = 5 * time.Second
	ackTimeout     = 3 * time.Second
	moveDLQTimeout = 3 * time.Second

	maxFlushRetries = 3
	maxAckRetries   = 3

	msgChanSize = batchSizeLimit * 4
)

type eventBatch struct {
	mu     sync.Mutex
	counts map[string]int64
	ids    []string
}

func newEventBatch() *eventBatch {
	return &eventBatch{
		counts: make(map[string]int64),
		ids:    make([]string, 0, batchSizeLimit),
	}
}

func (b *eventBatch) add(id string, eventType string) {
	b.mu.Lock()
	b.counts[eventType]++
	b.ids = append(b.ids, id)
	b.mu.Unlock()
}

func (b *eventBatch) snapshot() (map[string]int64, []string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.ids) == 0 {
		return nil, nil
	}

	countsCopy := make(map[string]int64, len(b.counts))
	for k, v := range b.counts {
		countsCopy[k] = v
	}

	idsCopy := make([]string, len(b.ids))
	copy(idsCopy, b.ids)

	return countsCopy, idsCopy
}

func (b *eventBatch) clear() {
	b.mu.Lock()
	b.counts = make(map[string]int64)
	b.ids = b.ids[:0]
	b.mu.Unlock()
}

func (b *eventBatch) size() int {
	b.mu.Lock()
	n := len(b.ids)
	b.mu.Unlock()
	return n
}

type InteractionWorker struct {
	redisClient *redis.Client
	repo        service.InteractionRepository
	consumerID  string
	logger      logger.Logger
	cb          *gobreaker.CircuitBreaker
	tracer      trace.Tracer

	flushInProgress atomic.Bool

	shutdownTimeout time.Duration
	flushInterval   time.Duration

	meter         metric.Meter
	eventCounter  metric.Int64Counter
	flushDuration metric.Float64Histogram
}

func NewInteractionWorker(redisClient *redis.Client, repo service.InteractionRepository, log logger.Logger, timeout time.Duration, flushInterval time.Duration) *InteractionWorker {
	// 1. Najpierw konfigurujemy Circuit Breaker
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "db-interaction-repo",
		MaxRequests: 10,
		Interval:    30 * time.Second,
		Timeout:     100 * time.Millisecond,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > 5
		},
	}) // <-- Tu kończy się Settings i Circuit Breaker

	// 2. Potem inicjalizujemy metryki (POZA Circuit Breakerem)
	meter := otel.Meter("interaction-worker")

	eventCounter, _ := meter.Int64Counter(
		"worker_events_processed_total",
		metric.WithDescription("Całkowita liczba przetworzonych zdarzeń"),
	)

	flushDuration, _ := meter.Float64Histogram(
		"worker_flush_duration_seconds",
		metric.WithDescription("Czas trwania zapisu do bazy danych"),
	)

	// 3. Na końcu budujemy i zwracamy strukturę workera
	return &InteractionWorker{
		redisClient:     redisClient,
		repo:            repo,
		consumerID:      "worker-" + uuid.NewString(),
		logger:          log,
		cb:              cb,
		tracer:          otel.Tracer("interaction-worker"),
		meter:           meter,
		eventCounter:    eventCounter,
		flushDuration:   flushDuration,
		shutdownTimeout: timeout,
		flushInterval:   flushInterval,
	}
}

func (w *InteractionWorker) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	msgChan := make(chan redis.XMessage, msgChanSize)
	batch := newEventBatch()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		w.runStreamLoop(ctx, msgChan)
	}()

	go func() {
		defer wg.Done()
		w.runReclaimLoop(ctx, msgChan)
	}()

	flushTicker := time.NewTicker(w.flushInterval)
	defer flushTicker.Stop()

	for {
		select {
		case msg := <-msgChan:
			w.processSingle(ctx, msg, batch)
			if batch.size() >= batchSizeLimit {
				w.safeFlush(ctx, batch)
			}
		case <-flushTicker.C:
			if batch.size() > 0 {
				w.safeFlush(ctx, batch)
			}
		case <-ctx.Done():
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), w.shutdownTimeout)
			if batch.size() > 0 {
				w.safeFlush(shutdownCtx, batch)
			}
			w.drainPending(shutdownCtx, msgChan)
			shutdownCancel()
			wg.Wait()
			return ctx.Err()
		}
	}
}

func (w *InteractionWorker) processSingle(ctx context.Context, msg redis.XMessage, batch *eventBatch) {
	// 1. Walidacja formatu
	val, ok := msg.Values["type"].(string)
	if !ok || val == "" {
		w.logger.Warn("invalid message format - moving to DLQ", zap.String("id", msg.ID))
		w.moveToDLQ(ctx, msg)
		return
	}

	isProcessed, err := w.redisClient.SIsMember(ctx, idemKey, msg.ID).Result()
	if err != nil {
		w.logger.Error("failed to check idempotency", zap.Error(err))
		return
	}

	if isProcessed {
		w.redisClient.XAck(ctx, streamName, groupName, msg.ID)
		return
	}

	batch.add(msg.ID, val)
}

func (w *InteractionWorker) safeFlush(ctx context.Context, batch *eventBatch) {
	counts, ids := batch.snapshot()
	if len(ids) == 0 {
		return
	}

	startTime := time.Now()

	w.flushInProgress.Store(true)
	defer w.flushInProgress.Store(false)

	flushCtx, cancel := context.WithTimeout(ctx, flushTimeout)
	defer cancel()

	flushCtx, span := w.tracer.Start(flushCtx, "FlushBatch", trace.WithAttributes(
		attribute.Int("batch_size", len(ids)),
		attribute.Int("types_count", len(counts)),
	))
	defer span.End()

	var flushErr error
	for i := 0; i < maxFlushRetries; i++ {
		_, flushErr = w.cb.Execute(func() (any, error) {
			for typ, amount := range counts {
				if err := w.repo.IncrementBy(flushCtx, typ, amount); err != nil {
					return nil, err
				}
			}
			return nil, nil
		})

		if flushErr == nil {
			break
		}

		select {
		case <-time.After(time.Duration(i+1) * 200 * time.Millisecond):
		case <-flushCtx.Done():
			w.logger.Warn("flush context cancelled during retries", zap.Error(flushCtx.Err()))
			span.RecordError(flushErr)
			return
		}
	}

	if flushErr != nil {
		w.logger.Warn("DB flush postponed - retrying in next cycle",
			zap.String("reason", flushErr.Error()),
			zap.Int("batch_size", len(ids)),
		)
		span.RecordError(flushErr)
		return
	}

	w.flushDuration.Record(flushCtx, time.Since(startTime).Seconds())
	w.eventCounter.Add(flushCtx, int64(len(ids)))

	w.markIdempotentBatch(flushCtx, ids)

	ackCtx, ackCancel := context.WithTimeout(context.Background(), ackTimeout)
	defer ackCancel()

	var ackErr error
	for i := 0; i < maxAckRetries; i++ {
		ackErr = w.ackIDs(ackCtx, ids)
		if ackErr == nil {
			break
		}

		w.logger.Warn("Redis ACK attempt failed", zap.Int("attempt", i+1), zap.Error(ackErr))

		select {
		case <-time.After(time.Duration(i+1) * 200 * time.Millisecond):
		case <-ackCtx.Done():
			w.logger.Error("ACK context timeout")
			span.RecordError(ackErr)
			return
		}
	}

	if ackErr != nil {
		w.logger.Error("Redis ACK failed final attempt", zap.Error(ackErr))
		span.RecordError(ackErr)
		return
	}

	batch.clear()
}

func (w *InteractionWorker) markIdempotentBatch(ctx context.Context, ids []string) {
	if len(ids) == 0 {
		return
	}

	pipe := w.redisClient.Pipeline()
	for _, id := range ids {
		pipe.SAdd(ctx, idemKey, id)
	}
	pipe.Expire(ctx, idemKey, idemTTL)

	_, err := pipe.Exec(ctx)
	if err != nil {
		w.logger.Error("failed to mark idempotent batch", zap.Error(err), zap.Int("count", len(ids)))
	}
}

func (w *InteractionWorker) ackIDs(ctx context.Context, ids []string) error {
	for start := 0; start < len(ids); start += 500 {
		end := start + 500
		if end > len(ids) {
			end = len(ids)
		}
		if err := w.redisClient.XAck(ctx, streamName, groupName, ids[start:end]...).Err(); err != nil {
			return err
		}
	}
	return nil
}

func (w *InteractionWorker) runStreamLoop(ctx context.Context, msgChan chan<- redis.XMessage) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		msgs, err := w.fetchMessages(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return
			}
			w.logger.Error("fetch error", zap.Error(err))
			select {
			case <-time.After(time.Second):
			case <-ctx.Done():
				return
			}
			continue
		}

		for _, m := range msgs {
			select {
			case msgChan <- m:
			case <-ctx.Done():
				return
			}
		}
	}
}

func (w *InteractionWorker) runReclaimLoop(ctx context.Context, msgChan chan<- redis.XMessage) {
	ticker := time.NewTicker(reclaimInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.reclaimPending(ctx, msgChan)
		}
	}
}

func (w *InteractionWorker) reclaimPending(ctx context.Context, msgChan chan<- redis.XMessage) {
	res, _, err := w.redisClient.XAutoClaim(ctx, &redis.XAutoClaimArgs{
		Stream:   streamName,
		Group:    groupName,
		Consumer: w.consumerID,
		MinIdle:  minIdleTime,
		Start:    "0-0",
		Count:    int64(maxReclaimPerLoop),
	}).Result()

	if err != nil {
		if errors.Is(err, redis.Nil) || strings.Contains(err.Error(), "NOGROUP") {
			return
		}
		w.logger.Error("XAUTOCLAIM failed", zap.Error(err))
		return
	}

	if len(res) == 0 {
		return
	}

	for _, m := range res {
		select {
		case msgChan <- m:
		case <-ctx.Done():
			return
		}
	}
}

func (w *InteractionWorker) fetchMessages(ctx context.Context) ([]redis.XMessage, error) {
	streams, err := w.redisClient.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    groupName,
		Consumer: w.consumerID,
		Streams:  []string{streamName, ">"},
		Count:    readBatchSize,
		Block:    readBlockTime,
	}).Result()

	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		if strings.Contains(err.Error(), "NOGROUP") {
			_ = w.redisClient.XGroupCreateMkStream(ctx, streamName, groupName, "0")
			return nil, nil
		}
		return nil, err
	}

	if len(streams) == 0 {
		// w.logger.Debug("reclaimed messages")s
		return nil, nil
	}

	return streams[0].Messages, nil
}

func (w *InteractionWorker) moveToDLQ(ctx context.Context, msg redis.XMessage) {
	ctx, cancel := context.WithTimeout(ctx, moveDLQTimeout)
	defer cancel()

	pipe := w.redisClient.Pipeline()
	pipe.XAdd(ctx, &redis.XAddArgs{Stream: dlqStream, Values: msg.Values})
	pipe.XAck(ctx, streamName, groupName, msg.ID)
	if _, err := pipe.Exec(ctx); err != nil {
		w.logger.Error("move to DLQ failed", zap.Error(err))
	}
}

func (w *InteractionWorker) drainPending(ctx context.Context, msgChan chan<- redis.XMessage) {
	for {
		pending, err := w.redisClient.XPending(ctx, streamName, groupName).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) || strings.Contains(err.Error(), "NOGROUP") {
				return
			}
			w.logger.Error("XPending in drain failed", zap.Error(err))
			return
		}

		if pending.Count == 0 {
			return
		}

		res, _, err := w.redisClient.XAutoClaim(ctx, &redis.XAutoClaimArgs{
			Stream:   streamName,
			Group:    groupName,
			Consumer: w.consumerID,
			MinIdle:  minIdleTime,
			Start:    "0-0",
			Count:    int64(maxReclaimPerLoop),
		}).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) || strings.Contains(err.Error(), "NOGROUP") {
				return
			}
			w.logger.Error("XAUTOCLAIM in drain failed", zap.Error(err))
			return
		}

		if len(res) == 0 {
			return
		}

		for _, m := range res {
			select {
			case msgChan <- m:
			case <-ctx.Done():
				return
			}
		}
	}
}
