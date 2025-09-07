package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zerodayz7/http-server/config"
	"github.com/zerodayz7/http-server/internal/handler"
)

func setupUserRoutes(app *fiber.App, h *handler.UserHandler) {
	_ = h

	users := app.Group("/users")
	protected := users.Group("/")
	protected.Use(config.NewLimiter("users"))

	// protected.Post("/", h.CreateUser)
	// protected.Patch("/:id", h.UpdateUser)
	// protected.Delete("/:id", h.DeleteUser)
}
