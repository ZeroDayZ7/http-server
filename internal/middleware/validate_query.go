package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zerodayz7/http-server/internal/errors"
)

func ValidateQuery[T any]() fiber.Handler {
	return func(c *fiber.Ctx) error {
		query := new(T)
		if err := c.QueryParser(query); err != nil {
			return errors.ErrInvalidRequest.WithErr(err)
		}

		if errs := ValidateStruct(query); len(errs) > 0 {
			meta := make(map[string]any, len(errs))
			for k, v := range errs {
				meta[k] = v
			}
			return errors.ErrValidationFailed.WithMeta(meta)
		}

		c.Locals("validatedQuery", *query)
		return c.Next()
	}
}
