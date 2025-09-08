package config

import (
	"time"

	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/mysql/v2"
)

var store *session.Store

// InitSessionStore inicjalizuje Fiber session store z MySQL
func InitSessionStore() *session.Store {
	if store != nil {
		return store
	}

	// isProd := AppConfig.Server.Env == "production"
	ttl := AppConfig.SessionTTL
	dsn := AppConfig.Database.DSN

	mysqlStorage := mysql.New(mysql.Config{
		ConnectionURI: dsn,
		Reset:         false,
		GCInterval:    10 * time.Second,
	})

	store = session.New(session.Config{
		Storage:        mysqlStorage,
		Expiration:     ttl,
		CookieSecure:   true,
		CookieHTTPOnly: true,
		CookieSameSite: "Lax",
	})

	return store
}

func SessionStore() *session.Store {
	return store
}
