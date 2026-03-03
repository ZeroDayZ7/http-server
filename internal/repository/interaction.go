package repository

import (
	"context"
)

type InteractionRepository interface {
	// Increment zwiększa licznik dla danego typu (like, dislike, visit)
	Increment(ctx context.Context, typ string) error

	// GetCount pobiera aktualny stan licznika dla danego typu
	GetCount(ctx context.Context, typ string) (int, error)
}
