package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type InteractionRepository interface {
	Increment(ctx context.Context, typ string) error
}

type StreamConsumer struct {
	redis *redis.Client
	repo  InteractionRepository
}

func NewStreamConsumer(
	r *redis.Client,
	repo InteractionRepository,
) *StreamConsumer {
	return &StreamConsumer{
		redis: r,
		repo:  repo,
	}
}

func (c *StreamConsumer) Start(ctx context.Context) {

	for {

		streams, err := c.redis.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    "interaction_workers",
			Consumer: "worker_1",
			Streams:  []string{"interaction_events", ">"},
			Count:    10,
			Block:    time.Second * 2,
		}).Result()

		if err != nil {
			continue
		}

		for _, stream := range streams {
			for _, msg := range stream.Messages {

				eventType, _ := msg.Values["type"].(string)
				// eventType := fmt.Sprint(msg.Values["type"])

				err := c.repo.Increment(ctx, eventType)
				if err != nil {
					continue
				}

				_ = c.redis.XAck(
					ctx,
					"interaction_events",
					"interaction_workers",
					msg.ID,
				)
			}
		}
	}
}
