package config

import (
	"time"

	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/zerodayz7/http-server/internal/middleware"
)

func NewFiberSessionStore(sessionSvc *middleware.SessionService) *session.Store {
	return session.New(session.Config{
		Expiration:     24 * time.Hour,
		CookieHTTPOnly: true,
		CookieSecure:   false,
		CookieSameSite: "Lax",
		Storage:        middleware.NewMySQLStore(sessionSvc),
	})
}
