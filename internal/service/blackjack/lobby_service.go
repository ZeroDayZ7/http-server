package blackjack

import (
	"encoding/json"
	"sync"

	"github.com/gofiber/websocket/v2"
)

type Lobby struct {
	Name       string
	Players    map[*websocket.Conn]*Player
	MaxPlayers int
	UseBots    bool
	Started    bool
	Mutex      sync.RWMutex
}

type LobbyService struct {
	Lobbies   map[string]*Lobby
	Mutex     sync.RWMutex
	Connected map[*websocket.Conn]string
}

func NewLobbyService() *LobbyService {
	return &LobbyService{
		Lobbies:   make(map[string]*Lobby),
		Connected: make(map[*websocket.Conn]string),
	}
}

func (ls *LobbyService) CreateLobby(name string, maxPlayers int, useBots bool) *Lobby {
	ls.Mutex.Lock()
	defer ls.Mutex.Unlock()
	lobby := &Lobby{
		Name:       name,
		Players:    make(map[*websocket.Conn]*Player),
		MaxPlayers: maxPlayers,
		UseBots:    useBots,
	}
	ls.Lobbies[name] = lobby
	return lobby
}

func (ls *LobbyService) GetLobby(name string) (*Lobby, bool) {
	ls.Mutex.RLock()
	defer ls.Mutex.RUnlock()
	l, ok := ls.Lobbies[name]
	return l, ok
}

func (ls *LobbyService) AddPlayer(conn *websocket.Conn, nick string, lobby *Lobby) {
	lobby.Mutex.Lock()
	defer lobby.Mutex.Unlock()

	player := &Player{Nick: nick, Hand: []Card{}, Score: 0}

	// Dodajemy do mapy tylko jeśli mamy websocket
	if conn != nil {
		lobby.Players[conn] = player
		ls.Connected[conn] = lobby.Name
	}

	// Zawsze możemy mieć "REST-only" listę graczy do zwracania
	// np. lobby.PlayerList = append(lobby.PlayerList, player)

	ls.BroadcastLobby(lobby)
}

func (ls *LobbyService) RemovePlayer(conn *websocket.Conn) {
	lobbyName, ok := ls.Connected[conn]
	if !ok {
		return
	}

	lobby, ok := ls.GetLobby(lobbyName)
	if !ok {
		return
	}

	lobby.Mutex.Lock()
	defer lobby.Mutex.Unlock()
	delete(lobby.Players, conn)
	delete(ls.Connected, conn)

	ls.BroadcastLobby(lobby)
}

func (ls *LobbyService) BroadcastLobby(lobby *Lobby) {
	lobby.Mutex.RLock()
	defer lobby.Mutex.RUnlock()

	players := []Player{}
	for _, p := range lobby.Players {
		players = append(players, *p)
	}

	data, _ := json.Marshal(map[string]interface{}{
		"type":    "lobby_state",
		"players": players,
		"name":    lobby.Name,
	})

	for conn := range lobby.Players {
		conn.WriteMessage(websocket.TextMessage, data)
	}
}
