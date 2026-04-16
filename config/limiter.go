package config

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	fiberRedis "github.com/gofiber/storage/redis/v3"

	"github.com/zerodayz7/http-server/config/env"
	"github.com/zerodayz7/http-server/internal/errors"
)

var (
	storage     fiber.Storage
	storageOnce sync.Once
)

func getStorage(cfg *env.Config) fiber.Storage {
	if cfg.Redis.Host == "" || cfg.Redis.Host == "memory" {
		return nil
	}

	storageOnce.Do(func() {
		storage = fiberRedis.New(fiberRedis.Config{
			Host:     cfg.Redis.Host,
			Port:     cfg.Redis.Port,
			Password: cfg.Redis.Password,
			Database: cfg.Redis.DB,
		})
	})

	return storage
}

func NewLimiter(cfg *env.Config, group string) fiber.Handler {
	presets := map[string]struct {
		Max    int
		Window time.Duration
	}{
		"global": {Max: cfg.RateLimit.Max, Window: cfg.RateLimit.Window},
		"health": {Max: 50, Window: 1 * time.Minute},
		"visits": {Max: 30, Window: 30 * time.Minute},
	}

	limit, ok := presets[group]
	if !ok {
		limit = presets["global"]
	}

	limiterConfig := limiter.Config{
		Max:        limit.Max,
		Expiration: limit.Window,
		Storage:    getStorage(cfg),
		KeyGenerator: func(c *fiber.Ctx) string {
			if testKey := c.Get("X-Test-IP"); testKey != "" {
				return testKey
			}
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return errors.ErrTooManyRequests.
				WithDetail("group", group).
				WithDetail("ip", c.IP()).
				WithDetail("path", c.Path())
		},
	}

	return limiter.New(limiterConfig)
}
