package main

import (
	"context"
	"os"

	"github.com/zerodayz7/http-server/config"
	"github.com/zerodayz7/http-server/internal/di"
	"github.com/zerodayz7/http-server/internal/redis"
	"github.com/zerodayz7/http-server/internal/router"
	"github.com/zerodayz7/http-server/internal/server"
	"github.com/zerodayz7/http-server/internal/shared/logger"
	"go.uber.org/zap"
)

func main() {
	log, _ := logger.InitLogger(os.Getenv("ENV"))

	// Load config
	if err := config.LoadConfigGlobal(); err != nil {
		log.Fatal("Config load failed", zap.Error(err))
	}

	// Init DB (SQL)
	db, closeDB := config.MustInitDB()
	defer closeDB()

	// Init Redis
	redisClient, closeRedis := config.MustInitRedis()
	defer closeRedis()

	// Redis Stream Group Setup
	log.Info("Setting up Redis Stream group")

	if err := redis.SetupStreamGroup(
		context.Background(),
		redisClient,
	); err != nil {
		log.Fatal("Redis stream setup failed", zap.Error(err))
	}

	log.Info("Redis Stream group ready")

	// Dependency Injection
	module, err := di.InitializeInteractionModule(db, redisClient, config.AppConfig.FingerprintSalt)
	if err != nil {
		log.Fatal("DI init failed", zap.Error(err))
	}

	log.Info("DI module created")

	if module == nil {
		log.Fatal("module is nil")
	}

	if module.Worker == nil {
		log.Fatal("worker is nil")
	}

	if module.Handler == nil {
		log.Fatal("handler is nil")
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Uruchomienie Workera z modułu
	go module.Worker.Start(ctx)

	// Fiber app
	app := config.NewFiberApp()

	// Podpięcie Route'ów przy użyciu Handlera z modułu
	router.SetupRoutes(app, module.Handler)

	// Graceful shutdown
	server.SetupGracefulShutdown(
		app,
		func() {
			log.Info("Shutting down services...")
			cancel()
			closeDB()
			closeRedis()
		},
		config.AppConfig.Shutdown,
	)

	log.Info(
		"Listening",
		zap.String("port", config.AppConfig.Server.Port),
	)

	if err := app.Listen(":" + config.AppConfig.Server.Port); err != nil {
		log.Fatal("Server failed", zap.Error(err))
	}
}
