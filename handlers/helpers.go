package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// RedactEmail redacts an email address for logging purposes to prevent PII leaks.
// It returns a pseudonymized identifier using a hash of the email prefix.
// Example: "john.smith@example.com" -> "customer_a1b2c3d4@example.com"
func RedactEmail(email string) string {
	if email == "" {
		return ""
	}

	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 {
		// Not a valid email format, return a fully redacted placeholder
		return "***@invalid"
	}

	// Hash the local part (before @) and take first 8 characters
	hash := sha256.Sum256([]byte(parts[0]))
	hashStr := hex.EncodeToString(hash[:])[:8]

	return "customer_" + hashStr + "@" + parts[1]
}

// TelemetryRecorder captures business metrics without creating import cycles.
type TelemetryRecorder interface {
	RecordInventory(ctx context.Context, variety string, count int)
	RecordFreshness(ctx context.Context, variety string, freshness float64)
	RecordRecipeView(ctx context.Context, recipeID, recipeName string)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func recordSpanError(span trace.Span, err error, errType, errCategory, message string) {
	if err != nil {
		span.RecordError(err)
	}
	span.SetAttributes(
		attribute.String("error.type", errType),
		attribute.String("error.category", errCategory),
	)
	span.SetStatus(codes.Error, message)
}
