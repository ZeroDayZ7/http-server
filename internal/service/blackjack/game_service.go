package blackjack

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/gofiber/websocket/v2"
)

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
	rand.Seed(time.Now().UnixNano())
	return &GameService{Deck: deck}
}

func (gs *GameService) DrawCard() Card {
	if len(gs.Deck) == 0 {
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

func (gs *GameService) PlayCard(conn *websocket.Conn, lobby *Lobby, player *Player) {
	lobby.Mutex.Lock()
	defer lobby.Mutex.Unlock()

	card := gs.DrawCard()
	player.Hand = append(player.Hand, card)
	player.Score = gs.CalculateScore(player.Hand)

	// Broadcast aktualnego stanu lobby
	lobbyService := &LobbyService{} // w Twoim przypadku możesz przekazać lub użyć z handlera
	lobbyService.BroadcastLobby(lobby)
}
