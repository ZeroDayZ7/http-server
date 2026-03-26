package service

import (
	"context"
	"time"
)

type InteractionRepository interface {
	GetCount(ctx context.Context, typ string) (int, error)
	Increment(ctx context.Context, typ string) error
}

type InteractionCache interface {
	TryRecordVisit(ctx context.Context, fp string, cooldown time.Duration) (bool, error)
	TryRecordInteraction(ctx context.Context, fp string, typ string, cooldown time.Duration) (bool, error)
	GetGlobalCount(ctx context.Context, typ string) (int, bool)
	SetGlobalCount(ctx context.Context, typ string, count int, ttl time.Duration) error
	GetUserChoice(ctx context.Context, fp string) (string, error)
}

type EventPublisher interface {
	PublishInteraction(ctx context.Context, typ string, fp string) error
}

type InteractionServiceInterface interface {
	GenerateFingerprint(ip, ua, lang string) string
	ProcessInitialVisit(ctx context.Context, fp string) (*StatsResponse, error)
	HandleInteraction(ctx context.Context, fp string, typ string) (*StatsResponse, error)
	GetStats(ctx context.Context, fp string) (*StatsResponse, error)
}
