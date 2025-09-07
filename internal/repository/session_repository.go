package repository

import "github.com/zerodayz7/http-server/internal/model"

type SessionRepository interface {
	CreateOrUpdate(session *model.Session) error
	GetBySessionID(sessionID string) (*model.Session, error)
	Delete(sessionID string) error
}
