package router

import (
	"github.com/zerodayz7/http-server/config"
	"github.com/zerodayz7/http-server/internal/handler"
	"github.com/zerodayz7/http-server/internal/middleware"
	"github.com/zerodayz7/http-server/internal/validator"

	"github.com/gofiber/fiber/v2"
)

func SetupStatsRoutes(app *fiber.App, h *handler.InteractionHandler) {
	stats := app.Group("/stats")
	stats.Use(config.NewLimiter("visits"))

	stats.Get("/interactions", h.RecordVisit)

	stats.Post("/interactions",
		middleware.ValidateBody[validator.InteractionRequest](),
		h.RecordLike,
	)

}
