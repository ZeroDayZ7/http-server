package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zerodayz7/http-server/internal/errors"
)

func ValidateBody[T any]() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body T
		if err := c.BodyParser(&body); err != nil {
			return errors.ErrInvalidJSON.WithErr(err)
		}

		if errs := ValidateStruct(body); len(errs) > 0 {
			meta := make(map[string]any, len(errs))
			for k, v := range errs {
				meta[k] = v
			}
			return errors.ErrValidationFailed.WithMeta(meta)
		}

		c.Locals("validatedBody", body)
		return c.Next()
	}
}
