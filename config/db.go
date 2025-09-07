package config

import (
	"fmt"

	"github.com/zerodayz7/http-server/internal/shared/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// MustInitDB inicjalizuje bazę i panicuje przy błędzie, zwraca *gorm.DB i funkcję do defer
func MustInitDB() (*gorm.DB, func()) {
	log := logger.GetLogger()
	cfg := AppConfig.Database

	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		panic(fmt.Errorf("failed to connect to database: %w", err))
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic(fmt.Errorf("failed to get database instance: %w", err))
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := sqlDB.Ping(); err != nil {
		panic(fmt.Errorf("database ping failed: %w", err))
	}

	log.Info("Successfully connected to MySQL")
	return db, func() { sqlDB.Close() }
}
