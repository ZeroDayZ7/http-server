package main

import (
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/zerodayz7/http-server/internal/router"
)

type DummyHandler struct{}

func TestHealthRoute(t *testing.T) {

	app := fiber.New()
	router.SetupRoutes(app, nil)
	req, _ := http.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
