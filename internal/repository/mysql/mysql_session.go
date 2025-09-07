package mysql

import (
	"errors"

	"github.com/zerodayz7/http-server/internal/model"
	"github.com/zerodayz7/http-server/internal/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ repository.SessionRepository = (*MySQLSessionRepo)(nil)

type MySQLSessionRepo struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) *MySQLSessionRepo {
	return &MySQLSessionRepo{db: db}
}

func (r *MySQLSessionRepo) CreateOrUpdate(session *model.Session) error {
	return r.db.Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "session_id"}, {Name: "user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"data", "updated_at", "expires_at"}),
		},
	).Create(session).Error
}

func (r *MySQLSessionRepo) GetBySessionID(sessionID string) (*model.Session, error) {
	var s model.Session
	if err := r.db.First(&s, "session_id = ?", sessionID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *MySQLSessionRepo) Delete(sessionID string) error {
	return r.db.Delete(&model.Session{}, "session_id = ?", sessionID).Error
}
