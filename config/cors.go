package config

import (
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func CorsConfig(allowOrigins string) cors.Config {
	return cors.Config{
		AllowOrigins: allowOrigins,
		AllowMethods: "GET,POST,PUT,DELETE",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}
}
