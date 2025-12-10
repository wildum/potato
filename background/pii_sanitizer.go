package background

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// SanitizeEmailForLog converts an email address to a privacy-preserving identifier
// suitable for logging. It returns a masked version of the email that still allows
// correlation within logs without exposing the actual email address.
//
// Example: "john.smith@example.com" -> "j***@e***.com"
func SanitizeEmailForLog(email string) string {
	if email == "" {
		return "unknown_user"
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		// Not a valid email format, return a hash instead
		return hashIdentifier(email)
	}

	localPart := parts[0]
	domainPart := parts[1]

	// Mask the local part: keep first character, mask the rest
	maskedLocal := maskString(localPart)

	// Mask the domain: keep first character and TLD
	maskedDomain := maskDomain(domainPart)

	return maskedLocal + "@" + maskedDomain
}

// GenerateUserLogID creates a short hash-based identifier for a user
// that can be used for log correlation without exposing PII.
//
// Example: "john.smith@example.com" -> "user_a1b2c3d4"
func GenerateUserLogID(email string) string {
	if email == "" {
		return "user_unknown"
	}
	return "user_" + hashIdentifier(email)[:8]
}

// maskString masks a string keeping only the first character visible
// Example: "john.smith" -> "j***"
func maskString(s string) string {
	if len(s) == 0 {
		return "***"
	}
	if len(s) == 1 {
		return s + "***"
	}
	return string(s[0]) + "***"
}

// maskDomain masks a domain keeping the first character and TLD
// Example: "example.com" -> "e***.com"
func maskDomain(domain string) string {
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return maskString(domain)
	}

	// Get the main domain part and TLD
	tld := parts[len(parts)-1]
	mainDomain := strings.Join(parts[:len(parts)-1], ".")

	return maskString(mainDomain) + "." + tld
}

// hashIdentifier creates a SHA256 hash of the input and returns
// the first 16 characters of the hex encoding
func hashIdentifier(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])[:16]
}
