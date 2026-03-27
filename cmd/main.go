package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/zerodayz7/http-server/config"
	"github.com/zerodayz7/http-server/internal/di"
	"github.com/zerodayz7/http-server/internal/redis"
	"github.com/zerodayz7/http-server/internal/router"
	"github.com/zerodayz7/http-server/internal/shared/logger"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func main() {
	// 1. Inicjalizacja loggera
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}
	log := logger.NewLogger(env)
	defer log.Sync()

	// 2. Ładowanie konfiguracji
	if err := config.LoadConfigGlobal(log); err != nil {
		log.Fatal("Config load failed", zap.Error(err))
	}
	cfg := &config.AppConfig

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 3. Inicjalizacja infrastruktury
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

	// 4. DI - Moduł interakcji
	module, err := di.InitializeInteractionModule(db, redisClient, cfg, log)
	if err != nil {
		log.Fatal("DI module initialization failed", zap.Error(err))
	}

	g, gCtx := errgroup.WithContext(ctx)

	// Dodajemy 'log'
	app := config.NewFiberApp(cfg, log)

	// Upewnij się, że SetupRoutes też ma 'log'
	router.SetupRoutes(app, module.Handler, cfg, log)
	// ----------------------

	// Serwer HTTP
	g.Go(func() error {
		log.Info("HTTP server starting", zap.String("port", cfg.Server.Port))
		return app.Listen(":" + cfg.Server.Port)
	})

	// Worker
	g.Go(func() error {
		log.Info("Interaction worker starting")
		module.Worker.Start(gCtx)
		return nil
	})

	// Graceful Shutdown
	g.Go(func() error {
		<-gCtx.Done()
		log.Info("Shutting down services...",
			zap.Duration("timeout", cfg.Shutdown),
		)
		return app.ShutdownWithTimeout(cfg.Shutdown)
	})

	if err := g.Wait(); err != nil {
		log.Error("Application exited with error", zap.Error(err))
		os.Exit(1)
	}

	log.Info("Application stopped cleanly")
}
