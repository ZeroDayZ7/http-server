package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	handler "github.com/zerodayz7/http-server/internal/handler/blackjack"
)

// SetupGameRoutes rejestruje wszystkie ścieżki gry
func SetupGameRoutes(app *fiber.App, h *handler.GameHandler) {
	// --- HTTP status ---
	status := app.Group("/status")
	status.Get("/ws", h.WSStatus)

	// --- WebSocket ---
	app.Get("/ws", websocket.New(h.HandleWS))

	// --- REST API dla lobby ---
	lobby := app.Group("/lobbies")
	lobby.Get("/", h.GetLobbies)   // <- delegujemy do handlera
	lobby.Post("/", h.CreateLobby) // <- delegujemy do handlera
}
