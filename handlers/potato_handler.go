package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/williamdumont/potato-demo/models"
	"github.com/williamdumont/potato-demo/service"
	"github.com/williamdumont/potato-demo/storage"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

var potatoTracer = otel.Tracer("github.com/williamdumont/potato-demo/handlers/potato")

type PotatoHandler struct {
	service *service.PotatoService
}

func NewPotatoHandler(service *service.PotatoService) *PotatoHandler {
	return &PotatoHandler{
		service: service,
	}
}

func (h *PotatoHandler) CreatePotato(w http.ResponseWriter, r *http.Request) {
	_, span := potatoTracer.Start(r.Context(), "PotatoHandler.CreatePotato")
	defer span.End()

	var potato models.Potato
	if err := json.NewDecoder(r.Body).Decode(&potato); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid request payload")
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	span.SetAttributes(attribute.String("potato.variety", potato.Variety))
	defer r.Body.Close()

	createdPotato, err := h.service.CreatePotato(potato)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
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
		span.RecordError(err)
		status := http.StatusInternalServerError
		msg := err.Error()
		if err == storage.ErrNotFound {
			status = http.StatusNotFound
			msg = "Potato not found"
		}
		span.SetStatus(codes.Error, msg)
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

	var potato models.Potato
	if err := json.NewDecoder(r.Body).Decode(&potato); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid request payload")
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	potato.ID = id
	updatedPotato, err := h.service.UpdatePotato(id, potato)
	if err != nil {
		span.RecordError(err)
		status := http.StatusBadRequest
		msg := err.Error()
		if err == storage.ErrNotFound {
			status = http.StatusNotFound
			msg = "Potato not found"
		}
		span.SetStatus(codes.Error, msg)
		respondWithError(w, status, msg)
		return
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

	if err := h.service.DeletePotato(id); err != nil {
		span.RecordError(err)
		status := http.StatusInternalServerError
		msg := err.Error()
		if err == storage.ErrNotFound {
			status = http.StatusNotFound
			msg = "Potato not found"
		}
		span.SetStatus(codes.Error, msg)
		respondWithError(w, status, msg)
		return
	}

	span.SetStatus(codes.Ok, "potato deleted")
	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func (h *PotatoHandler) GetInventory(w http.ResponseWriter, r *http.Request) {
	_, span := potatoTracer.Start(r.Context(), "PotatoHandler.GetInventory")
	defer span.End()

	summary := h.service.GetInventorySummary()
	span.SetAttributes(
		attribute.Int("inventory.total_potatoes", summary.TotalPotatoes),
		attribute.Int("inventory.variety_count", len(summary.ByVariety)),
	)
	span.SetStatus(codes.Ok, "inventory summary retrieved")
	respondWithJSON(w, http.StatusOK, summary)
}

func (h *PotatoHandler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	_, span := potatoTracer.Start(r.Context(), "PotatoHandler.GetAnalytics")
	defer span.End()

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

	potato, err := h.service.GetPotato(id)
	if err != nil {
		span.RecordError(err)
		status := http.StatusInternalServerError
		msg := err.Error()
		if err == storage.ErrNotFound {
			status = http.StatusNotFound
			msg = "Potato not found"
		}
		span.SetStatus(codes.Error, msg)
		respondWithError(w, status, msg)
		return
	}

	freshness := h.service.CalculateFreshness(potato)
	span.SetAttributes(attribute.String("potato.freshness", freshness))
	span.SetStatus(codes.Ok, "freshness calculated")
	respondWithJSON(w, http.StatusOK, map[string]string{
		"id":        potato.ID,
		"variety":   potato.Variety,
		"freshness": freshness,
	})
}
