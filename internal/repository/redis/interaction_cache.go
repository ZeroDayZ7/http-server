package redisrepo

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	redisinternal "github.com/zerodayz7/http-server/internal/redis"
	"github.com/zerodayz7/http-server/internal/service"
)

type RedisInteractionCache struct {
	client *redis.Client
	keys   redisinternal.RedisKeyBuilder
}

func NewRedisInteractionCache(client *redis.Client) *RedisInteractionCache {
	return &RedisInteractionCache{
		client: client,
		keys:   redisinternal.RedisKeyBuilder{},
	}
}

func (r *RedisInteractionCache) TryRecordVisit(ctx context.Context, fp string, cooldown time.Duration) (bool, error) {
	limitKey := r.keys.VisitCooldown(fp)
	statsKey := r.keys.GlobalStats(service.TypeVisit)

	// Pass cooldown in milliseconds to Lua script
	res, err := redisinternal.DefaultScripts.Visit.Run(
		ctx, r.client, []string{limitKey, statsKey}, int(cooldown.Milliseconds()),
	).Int()

	return res == 1, err
}

func (r *RedisInteractionCache) TryRecordInteraction(ctx context.Context, fp string, typ string, cooldown time.Duration) (bool, error) {
	cooldownKey := r.keys.UserInteraction(fp)
	statsKey := r.keys.GlobalStats(typ)

	res, err := redisinternal.DefaultScripts.Interaction.Run(
		ctx, r.client, []string{cooldownKey, statsKey}, int(cooldown.Milliseconds()),
	).Int()

	return res == 1, err
}

func (r *RedisInteractionCache) GetGlobalCount(ctx context.Context, typ string) (int, bool) {
	val, err := r.client.Get(ctx, r.keys.GlobalStats(typ)).Int()
	return val, err == nil
}

func (r *RedisInteractionCache) SetGlobalCount(ctx context.Context, typ string, count int, ttl time.Duration) error {
	return r.client.Set(ctx, r.keys.GlobalStats(typ), count, ttl).Err()
}

func (r *RedisInteractionCache) GetUserChoice(ctx context.Context, fp string) (string, error) {
	return r.client.Get(ctx, r.keys.UserInteraction(fp)).Result()
}
