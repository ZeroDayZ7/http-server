package service

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Stałe dla typów i cooldownu
const (
	Cooldown    = 1 * time.Hour
	TypeLike    = "like"
	TypeDislike = "dislike"
	TypeVisit   = "visit"
)

// InteractionRepository musi pasować do Twojego repozytorium w mysql!
type InteractionRepository interface {
	Increment(ctx context.Context, typ string) error
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

// StatsResponse struktura odpowiedzi dla frontendu
type StatsResponse struct {
	Likes    int    `json:"likes"`
	Dislikes int    `json:"dislikes"`
	Visits   int    `json:"visits"`
	Allowed  bool   `json:"allowed"`
	Message  string `json:"message"`
}

// HandleInteraction główna logika: sprawdza Redis (fingerprint) i aktualizuje MySQL
func (s *InteractionService) HandleInteraction(ctx context.Context, fingerprint string, typ string) (*StatsResponse, error) {
	// 1. Sprawdź blokadę w Redis (Fingerprint zastępuje IP - RODO SAFE)
	// Klucz np. "limit:like:fp_123"
	key := "limit:" + typ + ":" + fingerprint

	exists, err := s.redis.Exists(ctx, key).Result()
	if err == nil && exists > 0 {
		// Użytkownik już klikał w ciągu ostatniej godziny
		stats, _ := s.GetStats(ctx)
		stats.Allowed = false
		stats.Message = "Już zarejestrowano tę interakcję"
		return stats, nil
	}

	// 2. Jeśli to nie jest tylko pobieranie, zwiększ licznik w bazie
	if typ == TypeLike || typ == TypeDislike || typ == TypeVisit {
		if err := s.repo.Increment(ctx, typ); err != nil {
			return nil, err
		}
		// 3. Ustaw blokadę w Redis na 1h
		s.redis.Set(ctx, key, "true", Cooldown)
	}

	// 4. Pobierz aktualne statystyki
	stats, err := s.GetStats(ctx)
	if err != nil {
		return nil, err
	}
	stats.Allowed = false // Właśnie kliknął, więc teraz blokujemy
	stats.Message = "Interakcja zapisana"

	return stats, nil
}

// GetStats pobiera same liczby z bazy
func (s *InteractionService) GetStats(ctx context.Context) (*StatsResponse, error) {
	likes, _ := s.repo.GetCount(ctx, TypeLike)
	dislikes, _ := s.repo.GetCount(ctx, TypeDislike)
	visits, _ := s.repo.GetCount(ctx, TypeVisit)

	return &StatsResponse{
		Likes:    likes,
		Dislikes: dislikes,
		Visits:   visits,
		Allowed:  true,
	}, nil
}
