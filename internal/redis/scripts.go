package redis

import (
	_ "embed"

	"github.com/redis/go-redis/v9"
)

type Scripts struct {
	Visit       *redis.Script
	Interaction *redis.Script
	Stats       *redis.Script
}

//go:embed scripts/visit.lua
var visitScript string

//go:embed scripts/interaction.lua
var interactionScript string

//go:embed scripts/stats.lua
var statsScript string

var DefaultScripts = Scripts{
	Visit:       redis.NewScript(visitScript),
	Interaction: redis.NewScript(interactionScript),
	Stats:       redis.NewScript(statsScript),
}
