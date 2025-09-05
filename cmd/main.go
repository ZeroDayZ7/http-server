package main

import (
	"os"
	"time"

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

	cfg, err := config.LoadConfig(log)
	if err != nil {
		log.Error("Config load failed", zap.Error(err))
		return
	}
	log.Info("Starting server", zap.String("version", cfg.AppVersion))

	dbCfg := db.DBConfig{
		DSN:             cfg.MySQLDSN,
		MaxOpenConns:    cfg.DBMaxOpenConns,
		MaxIdleConns:    cfg.DBMaxIdleConns,
		ConnMaxLifetime: cfg.DBConnMaxLifetime,
	}

	conn, err := db.NewDB(dbCfg)
	if err != nil {
		log.Fatal("DB connection failed", zap.Error(err))
	}

	sqlDB, _ := conn.DB()
	defer sqlDB.Close()

	repo := mysqlrepo.NewUserRepository(conn)
	svc := service.NewUserService(repo)
	h := handler.NewUserHandler(svc)

	app := server.New(cfg)

	app.Use(middleware.RequestIDMiddleware())
	app.Use(fiberlogger.New())
	app.Use(recover.New())
	app.Use(config.BodyLimitMiddleware())
	app.Use(helmet.New(config.HelmetConfig()))
	app.Use(cors.New(config.CorsConfig(cfg.CORSAllowOrigins)))
	app.Use(limiter.New(config.LimiterConfig(cfg.RateLimitMax, cfg.RateLimitWindow)))

	server.SetupGracefulShutdown(app, sqlDB, time.Duration(cfg.ShutdownTimeoutSec)*time.Second)
	router.SetupRoutes(app, h)

	log.Info("Listening", zap.String("port", cfg.Port))
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Error("Fiber listen failed", zap.Error(err))
	}
}
