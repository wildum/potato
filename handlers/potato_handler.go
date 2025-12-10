package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/williamdumont/potato-demo/models"
	"github.com/williamdumont/potato-demo/service"
	"github.com/williamdumont/potato-demo/storage"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	logapi "go.opentelemetry.io/otel/log"
)

var potatoTracer = otel.Tracer("github.com/williamdumont/potato-demo/handlers/potato")

type PotatoHandler struct {
	service   *service.PotatoService
	telemetry TelemetryRecorder
	obs       ObservabilityLogger
}

type ObservabilityLogger interface {
	EmitDebugLog(ctx context.Context, message string, attrs ...logapi.KeyValue)
	EmitInfoLog(ctx context.Context, message string, attrs ...logapi.KeyValue)
}

func NewPotatoHandler(service *service.PotatoService, telemetry TelemetryRecorder, obs ObservabilityLogger) *PotatoHandler {
	return &PotatoHandler{
		service:   service,
		telemetry: telemetry,
		obs:       obs,
	}
}

func (h *PotatoHandler) CreatePotato(w http.ResponseWriter, r *http.Request) {
	_, span := potatoTracer.Start(r.Context(), "PotatoHandler.CreatePotato")
	defer span.End()

	var potato models.Potato
	if err := json.NewDecoder(r.Body).Decode(&potato); err != nil {
		recordSpanError(span, err, "validation_error", "client_error", "invalid request payload")
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	span.SetAttributes(attribute.String("potato.variety", potato.Variety))
	defer r.Body.Close()

	if h.obs != nil {
		h.obs.EmitDebugLog(r.Context(), "Creating new potato", 
			logapi.String("variety", potato.Variety),
			logapi.Float64("weight", potato.Weight))
	}

	createdPotato, err := h.service.CreatePotato(potato)
	if err != nil {
		recordSpanError(span, err, "validation_error", "client_error", err.Error())
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	if h.obs != nil {
		h.obs.EmitInfoLog(r.Context(), "Potato created successfully", 
			logapi.String("potato_id", createdPotato.ID))
	}

	span.SetAttributes(attribute.String("potato.id", createdPotato.ID))
	span.SetStatus(codes.Ok, "potato created")
	respondWithJSON(w, http.StatusCreated, createdPotato)
}

func (h *PotatoHandler) GetPotato(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	_, span := potatoTracer.Start(r.Context(), "PotatoHandler.GetPotato")
	defer span.End()
	span.SetAttributes(attribute.String("potato.id", id))

	potato, err := h.service.GetPotato(id)
	if err != nil {
		status := http.StatusInternalServerError
		msg := err.Error()
		errType := "storage_error"
		errCategory := "server_error"
		if err == storage.ErrNotFound {
			status = http.StatusNotFound
			msg = "Potato not found"
			errType = "not_found"
			errCategory = "client_error"
		}
		recordSpanError(span, err, errType, errCategory, msg)
		respondWithError(w, status, msg)
		return
	}

	span.SetStatus(codes.Ok, "potato retrieved")
	respondWithJSON(w, http.StatusOK, potato)
}

func (h *PotatoHandler) GetAllPotatoes(w http.ResponseWriter, r *http.Request) {
	variety := r.URL.Query().Get("variety")

	_, span := potatoTracer.Start(r.Context(), "PotatoHandler.GetAllPotatoes")
	defer span.End()
	if variety != "" {
		span.SetAttributes(attribute.String("potato.variety", variety))
	}

	if h.obs != nil {
		if variety != "" {
			h.obs.EmitDebugLog(r.Context(), "Fetching potatoes by variety", 
				logapi.String("variety", variety))
		} else {
			h.obs.EmitDebugLog(r.Context(), "Fetching all potatoes")
		}
	}

	var potatoes []models.Potato
	if variety != "" {
		potatoes = h.service.GetPotatoesByVariety(variety)
	} else {
		potatoes = h.service.GetAllPotatoes()
	}

	span.SetAttributes(attribute.Int("potato.count", len(potatoes)))
	span.SetStatus(codes.Ok, "potato list retrieved")
	respondWithJSON(w, http.StatusOK, potatoes)
}

func (h *PotatoHandler) UpdatePotato(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	_, span := potatoTracer.Start(r.Context(), "PotatoHandler.UpdatePotato")
	defer span.End()
	span.SetAttributes(attribute.String("potato.id", id))

	if h.obs != nil {
		h.obs.EmitDebugLog(r.Context(), "Updating potato", 
			logapi.String("potato_id", id))
	}

	var potato models.Potato
	if err := json.NewDecoder(r.Body).Decode(&potato); err != nil {
		recordSpanError(span, err, "validation_error", "client_error", "invalid request payload")
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	potato.ID = id
	updatedPotato, err := h.service.UpdatePotato(id, potato)
	if err != nil {
		status := http.StatusBadRequest
		msg := err.Error()
		errType := "validation_error"
		errCategory := "client_error"
		if err == storage.ErrNotFound {
			status = http.StatusNotFound
			msg = "Potato not found"
			errType = "not_found"
		}
		recordSpanError(span, err, errType, errCategory, msg)
		respondWithError(w, status, msg)
		return
	}

	if h.obs != nil {
		h.obs.EmitInfoLog(r.Context(), "Potato updated successfully", 
			logapi.String("potato_id", id))
	}

	span.SetStatus(codes.Ok, "potato updated")
	respondWithJSON(w, http.StatusOK, updatedPotato)
}

func (h *PotatoHandler) DeletePotato(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	_, span := potatoTracer.Start(r.Context(), "PotatoHandler.DeletePotato")
	defer span.End()
	span.SetAttributes(attribute.String("potato.id", id))

	if h.obs != nil {
		h.obs.EmitDebugLog(r.Context(), "Deleting potato", 
			logapi.String("potato_id", id))
	}

	if err := h.service.DeletePotato(id); err != nil {
		status := http.StatusInternalServerError
		msg := err.Error()
		errType := "storage_error"
		errCategory := "server_error"
		if err == storage.ErrNotFound {
			status = http.StatusNotFound
			msg = "Potato not found"
			errType = "not_found"
			errCategory = "client_error"
		}
		recordSpanError(span, err, errType, errCategory, msg)
		respondWithError(w, status, msg)
		return
	}

	if h.obs != nil {
		h.obs.EmitInfoLog(r.Context(), "Potato deleted successfully", 
			logapi.String("potato_id", id))
	}

	span.SetStatus(codes.Ok, "potato deleted")
	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func (h *PotatoHandler) GetInventory(w http.ResponseWriter, r *http.Request) {
	_, span := potatoTracer.Start(r.Context(), "PotatoHandler.GetInventory")
	defer span.End()

	if h.obs != nil {
		h.obs.EmitDebugLog(r.Context(), "Processing inventory request")
	}

	summary := h.service.GetInventorySummary()
	span.SetAttributes(
		attribute.Int("inventory.total_potatoes", summary.TotalPotatoes),
		attribute.Int("inventory.variety_count", len(summary.ByVariety)),
	)
	if h.telemetry != nil {
		for _, item := range summary.ByVariety {
			h.telemetry.RecordInventory(r.Context(), item.Variety, item.TotalQuantity)
		}
	}
	span.SetStatus(codes.Ok, "inventory summary retrieved")
	respondWithJSON(w, http.StatusOK, summary)
}

func (h *PotatoHandler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	_, span := potatoTracer.Start(r.Context(), "PotatoHandler.GetAnalytics")
	defer span.End()

	if h.obs != nil {
		h.obs.EmitDebugLog(r.Context(), "Calculating analytics")
	}

	analytics := h.service.GetAnalytics()
	if analytics.MostPopularVariety != "" {
		span.SetAttributes(attribute.String("analytics.most_popular", analytics.MostPopularVariety))
	}
	span.SetStatus(codes.Ok, "analytics retrieved")
	respondWithJSON(w, http.StatusOK, analytics)
}

func (h *PotatoHandler) CheckFreshness(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	_, span := potatoTracer.Start(r.Context(), "PotatoHandler.CheckFreshness")
	defer span.End()
	span.SetAttributes(attribute.String("potato.id", id))

	if h.obs != nil {
		h.obs.EmitDebugLog(r.Context(), "Checking freshness for potato")
	}

	potato, err := h.service.GetPotato(id)
	if err != nil {
		status := http.StatusInternalServerError
		msg := err.Error()
		errType := "storage_error"
		errCategory := "server_error"
		if err == storage.ErrNotFound {
			status = http.StatusNotFound
			msg = "Potato not found"
			errType = "not_found"
			errCategory = "client_error"
		}
		recordSpanError(span, err, errType, errCategory, msg)
		respondWithError(w, status, msg)
		return
	}

	freshness := h.service.CalculateFreshness(potato)
	span.SetAttributes(attribute.String("potato.freshness", freshness))
	if h.telemetry != nil {
		h.telemetry.RecordFreshness(r.Context(), potato.Variety, freshnessScoreForStatus(freshness))
	}
	span.SetStatus(codes.Ok, "freshness calculated")
	respondWithJSON(w, http.StatusOK, map[string]string{
		"id":        potato.ID,
		"variety":   potato.Variety,
		"freshness": freshness,
	})
}

func freshnessScoreForStatus(status string) float64 {
	switch strings.ToLower(status) {
	case "fresh":
		return 1.0
	case "good":
		return 0.75
	case "fair":
		return 0.5
	case "old":
		return 0.25
	default:
		return 0.0
	}
}
