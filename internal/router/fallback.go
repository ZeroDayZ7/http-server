package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zerodayz7/http-server/internal/shared/logger"
	"go.uber.org/zap"
)

func SetupFallbackHandlers(app *fiber.App, log logger.Logger) {
	// Obsługa favicon - zwracamy 204 No Content, aby nie zaśmiecać logów 404
	app.Get("/favicon.ico", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	// Globalny fallback dla nieistniejących ścieżek (404)
	app.Use(func(c *fiber.Ctx) error {
		log.Warn("404 - not found",
			zap.String("path", c.Path()),
			zap.String("method", c.Method()),
			zap.String("ip", c.IP()),
		)

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Resource not found",
		})
	})
}
