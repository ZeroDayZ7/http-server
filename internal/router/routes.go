package router

import (
	"github.com/zerodayz7/http-server/internal/handler"
	"github.com/zerodayz7/http-server/internal/service"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, authHandler *handler.AuthHandler, userHandler *handler.UserHandler, sessionSvc *service.SessionService) {
	SetupHealthRoutes(app)
	setupAuthRoutes(app, authHandler)
	setupUserRoutes(app, userHandler, sessionSvc)
	SetupFallbackHandlers(app)
}
