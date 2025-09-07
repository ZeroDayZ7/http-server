package service

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/zerodayz7/http-server/internal/model"
	"github.com/zerodayz7/http-server/internal/repository"
)

type SessionService struct {
	repo repository.SessionRepository
	ttl  time.Duration
}

func NewSessionService(repo repository.SessionRepository, ttl time.Duration) *SessionService {
	return &SessionService{
		repo: repo,
		ttl:  ttl,
	}
}

// Create a new session for a user
func (s *SessionService) CreateSession(userID uint, data map[string]any) (*model.Session, error) {
	session := &model.Session{
		SessionID: uuid.New().String(),
		UserID:    userID,
		Data:      marshalData(data),
		ExpiresAt: time.Now().Add(s.ttl),
	}
	err := s.repo.CreateOrUpdate(session)
	return session, err
}

// Get session by ID
func (s *SessionService) GetSession(sessionID string) (*model.Session, error) {
	return s.repo.GetBySessionID(sessionID)
}

// Delete session by ID
func (s *SessionService) DeleteSession(sessionID string) error {
	return s.repo.Delete(sessionID)
}

// Update session data
func (s *SessionService) UpdateSessionData(sessionID string, data map[string]any) error {
	session, err := s.repo.GetBySessionID(sessionID)
	if err != nil {
		return err
	}
	if session == nil {
		return nil // albo błąd "session not found"
	}
	session.Data = marshalData(data)
	session.UpdatedAt = time.Now()
	return s.repo.CreateOrUpdate(session)
}

// Helper to marshal data to JSON string
func marshalData(data map[string]any) string {
	// tutaj możesz użyć json.Marshal i error handling
	bytes, _ := json.Marshal(data)
	return string(bytes)
}
