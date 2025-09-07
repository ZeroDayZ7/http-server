package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

// SessionMiddlewareFromLocals pobiera sesję z c.Locals("store")
func SessionMiddlewareFromLocals() fiber.Handler {
	return func(c *fiber.Ctx) error {
		storeIfc := c.Locals("sessionStore")
		if storeIfc == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Session store not initialized",
			})
		}

		store, ok := storeIfc.(*session.Store)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Invalid session store type",
			})
		}

		sess, err := store.Get(c) // pobiera sesję powiązaną z ciasteczkiem session_id
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get session",
			})
		}

		c.Locals("session", sess) // teraz handler może używać c.Locals("session")
		return c.Next()
	}
}

// NewCSRFMiddlewareFromLocals weryfikuje CSRF token pobrany z sesji w Locals
func NewCSRFMiddlewareFromLocals() fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessIface := c.Locals("session")
		if sessIface == nil {
			return c.Status(403).JSON(fiber.Map{"error": "No session"})
		}
		sess := sessIface.(*session.Session)
		csrfToken := sess.Get("csrfToken")
		if csrfToken == nil || csrfToken != c.Get("X-CSRF-Token") {
			return c.Status(403).JSON(fiber.Map{"error": "Invalid CSRF token"})
		}
		return c.Next()
	}
}
