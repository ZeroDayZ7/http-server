package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/zerodayz7/http-server/internal/shared/logger"
	"golang.org/x/sync/singleflight"
)

type InteractionService struct {
	repo         InteractionRepository
	cache        InteractionCache
	publisher    EventPublisher
	identity     IdentityService
	log          logger.Logger
	requestGroup singleflight.Group
}

func NewInteractionService(
	repo InteractionRepository,
	cache InteractionCache,
	publisher EventPublisher,
	identity IdentityService,
	log logger.Logger,
) *InteractionService {
	return &InteractionService{
		repo:      repo,
		cache:     cache,
		publisher: publisher,
		identity:  identity,
		log:       log,
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
	likes, hasLikes := s.cache.GetGlobalCount(ctx, TypeLike)
	dislikes, hasDislikes := s.cache.GetGlobalCount(ctx, TypeDislike)
	visits, hasVisits := s.cache.GetGlobalCount(ctx, TypeVisit)

	if !hasLikes || !hasDislikes || !hasVisits {
		v, err, shared := s.requestGroup.Do("get_global_stats", func() (interface{}, error) {
			s.log.Infow("Cache miss, fetching from DB (SingleFlight leader)",
				"hasLikes", hasLikes, "hasDislikes", hasDislikes, "hasVisits", hasVisits)

			dbLikes, dbDislikes, dbVisits, err := s.repo.GetStats(ctx)
			if err != nil {
				return nil, err
			}

			s.cache.SetGlobalCount(ctx, TypeLike, dbLikes, GlobalStatsTTL)
			s.cache.SetGlobalCount(ctx, TypeDislike, dbDislikes, GlobalStatsTTL)
			s.cache.SetGlobalCount(ctx, TypeVisit, dbVisits, GlobalStatsTTL)

			return []int64{dbLikes, dbDislikes, dbVisits}, nil
		})

		if err != nil {
			s.log.Errorw("Failed to fetch stats from repository via singleflight", "err", err)
		} else {
			res := v.([]int64)
			likes, dislikes, visits = res[0], res[1], res[2]

			if shared {
				s.log.Debugw("Result shared via singleflight", "fp", fp)
			}
		}
	}

	val, foundChoice, err := s.cache.GetUserChoice(ctx, fp)
	if err != nil {
		s.log.Warnw("Redis GetUserChoice failed", "fp", fp, "err", err)
	}

	var choicePtr *string
	if foundChoice && val != "" {
		choicePtr = &val
	}

	return &StatsResponse{
		Likes:      likes,
		Dislikes:   dislikes,
		Visits:     visits,
		Allowed:    !foundChoice,
		UserChoice: choicePtr,
		Message:    "Success",
	}, nil
}
