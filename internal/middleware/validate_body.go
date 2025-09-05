package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zerodayz7/http-server/internal/errors"
)

func ValidateBody[T any]() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body T
		if err := c.BodyParser(&body); err != nil {
			return errors.SendAppError(c, &errors.AppError{
				Code: "INVALID_JSON",
				Type: errors.BadRequest,
				Meta: map[string]any{"field": "body"},
			})
		}

		if errs := Validate(body); len(errs) > 0 {
			return errors.SendAppError(c, &errors.AppError{
				Code: "VALIDATION_FAILED",
				Type: errors.Validation,
				Meta: errs,
			})
		}

		c.Locals("validatedBody", body)
		return c.Next()
	}
}
