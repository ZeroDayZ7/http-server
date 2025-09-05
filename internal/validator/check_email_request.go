package validator

type CheckEmailRequest struct {
	Email string `json:"email"`
}

func (r *CheckEmailRequest) Validate() map[string]any {
	errors := make(map[string]any)

	if msg := ValidateEmail(r.Email); msg != "" {
		errors["email"] = msg
	}

	return errors
}
