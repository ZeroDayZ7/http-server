package worker

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sony/gobreaker"
)

// --- SENIORALNE METRYKI ---
var (
	// Statusy (Gauges)
	workerRedisStatus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "worker_redis_up",
		Help: "Status polaczenia z Redis (1 = UP, 0 = DOWN)",
	})
	workerDBBreakerStatus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "worker_db_breaker_closed",
		Help: "Status Circuit Breakera bazy danych (1 = CLOSED, 0 = OPEN/HALF-OPEN)",
	})

	// 1. NOWOŚĆ: Licznik błędów (Counter) - tylko rośnie
	workerHealthErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "worker_health_check_errors_total",
		Help: "Calkowita liczba bledow wykrytych przez health check",
	})

	// 2. NOWOŚĆ: Czas trwania (Histogram) - idealny do heatmap w Grafanie
	workerHealthDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "worker_health_check_duration_seconds",
		Help:    "Czas trwania health checka w sekundach",
		Buckets: prometheus.DefBuckets, // Standardowe przedziały czasowe
	})
)

func init() {
	prometheus.MustRegister(workerRedisStatus)
	prometheus.MustRegister(workerDBBreakerStatus)
	prometheus.MustRegister(workerHealthErrors)
	prometheus.MustRegister(workerHealthDuration)
}

type HealthStatus struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks"`
	Time   time.Time         `json:"time"`
}

func (w *InteractionWorker) HealthCheck(ctx context.Context) (bool, HealthStatus) {
	// START POMIARU CZASU
	timer := prometheus.NewTimer(workerHealthDuration)
	defer timer.ObserveDuration()

	isHealthy := true
	checks := make(map[string]string)

	// 1. Sprawdź Redis
	if err := w.redisClient.Ping(ctx).Err(); err != nil {
		checks["redis"] = "DOWN: " + err.Error()
		isHealthy = false
		workerRedisStatus.Set(0)
		workerHealthErrors.Inc() // Zwiększ licznik błędów
	} else {
		checks["redis"] = "UP"
		workerRedisStatus.Set(1)
	}

	// 2. Sprawdź Circuit Breaker
	state := w.cb.State()
	checks["database_breaker"] = state.String()

	if state == gobreaker.StateOpen {
		isHealthy = false
		workerDBBreakerStatus.Set(0)
		workerHealthErrors.Inc() // Zwiększ licznik błędów
	} else {
		workerDBBreakerStatus.Set(1)
	}

	// 3. Status atomowy (Flush)
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
