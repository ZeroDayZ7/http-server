package blackjack

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	bjs "github.com/zerodayz7/http-server/internal/service/blackjack"
)

type GameHandler struct {
	LobbyService *bjs.LobbyService
	GameService  *bjs.GameService
}

func NewGameHandler() *GameHandler {
	return &GameHandler{
		LobbyService: bjs.NewLobbyService(),
		GameService:  bjs.NewGameService(),
	}
}

// WS Status
func (h *GameHandler) WSStatus(c *fiber.Ctx) error {
	h.LobbyService.Mutex.RLock()
	defer h.LobbyService.Mutex.RUnlock()

	return c.JSON(map[string]interface{}{
		"status":         "online",
		"players_online": len(h.LobbyService.Connected),
	})
}

// WS Event
func (h *GameHandler) HandleWS(conn *websocket.Conn) {
	defer h.LobbyService.RemovePlayer(conn)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			h.LobbyService.RemovePlayer(conn)
			break
		}

		var event bjs.Event
		if err := json.Unmarshal(msg, &event); err != nil {
			continue
		}

		switch event.Type {
		case "create_lobby":
			var data struct {
				Name       string `json:"name"`
				Nick       string `json:"nick"`
				MaxPlayers int    `json:"maxPlayers"`
				UseBots    bool   `json:"useBots"`
			}
			if err := json.Unmarshal(event.Data, &data); err != nil {
				continue
			}
			lobby := h.LobbyService.CreateLobby(data.Name, data.MaxPlayers, data.UseBots)
			h.LobbyService.AddPlayer(conn, data.Nick, lobby)

		case "join_lobby":
			var data struct {
				Name string `json:"name"`
				Nick string `json:"nick"`
			}
			if err := json.Unmarshal(event.Data, &data); err != nil {
				continue
			}
			lobby, ok := h.LobbyService.GetLobby(data.Name)
			if !ok {
				continue
			}
			h.LobbyService.AddPlayer(conn, data.Nick, lobby)

		case "play_card":
			// pobierz lobby po conn
			lobbyName, ok := h.LobbyService.Connected[conn]
			if !ok {
				continue
			}

			lobby, ok := h.LobbyService.GetLobby(lobbyName)
			if !ok {
				continue
			}

			player, ok := lobby.Players[conn]
			if !ok {
				continue
			}

			h.GameService.PlayCard(conn, lobby, player)

		}
	}
}

// REST endpoint do tworzenia lobby
func (h *GameHandler) CreateLobby(c *fiber.Ctx) error {
	var req struct {
		Nick       string `json:"nick"`
		Name       string `json:"lobbyName"`
		MaxPlayers int    `json:"playerCount"`
		UseBots    bool   `json:"useBots"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(map[string]any{
			"success": false,
			"error":   "invalid request body",
		})
	}

	lobby := h.LobbyService.CreateLobby(req.Name, req.MaxPlayers, req.UseBots)
	h.LobbyService.AddPlayer(nil, req.Nick, lobby)

	resp := map[string]any{
		"success": true,
		"data": map[string]any{
			"name":       lobby.Name,
			"players":    h.PlayerList(lobby),
			"maxPlayers": lobby.MaxPlayers,
			"useBots":    lobby.UseBots,
			"started":    lobby.Started,
		},
	}

	return c.JSON(resp)
}

// Pomocnicza funkcja do listy graczy
func (h *GameHandler) PlayerList(lobby *bjs.Lobby) []string {
	lobby.Mutex.RLock()
	defer lobby.Mutex.RUnlock()
	players := []string{}
	for _, p := range lobby.Players {
		players = append(players, p.Nick)
	}
	return players
}

func (h *GameHandler) GetLobbies(c *fiber.Ctx) error {
	h.LobbyService.Mutex.RLock()
	defer h.LobbyService.Mutex.RUnlock()

	lobbies := []map[string]any{}
	for _, l := range h.LobbyService.Lobbies {
		l.Mutex.RLock()
		players := []string{}
		for _, p := range l.Players {
			players = append(players, p.Nick)
		}
		lobbies = append(lobbies, map[string]any{
			"name":       l.Name,
			"players":    players,
			"maxPlayers": l.MaxPlayers,
			"useBots":    l.UseBots,
			"started":    l.Started,
		})
		l.Mutex.RUnlock()
	}

	return c.JSON(map[string]any{
		"success": true,
		"data":    lobbies,
	})
}
