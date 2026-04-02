package config

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/zerodayz7/http-server/config/env"
	"github.com/zerodayz7/http-server/internal/errors"
	"github.com/zerodayz7/http-server/internal/shared"
	"github.com/zerodayz7/http-server/internal/shared/logger"
)

func NewFiberApp(cfg *env.Config, log logger.Logger) *fiber.App {
	app := fiber.New(fiber.Config{
		ProxyHeader:             fiber.HeaderXForwardedFor,
		EnableTrustedProxyCheck: true,
		TrustedProxies:          []string{"127.0.0.1", "::1"},

		BodyLimit:             cfg.Server.BodyLimitMB * 1024 * 1024,
		ReadTimeout:           cfg.Server.ReadTimeout,
		WriteTimeout:          cfg.Server.WriteTimeout,
		IdleTimeout:           cfg.Server.IdleTimeout,
		Prefork:               cfg.Server.Prefork,
		CaseSensitive:         cfg.Server.CaseSensitive,
		DisableStartupMessage: true,
		EnableIPValidation:    true,
		ServerHeader:          cfg.Server.ServerHeader,
		AppName:               cfg.Server.AppName,

		RequestMethods: []string{"GET", "POST", "OPTIONS", "HEAD"},
		ErrorHandler:   errors.ErrorHandler(log),
	})

	app.Use(requestid.New())
	app.Use(recover.New())

	// Przekazujemy log do Middleware
	app.Use(shared.RequestLoggerMiddleware(log))

	app.Use(cors.New(CorsConfig(cfg)))
	app.Use(helmet.New(HelmetConfig()))
	// app.Use(NewLimiter(cfg, "global"))
	app.Use(compress.New(CompressConfig()))

	return app
}
