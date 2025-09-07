package router

import (
	"github.com/zerodayz7/http-server/internal/handler"
	"github.com/zerodayz7/http-server/internal/middleware"
	"github.com/zerodayz7/http-server/internal/validator"

	"github.com/gofiber/fiber/v2"
	"github.com/zerodayz7/http-server/config"
)

func setupAuthRoutes(app *fiber.App, h *handler.AuthHandler) {
	auth := app.Group("/auth")
	auth.Use(config.NewLimiter("auth"))

	auth.Get("/csrf-token", h.GetCSRFToken)

	auth.Post("/login",
		middleware.ValidateBody[validator.LoginRequest](),
		h.Login,
	)

	// auth.Post("/2fa-verify",
	// 	middleware.ValidateBody[validator.TwoFARequest](),
	// 	handler.Verify2FA)

	auth.Post("/register",
		middleware.ValidateBody[validator.RegisterRequest](),
		h.Register,
	)
}
