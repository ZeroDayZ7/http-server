package shared

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/zerodayz7/http-server/internal/shared/logger"
	"go.uber.org/zap"
)

func RequestLoggerMiddleware(log logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Method() == fiber.MethodOptions {
			return c.Next()
		}

		start := time.Now()

		log.Debug("Request started",
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
		)

		err := c.Next()

		latency := time.Since(start)
		status := c.Response().StatusCode()

		log.Info("Request Processed",
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("ip", c.IP()),
		)

		return err
	}
}
