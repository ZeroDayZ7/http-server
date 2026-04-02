package config

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	fiberRedis "github.com/gofiber/storage/redis/v3"

	"github.com/zerodayz7/http-server/config/env"
	"github.com/zerodayz7/http-server/internal/errors"
)

func NewLimiter(cfg *env.Config, group string) fiber.Handler {
	presets := map[string]struct {
		Max    int
		Window time.Duration
	}{
		"global": {Max: cfg.RateLimit.Max, Window: cfg.RateLimit.Window},
		"health": {Max: 5, Window: 5 * time.Minute},
		"visits": {Max: 30, Window: 30 * time.Minute},
	}

	limit, ok := presets[group]
	if !ok {
		limit = presets["global"]
	}

	limiterConfig := limiter.Config{
		Max:        limit.Max,
		Expiration: limit.Window,
		KeyGenerator: func(c *fiber.Ctx) string {
			testKey := c.Get("X-Test-IP")
			if testKey != "" {
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

	// --- LOGIKA STORAGE ---
	if cfg.Redis.Host != "" && cfg.Redis.Host != "memory" {
		limiterConfig.Storage = fiberRedis.New(fiberRedis.Config{
			Host:     cfg.Redis.Host,
			Port:     cfg.Redis.Port,
			Password: cfg.Redis.Password,
			Database: cfg.Redis.DB,
		})
	}

	return limiter.New(limiterConfig)
}
