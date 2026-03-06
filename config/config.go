package config

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/spf13/viper"
	"github.com/zerodayz7/http-server/internal/shared/logger"
	"go.uber.org/zap"
)

type ServerConfig struct {
	AppName       string
	Port          string
	BodyLimitMB   int
	Env           string
	AppVersion    string
	ServerHeader  string
	Prefork       bool
	CaseSensitive bool
	StrictRouting bool
	IdleTimeout   time.Duration
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
}

type DBConfig struct {
	User            string
	Password        string
	Host            string
	Port            string
	DBName          string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type RateLimitConfig struct {
	Max    int
	Window time.Duration
}

type Config struct {
	Server          ServerConfig
	Database        DBConfig
	Redis           RedisConfig
	RateLimit       RateLimitConfig
	CORSAllow       string
	Shutdown        time.Duration
	SessionTTL      time.Duration
	FingerprintSalt string
}

var AppConfig Config
var Store *session.Store

func LoadConfigGlobal() error {
	log := logger.GetLogger()

	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	viper.SetDefault("APP_NAME", "http-server")
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("BODY_LIMIT_MB", 2)
	viper.SetDefault("APP_VERSION", "0.1.0")
	viper.SetDefault("ENV", "development")
	viper.SetDefault("SERVER_HEADER", "ZeroDayZ7")
	viper.SetDefault("PREFORK", false)
	viper.SetDefault("CASE_SENSITIVE", true)
	viper.SetDefault("STRICT_ROUTING", true)
	viper.SetDefault("IDLE_TIMEOUT_SEC", 30)
	viper.SetDefault("READ_TIMEOUT_SEC", 15)
	viper.SetDefault("WRITE_TIMEOUT_SEC", 15)
	viper.SetDefault("DB_MAX_OPEN_CONNS", 50)
	viper.SetDefault("DB_MAX_IDLE_CONNS", 10)
	viper.SetDefault("DB_CONN_MAX_LIFETIME_MIN", 30)
	viper.SetDefault("RATE_LIMIT_MAX", 100)
	viper.SetDefault("RATE_LIMIT_WINDOW_SEC", 60)
	viper.SetDefault("CORS_ALLOW_ORIGINS", "*")
	viper.SetDefault("SHUTDOWN_TIMEOUT_SEC", 5)
	viper.SetDefault("SESSION_TTL_MINUTES", 1440)
	viper.SetDefault("FINGERPRINT_SALT", "default-secret-salt")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Error("Error loading .env", zap.Error(err))
			return fmt.Errorf("error loading .env: %v", err)
		}
	}

	AppConfig = Config{
		Server: ServerConfig{
			AppName:       viper.GetString("APP_NAME"),
			Port:          viper.GetString("PORT"),
			BodyLimitMB:   viper.GetInt("BODY_LIMIT_MB"),
			AppVersion:    viper.GetString("APP_VERSION"),
			Env:           viper.GetString("ENV"),
			ServerHeader:  viper.GetString("SERVER_HEADER"),
			Prefork:       viper.GetBool("PREFORK"),
			CaseSensitive: viper.GetBool("CASE_SENSITIVE"),
			StrictRouting: viper.GetBool("STRICT_ROUTING"),
			IdleTimeout:   time.Duration(viper.GetInt("IDLE_TIMEOUT_SEC")) * time.Second,
			ReadTimeout:   time.Duration(viper.GetInt("READ_TIMEOUT_SEC")) * time.Second,
			WriteTimeout:  time.Duration(viper.GetInt("WRITE_TIMEOUT_SEC")) * time.Second,
		},
		Database: DBConfig{
			User:     viper.GetString("MYSQL_USER"),
			Password: viper.GetString("MYSQL_PASSWORD"),
			Host:     viper.GetString("MYSQL_HOST"),
			Port:     viper.GetString("MYSQL_PORT"),
			DBName:   viper.GetString("MYSQL_DATABASE"),

			MaxOpenConns:    viper.GetInt("DB_MAX_OPEN_CONNS"),
			MaxIdleConns:    viper.GetInt("DB_MAX_IDLE_CONNS"),
			ConnMaxLifetime: time.Duration(viper.GetInt("DB_CONN_MAX_LIFETIME_MIN")) * time.Minute,
		},
		Redis: RedisConfig{
			Host:     viper.GetString("REDIS_HOST"),
			Port:     viper.GetString("REDIS_PORT"),
			Password: viper.GetString("REDIS_PASSWORD"),
			DB:       viper.GetInt("REDIS_DB"),
		},
		RateLimit: RateLimitConfig{
			Max:    viper.GetInt("RATE_LIMIT_MAX"),
			Window: time.Duration(viper.GetInt("RATE_LIMIT_WINDOW_SEC")) * time.Second,
		},
		CORSAllow:       viper.GetString("CORS_ALLOW_ORIGINS"),
		Shutdown:        time.Duration(viper.GetInt("SHUTDOWN_TIMEOUT_SEC")) * time.Second,
		SessionTTL:      time.Duration(viper.GetInt("SESSION_TTL_MINUTES")) * time.Minute,
		FingerprintSalt: viper.GetString("FINGERPRINT_SALT"),
	}

	log.Info("Configuration loaded")
	return nil
}
