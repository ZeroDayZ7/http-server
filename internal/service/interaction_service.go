package service

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	rediskeys "github.com/zerodayz7/http-server/internal/redis"
)

const (
	Cooldown       = 24 * time.Hour
	GlobalStatsTTL = 2 * time.Minute

	TypeLike    = "like"
	TypeDislike = "dislike"
	TypeVisit   = "visit"
)

type InteractionRepository interface {
	Increment(ctx context.Context, typ string) error
	GetCount(ctx context.Context, typ string) (int, error)
}

type InteractionService struct {
	repo      InteractionRepository
	redis     *redis.Client
	keys      rediskeys.RedisKeyBuilder
	publisher rediskeys.StreamPublisher
}

func NewInteractionService(
	repo InteractionRepository,
	redisClient *redis.Client,
	publisher rediskeys.StreamPublisher,
) *InteractionService {

	return &InteractionService{
		repo:      repo,
		redis:     redisClient,
		keys:      rediskeys.RedisKeyBuilder{},
		publisher: publisher,
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

func (s *InteractionService) ProcessInitialVisit(ctx context.Context, fp string) (*StatsResponse, error) {

	limitKey := s.keys.VisitCooldown(fp)
	statsKey := s.keys.GlobalStats(TypeVisit)

	result, err := rediskeys.DefaultScripts.Visit.Run(
		ctx,
		s.redis,
		[]string{
			limitKey,
			statsKey,
		},
		Cooldown.Milliseconds(),
	).Int()

	if err == nil && result == 1 {
		_ = s.publisher.PublishInteraction(ctx, TypeVisit, fp)
	}

	return s.GetStats(ctx, fp)
}

func (s *InteractionService) HandleInteraction(ctx context.Context, fp string, typ string) (*StatsResponse, error) {

	if fp == "" || (typ != TypeLike && typ != TypeDislike) {
		return nil, errors.New("invalid request")
	}

	cooldownKey := s.keys.UserInteraction(fp)
	statsKey := s.keys.GlobalStats(typ)

	result, err := rediskeys.DefaultScripts.Interaction.Run(
		ctx,
		s.redis,
		[]string{
			cooldownKey,
			statsKey,
		},
		Cooldown.Milliseconds(),
	).Int()

	if err == nil && result == 1 {
		_ = s.publisher.PublishInteraction(ctx, typ, fp)
	}

	return s.GetStats(ctx, fp)
}

func (s *InteractionService) getGlobalCount(ctx context.Context, typ string) int {
	cacheKey := s.keys.GlobalStats(typ)

	// 1. cache
	val, err := s.redis.Get(ctx, cacheKey).Int()
	if err == nil {
		return val
	}

	// 2. Cache miss
	count, err := s.repo.GetCount(ctx, typ)
	if err != nil {
		return 0
	}

	_ = s.redis.Set(ctx, cacheKey, count, GlobalStatsTTL).Err()

	return count
}

func (s *InteractionService) GetStats(ctx context.Context, fp string) (*StatsResponse, error) {

	likes := s.getGlobalCount(ctx, TypeLike)
	dislikes := s.getGlobalCount(ctx, TypeDislike)
	visits := s.getGlobalCount(ctx, TypeVisit)

	val, _ := s.redis.Get(ctx, s.keys.UserInteraction(fp)).Result()

	var choicePtr *string
	allowed := true

	if val != "" {
		v := val
		choicePtr = &v
		allowed = false
	}

	return &StatsResponse{
		Likes:      likes,
		Dislikes:   dislikes,
		Visits:     visits,
		Allowed:    allowed,
		UserChoice: choicePtr,
		Message:    "Success",
	}, nil
}
