package config

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/csrf"
)

func NewCSRFConfig(storage fiber.Storage) csrf.Config {
	// isProd := AppConfig.Server.Env == "production"
	ttl := AppConfig.SessionTTL
	return csrf.Config{
		// Storage: storage,
		Session:    SessionStore(),
		KeyLookup:  "header:X-CSRF-Token",
		CookieName: "__Host-csrf_",
		ContextKey: "csrf",
		Expiration: ttl,
		// KeyGenerator:      shared.GenerateCSRFToken,
		CookieSecure:   true,
		CookieHTTPOnly: false,
		CookieSameSite: "None", // Strict, Lax
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
