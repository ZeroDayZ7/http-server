package redis

import (
	_ "embed"

	"github.com/redis/go-redis/v9"
)

type Scripts struct {
	GetStats          *redis.Script
	RecordInteraction *redis.Script
	RecordVisit       *redis.Script
}

var getStatsScript string

var recordInteractionScript string

var recordVisitScript string

var DefaultScripts = &Scripts{
	GetStats:          redis.NewScript(getStatsScript),
	RecordInteraction: redis.NewScript(recordInteractionScript),
	RecordVisit:       redis.NewScript(recordVisitScript),
}
