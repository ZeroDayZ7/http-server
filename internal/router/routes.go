package router

import (
	"time"

	"github.com/zerodayz7/http-server/internal/handler"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

func SetupRoutes(app *fiber.App, authHandler *handler.AuthHandler, userHandler *handler.UserHandler, sessionStore *session.Store, sessionTTL time.Duration) {
	SetupHealthRoutes(app)
	setupAuthRoutes(app, authHandler)
	SetupUserRoutes(app, sessionStore, sessionTTL)
	SetupFallbackHandlers(app)
}
