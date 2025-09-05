package main

import (
	"os"

	"github.com/zerodayz7/http-server/config"
	"github.com/zerodayz7/http-server/internal/handler"
	"github.com/zerodayz7/http-server/internal/middleware"
	mysqlrepo "github.com/zerodayz7/http-server/internal/repository/mysql"
	"github.com/zerodayz7/http-server/internal/router"
	"github.com/zerodayz7/http-server/internal/server"
	"github.com/zerodayz7/http-server/internal/service"
	"github.com/zerodayz7/http-server/internal/shared/db"
	"github.com/zerodayz7/http-server/internal/shared/logger"
	"go.uber.org/zap"

	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// Inicjalizacja loggera
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	_, err := logger.InitLogger(env)
	if err != nil {
		panic("logger init failed: " + err.Error())
	}
	defer logger.GetLogger().Sync()
	log := logger.GetLogger()

	// Load config
	cfg, err := config.LoadConfig(log)
	if err != nil {
		log.Fatal("Config load failed", zap.Error(err))
		return
	}

	log.Info("Starting server", zap.String("version", cfg.Server.AppVersion))

	// DB setup
	dbCfg := db.DBConfig{
		DSN:             cfg.Database.DSN,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	}

	conn, err := db.NewDB(dbCfg)
	if err != nil {
		log.Fatal("DB connection failed", zap.Error(err))
	}

	sqlDB, _ := conn.DB()
	defer sqlDB.Close()

	// Repo, service, handler
	repo := mysqlrepo.NewUserRepository(conn)
	svc := service.NewUserService(repo)
	h := handler.NewUserHandler(svc)

	// Fiber app
	app := server.New(cfg)

	// Middleware
	app.Use(middleware.RequestIDMiddleware())
	app.Use(fiberlogger.New())
	app.Use(recover.New())
	app.Use(helmet.New(config.HelmetConfig()))
	app.Use(cors.New(config.CorsConfig(cfg.CORSAllow)))
	app.Use(limiter.New(config.LimiterConfig(cfg.RateLimit.Max, cfg.RateLimit.Window)))

	// Graceful shutdown
	server.SetupGracefulShutdown(app, sqlDB, cfg.Shutdown)

	// Routes
	router.SetupRoutes(app, h)

	log.Info("Listening", zap.String("port", cfg.Server.Port))
	if err := app.Listen(":" + cfg.Server.Port); err != nil {
		log.Fatal("Fiber listen failed", zap.Error(err))
	}
}
