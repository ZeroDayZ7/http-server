package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/zerodayz7/http-server/internal/errors"
)

const csrfHeader = "X-CSRF-Token"

func GenerateCSRFToken(c *fiber.Ctx) string {
	token := uuid.New().String()
	c.Locals("csrf_token", token)
	return token
}

func VerifyCSRFToken(c *fiber.Ctx) error {
	headerToken := c.Get(csrfHeader)
	contextTokenAny := c.Locals("csrf_token")

	contextToken, ok := contextTokenAny.(string)
	if !ok || headerToken == "" || headerToken != contextToken {
		return errors.SendAppError(c, errors.ErrCSRFInvalid)
	}

	return c.Next()
}
