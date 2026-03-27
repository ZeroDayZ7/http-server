package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zerodayz7/http-server/config/env"
	"github.com/zerodayz7/http-server/internal/handler"
	"github.com/zerodayz7/http-server/internal/shared/logger"
)

func SetupRoutes(
	app *fiber.App,
	interactionHandler *handler.InteractionHandler,
	cfg *env.Config,
	log logger.Logger,
) {
	SetupHealthRoutes(app, cfg)
	SetupStatsRoutes(app, interactionHandler, cfg)
	SetupFallbackHandlers(app, log)
}
