package config

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/zerodayz7/http-server/internal/cache"
)

func NewFiberApp(sessionStore *session.Store) *fiber.App {
	app := fiber.New()

	cache.Init()

	app.Use(requestid.New())
	app.Use(recover.New())
	app.Use(FiberLoggerMiddleware())
	app.Use(helmet.New(HelmetConfig()))
	app.Use(cors.New(CorsConfig()))
	app.Use(NewLimiter("global"))
	app.Use(compress.New(CompressConfig()))

	// global session
	app.Use(func(c *fiber.Ctx) error {
		sess, err := sessionStore.Get(c)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to get session"})
		}
		c.Locals("session", sess)
		return c.Next()
	})

	// global CSRF
	csrfCfg := NewCSRFConfig(sessionStore.Storage)
	app.Use(csrf.New(csrfCfg))

	return app
}
