package middleware

import (
	"encoding/json"
	"time"

	"github.com/zerodayz7/http-server/internal/service"
)

// MySQLStore implementuje fiber.Storage
type MySQLStore struct {
	sessionService *service.SessionService
}

func NewMySQLStore(sessSvc *service.SessionService) *MySQLStore {
	return &MySQLStore{sessionService: sessSvc}
}

// Get pobiera dane sesji
func (s *MySQLStore) Get(key string) ([]byte, error) {
	sess, err := s.sessionService.GetSession(key)
	if err != nil {
		// Jeśli błąd to "record not found", traktuj jako nieistniejącą sesję
		if err.Error() == "record not found" { // Lub sprawdź sql.ErrNoRows w service
			return []byte("{}"), nil
		}
		return nil, err
	}
	if sess == nil {
		return []byte("{}"), nil
	}
	return []byte(sess.Data), nil
}

// Set zapisuje dane sesji z expirem (optymalizacja: użyj exp dla expires_at)
func (s *MySQLStore) Set(key string, val []byte, exp time.Duration) error {
	var data map[string]any
	if len(val) > 0 {
		if err := json.Unmarshal(val, &data); err != nil {
			return err
		}
	}
	// Optymalizacja: Przekaż exp do service (dodaj parametr expiresAt := time.Now().Add(exp))
	return s.sessionService.UpdateSessionData(key, data) // Dodaj exp w service, jeśli nie masz
}

// Delete usuwa sesję
func (s *MySQLStore) Delete(key string) error {
	return s.sessionService.DeleteSession(key)
}

// Reset pusty
func (s *MySQLStore) Reset() error {
	return nil
}

// Close pusty
func (s *MySQLStore) Close() error {
	return nil
}
