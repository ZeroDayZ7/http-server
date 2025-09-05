package config

import (
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func BodyLimitMiddleware() fiber.Handler {
	limitMB := 2 // default 2MB
	if val := os.Getenv("BODY_LIMIT_MB"); val != "" {
		if l, err := strconv.Atoi(val); err == nil {
			limitMB = l
		}
	}

	maxSize := limitMB * 1024 * 1024

	return func(c *fiber.Ctx) error {
		if len(c.Body()) > maxSize {
			return c.Status(fiber.StatusRequestEntityTooLarge).SendString("Payload too large")
		}
		return c.Next()
	}
}
