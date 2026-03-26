package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zerodayz7/http-server/config"
	"github.com/zerodayz7/http-server/internal/handler"
	"github.com/zerodayz7/http-server/internal/middleware"
	"github.com/zerodayz7/http-server/internal/validator"
)

func SetupStatsRoutes(app *fiber.App, h *handler.InteractionHandler, cfg *config.Config) {
	stats := app.Group("/stats")
	stats.Use(config.NewLimiter(cfg, "visits"))

	stats.Get("/interactions",
		middleware.ValidateQuery[validator.FingerprintRequest](),
		h.GetStats,
	)

	// Zmieniono h.RecordLike na h.HandleLike
	stats.Post("/interactions",
		middleware.ValidateBody[validator.InteractionRequest](),
		h.HandleLike,
	)

	// Zmieniono h.InitializeSession na h.HandleVisit
	stats.Post("/init",
		h.HandleVisit,
	)
}
