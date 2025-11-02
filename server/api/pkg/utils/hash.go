package utils

import "golang.org/x/crypto/bcrypt"

// Hash password
func HashPassword(password string) (string, error) {
	// See https://stackoverflow.com/a/69568265
	// Set cost to 10 for optimal tradeoff between security and performance
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}

// Compare password with hash
func CheckPasswordHash(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
