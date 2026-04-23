package worker

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func (w *InteractionWorker) safeFlush(ctx context.Context, batch *eventBatch) {
	counts, ids := batch.snapshot()
	if len(ids) == 0 {
		return
	}

	w.flushInProgress.Store(true)
	defer w.flushInProgress.Store(false)

	startTime := time.Now()
	flushCtx, cancel := context.WithTimeout(ctx, flushTimeout)
	defer cancel()

	flushCtx, span := w.tracer.Start(flushCtx, "FlushBatch", trace.WithAttributes(
		attribute.Int("batch_size", len(ids)),
		attribute.Int("types_count", len(counts)),
	))
	defer span.End()

	var flushErr error
	for i := range maxFlushRetries {
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
			return
		}
	}

	if flushErr != nil {
		w.logger.Warn("DB flush failed - requeueing batch", zap.Error(flushErr), zap.Int("batch_size", len(ids)))
		return
	}

	batch.clear()
	w.flushDuration.Record(flushCtx, time.Since(startTime).Seconds())
	w.eventCounter.Add(flushCtx, int64(len(ids)))
	w.markIdempotentBatch(flushCtx, ids)

	// ACK Logic
	ackCtx, ackCancel := context.WithTimeout(context.Background(), ackTimeout)
	defer ackCancel()

	for i := range maxAckRetries {
		if err := w.ackIDs(ackCtx, ids); err == nil {
			break
		}
		select {
		case <-time.After(time.Duration(i+1) * 200 * time.Millisecond):
		case <-ackCtx.Done():
			return
		}
	}
}

func (w *InteractionWorker) drainPending(ctx context.Context, batch *eventBatch) {
	for {
		pending, err := w.redisClient.XPending(ctx, streamName, groupName).Result()
		if err != nil || pending.Count == 0 {
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

		if err != nil || len(res) == 0 {
			return
		}

		for _, m := range res {
			w.processSingle(ctx, m, batch)
			if batch.size() >= batchSizeLimit {
				w.safeFlush(ctx, batch)
			}
		}
	}
}
