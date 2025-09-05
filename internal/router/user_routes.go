package router

import (
	"time"

	"github.com/zerodayz7/http-server/internal/handler"
	"github.com/zerodayz7/http-server/internal/shared/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"go.uber.org/zap"
)

func setupUserRoutes(app *fiber.App, h *handler.UserHandler) {
	users := app.Group("/users")

	protected := users.Group("/")
	protected.Use(limiter.New(limiter.Config{
		Max:          5,
		Expiration:   1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string { return c.IP() },
		LimitReached: func(c *fiber.Ctx) error {
			logger.GetLogger().Warn("Users modification limit exceeded", zap.String("ip", c.IP()), zap.String("path", c.Path()))
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": "Too many requests"})
		},
	}))

	// protected.Post("/", h.CreateUser)
	// protected.Patch("/:id", h.UpdateUser)
	// protected.Delete("/:id", h.DeleteUser)
}
