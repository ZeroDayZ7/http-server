package router

import (
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/zerodayz7/http-server/internal/handler"
	"github.com/zerodayz7/http-server/internal/middleware"
	"github.com/zerodayz7/http-server/internal/validator"

	"github.com/gofiber/fiber/v2"
	"github.com/zerodayz7/http-server/config"
)

func setupAuthRoutes(app *fiber.App, h *handler.AuthHandler, sessionStore *session.Store) {
	auth := app.Group("/auth")
	auth.Use(config.NewLimiter("auth"))

	csrfCfg := config.NewCSRFConfig(sessionStore.Storage)
	auth.Use(csrf.New(csrfCfg))

	auth.Get("/init-session", h.InitSession)

	auth.Get("/csrf-token", h.GetCSRFToken)

	auth.Post("/login",
		middleware.ValidateBody[validator.LoginRequest](),
		h.Login,
	)

	auth.Post("/2fa-verify",
		middleware.ValidateBody[validator.TwoFARequest](),
		h.Verify2FA)

	auth.Post("/register",
		middleware.ValidateBody[validator.RegisterRequest](),
		h.Register,
	)
}
