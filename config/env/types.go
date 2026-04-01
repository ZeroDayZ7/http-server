package env

import "time"

type ServerConfig struct {
	AppName       string        `mapstructure:"APP_NAME" validate:"required"`
	Port          string        `mapstructure:"PORT" validate:"required,numeric"`
	HealthPort    string        `mapstructure:"HEALTH_PORT" validate:"required,numeric"`
	BodyLimitMB   int           `mapstructure:"BODY_LIMIT_MB"`
	Env           string        `mapstructure:"ENV" validate:"required,oneof=development staging production"`
	AppVersion    string        `mapstructure:"APP_VERSION"`
	ServerHeader  string        `mapstructure:"SERVER_HEADER"`
	Prefork       bool          `mapstructure:"PREFORK"`
	CaseSensitive bool          `mapstructure:"CASE_SENSITIVE"`
	StrictRouting bool          `mapstructure:"STRICT_ROUTING"`
	IdleTimeout   time.Duration `mapstructure:"IDLE_TIMEOUT"`
	ReadTimeout   time.Duration `mapstructure:"READ_TIMEOUT"`
	WriteTimeout  time.Duration `mapstructure:"WRITE_TIMEOUT"`
}

type DBConfig struct {
	User            string        `mapstructure:"MYSQL_USER" validate:"required"`
	Password        string        `mapstructure:"MYSQL_PASSWORD" validate:"required"`
	Host            string        `mapstructure:"MYSQL_HOST" validate:"required"`
	Port            string        `mapstructure:"MYSQL_PORT" validate:"required,numeric"`
	DBName          string        `mapstructure:"MYSQL_DATABASE" validate:"required"`
	MaxOpenConns    int           `mapstructure:"DB_MAX_OPEN_CONNS"`
	MaxIdleConns    int           `mapstructure:"DB_MAX_IDLE_CONNS"`
	ConnMaxLifetime time.Duration `mapstructure:"DB_CONN_MAX_LIFETIME"`
}

type RedisConfig struct {
	Host     string `mapstructure:"REDIS_HOST" validate:"required"`
	Port     string `mapstructure:"REDIS_PORT" validate:"required,numeric"`
	Password string `mapstructure:"REDIS_PASSWORD"`
	DB       int    `mapstructure:"REDIS_DB"`
}

type RateLimitConfig struct {
	Max    int           `mapstructure:"RATE_LIMIT_MAX"`
	Window time.Duration `mapstructure:"RATE_LIMIT_WINDOW"`
}

type OTELConfig struct {
	Enabled  bool   `mapstructure:"OTEL_ENABLED"`
	Endpoint string `mapstructure:"OTEL_EXPORTER_OTLP_ENDPOINT" validate:"required_if=Enabled true"`
}

type Config struct {
	Server              ServerConfig    `mapstructure:",squash"`
	Database            DBConfig        `mapstructure:",squash"`
	Redis               RedisConfig     `mapstructure:",squash"`
	RateLimit           RateLimitConfig `mapstructure:",squash"`
	OTEL                OTELConfig      `mapstructure:",squash"`
	CORSAllow           string          `mapstructure:"CORS_ALLOW_ORIGINS" validate:"required"`
	CORSMethods         string          `mapstructure:"CORS_ALLOW_METHODS" validate:"required"`
	CORSHeaders         string          `mapstructure:"CORS_ALLOW_HEADERS" validate:"required"`
	CORSCredentials     bool            `mapstructure:"CORS_ALLOW_CREDENTIALS"`
	Shutdown            time.Duration   `mapstructure:"SHUTDOWN_TIMEOUT"`
	SessionTTL          time.Duration   `mapstructure:"SESSION_TTL"`
	FingerprintSalt     string          `mapstructure:"FINGERPRINT_SALT" validate:"required,min=16"`
	WorkerFlushInterval time.Duration   `mapstructure:"WORKER_FLUSH_INTERVAL"`
}
