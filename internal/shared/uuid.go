package shared

import "github.com/google/uuid"

func GenerateUuid() string {
	return uuid.NewString()
}
