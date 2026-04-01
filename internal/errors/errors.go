package errors

import (
	"fmt"
	"maps"
)

type ErrorType string

const (
	Unauthorized    ErrorType = "UNAUTHORIZED"
	Validation      ErrorType = "VALIDATION"
	NotFound        ErrorType = "NOT_FOUND"
	Internal        ErrorType = "INTERNAL"
	BadRequest      ErrorType = "BAD_REQUEST"
	TooManyRequests ErrorType = "TOO_MANY_REQUESTS"
)

type AppError struct {
	Code    string
	Type    ErrorType
	Message string
	Err     error
	Meta    map[string]any
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) WithDetail(key string, value any) *AppError {
	newErr := *e

	newMeta := make(map[string]any)
	maps.Copy(newMeta, e.Meta)
	newMeta[key] = value

	newErr.Meta = newMeta
	return &newErr
}

func (e *AppError) WithMeta(meta map[string]any) *AppError {
	newErr := *e
	newErr.Meta = meta
	return &newErr
}

func (e *AppError) WithErr(err error) *AppError {
	newErr := *e
	newErr.Err = err
	return &newErr
}

var (
	ErrInvalidRequest   = &AppError{Code: "INVALID_REQUEST", Type: Validation, Message: "Invalid request data"}
	ErrInternal         = &AppError{Code: "SERVER_ERROR", Type: Internal, Message: "Internal server error"}
	ErrInvalidJSON      = &AppError{Code: "INVALID_JSON", Type: BadRequest, Message: "Invalid JSON in request body"}
	ErrValidationFailed = &AppError{Code: "VALIDATION_FAILED", Type: Validation, Message: "Request validation failed"}
	ErrUnauthorized     = &AppError{Code: "UNAUTHORIZED", Type: Unauthorized, Message: "Unauthorized access"}
	ErrUserNotFound     = &AppError{Code: "USER_NOT_FOUND", Type: NotFound, Message: "User not found"}
	ErrTooManyRequests  = &AppError{Code: "TOO_MANY_REQUESTS", Type: TooManyRequests, Message: "Too many requests"}
)
