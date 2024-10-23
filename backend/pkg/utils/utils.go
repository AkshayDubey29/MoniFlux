package utils

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"time"
)

// GenerateRandomString generates a random string of the specified length.
// It can be used for creating unique tokens, IDs, or secrets.
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// HTTPError sends a standardized JSON error response with the specified status code.
func HTTPError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write([]byte(`{"error": "` + message + `"}`))
}

// IsValidEmail validates the format of an email address.
func IsValidEmail(email string) bool {
	// Very basic regex for email validation. This can be replaced with a more sophisticated validation if needed.
	return len(email) >= 3 && len(email) <= 254 && emailContainsAtSymbol(email)
}

// emailContainsAtSymbol is a helper function to check if an email address contains an "@" symbol.
func emailContainsAtSymbol(email string) bool {
	for _, char := range email {
		if char == '@' {
			return true
		}
	}
	return false
}

// TimeNow returns the current time in UTC.
func TimeNow() time.Time {
	return time.Now().UTC()
}

// ParseDuration safely parses a string duration (e.g., "1h", "30m") and returns a time.Duration object.
// Returns an error if the format is invalid.
func ParseDuration(durationStr string) (time.Duration, error) {
	if duration, err := time.ParseDuration(durationStr); err != nil {
		return 0, err
	} else {
		return duration, nil
	}
}

// ValidateRequestMethod checks if the HTTP request method is valid for the given handler.
func ValidateRequestMethod(r *http.Request, allowedMethods []string) error {
	for _, method := range allowedMethods {
		if r.Method == method {
			return nil
		}
	}
	return errors.New("invalid request method")
}
