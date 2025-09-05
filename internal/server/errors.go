package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zerodayz7/http-server/internal/shared/logger"
	"go.uber.org/zap"
)

func ErrorHandler() fiber.ErrorHandler {
	log := logger.GetLogger()
	return func(c *fiber.Ctx, err error) error {
		if e, ok := err.(*fiber.Error); ok {
			log.Error("HTTP error", zap.Error(err), zap.String("path", c.Path()))
			return c.Status(e.Code).JSON(fiber.Map{"error": e.Message})
		}
		log.Error("Server error", zap.Error(err), zap.String("path", c.Path()))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}
}
