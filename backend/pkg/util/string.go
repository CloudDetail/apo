package util

import "strings"

// IsValidIdentifier checks if a string contains only letters, numbers, underscores, and hyphens
func IsValidIdentifier(s string) bool {
	for _, r := range s {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '-') {
			return false
		}
	}
	return true
}

// EscapeSQLString escapes special characters in SQL strings
func EscapeSQLString(s string) string {
	// Replace single quotes
	s = strings.ReplaceAll(s, "'", "''")
	// Replace backslashes
	s = strings.ReplaceAll(s, "\\", "\\\\")
	return s
}
