package validator

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *RegisterRequest) Validate() map[string]any {
	errors := make(map[string]any)

	if msg := ValidateUsername(r.Username); msg != "" {
		errors["username"] = msg
	}
	if msg := ValidateEmail(r.Email); msg != "" {
		errors["email"] = msg
	}
	if msg := ValidatePassword(r.Password); msg != "" {
		errors["password"] = msg
	}

	return errors
}
