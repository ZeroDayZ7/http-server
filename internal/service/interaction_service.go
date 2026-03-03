package service

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// Zmieniamy na dłuższy cooldown lub stałą blokadę, skoro używamy fingerprintu
	Cooldown    = 24 * time.Hour
	TypeLike    = "like"
	TypeDislike = "dislike"
	TypeVisit   = "visit"
)

type InteractionRepository interface {
	Increment(ctx context.Context, typ string) error
	// Najlepiej pobrać wszystko jednym zapytaniem sqlc, ale trzymamy Twój interfejs:
	GetCount(ctx context.Context, typ string) (int, error)
}

type InteractionService struct {
	repo  InteractionRepository
	redis *redis.Client
}

func NewInteractionService(repo InteractionRepository, redis *redis.Client) *InteractionService {
	return &InteractionService{
		repo:  repo,
		redis: redis,
	}
}

type StatsResponse struct {
	Likes      int     `json:"likes"`
	Dislikes   int     `json:"dislikes"`
	Visits     int     `json:"visits"`
	Allowed    bool    `json:"allowed"`
	UserChoice *string `json:"userChoice"`
	Message    string  `json:"message"`
}

// HandleInteraction - obsługuje kliknięcie like/dislike
func (s *InteractionService) HandleInteraction(ctx context.Context, fingerprint string, typ string) (*StatsResponse, error) {
	// UJEDNOLICONY KLUCZ
	limitKey := "user:interaction:" + fingerprint

	if typ == TypeVisit {
		_ = s.repo.Increment(ctx, typ)
		return s.GetStats(ctx, fingerprint)
	}

	// 1. Sprawdź blokadę (czy klucz istnieje)
	exists, _ := s.redis.Exists(ctx, limitKey).Result()
	if exists > 0 {
		return s.GetStats(ctx, fingerprint)
	}

	// 2. MySQL
	if err := s.repo.Increment(ctx, typ); err != nil {
		return nil, err
	}

	// 3. ZAPISZ WYBÓR (Wartość to 'like' lub 'dislike')
	_ = s.redis.Set(ctx, limitKey, typ, Cooldown).Err()

	return s.GetStats(ctx, fingerprint)
}

func (s *InteractionService) GetStats(ctx context.Context, fingerprint string) (*StatsResponse, error) {
	likes, _ := s.repo.GetCount(ctx, TypeLike)
	dislikes, _ := s.repo.GetCount(ctx, TypeDislike)
	visits, _ := s.repo.GetCount(ctx, TypeVisit)

	limitKey := "user:interaction:" + fingerprint
	val, err := s.redis.Get(ctx, limitKey).Result()

	allowed := true
	var choicePtr *string

	if err == nil && val != "" {
		allowed = false
		v := val
		choicePtr = &v
	}

	return &StatsResponse{
		Likes:      likes,
		Dislikes:   dislikes,
		Visits:     visits,
		Allowed:    allowed,
		UserChoice: choicePtr,
		Message:    "Statystyki pobrane",
	}, nil
}
