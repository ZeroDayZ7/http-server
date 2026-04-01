package worker

import (
	"context"
	"errors"
	"strings"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

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
		return nil, nil
	}

	return streams[0].Messages, nil
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

	for _, m := range res {
		select {
		case msgChan <- m:
		case <-ctx.Done():
			return
		}
	}
}
