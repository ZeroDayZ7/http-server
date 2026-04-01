package errors

import (
	stdErrors "errors"

	"github.com/gofiber/fiber/v2"
	"github.com/zerodayz7/http-server/internal/shared/logger"
	"go.uber.org/zap"
)

func ErrorHandler(log logger.Logger) fiber.ErrorHandler {
	statusMap := map[ErrorType]int{
		Validation:      fiber.StatusBadRequest,
		Unauthorized:    fiber.StatusUnauthorized,
		NotFound:        fiber.StatusNotFound,
		Internal:        fiber.StatusInternalServerError,
		BadRequest:      fiber.StatusBadRequest,
		TooManyRequests: fiber.StatusTooManyRequests,
	}

	return func(c *fiber.Ctx, err error) error {
		var appErr *AppError

		if stdErrors.As(err, &appErr) {
			status, ok := statusMap[appErr.Type]
			if !ok {
				status = fiber.StatusInternalServerError
			}

			fields := []zap.Field{
				zap.String("code", appErr.Code),
				zap.String("type", string(appErr.Type)),
				zap.String("path", c.Path()),
				zap.String("method", c.Method()),
			}

			if appErr.Meta != nil {
				fields = append(fields, zap.Any("meta", appErr.Meta))
			}

			if appErr.Err != nil {
				fields = append(fields, zap.Error(appErr.Err))
				log.Error("Application error", fields...)
			} else {
				log.Warn("Business logic warning", fields...)
			}

			return c.Status(status).JSON(fiber.Map{
				"code":    appErr.Code,
				"message": appErr.Message,
				"meta":    appErr.Meta,
			})
		}

		if e, ok := err.(*fiber.Error); ok {
			return c.Status(e.Code).JSON(fiber.Map{
				"code":    "HTTP_ERROR",
				"message": e.Message,
			})
		}

		log.Error("Uncaught server error",
			zap.Error(err),
			zap.String("path", c.Path()),
			zap.String("method", c.Method()),
		)

		internal := ErrInternal.WithErr(err)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    internal.Code,
			"message": internal.Message,
		})
	}
}
