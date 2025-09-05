package router

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

func SetupHealthRoutes(app *fiber.App) {
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"time":   time.Now().Format("2006-01-02 15:04:05"),
		})
	})
}
