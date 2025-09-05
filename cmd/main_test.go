package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/zerodayz7/http-server/internal/handler"
	"github.com/zerodayz7/http-server/internal/model"
	mysqlrepo "github.com/zerodayz7/http-server/internal/repository/mysql"
	"github.com/zerodayz7/http-server/internal/router"
	"github.com/zerodayz7/http-server/internal/service"
	"github.com/zerodayz7/http-server/internal/shared/db"
	"github.com/zerodayz7/http-server/internal/shared/logger"
	"go.uber.org/zap"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func createTestApp(t *testing.T) *fiber.App {
	// Init logger once
	_, err := logger.InitLogger("development")
	if err != nil {
		t.Fatalf("Logger init failed: %v", err)
	}
	log := logger.GetLogger()
	defer log.Sync()

	// DB config
	cfg := db.DBConfig{
		DSN:             "root:admin@tcp(localhost:3306)/portfolio_db?charset=utf8mb4&parseTime=True&loc=Local",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}

	conn, err := db.NewDB(cfg)
	if err != nil {
		t.Fatalf("DB connection failed: %v", err)
	}

	sqlDB, err := conn.DB()
	if err != nil {
		t.Fatalf("Failed to get sql.DB: %v", err)
	}
	defer sqlDB.Close()

	if err := conn.AutoMigrate(&model.User{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	// repo, service, handler
	repo := mysqlrepo.NewUserRepository(conn)
	svc := service.NewUserService(repo)
	h := handler.NewUserHandler(svc)

	// Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if e, ok := err.(*fiber.Error); ok {
				log.Error("HTTP error", zap.Error(err), zap.String("path", c.Path()))
				return c.Status(e.Code).JSON(fiber.Map{"error": e.Message})
			}
			log.Error("Server error", zap.Error(err), zap.String("path", c.Path()))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
		},
	})

	// Middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	app.Use(limiter.New(limiter.Config{
		Max:        1000,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	}))

	// Routes
	router.SetupRoutes(app, h)

	return app
}

func TestHealthcheck(t *testing.T) {
	app := createTestApp(t)

	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req, int((5*time.Second)/time.Millisecond))
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("JSON decode failed: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("Expected status 'ok', got %v", body["status"])
	}
}

func TestCORS(t *testing.T) {
	app := createTestApp(t)

	req := httptest.NewRequest("GET", "/health", nil)
	req.Header.Set("Origin", "http://example.com")

	resp, err := app.Test(req, int((5*time.Second)/time.Millisecond))
	if err != nil {
		t.Fatalf("CORS request failed: %v", err)
	}

	if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Fatalf("Expected CORS header to allow origin, got: %v", resp.Header.Get("Access-Control-Allow-Origin"))
	}
}

func TestRateLimiter(t *testing.T) {
	app := createTestApp(t)

	for range make([]struct{}, 10) {
		req := httptest.NewRequest("GET", "/health", nil)
		resp, _ := app.Test(req, int((5*time.Second)/time.Millisecond))
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200 from health endpoint, got %d", resp.StatusCode)
		}
	}
}
