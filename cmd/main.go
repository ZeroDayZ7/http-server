package main

import (
	"os"

	"github.com/zerodayz7/http-server/config"
	"github.com/zerodayz7/http-server/internal/di"
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

	// Init DB (GORM)
	db, closeDB := config.MustInitDB()
	defer closeDB()

	// Init Redis
	redisClient, closeRedis := config.MustInitRedis()
	defer closeRedis()

	// Dependency Injection (Wire)
	interactionHandler, err := di.InitializeInteractionModule(db, redisClient)
	if err != nil {
		log.Fatal("DI init failed", zap.Error(err))
	}

	// Fiber app
	app := config.NewFiberApp()

	// Routes
	router.SetupRoutes(app, interactionHandler)

	// Graceful shutdown
	server.SetupGracefulShutdown(
		app,
		func() {
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
