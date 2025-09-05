package validator

import (
	"regexp"
	"strings"
)

func ValidateUsername(username string) string {
	if len(username) < 3 || len(username) > 30 {
		return "Username must be 3-30 characters"
	}

	re := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	if !re.MatchString(username) {
		return "Username can only contain letters and numbers"
	}

	return ""
}

func ValidateEmail(email string) string {
	reEmail := regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
	if !reEmail.MatchString(email) {
		return "Invalid email address"
	}
	return ""
}

func ValidatePassword(password string) string {
	if len(password) < 8 {
		return "Password must be at least 8 characters"
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, c := range password {
		switch {
		case 'A' <= c && c <= 'Z':
			hasUpper = true
		case 'a' <= c && c <= 'z':
			hasLower = true
		case '0' <= c && c <= '9':
			hasDigit = true
		case strings.ContainsRune("!@#$%^&*()-_=+[]{}|;:,.<>/?~`", c):
			hasSpecial = true
		default:
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit || !hasSpecial {
		return "Password must include 1 uppercase, 1 lowercase, 1 number, and 1 special character"
	}
	return ""
}
