package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/zerodayz7/http-server/config"
	"github.com/zerodayz7/http-server/internal/di"
	"github.com/zerodayz7/http-server/internal/redis"
	"github.com/zerodayz7/http-server/internal/shared/logger"
	"github.com/zerodayz7/http-server/internal/shared/telemetry"
	"go.uber.org/zap"
)

func main() {
	envStr := os.Getenv("ENV")
	appEnv := logger.EnvDevelopment

	if envStr != "" {
		appEnv = logger.Env(envStr)
	}

	log := logger.NewLogger(appEnv)
	defer log.Sync()

	log.Info("Application started", zap.String("env", string(appEnv)))

	config.LoadConfigGlobal(log)
	cfg := &config.AppConfig

	if cfg.OTEL.Enabled {
		cleanup := telemetry.InitTelemetry(
			context.Background(),
			cfg.Server.AppName,
			cfg.OTEL.Endpoint,
			15*time.Second,
		)
		defer cleanup()
		log.Info("OTEL Telemetry initialized", zap.String("endpoint", cfg.OTEL.Endpoint))
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, _ := config.InitDB(ctx, cfg.Database, log)
	defer db.Close()

	redisClient, _ := config.InitRedis(ctx, cfg.Redis, log)
	defer redisClient.Close()

	redis.SetupStreamGroup(ctx, redisClient)

	module, _ := di.InitializeInteractionModule(db, redisClient, cfg, log)

	go func() {
		mux := http.NewServeMux()

		mux.Handle("/metrics", promhttp.Handler())

		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			hCtx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()

			isHealthy, report := module.Worker.HealthCheck(hCtx)

			w.Header().Set("Content-Type", "application/json")
			if !isHealthy {
				w.WriteHeader(http.StatusServiceUnavailable)
			}

			json.NewEncoder(w).Encode(report)
		})

		healthAddr := ":" + cfg.Server.HealthPort

		log.Info("Health & Metrics server listening", zap.String("addr", healthAddr))

		if err := http.ListenAndServe(healthAddr, mux); err != nil && err != http.ErrServerClosed {
			log.Error("Health check server failed", zap.Error(err))
		}
	}()

	log.Info("Background Worker starting...", zap.String("env", os.Getenv("ENV")))

	done := make(chan struct{})

	go func() {
		if err := module.Worker.Start(ctx); err != nil {
			log.Error("Worker execution stopped", zap.Error(err))
		}
		close(done)
	}()

	<-ctx.Done()
	log.Warn("Shutdown signal received. Cleaning up worker...")

	select {
	case <-done:
		log.Info("Worker finished cleanup successfully")
	case <-time.After(cfg.Shutdown):
		log.Error("Worker cleanup timed out! Forcing exit.")
	}

	log.Info("Worker process stopped cleanly")
}
