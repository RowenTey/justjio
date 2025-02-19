package util

import (
	"strings"
	"time"
)

// FormatUsername takes a name string and returns a formatted username
// by replacing spaces with underscores and adding today's date (DDMMYYYY)
func FormatUsername(name string) string {
	// Replace spaces with underscores
	formattedName := strings.ReplaceAll(strings.ToLower(name), " ", "_")

	// Get current date and format it
	// Format: DDMMYY
	currentDate := time.Now().Format("020106")

	// Combine name and date
	return formattedName + "_" + currentDate
}
