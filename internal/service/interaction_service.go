package service

import (
	"time"

	"github.com/zerodayz7/http-server/internal/repository/mysql"
)

// Constants
const (
	VisitCooldown = 1 * time.Hour // Ograniczenie wizyt do jednej na godzinę
	LikeCooldown  = 1 * time.Hour
	TypeLike      = "like"
	TypeDislike   = "dislike"
	TypeVisit     = "visit"
)

type InteractionService struct {
	repo *mysql.MySQLInteractionRepo
}

func NewInteractionService(repo *mysql.MySQLInteractionRepo) *InteractionService {
	return &InteractionService{repo: repo}
}

// Record zapisuje interakcję (visit, like, comment)
func (s *InteractionService) Record(ip string, userID *uint, typ string, value int, content *string) error {
	return s.repo.Add(ip, userID, typ, value, content, time.Now())
}

// CountByType zwraca liczbę interakcji danego typu
func (s *InteractionService) CountByType(typ string) (int, error) {
	return s.repo.Count(typ)
}

// GetLastInteractionByIP zwraca czas ostatniej interakcji danego typu dla IP
func (s *InteractionService) GetLastInteractionByIP(ip, typ string) (time.Time, error) {
	return s.repo.GetLastInteraction(ip, typ)
}

// GetLastVisitByIP zwraca czas ostatniej wizyty dla IP (można usunąć, jeśli używamy GetLastInteractionByIP)
func (s *InteractionService) GetLastVisitByIP(ip string) (time.Time, error) {
	return s.GetLastInteractionByIP(ip, TypeVisit)
}

// HandleInteraction obsługuje logikę biznesową dla interakcji
func (s *InteractionService) HandleInteraction(ip string, userID *uint, interactionType string, cooldown time.Duration, record bool) (map[string]interface{}, error) {
	resp := make(map[string]interface{})

	// Check last interaction
	var otherType string
	switch interactionType {
	case TypeLike:
		otherType = TypeDislike
	case TypeDislike:
		otherType = TypeLike
	}

	// Get stats
	likes, err := s.CountByType(TypeLike)
	if err != nil {
		return nil, err
	}
	dislikes, err := s.CountByType(TypeDislike)
	if err != nil {
		return nil, err
	}
	visits, err := s.CountByType(TypeVisit)
	if err != nil {
		return nil, err
	}

	// Check user's last interaction (like or dislike)
	var userChoice *string
	var allowed = true
	lastLike, _ := s.GetLastInteractionByIP(ip, TypeLike)
	lastDislike, _ := s.GetLastInteractionByIP(ip, TypeDislike)

	if !lastLike.IsZero() && time.Since(lastLike) < LikeCooldown {
		choice := TypeLike
		userChoice = &choice
		allowed = false
	} else if !lastDislike.IsZero() && time.Since(lastDislike) < LikeCooldown {
		choice := TypeDislike
		userChoice = &choice
		allowed = false
	}

	// Check if interaction is allowed
	shouldRecord := record
	if cooldown > 0 {
		last, err := s.GetLastInteractionByIP(ip, interactionType)
		if err != nil {
			return nil, err
		}
		if !last.IsZero() && time.Since(last) < cooldown {
			resp["allowed"] = interactionType != TypeVisit // Wizyty nie blokują like/dislike
			resp["message"] = "interakcja już zarejestrowana"
			shouldRecord = false
		}

		// Check other interaction type (for like/dislike)
		if otherType != "" {
			lastOther, err := s.GetLastInteractionByIP(ip, otherType)
			if err != nil {
				return nil, err
			}
			if !lastOther.IsZero() && time.Since(lastOther) < cooldown {
				resp["allowed"] = false
				resp["message"] = "inna interakcja już zarejestrowana"
				shouldRecord = false
			}
		}
	}

	// Record interaction if allowed
	if shouldRecord {
		if err := s.Record(ip, userID, interactionType, 0, nil); err != nil {
			return nil, err
		}
		// Refresh stats after recording
		likes, err = s.CountByType(TypeLike)
		if err != nil {
			return nil, err
		}
		dislikes, err = s.CountByType(TypeDislike)
		if err != nil {
			return nil, err
		}
		visits, err = s.CountByType(TypeVisit)
		if err != nil {
			return nil, err
		}
	}

	resp["ip"] = ip
	resp["visits"] = visits
	resp["likes"] = likes
	resp["dislikes"] = dislikes
	resp["allowed"] = allowed
	resp["userChoice"] = userChoice
	resp["message"] = "interakcja zarejestrowana"

	if shouldRecord && interactionType != TypeVisit {
		resp["userChoice"] = interactionType
		resp["allowed"] = false
	}

	return resp, nil
}
