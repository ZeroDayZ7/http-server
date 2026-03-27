package mysql

import (
	"context"

	"github.com/zerodayz7/http-server/internal/db"
)

type MySQLInteractionRepo struct {
	q *db.Queries
}

func NewInteractionRepository(q *db.Queries) *MySQLInteractionRepo {
	return &MySQLInteractionRepo{q: q}
}

func (r *MySQLInteractionRepo) Increment(ctx context.Context, typ string) error {
	return r.q.IncrementStat(ctx, typ)
}

func (r *MySQLInteractionRepo) IncrementBy(ctx context.Context, typ string, amount int64) error {
	return r.q.IncrementStatByAmount(ctx, db.IncrementStatByAmountParams{
		Type:         typ,
		CurrentCount: amount,
	})
}

func (r *MySQLInteractionRepo) GetStats(ctx context.Context) (int64, int64, int64, error) {
	row, err := r.q.GetStats(ctx)
	if err != nil {
		return 0, 0, 0, err
	}

	return row.Likes, row.Dislikes, row.Visits, nil
}
