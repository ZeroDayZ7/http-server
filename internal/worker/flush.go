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
	w.logger.Info("🚀 Starting batch flush to database",
		zap.Int("batch_size", len(ids)),
		zap.Any("event_counts", counts),
	)

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
					w.logger.Warn("Partial flush increment failed",
						zap.String("type", typ),
						zap.Int64("amount", amount),
						zap.Error(err),
						zap.Int("retry_attempt", i+1),
					)
					return nil, err
				}
			}
			return nil, nil
		})

		if flushErr == nil {
			break
		}

		w.logger.Warn("Batch flush attempt failed - retrying",
			zap.Int("attempt", i+1),
			zap.Error(flushErr),
		)

		select {
		case <-time.After(time.Duration(i+1) * 200 * time.Millisecond):
		case <-flushCtx.Done():
			w.logger.Error("Flush context cancelled during retries - DATA STUCK IN BATCH",
				zap.Error(flushCtx.Err()),
				zap.Int("batch_size", len(ids)),
			)
			return
		}
	}

	if flushErr != nil {
		w.logger.Error("🚨 CRITICAL: DB flush failed after all retries. Batch kept in memory.",
			zap.Error(flushErr),
			zap.Int("batch_size", len(ids)),
		)
		return
	}

	// Success! Clear batch and log metrics
	duration := time.Since(startTime)
	batch.clear()

	w.logger.Info("✅ Successfully flushed batch to MySQL",
		zap.Int("batch_size", len(ids)),
		zap.Duration("duration", duration),
	)

	w.flushDuration.Record(flushCtx, duration.Seconds())
	w.eventCounter.Add(flushCtx, int64(len(ids)))
	w.markIdempotentBatch(flushCtx, ids)

	// ACK Logic with Logging
	ackCtx, ackCancel := context.WithTimeout(context.Background(), ackTimeout)
	defer ackCancel()

	w.logger.Debug("Sending XACK for processed IDs", zap.Int("count", len(ids)))

	var ackErr error
	for i := range maxAckRetries {
		if ackErr = w.ackIDs(ackCtx, ids); ackErr == nil {
			w.logger.Info("Successfully acknowledged messages in Redis", zap.Int("count", len(ids)))
			break
		}

		w.logger.Warn("XACK attempt failed", zap.Int("attempt", i+1), zap.Error(ackErr))

		select {
		case <-time.After(time.Duration(i+1) * 200 * time.Millisecond):
		case <-ackCtx.Done():
			w.logger.Error("ACK context expired - messages might be re-processed by XAUTOCLAIM")
			return
		}
	}
}

func (w *InteractionWorker) drainPending(ctx context.Context, batch *eventBatch) {
	w.logger.Info("Starting drain of pending messages (XPENDING)...")

	for {
		pending, err := w.redisClient.XPending(ctx, streamName, groupName).Result()
		if err != nil {
			if err != redis.Nil {
				w.logger.Error("XPENDING check failed", zap.Error(err))
			}
			return
		}

		if pending.Count == 0 {
			w.logger.Info("No more pending messages to drain")
			return
		}

		w.logger.Info("Found pending messages, attempting to reclaim", zap.Int64("count", pending.Count))

		res, _, err := w.redisClient.XAutoClaim(ctx, &redis.XAutoClaimArgs{
			Stream:   streamName,
			Group:    groupName,
			Consumer: w.consumerID,
			MinIdle:  minIdleTime,
			Start:    "0-0",
			Count:    int64(maxReclaimPerLoop),
		}).Result()

		if err != nil {
			w.logger.Error("XAutoClaim failed during drain", zap.Error(err))
			return
		}

		if len(res) == 0 {
			w.logger.Info("XAutoClaim returned no messages - waiting for minIdleTime")
			return
		}

		w.logger.Info("Reclaimed messages for processing", zap.Int("count", len(res)))

		for _, m := range res {
			w.processSingle(ctx, m, batch)
			if batch.size() >= batchSizeLimit {
				w.logger.Info("Batch limit reached during drain - triggering flush")
				w.safeFlush(ctx, batch)
			}
		}
	}
}
