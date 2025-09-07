package router

import (
	"github.com/zerodayz7/http-server/internal/handler"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, authHandler *handler.AuthHandler, userHandler *handler.UserHandler) {
	SetupHealthRoutes(app)
	setupAuthRoutes(app, authHandler)
	SetupUserRoutes(app, userHandler)
	SetupFallbackHandlers(app)
}
