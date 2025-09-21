package config

import (
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func CorsConfig() cors.Config {
	allowOrigins := AppConfig.CORSAllow
	return cors.Config{
		AllowOrigins: allowOrigins,
		AllowMethods: "GET,POST,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept",
	}
}
