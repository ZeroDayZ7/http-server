package config

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/zerodayz7/http-server/internal/server"
	"github.com/zerodayz7/http-server/internal/shared"
)

func NewFiberApp() *fiber.App {
	app := fiber.New(fiber.Config{
		ProxyHeader:             fiber.HeaderXForwardedFor,
		EnableTrustedProxyCheck: true,
		TrustedProxies: []string{
			"127.0.0.1",
			"::1",
		},
		BodyLimit:             AppConfig.Server.BodyLimitMB * 1024 * 1024,
		ReadTimeout:           AppConfig.Server.ReadTimeout,
		WriteTimeout:          AppConfig.Server.WriteTimeout,
		IdleTimeout:           AppConfig.Server.IdleTimeout,
		Prefork:               AppConfig.Server.Prefork,
		CaseSensitive:         AppConfig.Server.CaseSensitive,
		DisableStartupMessage: true,
		EnableIPValidation:    true,
		ServerHeader:          AppConfig.Server.ServerHeader,
		AppName:               AppConfig.Server.AppName,
		RequestMethods:        []string{"GET", "POST", "OPTIONS", "HEAD"},
		ErrorHandler:          server.ErrorHandler(),
	})

	app.Use(requestid.New())
	app.Use(recover.New())
	// app.Use(FiberLoggerMiddleware())
	app.Use(shared.RequestLoggerMiddleware())
	app.Use(cors.New(CorsConfig()))
	app.Use(helmet.New(HelmetConfig()))
	app.Use(NewLimiter("global"))
	app.Use(compress.New(CompressConfig()))

	return app
}
