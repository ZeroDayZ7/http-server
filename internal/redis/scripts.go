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

//go:embed scripts/get_stats.lua
var getStatsScript string

//go:embed scripts/record_interaction.lua
var recordInteractionScript string

//go:embed scripts/visit.lua
var recordVisitScript string

var DefaultScripts = &Scripts{
	GetStats:          redis.NewScript(getStatsScript),
	RecordInteraction: redis.NewScript(recordInteractionScript),
	RecordVisit:       redis.NewScript(recordVisitScript),
}
