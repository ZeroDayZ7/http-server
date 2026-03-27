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

// 1. Definiujemy provider pomocniczy POZA funkcją Initialize
func provideInteractionWorker(
	rdb *redis.Client,
	repo service.InteractionRepository,
	log logger.Logger,
	cfg *env.Config,
) *worker.InteractionWorker {
	return worker.NewInteractionWorker(
		rdb,
		repo,
		log,
		cfg.Shutdown,            // To idzie do 'timeout'
		cfg.WorkerFlushInterval, // To idzie do 'flushInterval'
	)
}

func InitializeInteractionModule(
	sqlDB *sql.DB,
	redisClient *redis.Client,
	cfg *env.Config,
	log logger.Logger,
) (*InteractionModule, error) {
	panic(wire.Build(
		// Zostawiamy tylko te pola, które nie kolidują typami
		wire.FieldsOf(new(*env.Config), "FingerprintSalt"),

		wire.Bind(new(db.DBTX), new(*sql.DB)),
		db.New,

		mysqlrepo.NewInteractionRepository,
		wire.Bind(new(service.InteractionRepository), new(*mysqlrepo.MySQLInteractionRepo)),

		redisrepo.NewRedisInteractionCache,
		wire.Bind(new(service.InteractionCache), new(*redisrepo.RedisInteractionCache)),

		intRedis.NewStreamProducer,
		wire.Bind(new(service.EventPublisher), new(*intRedis.StreamProducer)),

		service.NewIdentityService,
		service.NewInteractionService,
		wire.Bind(new(service.InteractionServiceInterface), new(*service.InteractionService)),

		handler.NewInteractionHandler,

		provideInteractionWorker,

		NewInteractionModule,
	))
}
