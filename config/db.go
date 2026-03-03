package config

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/zerodayz7/http-server/internal/shared/logger"
)

func MustInitDB() (*sql.DB, func()) {
	log := logger.GetLogger()
	cfg := AppConfig.Database

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := db.Ping(); err != nil {
		panic(err)
	}

	RunMigrations(db)

	log.Info("MySQL connected via database/sql")

	return db, func() {
		_ = db.Close()
	}
}
