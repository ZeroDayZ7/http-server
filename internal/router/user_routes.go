package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zerodayz7/http-server/config"
	"github.com/zerodayz7/http-server/internal/handler"
	"github.com/zerodayz7/http-server/internal/middleware"
	"github.com/zerodayz7/http-server/internal/service"
)

func setupUserRoutes(app *fiber.App, h *handler.UserHandler, sessionSvc *service.SessionService) {
	users := app.Group("/users")
	protected := users.Group("/")
	protected.Use(config.NewLimiter("users"))

	// Middleware sesji dla wszystkich tras w grupie
	protected.Use(middleware.SessionMiddleware(sessionSvc))

	// Trasa testowa – nie wymaga CSRF
	protected.Get("/test-session", func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(uint)
		return c.JSON(fiber.Map{
			"message": "Middleware działa!",
			"userID":  userID,
		})
	})

	// Trasa testowa CSRF – wymaga tokena
	csrf := middleware.NewCSRFMiddleware(sessionSvc) // konstruktor
	protected.Get("/test-csrf", csrf.VerifyCSRFToken, func(c *fiber.Ctx) error {
		sessionID := c.Cookies("session_id")
		return c.JSON(fiber.Map{
			"message":    "CSRF token zweryfikowany!",
			"session_id": sessionID,
		})
	})

	// Tu możesz potem podpiąć normalne trasy użytkowników
	// protected.Post("/", h.CreateUser)
	// protected.Patch("/:id", h.UpdateUser)
	// protected.Delete("/:id", h.DeleteUser)
}
