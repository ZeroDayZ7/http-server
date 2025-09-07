package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zerodayz7/http-server/internal/errors"
	"github.com/zerodayz7/http-server/internal/middleware"
	"github.com/zerodayz7/http-server/internal/model"
	"github.com/zerodayz7/http-server/internal/service"
	"github.com/zerodayz7/http-server/internal/shared"
	"github.com/zerodayz7/http-server/internal/validator"
)

type AuthHandler struct {
	authService    *service.AuthService
	sessionService *service.SessionService
	csrfMiddleware *middleware.CSRFMiddleware
}

func NewAuthHandler(authService *service.AuthService, sessionService *service.SessionService) *AuthHandler {
	csrf := middleware.NewCSRFMiddleware(sessionService)
	return &AuthHandler{
		authService:    authService,
		sessionService: sessionService,
		csrfMiddleware: csrf,
	}
}

func (h *AuthHandler) GetCSRFToken(c *fiber.Ctx) error {
	token := h.csrfMiddleware.GenerateCSRFToken(c)
	return c.JSON(fiber.Map{"csrf_token": token})
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

	// Pobierz istniejącą sesję (anonimową)
	session := c.Locals("session").(*model.Session)

	// Przypisz userID do sesji
	session.UserID = user.ID
	if err := h.sessionService.UpdateSessionData(session.SessionID, nil); err != nil {
		return errors.SendAppError(c, errors.ErrInternal)
	}

	// Ustaw ciasteczko sesji
	shared.SetSessionCookie(c, session.SessionID)

	// Generuj CSRF token
	csrfToken := h.csrfMiddleware.GenerateCSRFToken(c)

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
