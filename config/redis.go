package config

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zerodayz7/http-server/internal/shared/logger"
	"go.uber.org/zap"
)

// InitRedis przyjmuje teraz config i logger jako jawne zależności
func InitRedis(ctx context.Context, cfg RedisConfig, log logger.Logger) (*redis.Client, error) {
	var addr string
	var network string

	// Logika wyboru protokołu (TCP vs UNIX Socket)
	if cfg.Port == "0" || cfg.Port == "" {
		network = "unix"
		addr = cfg.Host
	} else {
		network = "tcp"
		addr = fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	}

	client := redis.NewClient(&redis.Options{
		Network:  network,
		Addr:     addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Sprawdzenie połączenia z timeoutem
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed (network: %s, addr: %s): %w", network, addr, err)
	}

	log.Info("Redis connected",
		zap.String("network", network),
		zap.String("addr", addr),
		zap.Int("db", cfg.DB),
	)

	return client, nil
}
