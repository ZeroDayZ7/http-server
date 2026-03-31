package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zerodayz7/http-server/config"
	"github.com/zerodayz7/http-server/config/env"
	"github.com/zerodayz7/http-server/internal/handler"
	"github.com/zerodayz7/http-server/internal/middleware"
	"github.com/zerodayz7/http-server/internal/validator"
)

func SetupStatsRoutes(app *fiber.App, h *handler.InteractionHandler, cfg *env.Config) {
	stats := app.Group("/stats")
	stats.Use(config.NewLimiter(cfg, "visits"))

	stats.Get("/interactions",
		middleware.ValidateQuery[validator.FingerprintRequest](),
		h.GetStats,
	)

	stats.Post("/interactions",
		middleware.ValidateBody[validator.InteractionRequest](),
		h.HandleInteraction,
	)

	stats.Post("/init",
		h.HandleVisit,
	)
}
