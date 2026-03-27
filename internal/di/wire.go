//go:build wireinject
// +build wireinject

package di

import (
	"database/sql"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"github.com/zerodayz7/http-server/config/env"
	"github.com/zerodayz7/http-server/internal/db"
	"github.com/zerodayz7/http-server/internal/handler"
	intRedis "github.com/zerodayz7/http-server/internal/redis"
	mysqlrepo "github.com/zerodayz7/http-server/internal/repository/mysql"
	redisrepo "github.com/zerodayz7/http-server/internal/repository/redis"
	"github.com/zerodayz7/http-server/internal/service"
	"github.com/zerodayz7/http-server/internal/shared/logger"
	"github.com/zerodayz7/http-server/internal/worker"
)

func InitializeInteractionModule(
	sqlDB *sql.DB,
	redisClient *redis.Client,
	cfg *env.Config,
	log logger.Logger,
) (*InteractionModule, error) {
	panic(wire.Build(
		wire.FieldsOf(new(*env.Config), "FingerprintSalt"),

		// FIX 1: Mapowanie *sql.DB na db.DBTX (wymagane przez SQLc)
		wire.Bind(new(db.DBTX), new(*sql.DB)),
		db.New,

		// Repozytorium MySQL
		mysqlrepo.NewInteractionRepository,
		wire.Bind(new(service.InteractionRepository), new(*mysqlrepo.MySQLInteractionRepo)),

		// Cache Redis
		redisrepo.NewRedisInteractionCache,
		wire.Bind(new(service.InteractionCache), new(*redisrepo.RedisInteractionCache)),

		// Events
		intRedis.NewStreamProducer,
		wire.Bind(new(service.EventPublisher), new(*intRedis.StreamProducer)),

		// Serwisy
		service.NewIdentityService,
		service.NewInteractionService,
		wire.Bind(new(service.InteractionServiceInterface), new(*service.InteractionService)),

		// Handler & Worker
		handler.NewInteractionHandler,
		worker.NewInteractionWorker,

		NewInteractionModule,
	))
}
