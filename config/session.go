package config

import (
	"time"

	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/mysql/v2"
	"gorm.io/gorm"
)

var store *session.Store

// InitSessionStore initializes the Fiber session store with MySQL
func InitSessionStore(db *gorm.DB) *session.Store {
	if store != nil {
		return store
	}

	sqlDB, err := db.DB() // conversion *gorm.DB → *sql.DB
	if err != nil {
		panic(err)
	}

	// isProd := AppConfig.Server.Env == "production"
	ttl := AppConfig.SessionTTL

	mysqlStorage := mysql.New(mysql.Config{
		Db:         sqlDB,
		Reset:      false,
		GCInterval: 10 * time.Second,
	})

	store = session.New(session.Config{
		Storage:        mysqlStorage,
		KeyLookup:      "cookie:session_id",
		Expiration:     ttl,
		CookieSecure:   true,
		CookieHTTPOnly: true,
		CookieSameSite: "None",
	})

	return store
}

func SessionStore() *session.Store {
	return store
}
