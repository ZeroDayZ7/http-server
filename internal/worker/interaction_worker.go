package worker

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
	"github.com/zerodayz7/http-server/internal/shared/logger"
	"go.uber.org/zap"
)

const (
	streamName = "interaction_events"
	groupName  = "interaction_workers"
)

type InteractionRepository interface {
	Increment(ctx context.Context, typ string) error
}

type InteractionWorker struct {
	redisClient *goredis.Client
	repo        InteractionRepository
	consumerID  string
	logger      logger.Logger
}

func NewInteractionWorker(redisClient *goredis.Client, repo InteractionRepository, log logger.Logger) *InteractionWorker {
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

	w.logger.Info("Interaction worker started")

	iteration := 0
	for ctx.Err() == nil {
		iteration++
		w.logger.Debug("Worker loop iteration", zap.Int("iteration", iteration))

		start := time.Now()
		w.consumeBatch(ctx)

		w.logger.Debug("Worker loop done",
			zap.Int("iteration", iteration),
			zap.Duration("duration", time.Since(start)),
		)
	}

	w.logger.Info("Context canceled. Worker shutting down cleanly.")
}

func (w *InteractionWorker) consumeBatch(ctx context.Context) {
	w.logger.Debug("Waiting for messages",
		zap.String("stream", streamName),
		zap.String("group", groupName),
	)

	streams, err := w.redisClient.XReadGroup(ctx, &goredis.XReadGroupArgs{
		Group:    groupName,
		Consumer: w.consumerID,
		Streams:  []string{streamName, ">"},
		Count:    10,
		Block:    time.Second * 2,
	}).Result()

	if err != nil {
		if errors.Is(err, goredis.Nil) {
			w.logger.Debug("No messages in stream (timeout)")
			return
		}
		w.logger.Error("XReadGroup failed", zap.Error(err))
		time.Sleep(time.Second)
		return
	}

	for _, stream := range streams {
		w.logger.Info("Batch received",
			zap.String("stream", stream.Stream),
			zap.Int("count", len(stream.Messages)),
		)

		for _, msg := range stream.Messages {
			w.handleMessage(ctx, msg)
		}
	}
}

func (w *InteractionWorker) handleMessage(ctx context.Context, msg goredis.XMessage) {
	defer func() {
		if r := recover(); r != nil {
			w.logger.Error("Message processing panic",
				zap.String("msg_id", msg.ID),
				zap.Any("panic_info", r),
			)
		}
	}()

	eventType := w.safeString(msg.Values["type"])
	fp := w.safeString(msg.Values["fp"])

	if eventType == "" {
		w.logger.Warn("Missing event type, skipping and acking", zap.String("msg_id", msg.ID))
		w.ackMessage(ctx, msg.ID)
		return
	}

	err := w.repo.Increment(ctx, eventType)
	if err != nil {
		w.logger.Error("Increment failed",
			zap.String("type", eventType),
			zap.String("fp", fp),
			zap.Error(err),
			zap.String("msg_id", msg.ID),
		)
		return
	}

	w.logger.Debug("Increment OK", zap.String("type", eventType), zap.String("msg_id", msg.ID))
	w.ackMessage(ctx, msg.ID)
}

func (w *InteractionWorker) ackMessage(ctx context.Context, msgID string) {
	err := w.redisClient.XAck(ctx, streamName, groupName, msgID).Err()
	if err != nil {
		w.logger.Error("ACK error", zap.String("msg_id", msgID), zap.Error(err))
		return
	}
}

func (w *InteractionWorker) safeString(v any) string {
	if v == nil {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		w.logger.Warn("Value is not a string", zap.Any("type", v))
		return ""
	}
	return s
}
