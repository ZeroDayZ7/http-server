package shared

import "github.com/google/uuid"

// GenerateCSRFToken zwraca losowy token CSRF
func GenerateCSRFToken() string {
	return uuid.NewString()
}
