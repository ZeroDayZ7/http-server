package cache

import (
	"time"

	"github.com/patrickmn/go-cache"
)

var InteractionCache *cache.Cache

func Init() {
	// domyślny TTL 1h, czyszczenie co 10 minut
	InteractionCache = cache.New(1*time.Hour, 10*time.Minute)
}
