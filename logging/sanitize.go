// Package logging provides utilities for secure logging with PII sanitization.
package logging

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
)

// Salt for hashing emails - in production, this should be configured securely
const defaultSalt = "potato-service-pii-salt-2024"

var (
	// emailRegex matches common email patterns
	emailRegex = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
)

// HashEmail creates a one-way SHA-256 hash of an email address with salt.
// The resulting hash is truncated to 12 characters for readability while
// maintaining sufficient uniqueness for debugging purposes.
func HashEmail(email string) string {
	return HashEmailWithSalt(email, defaultSalt)
}

// HashEmailWithSalt creates a one-way SHA-256 hash of an email address with a custom salt.
func HashEmailWithSalt(email, salt string) string {
	// Normalize email to lowercase for consistent hashing
	normalized := strings.ToLower(strings.TrimSpace(email))
	data := salt + normalized

	hash := sha256.Sum256([]byte(data))
	fullHash := hex.EncodeToString(hash[:])

	// Return truncated hash prefixed with "user_" for identification
	return "user_" + fullHash[:12]
}

// SanitizeLogMessage scans a log message for email patterns and replaces them
// with hashed identifiers to prevent PII leakage.
func SanitizeLogMessage(message string) string {
	return emailRegex.ReplaceAllStringFunc(message, func(email string) string {
		return HashEmail(email)
	})
}

// SanitizeLogMessageWithSalt scans a log message for email patterns and replaces
// them with hashed identifiers using a custom salt.
func SanitizeLogMessageWithSalt(message, salt string) string {
	return emailRegex.ReplaceAllStringFunc(message, func(email string) string {
		return HashEmailWithSalt(email, salt)
	})
}

// ContainsEmail checks if a string contains any email addresses.
// Useful for validation and testing purposes.
func ContainsEmail(message string) bool {
	return emailRegex.MatchString(message)
}

// UserIdentifier represents an anonymized user identifier for logging purposes.
type UserIdentifier struct {
	ID       string // Unique identifier (e.g., numeric ID or UUID)
	HashID   string // Hashed version of original email for correlation if needed
}

// NewUserIdentifier creates a UserIdentifier from a user ID.
func NewUserIdentifier(userID string) UserIdentifier {
	return UserIdentifier{
		ID:     userID,
		HashID: "",
	}
}

// NewUserIdentifierFromEmail creates a UserIdentifier by hashing an email address.
// This is useful for transitioning from email-based logging to ID-based logging.
func NewUserIdentifierFromEmail(email string) UserIdentifier {
	hashedID := HashEmail(email)
	return UserIdentifier{
		ID:     hashedID,
		HashID: hashedID,
	}
}

// String returns the user identifier for logging.
func (u UserIdentifier) String() string {
	return u.ID
}
