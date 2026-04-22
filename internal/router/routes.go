package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zerodayz7/http-server/config"
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
	SetupFavicon(app)

	SetupHealthRoutes(app, cfg)
	SetupStatsRoutes(app, interactionHandler, cfg)

	api := app.Group("/")
	api.Use(config.GetLimiter(cfg, config.LimitGlobal))

	SetupNotFoundHandler(app, log)
}
