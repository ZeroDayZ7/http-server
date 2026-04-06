package env

import (
	"time"

	"github.com/spf13/viper"
)

func setDefaults() {
	// Server
	viper.SetDefault("APP_NAME", "http-server")
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("WORKER_PORT", "8081")
	viper.SetDefault("BODY_LIMIT_MB", 2)
	viper.SetDefault("APP_VERSION", "0.1.0")
	viper.SetDefault("ENV", "development")
	viper.SetDefault("SERVER_HEADER", "GoServer")
	viper.SetDefault("PREFORK", false)
	viper.SetDefault("CASE_SENSITIVE", true)
	viper.SetDefault("STRICT_ROUTING", true)
	viper.SetDefault("IDLE_TIMEOUT", 30*time.Second)
	viper.SetDefault("READ_TIMEOUT", 15*time.Second)
	viper.SetDefault("WRITE_TIMEOUT", 15*time.Second)

	// Database
	viper.SetDefault("DB_USER", "root")
	viper.SetDefault("DB_PASSWORD", "password")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "3306")
	viper.SetDefault("DB_NAME", "appdb")
	viper.SetDefault("DB_MAX_OPEN_CONNS", 50)
	viper.SetDefault("DB_MAX_IDLE_CONNS", 10)
	viper.SetDefault("DB_CONN_MAX_LIFETIME", 30*time.Minute)

	// Redis
	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", 6379)
	viper.SetDefault("REDIS_PASSWORD", "")
	viper.SetDefault("REDIS_DB", 0)

	// Rate limit
	viper.SetDefault("RATE_LIMIT_MAX", 100)
	viper.SetDefault("RATE_LIMIT_WINDOW", 60*time.Second)

	// CORS
	viper.SetDefault("CORS_ALLOW_ORIGINS", "*")
	viper.SetDefault("CORS_ALLOW_METHODS", "GET,POST,OPTIONS,HEAD")
	viper.SetDefault("CORS_ALLOW_HEADERS", "Origin, Content-Type, Accept, Authorization")
	viper.SetDefault("CORS_ALLOW_CREDENTIALS", false)

	// App settings
	viper.SetDefault("SHUTDOWN_TIMEOUT", 5*time.Second)
	viper.SetDefault("SESSION_TTL", 24*time.Hour)
	viper.SetDefault("FINGERPRINT_SALT", "default-secret-salt-1234")
	viper.SetDefault("WORKER_FLUSH_INTERVAL", 10*time.Second)

	// OTEL
	viper.SetDefault("OTEL_ENABLED", false)
	viper.SetDefault("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317")
}
