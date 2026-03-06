package redis

import (
	"context"
	"strings"

	"github.com/redis/go-redis/v9"
)

func SetupStreamGroup(ctx context.Context, r *redis.Client) error {

	_, err := r.XGroupCreateMkStream(
		ctx,
		"interaction_events",
		"interaction_workers",
		"0",
	).Result()

	// Redis zwraca błąd jeśli grupa już istnieje
	if err != nil {
		if strings.Contains(err.Error(), "BUSYGROUP") {
			return nil
		}
		return err
	}

	return nil
}
