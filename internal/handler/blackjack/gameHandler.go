package blackjack

import (
	"encoding/json"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/websocket/v2"
)

// --- MODELE ---
type Card struct {
	Suit  string `json:"suit"`
	Value string `json:"value"`
}

type Player struct {
	Nick  string `json:"nick"`
	Hand  []Card `json:"hand"`
	Score int    `json:"score"`
	IsBot bool   `json:"isBot"`
}

type Lobby struct {
	Name       string
	Players    map[*websocket.Conn]*Player
	MaxPlayers int
	UseBots    bool
	Started    bool
	Mutex      sync.RWMutex
	Timer      *time.Timer
}

// --- SERWIS LOBBY ---
type LobbyService struct {
	Lobbies map[string]*Lobby
	Mutex   sync.RWMutex
}

func NewLobbyService() *LobbyService {
	return &LobbyService{
		Lobbies: make(map[string]*Lobby),
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

// --- SERWIS GRY ---
type GameService struct {
	Deck []Card
}

func NewGameService() *GameService {
	suits := []string{"Hearts", "Diamonds", "Clubs", "Spades"}
	values := []string{"2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A"}
	deck := make([]Card, 0, 52)
	for _, suit := range suits {
		for _, value := range values {
			deck = append(deck, Card{Suit: suit, Value: value})
		}
	}
	return &GameService{Deck: deck}
}

func (gs *GameService) DrawCard() Card {
	if len(gs.Deck) == 0 {
		// przetasuj ponownie lub reset
		return Card{}
	}
	idx := rand.Intn(len(gs.Deck))
	card := gs.Deck[idx]
	gs.Deck = append(gs.Deck[:idx], gs.Deck[idx+1:]...)
	return card
}

func (gs *GameService) CalculateScore(hand []Card) int {
	score := 0
	aces := 0
	for _, card := range hand {
		switch card.Value {
		case "A":
			aces++
			score += 11
		case "K", "Q", "J", "10":
			score += 10
		default:
			val, _ := strconv.Atoi(card.Value)
			score += val
		}
	}
	for aces > 0 && score > 21 {
		score -= 10
		aces--
	}
	return score
}

// --- HANDLER ---
type GameHandler struct {
	LobbyService *LobbyService
	GameService  *GameService
	Connected    map[*websocket.Conn]string // ws -> lobbyName
	Mutex        sync.RWMutex
}

func NewGameHandler() *GameHandler {
	return &GameHandler{
		LobbyService: NewLobbyService(),
		GameService:  NewGameService(),
		Connected:    make(map[*websocket.Conn]string),
	}
}

// --- WS ---
func (h *GameHandler) HandleWS(conn *websocket.Conn) {
	defer conn.Close()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			h.removePlayer(conn)
			break
		}

		var payload map[string]interface{}
		if err := json.Unmarshal(msg, &payload); err != nil {
			continue
		}

		switch payload["type"] {
		case "create_lobby":
			name := payload["name"].(string)
			maxPlayers := int(payload["maxPlayers"].(float64))
			useBots := payload["useBots"].(bool)
			lobby := h.LobbyService.CreateLobby(name, maxPlayers, useBots)
			h.addPlayerToLobby(conn, payload["nick"].(string), lobby)
			h.startLobbyTimer(lobby)
		case "join_lobby":
			name := payload["name"].(string)
			nick := payload["nick"].(string)
			if lobby, ok := h.LobbyService.GetLobby(name); ok {
				h.addPlayerToLobby(conn, nick, lobby)
			}
		case "play_card":
			h.playCard(conn)
		}
	}
}

// --- FUNKCJE WS ---
func (h *GameHandler) addPlayerToLobby(conn *websocket.Conn, nick string, lobby *Lobby) {
	lobby.Mutex.Lock()
	defer lobby.Mutex.Unlock()

	lobby.Players[conn] = &Player{Nick: nick, Hand: []Card{}, Score: 0, IsBot: false}

	h.Mutex.Lock()
	h.Connected[conn] = lobby.Name
	h.Mutex.Unlock()

	h.broadcastLobbyState(lobby)
}

func (h *GameHandler) removePlayer(conn *websocket.Conn) {
	h.Mutex.Lock()
	lobbyName, ok := h.Connected[conn]
	delete(h.Connected, conn)
	h.Mutex.Unlock()

	if !ok {
		return
	}

	if lobby, ok := h.LobbyService.GetLobby(lobbyName); ok {
		lobby.Mutex.Lock()
		delete(lobby.Players, conn)
		lobby.Mutex.Unlock()
		h.broadcastLobbyState(lobby)
	}
}

func (h *GameHandler) broadcastLobbyState(lobby *Lobby) {
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

// --- TIMER ---
func (h *GameHandler) startLobbyTimer(lobby *Lobby) {
	lobby.Mutex.Lock()
	if lobby.Started || lobby.Timer != nil {
		lobby.Mutex.Unlock()
		return
	}
	lobby.Timer = time.AfterFunc(30*time.Second, func() {
		h.startGame(lobby)
	})
	lobby.Mutex.Unlock()
}

// --- START GRY ---
func (h *GameHandler) startGame(lobby *Lobby) {
	lobby.Mutex.Lock()
	lobby.Started = true

	// Dodaj botów jeśli brakuje graczy
	for len(lobby.Players) < lobby.MaxPlayers && lobby.UseBots {
		bot := &Player{Nick: "Bot" + strconv.Itoa(len(lobby.Players)+1), Hand: []Card{}, Score: 0, IsBot: true}
		lobby.Players[nil] = bot // nil bo bot nie ma WS
	}

	// Rozdaj pierwsze karty
	for conn, player := range lobby.Players {
		for i := 0; i < 2; i++ {
			card := h.GameService.DrawCard()
			player.Hand = append(player.Hand, card)
		}
		player.Score = h.GameService.CalculateScore(player.Hand)
		if conn != nil {
			h.broadcastLobbyState(lobby)
		}
	}
	lobby.Mutex.Unlock()
}

// --- GRA ---
func (h *GameHandler) playCard(conn *websocket.Conn) {
	h.Mutex.RLock()
	lobbyName := h.Connected[conn]
	h.Mutex.RUnlock()

	lobby, ok := h.LobbyService.GetLobby(lobbyName)
	if !ok {
		return
	}

	lobby.Mutex.Lock()
	defer lobby.Mutex.Unlock()

	player, ok := lobby.Players[conn]
	if !ok {
		return
	}

	card := h.GameService.DrawCard()
	player.Hand = append(player.Hand, card)
	player.Score = h.GameService.CalculateScore(player.Hand)

	h.broadcastLobbyState(lobby)
}
