package redis

import (
	_ "embed"

	"github.com/redis/go-redis/v9"
)

// Scripts przechowuje wszystkie skrypty Lua
type Scripts struct {
	GetStats          *redis.Script
	RecordInteraction *redis.Script
	RecordVisit       *redis.Script
}

// Embedujemy skrypty
//
//go:embed scripts/get_stats.lua
var getStatsScript string

//go:embed scripts/record_interaction.lua
var recordInteractionScript string

//go:embed scripts/visit.lua
var recordVisitScript string

// DefaultScripts - gotowy zestaw skryptów
var DefaultScripts = &Scripts{
	GetStats:          redis.NewScript(getStatsScript),
	RecordInteraction: redis.NewScript(recordInteractionScript),
	RecordVisit:       redis.NewScript(recordVisitScript),
}
