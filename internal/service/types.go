package service

import "time"

const (
	Cooldown       = 24 * time.Hour
	GlobalStatsTTL = 2 * time.Minute

	TypeLike    = "like"
	TypeDislike = "dislike"
	TypeVisit   = "visit"
)

type StatsResponse struct {
	Likes      int     `json:"likes"`
	Dislikes   int     `json:"dislikes"`
	Visits     int     `json:"visits"`
	Allowed    bool    `json:"allowed"`
	UserChoice *string `json:"userChoice"`
	Message    string  `json:"message"`
}
