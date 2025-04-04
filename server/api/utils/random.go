package utils

import (
	"crypto/rand"
	"encoding/base64"
)

// GenerateRandomString generates a secure random string of the specified length.
func GenerateRandomString(length int) string {
	if length <= 0 {
		return ""
	}

	// Calculate the number of bytes needed to generate the desired length
	numBytes := (length * 3) / 4

	// Generate random bytes
	randomBytes := make([]byte, numBytes)
	if _, err := rand.Read(randomBytes); err != nil {
		return ""
	}

	// Encode the random bytes to a base64 string
	randomString := base64.URLEncoding.EncodeToString(randomBytes)

	// Trim the string to the desired length
	return randomString[:length]
}
