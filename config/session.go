package config

import (
	"time"

	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/mysql/v2"
)

var store *session.Store

// InitSessionStore inicjalizuje Fiber session store z MySQL
func InitSessionStore(dsn string, ttl time.Duration) *session.Store {
	if store != nil {
		return store
	}

	mysqlStorage := mysql.New(mysql.Config{
		ConnectionURI: dsn,
		Reset:         false,
		GCInterval:    10 * time.Second,
	})

	store = session.New(session.Config{
		Storage:        mysqlStorage,
		Expiration:     ttl,
		CookieSecure:   false,
		CookieHTTPOnly: true,
		CookieSameSite: "Strict",
	})

	return store
}

func SessionStore() *session.Store {
	return store
}
