package worker

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
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
}

func NewInteractionWorker(redisClient *goredis.Client, repo InteractionRepository) *InteractionWorker {
	id := "worker-" + uuid.NewString()

	log.Printf("[WORKER INIT] Creating worker instance ID=%s\n", id)

	return &InteractionWorker{
		redisClient: redisClient,
		repo:        repo,
		consumerID:  id,
	}
}

func (w *InteractionWorker) Start(ctx context.Context) {

	defer func() {
		if r := recover(); r != nil {
			log.Printf("[WORKER PANIC] Worker crashed: %v\n", r)
		}
	}()

	log.Printf("[WORKER START] Interaction worker started. consumerID=%s\n", w.consumerID)

	iteration := 0

	for ctx.Err() == nil {

		iteration++
		log.Printf("[WORKER LOOP] iteration=%d consumerID=%s\n", iteration, w.consumerID)

		start := time.Now()

		w.consumeBatch(ctx)

		log.Printf(
			"[WORKER LOOP DONE] iteration=%d duration=%s\n",
			iteration,
			time.Since(start),
		)
	}

	log.Println("[WORKER SHUTDOWN] Context canceled. Worker shutting down cleanly.")
}

func (w *InteractionWorker) consumeBatch(ctx context.Context) {

	log.Printf("[REDIS WAIT] Waiting for messages from stream=%s group=%s consumer=%s\n",
		streamName,
		groupName,
		w.consumerID,
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
			log.Println("[REDIS EMPTY] No messages in stream (timeout)")
			return
		}

		log.Printf("[REDIS ERROR] XReadGroup failed: %v\n", err)

		time.Sleep(time.Second)
		return
	}

	log.Printf("[REDIS RESPONSE] Streams returned=%d\n", len(streams))

	for _, stream := range streams {

		log.Printf("[STREAM] name=%s messages=%d\n",
			stream.Stream,
			len(stream.Messages),
		)

		for _, msg := range stream.Messages {

			log.Printf("[MESSAGE RECEIVED] id=%s payload=%v\n",
				msg.ID,
				msg.Values,
			)

			w.handleMessage(ctx, msg)
		}
	}
}

func (w *InteractionWorker) handleMessage(ctx context.Context, msg goredis.XMessage) {

	defer func() {
		if r := recover(); r != nil {
			log.Printf("[MESSAGE PANIC] msgID=%s panic=%v\n", msg.ID, r)
		}
	}()

	log.Printf("[MESSAGE START] Processing msgID=%s\n", msg.ID)

	eventType := safeString(msg.Values["type"])
	fp := safeString(msg.Values["fp"])

	log.Printf(
		"[MESSAGE PARSED] msgID=%s type=%s fingerprint=%s\n",
		msg.ID,
		eventType,
		fp,
	)

	if eventType == "" {

		log.Printf("[MESSAGE INVALID] Missing event type. msgID=%s\n", msg.ID)

		w.ackMessage(ctx, msg.ID)

		return
	}

	log.Printf("[DB OPERATION] Increment counter type=%s\n", eventType)

	err := w.repo.Increment(ctx, eventType)

	if err != nil {

		log.Printf(
			"[DB ERROR] Increment failed type=%s fp=%s error=%v msgID=%s\n",
			eventType,
			fp,
			err,
			msg.ID,
		)

		return
	}

	log.Printf("[DB SUCCESS] Increment OK type=%s msgID=%s\n", eventType, msg.ID)

	w.ackMessage(ctx, msg.ID)
}

func (w *InteractionWorker) ackMessage(ctx context.Context, msgID string) {

	log.Printf("[ACK START] msgID=%s\n", msgID)

	err := w.redisClient.XAck(ctx, streamName, groupName, msgID).Err()

	if err != nil {

		log.Printf("[ACK ERROR] msgID=%s error=%v\n", msgID, err)
		return
	}

	log.Printf("[ACK SUCCESS] msgID=%s\n", msgID)
}

func safeString(v any) string {

	if v == nil {
		return ""
	}

	s, ok := v.(string)

	if !ok {
		log.Printf("[TYPE WARNING] value is not string: %T\n", v)
		return ""
	}

	return s
}
