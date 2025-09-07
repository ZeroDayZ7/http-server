package config

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/csrf"
)

// NewCSRFConfig zwraca skonfigurowany CSRF z Storage (double submit)
func NewCSRFConfig(storage fiber.Storage) csrf.Config {
	return csrf.Config{
		Storage:        storage, // Użyj Storage zamiast Session dla double submit
		KeyLookup:      "header:X-CSRF-Token",
		CookieName:     "csrf_",
		ContextKey:     "csrf",
		Expiration:     1 * time.Hour, // Lub cfg.SessionTTL
		CookieSecure:   false,         // true w prod
		CookieHTTPOnly: true,
		CookieSameSite: "Strict",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if strings.HasPrefix(c.Path(), "/auth/") || c.Accepts("json") == "json" {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "Invalid CSRF token",
				})
			}
			return c.Status(fiber.StatusForbidden).SendString("Forbidden")
		},
	}
}
