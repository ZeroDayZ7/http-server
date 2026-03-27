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
	Likes      int64   `json:"likes"`
	Dislikes   int64   `json:"dislikes"`
	Visits     int64   `json:"visits"`
	Allowed    bool    `json:"allowed"`
	UserChoice *string `json:"user_choice"`
	Message    string  `json:"message"`
}

type InteractionStatsDTO struct {
	Likes    int64
	Dislikes int64
	Visits   int64
}
