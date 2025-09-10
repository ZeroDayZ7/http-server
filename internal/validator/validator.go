package validator

import (
	"github.com/zerodayz7/http-server/internal/middleware"
)

type InteractionRequest struct {
	Type string `json:"type" validate:"required,oneof=like dislike"`
}

func ValidateStruct(s any) map[string]string {
	return middleware.ValidateStruct(s)
}
