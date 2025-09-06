package router

import (
	"github.com/zerodayz7/http-server/internal/handler"
	"github.com/zerodayz7/http-server/internal/middleware"
	"github.com/zerodayz7/http-server/internal/validator"

	"github.com/gofiber/fiber/v2"
	"github.com/zerodayz7/http-server/config"
)

func setupAuthRoutes(app *fiber.App, h *handler.UserHandler) {
	auth := app.Group("/auth")
	auth.Use(config.NewLimiter("auth"))

	auth.Get("/csrf-token", h.GetCSRFToken)

	auth.Post("/check-email",
		middleware.ValidateBody[validator.CheckEmailRequest](),
		h.CheckEmail,
	)

	auth.Post("/register",
		middleware.ValidateBody[validator.RegisterRequest](),
		h.Register,
	)
}
