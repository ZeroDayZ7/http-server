package blackjack

type Player struct {
	Nick  string `json:"nick"`
	Hand  []Card `json:"hand"`
	Score int    `json:"score"`
}
