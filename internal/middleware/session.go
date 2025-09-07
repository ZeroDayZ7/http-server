package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zerodayz7/http-server/internal/errors"
	"github.com/zerodayz7/http-server/internal/model"
	"github.com/zerodayz7/http-server/internal/service"
	"github.com/zerodayz7/http-server/internal/shared"
)

// SessionMiddleware zwraca middleware Fiber, który weryfikuje sesję
func SessionMiddleware(sessionService *service.SessionService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessionID := c.Cookies("session_id")
		var session *model.Session
		var err error

		if sessionID != "" {
			session, err = sessionService.GetSession(sessionID)
			if err != nil {
				return errors.SendAppError(c, errors.ErrInternal)
			}
		}

		// Jeśli brak sesji → twórz anonimową
		if session == nil {
			session, err = sessionService.CreateSession(0, nil)
			if err != nil {
				return errors.SendAppError(c, errors.ErrInternal)
			}
			shared.SetSessionCookie(c, session.SessionID)
		}

		c.Locals("session", session)
		c.Locals("userID", session.UserID)

		return c.Next()
	}
}
