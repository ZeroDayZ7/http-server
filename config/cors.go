package config

import (
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func CorsConfig() cors.Config {
	return cors.Config{
		AllowOrigins:     AppConfig.CORSAllow,
		AllowMethods:     AppConfig.CORSMethods,
		AllowHeaders:     AppConfig.CORSHeaders,
		AllowCredentials: AppConfig.CORSCredentials,
		ExposeHeaders:    "Content-Length",
		MaxAge:           86400,
	}
}
