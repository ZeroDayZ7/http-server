package mysql

import (
	"context"

	"github.com/zerodayz7/http-server/internal/db"
	"github.com/zerodayz7/http-server/internal/service"
)

var _ service.InteractionRepository = (*MySQLInteractionRepo)(nil)

type MySQLInteractionRepo struct {
	q *db.Queries
}

func NewInteractionRepository(q *db.Queries) *MySQLInteractionRepo {
	return &MySQLInteractionRepo{q: q}
}

func (r *MySQLInteractionRepo) Increment(ctx context.Context, typ string) error {
	return r.q.IncrementCounter(ctx, typ)
}

func (r *MySQLInteractionRepo) GetCount(ctx context.Context, typ string) (int, error) {
	count, err := r.q.GetCountByType(ctx, typ)
	return int(count), err
}
