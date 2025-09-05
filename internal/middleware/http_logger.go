package middleware

import (
	"github.com/zerodayz7/http-server/internal/shared/logger"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func LoggerMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		log := logger.GetLogger()

		requestID := c.Locals("requestid")
		reqIDStr := ""
		if requestID != nil {
			reqIDStr = requestID.(string)
		}

		log.Debug("Incoming request",
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.String("requestID", reqIDStr),
			zap.String("ip", c.IP()),
		)

		err := c.Next()

		log.Debug("Response",
			zap.Int("status", c.Response().StatusCode()),
			zap.String("requestID", reqIDStr),
		)

		return err
	}
}
