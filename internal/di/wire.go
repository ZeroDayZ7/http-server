//go:build wireinject
// +build wireinject

package di

import (
	"database/sql"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"github.com/zerodayz7/http-server/internal/db"
	"github.com/zerodayz7/http-server/internal/handler"
	redisrepo "github.com/zerodayz7/http-server/internal/redis"
	mysqlrepo "github.com/zerodayz7/http-server/internal/repository/mysql"
	"github.com/zerodayz7/http-server/internal/service"
	"github.com/zerodayz7/http-server/internal/worker"
)

func InitializeInteractionModule(
	sqlDB *sql.DB,
	redisClient *redis.Client,
	salt string,
) (*InteractionModule, error) {
	panic(wire.Build(
		// DB & Repository
		wire.Bind(new(db.DBTX), new(*sql.DB)),
		db.New,
		mysqlrepo.NewInteractionRepository,

		// Interface Binds
		wire.Bind(new(service.InteractionRepository), new(*mysqlrepo.MySQLInteractionRepo)),
		wire.Bind(new(worker.InteractionRepository), new(*mysqlrepo.MySQLInteractionRepo)),

		// Infrastructure & Services
		redisrepo.NewStreamProducer,
		service.NewInteractionService,

		// Components
		handler.NewInteractionHandler,
		worker.NewInteractionWorker,

		// Module Aggregator
		NewInteractionModule,
	))
}
