package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/zerodayz7/http-server/config"
	"github.com/zerodayz7/http-server/internal/errors"
	"github.com/zerodayz7/http-server/internal/service"
	"github.com/zerodayz7/http-server/internal/shared"
	"github.com/zerodayz7/http-server/internal/validator"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) InitSession(c *fiber.Ctx) error {
	// Pobierz sesję, Fiber automatycznie ją tworzy
	sess := c.Locals("session").(*session.Session)

	// Możesz tu opcjonalnie ustawić dane w sesji
	UserID := shared.GenerateUuid()
	sess.Set("UserID", UserID)
	if err := sess.Save(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save session",
		})
	}

	// Zwróć JSON, nawet jeśli niczego nie potrzebujesz
	return c.JSON(fiber.Map{
		"sessionInitialized": true,
		"UserID":             UserID,
	})
}

func (h *AuthHandler) GetCSRFToken(c *fiber.Ctx) error {
	sess := c.Locals("session").(*session.Session)

	csrfToken := sess.Get("csrfToken")
	if csrfToken == nil {
		csrfToken = shared.GenerateUuid()
		sess.Set("csrfToken", csrfToken)
		if err := sess.Save(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to save session",
			})
		}
	}

	return c.JSON(fiber.Map{"csrf_token": csrfToken})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	body := c.Locals("validatedBody").(validator.LoginRequest)

	user, err := h.authService.GetUserByEmail(body.Email)
	if err != nil {
		return errors.SendAppError(c, errors.ErrInvalidCredentials)
	}

	valid, err := h.authService.VerifyPassword(user, body.Password)
	if err != nil || !valid {
		return errors.SendAppError(c, errors.ErrInvalidCredentials)
	}

	if user.TwoFactorEnabled {
		return c.JSON(fiber.Map{"2fa_required": true})
	}

	sess, err := config.SessionStore().Get(c)
	if err != nil {
		return errors.SendAppError(c, errors.ErrInternal)
	}

	sess.Set("userID", user.ID)

	csrfToken := sess.Get("csrfToken")
	if csrfToken == nil {
		csrfToken = shared.GenerateUuid()
		sess.Set("csrfToken", csrfToken)
	}

	if err := sess.Save(); err != nil {
		return errors.SendAppError(c, errors.ErrInternal)
	}

	return c.JSON(fiber.Map{
		"2fa_required": false,
		"csrf_token":   csrfToken,
	})
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	body := c.Locals("validatedBody").(validator.RegisterRequest)

	user, err := h.authService.Register(body.Username, body.Email, body.Password)
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

func (h *AuthHandler) Verify2FA(c *fiber.Ctx) error {
	body := c.Locals("validatedBody").(validator.TwoFARequest)

	sess := c.Locals("session").(*session.Session)
	userID := sess.Get("userID").(uint)

	ok, err := h.authService.Verify2FACodeByID(userID, body.Code)
	if err != nil {
		return errors.SendAppError(c, errors.ErrInvalidCredentials)
	}
	if !ok {
		return errors.SendAppError(c, errors.ErrInvalid2FACode)
	}

	return c.JSON(fiber.Map{
		"message": "2FA verified successfully",
	})
}
