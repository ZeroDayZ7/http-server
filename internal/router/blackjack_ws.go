package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	handler "github.com/zerodayz7/http-server/internal/handler/blackjack"
	"github.com/zerodayz7/http-server/internal/shared/logger"
	"go.uber.org/zap"
)

func SetupGameRoutes(app *fiber.App, h *handler.GameHandler) {
	// Status HTTP
	status := app.Group("/status")
	status.Get("/ws", h.WSStatus)

	// WebSocket
	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		h.Mutex.Lock()
		h.ConnectedClients[c] = true
		h.Mutex.Unlock()

		defer func() {
			h.Mutex.Lock()
			delete(h.ConnectedClients, c)
			h.Mutex.Unlock()
			c.Close()
		}()

		for {
			mt, msg, err := c.ReadMessage()
			if err != nil {
				logger.GetLogger().Warn("WebSocket read error", zap.Error(err))
				break
			}

			h.Mutex.RLock()
			for client := range h.ConnectedClients {
				if err := client.WriteMessage(mt, msg); err != nil {
					logger.GetLogger().Warn("WebSocket write error", zap.Error(err))
					h.Mutex.RUnlock()
					h.Mutex.Lock()
					delete(h.ConnectedClients, client)
					h.Mutex.Unlock()
					h.Mutex.RLock()
					client.Close()
				}
			}
			h.Mutex.RUnlock()
		}
	}, websocket.Config{
		HandshakeTimeout: 0,
		Origins:          []string{"*"}, // Allows connections from any origin (for testing)
	}))
}
