package config

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zerodayz7/http-server/internal/shared/logger"
	"go.uber.org/zap"
)

func MustInitRedis() (*redis.Client, func()) {
	log := logger.GetLogger()
	cfg := AppConfig.Redis

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		panic(fmt.Errorf("redis connection failed: %w", err))
	}

	log.Info("Redis connected", zap.String("addr", addr))

	return client, func() {
		log.Info("Closing Redis connection...")
		_ = client.Close()
	}
}
