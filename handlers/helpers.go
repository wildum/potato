package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

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
