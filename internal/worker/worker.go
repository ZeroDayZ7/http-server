package worker

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sony/gobreaker"
	"github.com/zerodayz7/http-server/internal/service"
	"github.com/zerodayz7/http-server/internal/shared/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type InteractionWorker struct {
	redisClient     *redis.Client
	repo            service.InteractionRepository
	consumerID      string
	logger          logger.Logger
	cb              *gobreaker.CircuitBreaker
	tracer          trace.Tracer
	flushInProgress atomic.Bool
	shutdownTimeout time.Duration
	flushInterval   time.Duration

	meter         metric.Meter
	eventCounter  metric.Int64Counter
	flushDuration metric.Float64Histogram
}

func NewInteractionWorker(redisClient *redis.Client, repo service.InteractionRepository, log logger.Logger, timeout time.Duration, flushInterval time.Duration) *InteractionWorker {
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "db-interaction-repo",
		MaxRequests: 10,
		Interval:    30 * time.Second,
		Timeout:     5 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool { return counts.ConsecutiveFailures > 5 },
	})

	meter := otel.Meter("interaction-worker")
	eventCounter, _ := meter.Int64Counter("worker_events_processed_total")
	flushDuration, _ := meter.Float64Histogram("worker_flush_duration_seconds")

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
	batch := newEventBatch(w.logger)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() { defer wg.Done(); w.runStreamLoop(ctx, msgChan) }()
	go func() { defer wg.Done(); w.runReclaimLoop(ctx, msgChan) }()

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
			defer shutdownCancel()
			if batch.size() > 0 {
				w.safeFlush(shutdownCtx, batch)
			}
			w.drainPending(shutdownCtx, batch)
			wg.Wait()
			return ctx.Err()
		}
	}
}

func (w *InteractionWorker) processSingle(ctx context.Context, msg redis.XMessage, batch *eventBatch) {
	val, ok := msg.Values["type"].(string)
	if !ok || val == "" {
		w.moveToDLQ(ctx, msg)
		return
	}

	isProcessed, err := w.redisClient.SIsMember(ctx, idemKey, msg.ID).Result()
	if err != nil || isProcessed {
		if isProcessed {
			w.redisClient.XAck(ctx, streamName, groupName, msg.ID)
		}
		return
	}

	batch.add(msg.ID, val)
}

func (w *InteractionWorker) runStreamLoop(ctx context.Context, msgChan chan<- redis.XMessage) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msgs, err := w.fetchMessages(ctx)
			if err != nil {
				time.Sleep(time.Second)
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
