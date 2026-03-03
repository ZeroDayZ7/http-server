//go:build wireinject
// +build wireinject

package di

import (
	"database/sql"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"github.com/zerodayz7/http-server/internal/db"
	"github.com/zerodayz7/http-server/internal/handler"
	mysqlrepo "github.com/zerodayz7/http-server/internal/repository/mysql"
	"github.com/zerodayz7/http-server/internal/service"
)

func InitializeInteractionModule(
	sqlDB *sql.DB,
	redisClient *redis.Client,
) (*handler.InteractionHandler, error) {
	wire.Build(
		// To rozwiązuje błąd DBTX:
		// Mówimy Wire: "Jeśli ktoś chce DBTX (a chce go db.New), daj mu sqlDB (*sql.DB)"
		wire.Bind(new(db.DBTX), new(*sql.DB)),

		db.New,
		mysqlrepo.NewInteractionRepository,

		wire.Bind(
			new(service.InteractionRepository),
			new(*mysqlrepo.MySQLInteractionRepo),
		),

		service.NewInteractionService,
		handler.NewInteractionHandler,
	)

	return nil, nil
}
