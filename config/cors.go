package config

import (
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/zerodayz7/http-server/config/env"
)

func CorsConfig(cfg *env.Config) cors.Config {
	return cors.Config{
		AllowOrigins:     cfg.CORSAllow,
		AllowMethods:     cfg.CORSMethods,
		AllowHeaders:     cfg.CORSHeaders,
		AllowCredentials: cfg.CORSCredentials,
		ExposeHeaders:    "Content-Length",
		MaxAge:           86400,
	}
}
