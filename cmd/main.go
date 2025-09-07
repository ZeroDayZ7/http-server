package main

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/zerodayz7/http-server/config"
	"github.com/zerodayz7/http-server/internal/handler"
	mysqlrepo "github.com/zerodayz7/http-server/internal/repository/mysql"
	"github.com/zerodayz7/http-server/internal/router"
	"github.com/zerodayz7/http-server/internal/server"
	"github.com/zerodayz7/http-server/internal/service"
	"github.com/zerodayz7/http-server/internal/shared/db"
	"github.com/zerodayz7/http-server/internal/shared/logger"
	"go.uber.org/zap"
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

	err = config.LoadConfigGlobal()
	if err != nil {
		log.Fatal("Config load failed", zap.Error(err))
		return
	}

	log.Info("Starting server", zap.String("version", config.AppConfig.Server.AppVersion))

	dbCfg := db.DBConfig{
		DSN:             config.AppConfig.Database.DSN,
		MaxOpenConns:    config.AppConfig.Database.MaxOpenConns,
		MaxIdleConns:    config.AppConfig.Database.MaxIdleConns,
		ConnMaxLifetime: config.AppConfig.Database.ConnMaxLifetime,
	}

	conn, err := db.NewDB(dbCfg)
	if err != nil {
		log.Fatal("DB connection failed", zap.Error(err))
	}

	sqlDB, _ := conn.DB()
	defer sqlDB.Close()

	userRepo := mysqlrepo.NewUserRepository(conn)
	authSvc := service.NewAuthService(userRepo)
	userSvc := service.NewUserService(userRepo)

	authHandler := handler.NewAuthHandler(authSvc)
	userHandler := handler.NewUserHandler(userSvc)

	sessionStore := config.InitSessionStore(config.AppConfig.Database.DSN, config.AppConfig.SessionTTL)

	app := server.New(config.AppConfig)
	app.Use(requestid.New())
	app.Use(recover.New())
	app.Use(config.FiberLoggerMiddleware())
	app.Use(helmet.New(config.HelmetConfig()))
	app.Use(cors.New(config.CorsConfig(config.AppConfig.CORSAllow)))
	app.Use(config.NewLimiter("global"))
	app.Use(compress.New(config.CompressConfig()))

	app.Use(func(c *fiber.Ctx) error {
		sess, err := sessionStore.Get(c)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to get session"})
		}
		c.Locals("session", sess)
		return c.Next()
	})

	csrfConfig := config.NewCSRFConfig(sessionStore.Storage, config.AppConfig.SessionTTL)
	app.Use(csrf.New(csrfConfig))

	server.SetupGracefulShutdown(app, sqlDB, config.AppConfig.Shutdown)
	router.SetupRoutes(app, authHandler, userHandler, sessionStore, config.AppConfig.SessionTTL)

	log.Info("Listening", zap.String("port", config.AppConfig.Server.Port))
	if err := app.Listen(":" + config.AppConfig.Server.Port); err != nil {
		log.Fatal("Fiber listen failed", zap.Error(err))
	}
}
