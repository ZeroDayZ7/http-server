package db

import (
	"fmt"
	"time"

	"github.com/zerodayz7/http-server/internal/shared/logger"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DBConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func NewDB(cfg DBConfig) (*gorm.DB, error) {
	log := logger.GetLogger()
	log.Info("Connecting to MySQL database...")

	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		log.Error("Failed to connect to MySQL", zap.Error(err))
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Error("Failed to get database instance", zap.Error(err))
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := sqlDB.Ping(); err != nil {
		log.Error("Database ping failed", zap.Error(err))
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	log.Info("Successfully connected to MySQL")
	return db, nil
}
