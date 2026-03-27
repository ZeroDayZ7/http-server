package service

import (
	"context"
	"time"
)

type InteractionRepository interface {
	Increment(ctx context.Context, typ string) error
	IncrementBy(ctx context.Context, typ string, amount int64) error
	GetStats(ctx context.Context) (likes int64, dislikes int64, visits int64, err error)
}

type InteractionCache interface {
	TryRecordVisit(ctx context.Context, fp string, cooldown time.Duration) (bool, error)
	TryRecordInteraction(ctx context.Context, fp string, typ string, cooldown time.Duration) (bool, error)
	GetGlobalCount(ctx context.Context, typ string) (int64, bool)
	SetGlobalCount(ctx context.Context, typ string, count int64, ttl time.Duration) error
	GetUserChoice(ctx context.Context, fp string) (string, bool, error)
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

type IdentityService interface {
	GenerateFingerprint(ip, ua, lang string) string
}
