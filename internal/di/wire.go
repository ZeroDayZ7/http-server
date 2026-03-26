//go:build wireinject
// +build wireinject

package di

import (
	"database/sql"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"github.com/zerodayz7/http-server/config"
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
	cfg *config.Config,
	log logger.Logger,
) (*InteractionModule, error) {
	panic(wire.Build(
		// 1. Wyciąganie pól z configu
		wire.FieldsOf(new(*config.Config), "FingerprintSalt"),

		// 2. Baza danych
		wire.Bind(new(db.DBTX), new(*sql.DB)),
		db.New,

		// 3. Repozytoria (MySQL)
		mysqlrepo.NewInteractionRepository,
		wire.Bind(new(service.InteractionRepository), new(*mysqlrepo.MySQLInteractionRepo)),
		wire.Bind(new(worker.InteractionRepository), new(*mysqlrepo.MySQLInteractionRepo)),

		// 4. Cache (Redis)
		redisrepo.NewRedisInteractionCache,
		wire.Bind(new(service.InteractionCache), new(*redisrepo.RedisInteractionCache)),

		// 5. Eventy (Producer)
		intRedis.NewStreamProducer,
		wire.Bind(new(service.EventPublisher), new(*intRedis.StreamProducer)),

		// 6. Serwisy
		service.NewIdentityService,
		service.NewInteractionService,
		// TO POŁĄCZENIE JEST KLUCZOWE:
		wire.Bind(new(service.InteractionServiceInterface), new(*service.InteractionService)),

		// 7. Handler i Worker
		handler.NewInteractionHandler,
		worker.NewInteractionWorker,

		// 8. Finalny moduł
		NewInteractionModule,
	))
}
