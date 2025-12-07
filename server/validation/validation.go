package validation

import (
	"regexp"
	"strings"
	"unicode"
)

const (
	MinUsernameLength = 3
	MaxUsernameLength = 50
)

const (
	MinPasswordLength = 8
	MaxPasswordLength = 128
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

func ValidateUsername(username string) (bool, string) {
	username = strings.TrimSpace(username)

	if username == "" {
		return false, "Username is required"
	}

	if len(username) < MinUsernameLength {
		return false, "Username must be at least 3 characters"
	}

	if len(username) > MaxUsernameLength {
		return false, "Username must not exceed 50 characters"
	}

	if !usernameRegex.MatchString(username) {
		return false, "Username can only contain letters, numbers, and underscores"
	}

	return true, ""
}

func ValidatePassword(password string) (bool, string) {
	if password == "" {
		return false, "Password is required"
	}

	if len(password) < MinPasswordLength {
		return false, "Password must be at least 8 characters"
	}

	if len(password) > MaxPasswordLength {
		return false, "Password must not exceed 128 characters"
	}

	var (
		hasUpper bool
		hasLower bool
		hasDigit bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		}
	}

	if !hasUpper {
		return false, "Password must contain at least one uppercase letter"
	}

	if !hasLower {
		return false, "Password must contain at least one lowercase letter"
	}

	if !hasDigit {
		return false, "Password must contain at least one digit"
	}

	return true, ""
}
