package config

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/zerodayz7/http-server/config/env"
	"github.com/zerodayz7/http-server/internal/errors"
)

func NewLimiter(cfg *env.Config, group string) fiber.Handler {
	presets := map[string]struct {
		Max    int
		Window time.Duration
	}{
		"global": {Max: cfg.RateLimit.Max, Window: cfg.RateLimit.Window},
		"health": {Max: 20, Window: 60 * time.Minute},
		"visits": {Max: 30, Window: 30 * time.Minute},
	}

	limit, ok := presets[group]
	if !ok {
		limit = presets["global"]
	}

	return limiter.New(limiter.Config{
		Max:        limit.Max,
		Expiration: limit.Window,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return errors.ErrTooManyRequests.
				WithDetail("group", group).
				WithDetail("ip", c.IP()).
				WithDetail("path", c.Path())
		},
	})
}
