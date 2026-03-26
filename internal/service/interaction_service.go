package service

import (
	"context"
	"errors"
	"fmt"
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
	if err != nil {
		return nil, fmt.Errorf("visit cache failed: %w", err)
	}

	if recorded {
		if err := s.publisher.PublishInteraction(ctx, TypeVisit, fp); err != nil {
			return nil, fmt.Errorf("publish visit failed: %w", err)
		}
	}

	return s.GetStats(ctx, fp)
}

func (s *InteractionService) HandleInteraction(ctx context.Context, fp string, typ string) (*StatsResponse, error) {
	if fp == "" || (typ != TypeLike && typ != TypeDislike) {
		return nil, errors.New("invalid request")
	}

	recorded, err := s.cache.TryRecordInteraction(ctx, fp, typ, Cooldown)
	if err != nil {
		return nil, fmt.Errorf("interaction cache failed: %w", err)
	}

	if recorded {
		if err := s.publisher.PublishInteraction(ctx, typ, fp); err != nil {
			return nil, fmt.Errorf("publish interaction failed: %w", err)
		}
	}

	return s.GetStats(ctx, fp)
}

func (s *InteractionService) GetStats(ctx context.Context, fp string) (*StatsResponse, error) {
	likes, err := s.getGlobalCount(ctx, TypeLike)
	if err != nil {
		return nil, err
	}

	dislikes, err := s.getGlobalCount(ctx, TypeDislike)
	if err != nil {
		return nil, err
	}

	visits, err := s.getGlobalCount(ctx, TypeVisit)
	if err != nil {
		return nil, err
	}

	val, err := s.cache.GetUserChoice(ctx, fp)
	if err != nil {
		val = ""
	}

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

func (s *InteractionService) getGlobalCount(ctx context.Context, typ string) (int, error) {
	if val, ok := s.cache.GetGlobalCount(ctx, typ); ok {
		return val, nil
	}

	count, err := s.repo.GetCount(ctx, typ)
	if err != nil {
		return 0, fmt.Errorf("db get count failed (%s): %w", typ, err)
	}

	if err := s.cache.SetGlobalCount(ctx, typ, count, GlobalStatsTTL); err != nil {
		return count, nil
	}

	return count, nil
}
