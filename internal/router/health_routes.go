package router

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/zerodayz7/http-server/config"
	"github.com/zerodayz7/http-server/config/env"
)

func SetupHealthRoutes(app *fiber.App, cfg *env.Config) {
	health := app.Group("/health")

	health.Use(config.NewLimiter(cfg, "health"))

	health.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"time":   time.Now().Format("2006-01-02 15:04:05"),
		})
	})
}
