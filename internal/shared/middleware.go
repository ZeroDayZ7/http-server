package shared

import (
	stdErrors "errors"
	"time"

	"github.com/gofiber/fiber/v2"
	apperrors "github.com/zerodayz7/http-server/internal/errors"
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

		if err != nil {
			var appErr *apperrors.AppError
			if stdErrors.As(err, &appErr) {
				switch appErr.Type {
				case apperrors.Validation, apperrors.BadRequest:
					status = fiber.StatusBadRequest
				case apperrors.Unauthorized:
					status = fiber.StatusUnauthorized
				case apperrors.NotFound:
					status = fiber.StatusNotFound
				default:
					status = fiber.StatusInternalServerError
				}
			} else if fe, ok := err.(*fiber.Error); ok {
				status = fe.Code
			} else {
				status = fiber.StatusInternalServerError
			}
		}

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
