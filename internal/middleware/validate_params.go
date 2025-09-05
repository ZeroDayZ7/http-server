package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zerodayz7/http-server/internal/errors"
)

func ValidateParams[T any]() fiber.Handler {
	return func(c *fiber.Ctx) error {
		params := new(T)
		if err := c.ParamsParser(params); err != nil {
			return errors.SendAppError(c, &errors.AppError{
				Code: "INVALID_PARAMS",
				Type: errors.BadRequest,
			})
		}

		if errs := Validate(params); len(errs) > 0 {
			return errors.SendAppError(c, &errors.AppError{
				Code: "VALIDATION_FAILED",
				Type: errors.Validation,
				Meta: errs,
			})
		}

		c.Locals("validatedParams", params)
		return c.Next()
	}
}
