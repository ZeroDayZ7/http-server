package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zerodayz7/http-server/config"
	"github.com/zerodayz7/http-server/internal/di"
	"github.com/zerodayz7/http-server/internal/redis"
	"github.com/zerodayz7/http-server/internal/shared/logger"
	"github.com/zerodayz7/http-server/internal/shared/telemetry"
	"go.uber.org/zap"
)

func main() {
	log := logger.NewLogger(os.Getenv("ENV"))
	defer log.Sync()

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
