package redis

const (
	RedisUserInteractionPrefix = "user:interaction:"
	RedisGlobalStatsPrefix     = "global:stats:"
	RedisVisitCooldownPrefix   = "visit:cooldown:"
)

type RedisKeyBuilder struct{}

func (r RedisKeyBuilder) UserInteraction(fp string) string {
	return RedisUserInteractionPrefix + fp
}

func (r RedisKeyBuilder) GlobalStats(typ string) string {
	return RedisGlobalStatsPrefix + typ
}

func (r RedisKeyBuilder) VisitCooldown(fp string) string {
	return RedisVisitCooldownPrefix + fp
}
