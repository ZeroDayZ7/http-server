package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type StreamPublisher interface {
	PublishInteraction(
		ctx context.Context,
		eventType string,
		fp string,
	) error
}

type StreamProducer struct {
	redis *redis.Client
}

func NewStreamProducer(r *redis.Client) *StreamProducer {
	return &StreamProducer{redis: r}
}

func (p *StreamProducer) PublishInteraction(
	ctx context.Context,
	eventType string,
	fp string,
) error {

	_, err := p.redis.XAdd(ctx, &redis.XAddArgs{
		Stream: "interaction_events",
		Values: map[string]interface{}{
			"type": eventType,
			"fp":   fp,
			"ts":   time.Now().Unix(),
		},
	}).Result()

	return err
}
