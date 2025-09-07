package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/zerodayz7/http-server/config"
	"github.com/zerodayz7/http-server/internal/middleware" // <- dodaj import własnego middleware
)

// setupUserRoutes ustawia wszystkie trasy dla użytkowników
func setupUserRoutes(app *fiber.App) {
	users := app.Group("/users")
	protected := users.Group("/")
	protected.Use(config.NewLimiter("users"))

	// Middleware sesji z Locals
	protected.Use(middleware.SessionMiddlewareFromLocals())

	// Trasa testowa
	protected.Get("/test-session", func(c *fiber.Ctx) error {
		sess := c.Locals("session").(*session.Session)
		userID := sess.Get("userID")
		return c.JSON(fiber.Map{
			"message": "Middleware działa!",
			"userID":  userID,
		})
	})

	// CSRF test
	protected.Get("/test-csrf", middleware.NewCSRFMiddlewareFromLocals(), func(c *fiber.Ctx) error {
		sessionID := c.Cookies("session_id")
		return c.JSON(fiber.Map{
			"message":    "CSRF token zweryfikowany!",
			"session_id": sessionID,
		})
	})
}
