package middleware

import (
	"github.com/gofiber/fiber/v2"
)

func ValidateBody[T any]() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body T
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "INVALID_JSON",
				"meta":  map[string]any{"field": "body"},
			})
		}

		if errs := ValidateStruct(body); len(errs) > 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "VALIDATION_FAILED",
				"meta":  errs,
			})
		}

		c.Locals("validatedBody", body)
		return c.Next()
	}
}
