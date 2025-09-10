package mysql

import (
	"time"

	"github.com/zerodayz7/http-server/internal/model"
	"github.com/zerodayz7/http-server/internal/repository"
	"gorm.io/gorm"
)

var _ repository.InteractionRepository = (*MySQLInteractionRepo)(nil)

// Typy interakcji
const (
	InteractionVisit   = "visit"
	InteractionLike    = "like"
	InteractionDislike = "dislike"
	InteractionComment = "comment"
)

type MySQLInteractionRepo struct {
	db *gorm.DB
}

// NewInteractionRepository tworzy repozytorium interakcji/statystyk
func NewInteractionRepository(db *gorm.DB) *MySQLInteractionRepo {
	return &MySQLInteractionRepo{db: db}
}

// Add zapisuje interakcję
func (r *MySQLInteractionRepo) Add(ip string, userID *uint, typ string, value int, content *string, at time.Time) error {
	return r.db.Create(&model.Interaction{
		IP:        ip,
		UserID:    userID,
		Type:      typ,
		Value:     value,
		Content:   content,
		CreatedAt: at,
	}).Error
}

// Count zwraca sumę wszystkich interakcji danego typu
func (r *MySQLInteractionRepo) Count(typ string) (int, error) {
	var count int64
	if err := r.db.Model(&model.Interaction{}).Where("type = ?", typ).Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

// CountByIP zwraca sumę interakcji danego typu dla określonego IP
func (r *MySQLInteractionRepo) CountByIP(ip, typ string) (int, error) {
	var count int64
	if err := r.db.Model(&model.Interaction{}).Where("type = ? AND ip = ?", typ, ip).Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

// GetLastVisit zwraca czas ostatniej wizyty dla danego IP
func (r *MySQLInteractionRepo) GetLastVisit(ip string) (time.Time, error) {
	var last model.Interaction
	err := r.db.Model(&model.Interaction{}).
		Where("type = ? AND ip = ?", InteractionVisit, ip).
		Order("created_at DESC").
		Limit(1).
		First(&last).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return time.Time{}, nil // brak wizyt
		}
		return time.Time{}, err
	}
	return last.CreatedAt, nil
}

func (r *MySQLInteractionRepo) GetLastInteraction(ip, typ string) (time.Time, error) {
	var last model.Interaction
	err := r.db.Model(&model.Interaction{}).
		Where("type = ? AND ip = ?", typ, ip).
		Order("created_at DESC").
		Limit(1).
		First(&last).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return time.Time{}, nil
		}
		return time.Time{}, err
	}
	return last.CreatedAt, nil
}
