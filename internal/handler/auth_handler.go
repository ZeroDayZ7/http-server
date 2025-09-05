package handler

import (
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

func (h *UserHandler) CheckEmail(c *fiber.Ctx) error {
	body := c.Locals("validatedBody").(validator.CheckEmailRequest)
	exists, err := h.service.IsEmailExists(body.Email)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Internal server error")
	}
	return c.JSON(fiber.Map{"exists": exists})
}

func (h *UserHandler) Register(c *fiber.Ctx) error {
	body := c.Locals("validatedBody").(validator.RegisterRequest)

	emailExists, usernameExists, err := h.service.IsEmailOrUsernameExists(body.Email, body.Username)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Internal server error")
	}

	if emailExists {
		return fiber.NewError(fiber.StatusBadRequest, "Email already exists")
	}
	if usernameExists {
		return fiber.NewError(fiber.StatusBadRequest, "Username already exists")
	}

	user, err := h.service.Register(body.Username, body.Email, body.Password)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Internal server error")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"user":    user,
	})
}
