package logging

import (
	"strings"
	"testing"
)

func TestHashEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		wantLen  int
		wantPref string
	}{
		{
			name:     "standard email",
			email:    "john.smith@example.com",
			wantLen:  17, // "user_" (5) + 12 char hash
			wantPref: "user_",
		},
		{
			name:     "uppercase email normalized",
			email:    "John.Smith@Example.COM",
			wantLen:  17,
			wantPref: "user_",
		},
		{
			name:     "email with spaces trimmed",
			email:    "  john.smith@example.com  ",
			wantLen:  17,
			wantPref: "user_",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HashEmail(tt.email)
			if len(got) != tt.wantLen {
				t.Errorf("HashEmail() len = %d, want %d", len(got), tt.wantLen)
			}
			if !strings.HasPrefix(got, tt.wantPref) {
				t.Errorf("HashEmail() = %v, want prefix %v", got, tt.wantPref)
			}
		})
	}
}

func TestHashEmailConsistency(t *testing.T) {
	email := "test@example.com"
	hash1 := HashEmail(email)
	hash2 := HashEmail(email)

	if hash1 != hash2 {
		t.Errorf("HashEmail() not consistent: %s != %s", hash1, hash2)
	}

	// Different emails should produce different hashes
	differentEmail := "different@example.com"
	hash3 := HashEmail(differentEmail)
	if hash1 == hash3 {
		t.Errorf("Different emails produced same hash")
	}
}

func TestHashEmailCaseInsensitive(t *testing.T) {
	email1 := "Test@Example.COM"
	email2 := "test@example.com"

	hash1 := HashEmail(email1)
	hash2 := HashEmail(email2)

	if hash1 != hash2 {
		t.Errorf("Case normalization failed: %s != %s", hash1, hash2)
	}
}

func TestSanitizeLogMessage(t *testing.T) {
	tests := []struct {
		name    string
		message string
		wantPII bool // should the result NOT contain email patterns
	}{
		{
			name:    "message with email",
			message: "User john.smith@example.com performed action",
			wantPII: false,
		},
		{
			name:    "message with multiple emails",
			message: "Users john@example.com and jane@company.org logged in",
			wantPII: false,
		},
		{
			name:    "message without email",
			message: "User user_1001 performed action",
			wantPII: false,
		},
		{
			name:    "inventory adjustment with email",
			message: "Inventory adjustment: Removed potato from inventory. Processed by user: admin@potato-warehouse.internal",
			wantPII: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeLogMessage(tt.message)
			if ContainsEmail(got) {
				t.Errorf("SanitizeLogMessage() still contains email: %s", got)
			}
			// Verify the hash format is present if original had email
			if ContainsEmail(tt.message) && !strings.Contains(got, "user_") {
				t.Errorf("SanitizeLogMessage() should contain user_ prefix: %s", got)
			}
		})
	}
}

func TestContainsEmail(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    bool
	}{
		{
			name:    "simple email",
			message: "Contact: john@example.com",
			want:    true,
		},
		{
			name:    "email with subdomain",
			message: "admin@mail.company.org",
			want:    true,
		},
		{
			name:    "email with plus",
			message: "john+test@example.com",
			want:    true,
		},
		{
			name:    "no email",
			message: "User user_1001 logged in",
			want:    false,
		},
		{
			name:    "at sign without domain",
			message: "Price @ $5.00",
			want:    false,
		},
		{
			name:    "hashed user id",
			message: "Processed by user_abc123def456",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContainsEmail(tt.message); got != tt.want {
				t.Errorf("ContainsEmail() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserIdentifier(t *testing.T) {
	t.Run("NewUserIdentifier", func(t *testing.T) {
		userID := NewUserIdentifier("user_1001")
		if userID.String() != "user_1001" {
			t.Errorf("NewUserIdentifier().String() = %v, want user_1001", userID.String())
		}
		if userID.HashID != "" {
			t.Errorf("NewUserIdentifier().HashID should be empty")
		}
	})

	t.Run("NewUserIdentifierFromEmail", func(t *testing.T) {
		userID := NewUserIdentifierFromEmail("john@example.com")
		if !strings.HasPrefix(userID.String(), "user_") {
			t.Errorf("NewUserIdentifierFromEmail().String() should start with user_")
		}
		if userID.ID != userID.HashID {
			t.Errorf("NewUserIdentifierFromEmail() ID should equal HashID")
		}
		// Verify no email in the identifier
		if ContainsEmail(userID.String()) {
			t.Errorf("UserIdentifier should not contain email")
		}
	})
}

func TestSanitizeLogMessagePreservesNonPII(t *testing.T) {
	// Verify that non-PII content is preserved
	original := "Inventory adjustment: Removed potato p1234 from inventory. Weight: 0.5kg"
	sanitized := SanitizeLogMessage(original)

	if original != sanitized {
		t.Errorf("Non-PII message was modified: got %s", sanitized)
	}
}

func TestSanitizeLogMessageFormatsCorrectly(t *testing.T) {
	// Simulate the exact log message pattern that was leaking emails
	original := "Inventory adjustment: Removed potato from inventory. Processed by user: john.smith@example.com"
	sanitized := SanitizeLogMessage(original)

	// Should contain the hashed user ID
	if !strings.Contains(sanitized, "user_") {
		t.Errorf("Sanitized message should contain user_ prefix: %s", sanitized)
	}

	// Should NOT contain the email
	if strings.Contains(sanitized, "@") {
		t.Errorf("Sanitized message should not contain email: %s", sanitized)
	}

	// Should preserve the rest of the message
	if !strings.Contains(sanitized, "Inventory adjustment: Removed potato from inventory. Processed by user: ") {
		t.Errorf("Sanitized message should preserve non-PII content: %s", sanitized)
	}
}
