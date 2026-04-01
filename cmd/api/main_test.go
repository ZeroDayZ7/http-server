package main

import (
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/zerodayz7/http-server/config/env"
	"github.com/zerodayz7/http-server/internal/router"
	"github.com/zerodayz7/http-server/internal/shared/logger"
)

func loadTestConfig() (*env.Config, error) {
	_ = os.Setenv("REDIS_HOST", "memory")
	_ = os.Setenv("RATE_LIMIT_MAX", "5")
	_ = os.Setenv("RATE_LIMIT_WINDOW", "1m")

	viper.Reset()
	viper.AutomaticEnv()

	var cfg env.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func TestIntegration_AppFlow(t *testing.T) {
	testLog := logger.NewLogger(logger.EnvTest)

	cfg, err := loadTestConfig()
	assert.NoError(t, err)

	app := fiber.New()
	router.SetupRoutes(app, nil, cfg, testLog)

	t.Run("Status_200_On_Health_Endpoint", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/health", nil)
		resp, err := app.Test(req, -1)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "ok")
	})

	t.Run("Not_Found_Handler", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/undefined-route-123", nil)
		resp, _ := app.Test(req, -1)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
