package config

import (
	"fmt"
	"time"

	fiberMysql "github.com/gofiber/storage/mysql/v2"
	"github.com/zerodayz7/http-server/internal/shared/logger"
	gormMysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var globalStorage *fiberMysql.Storage

// MustInitDB inicjalizuje bazę i panicuje przy błędzie, zwraca *gorm.DB i funkcję do defer
func MustInitDB() (*gorm.DB, func()) {
	log := logger.GetLogger()
	cfg := AppConfig.Database

	db, err := gorm.Open(gormMysql.Open(cfg.DSN), &gorm.Config{})
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

	globalStorage = fiberMysql.New(fiberMysql.Config{
		Db:         sqlDB,
		Table:      "fiber_storage",
		Reset:      false,
		GCInterval: 30 * time.Second,
	})

	log.Info("Successfully connected to MySQL")
	return db, func() { sqlDB.Close() }
}

func Storage() *fiberMysql.Storage {
	if globalStorage == nil {
		panic("storage not initialized – call MustInitDB first")
	}
	return globalStorage
}
