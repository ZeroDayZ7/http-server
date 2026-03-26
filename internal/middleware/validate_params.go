package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zerodayz7/http-server/internal/errors"
)

func ValidateParams[T any]() fiber.Handler {
	return func(c *fiber.Ctx) error {
		params := new(T)
		if err := c.ParamsParser(params); err != nil {
			return errors.ErrInvalidRequest.WithErr(err)
		}

		if errs := ValidateStruct(params); len(errs) > 0 {
			meta := make(map[string]any, len(errs))
			for k, v := range errs {
				meta[k] = v
			}
			return errors.ErrValidationFailed.WithMeta(meta)
		}

		c.Locals("validatedParams", *params)
		return c.Next()
	}
}
