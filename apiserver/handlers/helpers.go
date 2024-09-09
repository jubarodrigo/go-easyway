package handlers

import (
	"errors"
	"regexp"
)

const (
	minPasswordLength = 8
	ScopeSA           = "SERVICE_ACCOUNT"
)

func IsValidPassword(password string) error {
	if len(password) < minPasswordLength {
		return errors.New("password must be at least 8 characters long")
	}

	hasNumber := false
	hasSpecial := false
	hasUpperCase := false
	hasLowerCase := false

	for _, char := range password {
		switch {
		case '0' <= char && char <= '9':
			hasNumber = true
		case 'A' <= char && char <= 'Z':
			hasUpperCase = true
		case 'a' <= char && char <= 'z':
			hasLowerCase = true
		case char == '!' || char == '@' || char == '#' || char == '$' || char == '%' || char == '^' || char == '&' || char == '*':
			hasSpecial = true
		}
	}

	if !hasNumber {
		return errors.New("password must contain at least one number")
	}
	if !hasSpecial {
		return errors.New("password must contain at least one special character")
	}
	if !hasUpperCase {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !hasLowerCase {
		return errors.New("password must contain at least one lowercase letter")
	}

	return nil
}

func IsValidEmail(email string) bool {
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	match, _ := regexp.MatchString(emailRegex, email)
	return match
}
