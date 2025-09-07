package config

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/csrf"
)

func NewCSRFConfig(storage fiber.Storage, ttl time.Duration) csrf.Config {
	return csrf.Config{
		Storage:        storage,
		KeyLookup:      "header:X-CSRF-Token",
		CookieName:     "csrf_",
		ContextKey:     "csrf",
		Expiration:     ttl,
		CookieSecure:   false,
		CookieHTTPOnly: false,
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
