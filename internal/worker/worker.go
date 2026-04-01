package worker

import (
	"context"
	"time"

	"github.com/sony/gobreaker"
)

type HealthStatus struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks"`
	Time   time.Time         `json:"time"`
}

func (w *InteractionWorker) HealthCheck(ctx context.Context) (bool, HealthStatus) {
	isHealthy := true
	checks := make(map[string]string)

	// 1. Sprawdź Redis - używamy poprawnej nazwy pola: redisClient
	if err := w.redisClient.Ping(ctx).Err(); err != nil {
		checks["redis"] = "DOWN: " + err.Error()
		isHealthy = false
	} else {
		checks["redis"] = "UP"
	}

	// 2. Sprawdź Circuit Breaker (bazę danych)
	state := w.cb.State()
	checks["database_breaker"] = state.String()

	// Jeśli Circuit Breaker jest otwarty, uznajemy serwis za DOWN
	if state == gobreaker.StateOpen {
		isHealthy = false
	}

	// 3. Sprawdź status atomowy (czy nie wisi procesowanie)
	if w.flushInProgress.Load() {
		checks["flush_status"] = "BUSY"
	} else {
		checks["flush_status"] = "IDLE"
	}

	status := "UP"
	if !isHealthy {
		status = "DOWN"
	}

	return isHealthy, HealthStatus{
		Status: status,
		Checks: checks,
		Time:   time.Now(),
	}
}
