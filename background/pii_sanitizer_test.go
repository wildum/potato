package background

import (
	"strings"
	"testing"
)

func TestSanitizeEmailForLog(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		wantMask string // substring that should be present
		noEmail  bool   // should not contain the original email
	}{
		{
			name:     "standard email",
			email:    "john.smith@example.com",
			wantMask: "j***@e***.com",
			noEmail:  true,
		},
		{
			name:     "short local part",
			email:    "a@b.com",
			wantMask: "a***@b***.com",
			noEmail:  true,
		},
		{
			name:     "empty email",
			email:    "",
			wantMask: "unknown_user",
			noEmail:  true,
		},
		{
			name:     "subdomain email",
			email:    "admin@potato-warehouse.internal",
			wantMask: "a***@p***.internal",
			noEmail:  true,
		},
		{
			name:     "complex email",
			email:    "bob.wilson@email.net",
			wantMask: "b***@e***.net",
			noEmail:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeEmailForLog(tt.email)

			if got != tt.wantMask {
				t.Errorf("SanitizeEmailForLog(%q) = %q, want %q", tt.email, got, tt.wantMask)
			}

			// Verify the original email is not present in the output
			if tt.noEmail && tt.email != "" && strings.Contains(got, tt.email) {
				t.Errorf("SanitizeEmailForLog(%q) = %q, should not contain original email", tt.email, got)
			}
		})
	}
}

func TestGenerateUserLogID(t *testing.T) {
	tests := []struct {
		name  string
		email string
	}{
		{
			name:  "standard email",
			email: "john.smith@example.com",
		},
		{
			name:  "empty email",
			email: "",
		},
		{
			name:  "another email",
			email: "bob.wilson@email.net",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateUserLogID(tt.email)

			// Should start with "user_"
			if !strings.HasPrefix(got, "user_") {
				t.Errorf("GenerateUserLogID(%q) = %q, should start with 'user_'", tt.email, got)
			}

			// Should not contain the original email
			if tt.email != "" && strings.Contains(got, tt.email) {
				t.Errorf("GenerateUserLogID(%q) = %q, should not contain original email", tt.email, got)
			}

			// Should be consistent (same input = same output)
			got2 := GenerateUserLogID(tt.email)
			if got != got2 {
				t.Errorf("GenerateUserLogID(%q) is not consistent: %q != %q", tt.email, got, got2)
			}

			// Should not contain @ symbol (no email parts leaked)
			if strings.Contains(got, "@") {
				t.Errorf("GenerateUserLogID(%q) = %q, should not contain '@'", tt.email, got)
			}
		})
	}
}

func TestGenerateUserLogID_Uniqueness(t *testing.T) {
	// Different emails should produce different IDs
	emails := []string{
		"john.smith@example.com",
		"alice.johnson@company.org",
		"bob.wilson@email.net",
		"sarah.davis@corporate.io",
	}

	ids := make(map[string]string)
	for _, email := range emails {
		id := GenerateUserLogID(email)
		if existingEmail, exists := ids[id]; exists {
			t.Errorf("Collision detected: %q and %q both produce ID %q", email, existingEmail, id)
		}
		ids[id] = email
	}
}

func TestNoEmailLeakage(t *testing.T) {
	// Comprehensive test to ensure no part of the email is leaked
	testEmails := []string{
		"john.smith@example.com",
		"alice.johnson@company.org",
		"bob.wilson@email.net",
		"sarah.davis@corporate.io",
		"support@freshpotatoes.com",
		"admin@potato-warehouse.internal",
	}

	for _, email := range testEmails {
		sanitized := SanitizeEmailForLog(email)
		userID := GenerateUserLogID(email)

		// Extract parts of the email that should not appear
		parts := strings.Split(email, "@")
		if len(parts) == 2 {
			localPart := parts[0]
			domainPart := parts[1]

			// Local part (minus first char) should not appear
			if len(localPart) > 1 && strings.Contains(sanitized, localPart[1:]) {
				t.Errorf("Sanitized email %q leaks local part from %q", sanitized, email)
			}

			// Domain (minus first char and TLD) should not appear in full
			domainParts := strings.Split(domainPart, ".")
			if len(domainParts) > 0 && len(domainParts[0]) > 1 {
				if strings.Contains(sanitized, domainParts[0][1:]) {
					t.Errorf("Sanitized email %q leaks domain from %q", sanitized, email)
				}
			}
		}

		// User ID should not contain @ or full email
		if strings.Contains(userID, "@") || strings.Contains(userID, email) {
			t.Errorf("User ID %q leaks email information from %q", userID, email)
		}
	}
}
