package validator

import (
	"github.com/zerodayz7/http-server/internal/middleware"
)

type RegisterRequest struct {
	Username string `json:"username" validate:"required,alphanum,min=3,max=30"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,passwd"`
}

type CheckEmailRequest struct {
	Email string `json:"email" validate:"required,email"`
}

func ValidateStruct(s any) map[string]string {
	return middleware.ValidateStruct(s)
}
