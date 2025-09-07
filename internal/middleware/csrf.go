package middleware

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/zerodayz7/http-server/internal/service"
)

type CSRFMiddleware struct {
	SessionService *service.SessionService
}

func NewCSRFMiddleware(sessSvc *service.SessionService) *CSRFMiddleware {
	return &CSRFMiddleware{
		SessionService: sessSvc,
	}
}

// GenerateCSRFToken generuje token i zapisuje w sesji
func (m *CSRFMiddleware) GenerateCSRFToken(c *fiber.Ctx) string {
	token := uuid.New().String()
	sessionID := c.Cookies("session_id")
	if sessionID != "" {
		session, err := m.SessionService.GetSession(sessionID)
		if err == nil && session != nil {
			var data map[string]any
			if session.Data != "" {
				// Parsujemy istniejące dane JSON
				err := json.Unmarshal([]byte(session.Data), &data)
				if err != nil || data == nil {
					data = make(map[string]any)
				}
			} else {
				// Jeśli pusty string → inicjalizujemy mapę
				data = make(map[string]any)
			}

			data["csrf_token"] = token

			_ = m.SessionService.UpdateSessionData(sessionID, data)
		}
	}
	return token
}

// VerifyCSRFToken weryfikuje token z sesji
func (m *CSRFMiddleware) VerifyCSRFToken(c *fiber.Ctx) error {
	headerToken := c.Get("X-CSRF-Token")
	sessionID := c.Cookies("session_id")
	if sessionID == "" || headerToken == "" {
		return fiber.ErrUnauthorized
	}

	session, err := m.SessionService.GetSession(sessionID)
	if err != nil || session == nil {
		return fiber.ErrUnauthorized
	}

	var data map[string]any
	_ = json.Unmarshal([]byte(session.Data), &data)
	savedToken, _ := data["csrf_token"].(string)

	if savedToken != headerToken {
		return fiber.ErrUnauthorized
	}

	return c.Next()
}
