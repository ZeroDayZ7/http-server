package validator

import (
	"github.com/zerodayz7/http-server/internal/middleware"
)

type InteractionRequest struct {
	Type        string `json:"type" validate:"required,oneof=like dislike visit"`
	Fingerprint string `json:"fingerprint" validate:"required"`
}

type FingerprintRequest struct {
	Fingerprint string `json:"fingerprint" validate:"required"`
}

func ValidateStruct(s any) map[string]string {
	return middleware.ValidateStruct(s)
}
