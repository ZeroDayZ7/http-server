package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/zerodayz7/http-server/internal/shared/logger"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type Config struct {
	AppVersion         string
	MySQLDSN           string
	Port               string
	Env                string
	CORSAllowOrigins   string
	RateLimitMax       int
	RateLimitWindow    time.Duration
	BodyLimitMB        int
	ShutdownTimeoutSec int
	DBMaxOpenConns     int
	DBMaxIdleConns     int
	DBConnMaxLifetime  time.Duration
}

func LoadConfig(log *logger.Logger) (Config, error) {
	log.Info("Loading .env file")
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) {
		log.Error("Error loading .env", zap.Error(err))
		return Config{}, fmt.Errorf("error loading .env: %v", err)
	}

	rateMax := 100
	rateWindow := time.Minute
	bodyLimit := 2
	shutdownTimeout := 5
	dbMaxOpen := 50
	dbMaxIdle := 10
	dbMaxLifetime := 30 * time.Minute

	if val := os.Getenv("BODY_LIMIT_MB"); val != "" {
		if l, err := strconv.Atoi(val); err == nil {
			bodyLimit = l
		} else {
			log.Warn("Invalid BODY_LIMIT_MB, using default 2MB")
		}
	}

	if val := os.Getenv("SHUTDOWN_TIMEOUT"); val != "" {
		if t, err := strconv.Atoi(val); err == nil {
			shutdownTimeout = t
		} else {
			log.Warn("Invalid SHUTDOWN_TIMEOUT, using default 5s")
		}
	}

	if val := os.Getenv("DB_MAX_OPEN_CONNS"); val != "" {
		if v, err := strconv.Atoi(val); err == nil {
			dbMaxOpen = v
		} else {
			log.Warn("Invalid DB_MAX_OPEN_CONNS, using default 50")
		}
	}

	if val := os.Getenv("DB_MAX_IDLE_CONNS"); val != "" {
		if v, err := strconv.Atoi(val); err == nil {
			dbMaxIdle = v
		} else {
			log.Warn("Invalid DB_MAX_IDLE_CONNS, using default 10")
		}
	}

	if val := os.Getenv("DB_CONN_MAX_LIFETIME_MIN"); val != "" {
		if v, err := strconv.Atoi(val); err == nil {
			dbMaxLifetime = time.Duration(v) * time.Minute
		} else {
			log.Warn("Invalid DB_CONN_MAX_LIFETIME_MIN, using default 30min")
		}
	}

	config := Config{
		AppVersion:         os.Getenv("APP_VERSION"),
		MySQLDSN:           os.Getenv("MYSQL_DSN"),
		Port:               os.Getenv("PORT"),
		Env:                os.Getenv("ENV"),
		CORSAllowOrigins:   os.Getenv("CORS_ALLOW_ORIGINS"),
		RateLimitMax:       rateMax,
		RateLimitWindow:    rateWindow,
		BodyLimitMB:        bodyLimit,
		ShutdownTimeoutSec: shutdownTimeout,
		DBMaxOpenConns:     dbMaxOpen,
		DBMaxIdleConns:     dbMaxIdle,
		DBConnMaxLifetime:  dbMaxLifetime,
	}

	if config.AppVersion == "" {
		config.AppVersion = "0.1.0"
	}
	if config.MySQLDSN == "" {
		log.Error("MYSQL_DSN environment variable is required")
		return Config{}, fmt.Errorf("MYSQL_DSN environment variable is required")
	}
	if config.Port == "" {
		log.Error("PORT environment variable is required")
		return Config{}, fmt.Errorf("PORT environment variable is required")
	}
	if config.Env == "" {
		log.Warn("ENV not set, defaulting to development")
		config.Env = "development"
	}
	if config.CORSAllowOrigins == "" {
		config.CORSAllowOrigins = "*"
	}

	log.Info("Configuration loaded",
		zap.String("env", config.Env),
		zap.String("port", config.Port),
		zap.Int("bodyLimitMB", config.BodyLimitMB),
		zap.Int("shutdownTimeoutSec", config.ShutdownTimeoutSec),
		zap.Int("dbMaxOpenConns", config.DBMaxOpenConns),
		zap.Int("dbMaxIdleConns", config.DBMaxIdleConns),
		zap.Duration("dbConnMaxLifetime", config.DBConnMaxLifetime),
	)

	return config, nil
}
