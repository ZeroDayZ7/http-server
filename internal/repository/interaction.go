package repository

import "time"

type InteractionRepository interface {
	// Add dodaje nową interakcję: visit, like, dislike, comment itd.
	Add(ip string, userID *uint, typ string, value int, content *string, at time.Time) error

	// Count zwraca liczbę wszystkich interakcji danego typu
	Count(typ string) (int, error)

	// CountByIP zwraca liczbę interakcji danego typu dla konkretnego IP
	CountByIP(ip, typ string) (int, error)
}
