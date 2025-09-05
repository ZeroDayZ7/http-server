package router

import (
	"time"

	"github.com/zerodayz7/http-server/internal/handler"
	"github.com/zerodayz7/http-server/internal/middleware"
	"github.com/zerodayz7/http-server/internal/shared/logger"
	"github.com/zerodayz7/http-server/internal/validator"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"go.uber.org/zap"
)

func setupAuthRoutes(app *fiber.App, h *handler.UserHandler) {
	auth := app.Group("/auth")

	auth.Use(limiter.New(limiter.Config{
		Max:        10,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			logger.GetLogger().Warn("Auth limit exceeded", zap.String("ip", c.IP()), zap.String("path", c.Path()))
			return fiber.NewError(fiber.StatusTooManyRequests, "Too many requests")
		},
	}))

	auth.Post("/check-email",
		middleware.ValidateBody[validator.CheckEmailRequest](),
		h.CheckEmail,
	)

	auth.Post("/register",
		middleware.ValidateBody[validator.RegisterRequest](),
		h.Register,
	)
}
