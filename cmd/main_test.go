package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/zerodayz7/http-server/config"
	"github.com/zerodayz7/http-server/internal/router"
	"github.com/zerodayz7/http-server/internal/shared/logger" // Import loggera
)

func TestHealthRoute(t *testing.T) {
	// 1. Tworzymy atrapę konfiguracji
	cfg := &config.Config{
		RateLimit: config.RateLimitConfig{ // Użyj właściwego typu z pakietu config
			Max:    100,
			Window: 60 * time.Second,
		},
	}

	// 2. Tworzymy logger testowy (No-op), żeby nie spamował konsoli podczas testów
	// Jeśli Twój pakiet logger ma funkcję NewTestLogger lub podobną, użyj jej.
	// W najprostszym wydaniu możemy zainicjalizować go w trybie development:
	testLog := logger.NewLogger("development")

	app := fiber.New()

	// 3. Przekazujemy wszystkie 4 wymagane argumenty:
	// app, handler (nil bo health go nie potrzebuje), cfg, log
	router.SetupRoutes(app, nil, cfg, testLog)

	req, _ := http.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req, -1)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
