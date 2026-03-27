package worker

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/zerodayz7/http-server/internal/service"
	"github.com/zerodayz7/http-server/internal/shared/logger"
	"go.uber.org/zap"
)

const (
	// Redis Stream configuration for interaction events
	streamName = "interaction_events"
	dlqStream  = "interaction_events_dlq" // Dead Letter Queue for failed messages
	groupName  = "interaction_workers"

	// Retry logic and failure handling
	maxRetries = 5 // Number of processing attempts before moving a message to DLQ

	// Performance and batching thresholds
	bulkThreshold  = 50               // Message count trigger to switch from individual to bulk status checks
	flushInterval  = 10 * time.Second // Maximum time to wait before forcing a database synchronization
	readBlockTime  = 2 * time.Second
	batchSizeLimit = 100 // Batch size threshold that triggers an immediate flush to DB
	readBatchSize  = 10  // Number of messages to pull from Redis in a single XReadGroup call
)

type InteractionWorker struct {
	redisClient     *redis.Client
	repo            service.InteractionRepository
	consumerID      string
	logger          logger.Logger
	shutdownTimeout time.Duration
}

func NewInteractionWorker(redisClient *redis.Client, repo service.InteractionRepository, log logger.Logger, timeout time.Duration) *InteractionWorker {
	id := "worker-" + uuid.NewString()
	log.Info("Creating worker instance", zap.String("worker_id", id))

	return &InteractionWorker{
		redisClient:     redisClient,
		repo:            repo,
		consumerID:      id,
		logger:          log,
		shutdownTimeout: timeout,
	}
}

func (w *InteractionWorker) Start(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			w.logger.Error("Worker crashed (panic recovery)", zap.Any("panic_info", r))
		}
	}()

	w.logger.Info("Interaction worker started", zap.String("id", w.consumerID))

	batch := make(map[string]int64)
	var pendingMsgIDs []string
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	readOld := true

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Shutting down. Flushing final batch with timeout...")
			shutdownCtx, cancel := context.WithTimeout(context.Background(), w.shutdownTimeout)
			defer cancel()
			w.flush(shutdownCtx, batch, pendingMsgIDs)
			return

		case <-ticker.C:
			pendingMsgIDs = w.flush(ctx, batch, pendingMsgIDs)

		default:
		}

		lastID := ">"
		if readOld {
			lastID = "0"
		}

		msgs := w.fetchMessages(ctx, lastID)

		if readOld && len(msgs) == 0 {
			readOld = false
			continue
		}

		var retryMap map[string]int64
		if readOld && len(msgs) >= bulkThreshold {
			pending, err := w.redisClient.XPendingExt(ctx, &redis.XPendingExtArgs{
				Stream: streamName,
				Group:  groupName,
				Start:  "-",
				End:    "+",
				Count:  int64(len(msgs)),
			}).Result()

			if err == nil {
				retryMap = make(map[string]int64)
				for _, p := range pending {
					retryMap[p.ID] = p.RetryCount
				}
			}
		}

		for _, msg := range msgs {
			if readOld {
				var rCount int64
				if retryMap != nil {
					rCount = retryMap[msg.ID]
				} else {
					p, err := w.redisClient.XPendingExt(ctx, &redis.XPendingExtArgs{
						Stream: streamName,
						Group:  groupName,
						Start:  msg.ID,
						End:    msg.ID,
						Count:  1,
					}).Result()
					if err == nil && len(p) > 0 {
						rCount = p[0].RetryCount
					}
				}

				if rCount > int64(maxRetries) {
					w.logger.Error("Message exceeded max retries, moving to DLQ",
						zap.String("msg_id", msg.ID),
						zap.Int64("retries", rCount),
					)
					w.moveToDLQ(ctx, msg)
					continue
				}
			}

			eventType := w.safeString(msg.Values["type"])
			if eventType != "" {
				batch[eventType]++
				pendingMsgIDs = append(pendingMsgIDs, msg.ID)
			} else {
				w.ackMessage(ctx, msg.ID)
			}
		}

		if len(pendingMsgIDs) >= batchSizeLimit {
			pendingMsgIDs = w.flush(ctx, batch, pendingMsgIDs)
		}
	}
}

func (w *InteractionWorker) fetchMessages(ctx context.Context, lastID string) []redis.XMessage {
	streams, err := w.redisClient.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    groupName,
		Consumer: w.consumerID,
		Streams:  []string{streamName, lastID},
		Count:    readBatchSize,
		Block:    readBlockTime,
	}).Result()

	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil
		}
		if strings.Contains(err.Error(), "NOGROUP") {
			w.logger.Warn("Redis group missing, recreating...")
			w.redisClient.XGroupCreateMkStream(ctx, streamName, groupName, "0")
			return nil
		}
		w.logger.Error("XReadGroup failed", zap.Error(err))
		time.Sleep(time.Second)
		return nil
	}

	if len(streams) > 0 {
		return streams[0].Messages
	}
	return nil
}

func (w *InteractionWorker) flush(ctx context.Context, batch map[string]int64, msgIDs []string) []string {
	if len(batch) == 0 {
		return nil
	}

	success := true
	for typ, count := range batch {
		err := w.repo.IncrementBy(ctx, typ, count)
		if err != nil {
			w.logger.Error("Database sync failed",
				zap.Error(err),
				zap.String("type", typ),
				zap.Int64("count", count),
			)
			success = false
			break
		}
	}

	if !success {
		return msgIDs
	}

	for _, id := range msgIDs {
		w.ackMessage(ctx, id)
	}

	for k := range batch {
		delete(batch, k)
	}

	return nil
}

func (w *InteractionWorker) moveToDLQ(ctx context.Context, msg redis.XMessage) {
	w.logger.Warn("Moving message to DLQ", zap.String("msg_id", msg.ID))

	err := w.redisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: dlqStream,
		Values: msg.Values,
	}).Err()

	if err != nil {
		w.logger.Error("Failed to move to DLQ", zap.Error(err))
	}

	w.ackMessage(ctx, msg.ID)
}

func (w *InteractionWorker) ackMessage(ctx context.Context, msgID string) {
	if err := w.redisClient.XAck(ctx, streamName, groupName, msgID).Err(); err != nil {
		w.logger.Error("ACK error", zap.String("msg_id", msgID), zap.Error(err))
	}
}

func (w *InteractionWorker) safeString(v any) string {
	if v == nil {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		w.logger.Warn("Invalid type for string field", zap.Any("value", v))
		return ""
	}
	return s
}
