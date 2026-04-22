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

type LimitGroup string

const (
	LimitApp    LimitGroup = "app"
	LimitGlobal LimitGroup = "global"
	LimitHealth LimitGroup = "health"
	LimitVisits LimitGroup = "visits"
)

var (
	storage     fiber.Storage
	storageOnce sync.Once

	limiters     = make(map[LimitGroup]fiber.Handler)
	limitersLock sync.RWMutex
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

func GetLimiter(cfg *env.Config, group LimitGroup) fiber.Handler {

	limitersLock.RLock()
	if l, exists := limiters[group]; exists {
		limitersLock.RUnlock()
		return l
	}
	limitersLock.RUnlock()

	limitersLock.Lock()
	defer limitersLock.Unlock()

	if l, exists := limiters[group]; exists {
		return l
	}

	l := createLimiter(cfg, group)
	limiters[group] = l

	return l
}

func createLimiter(cfg *env.Config, group LimitGroup) fiber.Handler {
	presets := map[LimitGroup]struct {
		Max    int
		Window time.Duration
	}{
		LimitApp:    {Max: cfg.RateLimit.Max, Window: cfg.RateLimit.Window},
		LimitGlobal: {Max: 100, Window: 1 * time.Minute},
		LimitHealth: {Max: 50, Window: 1 * time.Minute},
		LimitVisits: {Max: 30, Window: 30 * time.Minute},
	}

	limit, ok := presets[group]
	if !ok {
		limit = presets[LimitGlobal]
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
				WithDetail("group", string(group)).
				WithDetail("ip", c.IP()).
				WithDetail("path", c.Path())
		},
	}

	return limiter.New(limiterConfig)
}
