package router

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/zerodayz7/http-server/config"
)

func SetupUserRoutes(app *fiber.App, sessionStore *session.Store, sessionTTL time.Duration) {
	users := app.Group("/users")
	protected := users.Group("/")
	protected.Use(config.NewLimiter("users"))

	// Test sesji
	protected.Get("/test-session", func(c *fiber.Ctx) error {
		sess, err := sessionStore.Get(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get session",
			})
		}
		userID := sess.Get("userID")
		return c.JSON(fiber.Map{
			"message": "Middleware działa!",
			"userID":  userID,
		})
	})

	// Middleware CSRF
	csrfMiddleware := csrf.New(config.NewCSRFConfig(sessionStore.Storage, sessionTTL))
	protected.Use(csrfMiddleware)

	// Test CSRF
	protected.Post("/test-csrf", func(c *fiber.Ctx) error {
		token := c.Locals("csrf")
		return c.JSON(fiber.Map{
			"message":    "CSRF token zweryfikowany!",
			"csrf_token": token,
		})
	})
}
