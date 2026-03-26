package service

import (
	"context"
	"errors"
)

type InteractionService struct {
	repo      InteractionRepository
	cache     InteractionCache
	publisher EventPublisher
	identity  IdentityService
}

func NewInteractionService(
	repo InteractionRepository,
	cache InteractionCache,
	publisher EventPublisher,
	identity IdentityService,
) *InteractionService {
	return &InteractionService{
		repo:      repo,
		cache:     cache,
		publisher: publisher,
		identity:  identity,
	}
}

func (s *InteractionService) GenerateFingerprint(ip, ua, lang string) string {
	return s.identity.GenerateFingerprint(ip, ua, lang)
}

func (s *InteractionService) ProcessInitialVisit(ctx context.Context, fp string) (*StatsResponse, error) {
	recorded, err := s.cache.TryRecordVisit(ctx, fp, Cooldown)

	if err == nil && recorded {
		_ = s.publisher.PublishInteraction(ctx, TypeVisit, fp)
	}

	return s.GetStats(ctx, fp)
}

func (s *InteractionService) HandleInteraction(ctx context.Context, fp string, typ string) (*StatsResponse, error) {
	if fp == "" || (typ != TypeLike && typ != TypeDislike) {
		return nil, errors.New("invalid request")
	}

	recorded, err := s.cache.TryRecordInteraction(ctx, fp, typ, Cooldown)

	if err == nil && recorded {
		_ = s.publisher.PublishInteraction(ctx, typ, fp)
	}

	return s.GetStats(ctx, fp)
}

func (s *InteractionService) GetStats(ctx context.Context, fp string) (*StatsResponse, error) {
	likes := s.getGlobalCount(ctx, TypeLike)
	dislikes := s.getGlobalCount(ctx, TypeDislike)
	visits := s.getGlobalCount(ctx, TypeVisit)

	val, _ := s.cache.GetUserChoice(ctx, fp)

	var choicePtr *string
	allowed := true
	if val != "" {
		choicePtr = &val
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

func (s *InteractionService) getGlobalCount(ctx context.Context, typ string) int {
	if val, ok := s.cache.GetGlobalCount(ctx, typ); ok {
		return val
	}

	count, err := s.repo.GetCount(ctx, typ)
	if err != nil {
		return 0
	}

	_ = s.cache.SetGlobalCount(ctx, typ, count, GlobalStatsTTL)
	return count
}
