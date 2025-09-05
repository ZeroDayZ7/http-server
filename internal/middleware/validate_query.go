package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zerodayz7/http-server/internal/errors"
)

func ValidateQuery[T any]() fiber.Handler {
	return func(c *fiber.Ctx) error {
		query := new(T)
		if err := c.QueryParser(query); err != nil {
			return errors.SendAppError(c, &errors.AppError{
				Code: "INVALID_QUERY",
				Type: errors.BadRequest,
			})
		}

		if errs := Validate(query); len(errs) > 0 {
			return errors.SendAppError(c, &errors.AppError{
				Code: "VALIDATION_FAILED",
				Type: errors.Validation,
				Meta: errs,
			})
		}

		c.Locals("validatedQuery", query)
		return c.Next()
	}
}
