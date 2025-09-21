package config

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/zerodayz7/http-server/internal/server"
	"github.com/zerodayz7/http-server/internal/shared/logger"
	"go.uber.org/zap"
)

func NewFiberApp() *fiber.App {
	log := logger.GetLogger()
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
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c *fiber.Ctx, err interface{}) {
			log.Error("Panic recovered in handler",
				zap.Any("error", err),
				zap.String("path", c.Path()),
			)

			// opcjonalnie: zwróć klientowi 500
			c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   "Internal server error",
			})
		},
	}))
	app.Use(FiberLoggerMiddleware())
	app.Use(helmet.New(HelmetConfig()))
	app.Use(cors.New(CorsConfig()))
	app.Use(NewLimiter("global"))
	app.Use(compress.New(CompressConfig()))

	return app
}
