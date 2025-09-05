// errors/helpers.go
package errors

import "github.com/gofiber/fiber/v2"

func Send(c *fiber.Ctx, err *AppError) error {
	return SendAppError(c, err)
}
