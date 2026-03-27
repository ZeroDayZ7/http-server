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
	streamName = "interaction_events"
	groupName  = "interaction_workers"
)

type InteractionWorker struct {
	redisClient *redis.Client
	repo        service.InteractionRepository
	consumerID  string
	logger      logger.Logger
}

func NewInteractionWorker(redisClient *redis.Client, repo service.InteractionRepository, log logger.Logger) *InteractionWorker {
	id := "worker-" + uuid.NewString()
	log.Info("Creating worker instance", zap.String("worker_id", id))

	return &InteractionWorker{
		redisClient: redisClient,
		repo:        repo,
		consumerID:  id,
		logger:      log,
	}
}

func (w *InteractionWorker) Start(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			w.logger.Error("Worker crashed (panic recovery)", zap.Any("panic_info", r))
		}
	}()

	w.logger.Info("Interaction worker started (safe batching)")

	// ZMIANA TUTAJ: int na int64
	batch := make(map[string]int64)
	var pendingMsgIDs []string

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Shutting down. Flushing final batch...")
			w.flush(context.Background(), batch, pendingMsgIDs)
			return

		case <-ticker.C:
			pendingMsgIDs = w.flush(ctx, batch, pendingMsgIDs)

		default:
			msgs := w.fetchMessages(ctx)
			if msgs == nil {
				continue
			}

			for _, msg := range msgs {
				eventType := w.safeString(msg.Values["type"])
				if eventType != "" {
					batch[eventType]++
					pendingMsgIDs = append(pendingMsgIDs, msg.ID)
				} else {
					w.ackMessage(ctx, msg.ID)
				}
			}
		}
	}
}

func (w *InteractionWorker) fetchMessages(ctx context.Context) []redis.XMessage {
	streams, err := w.redisClient.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    groupName,
		Consumer: w.consumerID,
		Streams:  []string{streamName, ">"},
		Count:    10,
		Block:    2 * time.Second,
	}).Result()

	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil
		}

		if strings.Contains(err.Error(), "NOGROUP") {
			w.logger.Warn("Redis group/stream missing, recreating...")
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
