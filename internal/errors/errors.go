package errors

type ErrorType string

const (
	Unauthorized ErrorType = "UNAUTHORIZED"
	Validation   ErrorType = "VALIDATION"
	NotFound     ErrorType = "NOT_FOUND"
	Internal     ErrorType = "INTERNAL"
	BadRequest   ErrorType = "BAD_REQUEST"
)

var ErrorMessages = map[ErrorType]string{
	Unauthorized: "Brak autoryzacji.",
	Validation:   "Nieprawidłowe dane.",
	NotFound:     "Zasób nie został znaleziony.",
	Internal:     "Wewnętrzny błąd serwera.",
	BadRequest:   "Błędne żądanie.",
}

type AppError struct {
	Code    string
	Type    ErrorType
	Message string
	Err     error
	Meta    map[string]any
}

func (e *AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return ErrorMessages[e.Type]
}

var (
	ErrEmailExists      = &AppError{Code: "EMAIL_EXISTS", Type: Validation, Message: "Email already registered"}
	ErrUsernameExists   = &AppError{Code: "USERNAME_EXISTS", Type: Validation, Message: "Username already exist"}
	ErrPasswordTooShort = &AppError{Code: "PASSWORD_TOO_SHORT", Type: Validation, Message: "Password must be at least 8 characters"}
)
