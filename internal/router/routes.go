package router

import (
	"github.com/zerodayz7/http-server/internal/handler"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, h *handler.UserHandler) {
	SetupHealthRoutes(app)
	setupAuthRoutes(app, h)
	setupUserRoutes(app, h)
	SetupFallbackHandlers(app)
}
