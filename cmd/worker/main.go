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
	"golang.org/x/sync/errgroup"
)

func main() {
	// ENV / logger
	envStr := os.Getenv("ENV")
	appEnv := logger.EnvDevelopment

	if envStr != "" {
		appEnv = logger.Env(envStr)
	}

	log := logger.NewLogger(appEnv)
	defer log.Sync()

	log.Info("Application started", zap.String("env", string(appEnv)))

	// config
	config.LoadConfigGlobal(log)
	cfg := &config.AppConfig

	// telemetry
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

	// root context + signals
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// errgroup lifecycle
	g, gCtx := errgroup.WithContext(ctx)

	// infra init
	db, err := config.InitDB(ctx, cfg.Database, log)
	if err != nil {
		log.Fatal("Database initialization failed", zap.Error(err))
	}
	defer db.Close()

	redisClient, err := config.InitRedis(ctx, cfg.Redis, log)
	if err != nil {
		log.Fatal("Redis initialization failed", zap.Error(err))
	}
	defer redisClient.Close()

	if err := redis.SetupStreamGroup(ctx, redisClient); err != nil {
		log.Fatal("Redis stream setup failed", zap.Error(err))
	}

	module, err := di.InitializeInteractionModule(db, redisClient, cfg, log)
	if err != nil {
		log.Fatal("DI module initialization failed", zap.Error(err))
	}

	// -------------------------
	// HTTP Health Server
	// -------------------------
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

		_ = json.NewEncoder(w).Encode(report)
	})

	healthAddr := ":" + cfg.Server.WorkerPort

	server := &http.Server{
		Addr:    healthAddr,
		Handler: mux,
	}

	// start HTTP server
	g.Go(func() error {
		log.Info("Health server starting", zap.String("addr", healthAddr))
		return server.ListenAndServe()
	})

	// graceful shutdown HTTP server
	g.Go(func() error {
		<-gCtx.Done()

		log.Info("Shutting down health server...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Shutdown)
		defer cancel()

		return server.Shutdown(shutdownCtx)
	})

	// -------------------------
	// Worker lifecycle
	// -------------------------
	g.Go(func() error {
		log.Info("Worker starting...")
		err := module.Worker.Start(gCtx)
		if err != nil {
			log.Error("Worker execution stopped", zap.Error(err))
		}
		return err
	})

	// -------------------------
	// Wait for everything
	// -------------------------
	if err := g.Wait(); err != nil {
		log.Error("Application exited with error", zap.Error(err))
		os.Exit(1)
	}

	log.Info("Application stopped cleanly")
}
