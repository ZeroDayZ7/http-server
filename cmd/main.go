package main

import (
	"os"

	"github.com/zerodayz7/http-server/config"
	"github.com/zerodayz7/http-server/internal/handler"
	mysqlrepo "github.com/zerodayz7/http-server/internal/repository/mysql"
	"github.com/zerodayz7/http-server/internal/router"
	"github.com/zerodayz7/http-server/internal/server"
	"github.com/zerodayz7/http-server/internal/service"
	"github.com/zerodayz7/http-server/internal/shared/logger"
	"go.uber.org/zap"
)

func main() {
	_, _ = logger.InitLogger(os.Getenv("ENV"))

	// config
	if err := config.LoadConfigGlobal(); err != nil {
		logger.GetLogger().Fatal("Config load failed", zap.Error(err))
	}

	// DB
	conn, closeDB := config.MustInitDB()
	defer closeDB()

	// Repos, service, handlers
	userRepo := mysqlrepo.NewUserRepository(conn)
	interactionRepo := mysqlrepo.NewInteractionRepository(conn)
	authSvc := service.NewAuthService(userRepo)
	userSvc := service.NewUserService(userRepo)
	interactionSvc := service.NewInteractionService(interactionRepo)

	authHandler := handler.NewAuthHandler(authSvc)
	userHandler := handler.NewUserHandler(userSvc)
	interactionHandler := handler.NewInteractionHandler(interactionSvc)

	// session
	sessionStore := config.InitSessionStore(conn)

	// Fiber
	app := config.NewFiberApp(sessionStore)

	// routes
	router.SetupRoutes(app, authHandler, userHandler, interactionHandler, sessionStore)

	// graceful shutdown
	server.SetupGracefulShutdown(app, closeDB, config.AppConfig.Shutdown)

	logger.GetLogger().Info("Listening", zap.String("port", config.AppConfig.Server.Port))
	app.Listen(":" + config.AppConfig.Server.Port)
}
