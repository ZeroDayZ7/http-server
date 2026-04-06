package config

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/zerodayz7/http-server/config/env"
	"github.com/zerodayz7/http-server/internal/shared/logger"
	"go.uber.org/zap"
)

func InitDB(ctx context.Context, cfg env.DBConfig, log logger.Logger, runMigrations bool) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("db open failed: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		return nil, fmt.Errorf("db ping failed: %w", err)
	}

	if runMigrations {
		if err := RunMigrations(db, log); err != nil {
			return nil, fmt.Errorf("migrations failed: %w", err)
		}
		log.Info("MySQL connected and migrations applied", zap.String("database", cfg.DBName))
	} else {
		log.Info("MySQL connected (migrations skipped)", zap.String("database", cfg.DBName))
	}

	log.Info("MySQL connected and migrations applied", zap.String("database", cfg.DBName))
	return db, nil
}
