package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/zerodayz7/http-server/config"
	"github.com/zerodayz7/http-server/internal/handler"
	mysqlrepo "github.com/zerodayz7/http-server/internal/repository/mysql"
	"github.com/zerodayz7/http-server/internal/router"
	"github.com/zerodayz7/http-server/internal/service"
	"github.com/zerodayz7/http-server/internal/shared/logger"

	"github.com/gofiber/fiber/v2"
)

func createTestApp(t *testing.T) *fiber.App {
	_ = t
	_, _ = logger.InitLogger("development")
	log := logger.GetLogger()
	defer log.Sync()

	// DB + Session z config
	conn, closeDB := config.MustInitDB()
	defer closeDB()
	sessionStore := config.InitSessionStore(conn)

	userRepo := mysqlrepo.NewUserRepository(conn)
	interactionRepo := mysqlrepo.NewInteractionRepository(conn)
	authSvc := service.NewAuthService(userRepo)
	userSvc := service.NewUserService(userRepo)
	interactionSvc := service.NewInteractionService(interactionRepo)

	authHandler := handler.NewAuthHandler(authSvc)
	userHandler := handler.NewUserHandler(userSvc)
	interactionHandler := handler.NewInteractionHandler(interactionSvc)

	app := config.NewFiberApp(sessionStore)
	router.SetupRoutes(app, authHandler, userHandler, interactionHandler, sessionStore)
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
