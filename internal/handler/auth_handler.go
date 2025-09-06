package handler

import (
	"github.com/zerodayz7/http-server/internal/errors"
	"github.com/zerodayz7/http-server/internal/middleware"
	"github.com/zerodayz7/http-server/internal/service"
	"github.com/zerodayz7/http-server/internal/validator"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) GetCSRFToken(c *fiber.Ctx) error {
	token := middleware.GenerateCSRFToken(c)
	return c.JSON(fiber.Map{
		"csrf_token": token,
	})
}

func (h *UserHandler) CheckEmail(c *fiber.Ctx) error {
	body := c.Locals("validatedBody").(validator.CheckEmailRequest)
	exists, err := h.service.IsEmailExists(body.Email)
	if err != nil {
		return errors.SendAppError(c, errors.ErrInternal)
	}
	return c.JSON(fiber.Map{"exists": exists})
}

func (h *UserHandler) Register(c *fiber.Ctx) error {
	body := c.Locals("validatedBody").(validator.RegisterRequest)

	user, err := h.service.Register(body.Username, body.Email, body.Password)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			errors.AttachRequestMeta(c, appErr, "requestID")
			return appErr
		}
		return errors.ErrInternal
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"user":    user,
	})
}
