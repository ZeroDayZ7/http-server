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

	if err != nil {
		if isBusyGroupErr(err) {
			return nil
		}
		return err
	}

	return nil
}

func isBusyGroupErr(err error) bool {
	return strings.HasPrefix(err.Error(), "BUSYGROUP")
}
