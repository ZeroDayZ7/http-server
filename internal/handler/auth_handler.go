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

func (h *UserHandler) Login(c *fiber.Ctx) error {
	body := c.Locals("validatedBody").(validator.LoginRequest)

	user, err := h.service.GetUserByEmail(body.Email)
	if err != nil {
		return errors.SendAppError(c, errors.ErrInvalidCredentials)
	}

	valid, err := h.service.VerifyPassword(user, body.Password)
	if err != nil || !valid {
		return errors.SendAppError(c, errors.ErrInvalidCredentials)
	}

	if user.TwoFactorEnabled {
		return c.JSON(fiber.Map{"2fa_required": true})
	}

	// token, err := h.service.GenerateToken(user)
	// if err != nil {
	// 	return errors.SendAppError(c, errors.ErrInternal)
	// }

	return c.JSON(fiber.Map{
		"2fa_required": false,
		// "token":        token,
	})
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

// func (h *UserHandler) Verify2FA(c *fiber.Ctx) error {
// 	type Req struct {
// 		Email string `json:"user_id"`
// 		Code   string `json:"code"`
// 	}
// 	var body Req
// 	if err := c.BodyParser(&body); err != nil {
// 		return errors.SendAppError(c, errors.ErrInvalidRequest)
// 	}

// 	ok, err := service.Verify2FACode(body.Email, body.Code)
// 	if err != nil {
// 		return errors.SendAppError(c, errors.ErrInvalid2FACode)
// 	}

// 	if !ok {
// 		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Invalid 2FA code",
// 		})
// 	}

// 	return c.JSON(fiber.Map{
// 		"success": true,
// 		"message": "2FA verified successfully",
// 	})
// }
