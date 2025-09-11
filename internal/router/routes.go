package router

import (
	"github.com/zerodayz7/http-server/internal/handler"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(
	app *fiber.App,
	interactionHandler *handler.InteractionHandler,
) {
	SetupHealthRoutes(app)
	SetupStatsRoutes(app, interactionHandler)

	SetupFallbackHandlers(app)
}
