package utils

import (
	"errors"
	"regexp"
	"strings"
)

var (
	// Email regex pattern (basic validation)
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	
	// Username regex: alphanumeric, underscore, hyphen, 3-20 characters
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,20}$`)
)

// ValidateEmail checks if the email is in a valid format
func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return errors.New("email cannot be empty")
	}
	if len(email) > 254 {
		return errors.New("email is too long")
	}
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}
	return nil
}

// ValidateUsername checks if the username is valid
func ValidateUsername(username string) error {
	username = strings.TrimSpace(username)
	if username == "" {
		return errors.New("username cannot be empty")
	}
	if len(username) < 3 {
		return errors.New("username must be at least 3 characters")
	}
	if len(username) > 20 {
		return errors.New("username must not exceed 20 characters")
	}
	if !usernameRegex.MatchString(username) {
		return errors.New("username can only contain letters, numbers, underscores, and hyphens")
	}
	return nil
}

// ValidatePassword checks if the password meets minimum requirements
func ValidatePassword(password string) error {
	if password == "" {
		return errors.New("password cannot be empty")
	}
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	if len(password) > 72 {
		// bcrypt has a maximum password length of 72 bytes
		return errors.New("password must not exceed 72 characters")
	}
	return nil
}

// ValidateOTP checks if the OTP is valid
func ValidateOTP(otp string) error {
	otp = strings.TrimSpace(otp)
	if otp == "" {
		return errors.New("OTP cannot be empty")
	}
	if len(otp) != 6 {
		return errors.New("OTP must be 6 digits")
	}
	// Check if all characters are digits
	for _, char := range otp {
		if char < '0' || char > '9' {
			return errors.New("OTP must contain only digits")
		}
	}
	return nil
}

// SanitizeString removes leading/trailing whitespace and limits length
func SanitizeString(input string, maxLength int) string {
	input = strings.TrimSpace(input)
	if len(input) > maxLength {
		return input[:maxLength]
	}
	return input
}
