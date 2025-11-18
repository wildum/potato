package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/williamdumont/potato-demo/models"
	"github.com/williamdumont/potato-demo/service"
	"github.com/williamdumont/potato-demo/storage"
)

type PotatoHandler struct {
	service *service.PotatoService
}

func NewPotatoHandler(service *service.PotatoService) *PotatoHandler {
	return &PotatoHandler{
		service: service,
	}
}

func (h *PotatoHandler) CreatePotato(w http.ResponseWriter, r *http.Request) {
	var potato models.Potato
	if err := json.NewDecoder(r.Body).Decode(&potato); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	createdPotato, err := h.service.CreatePotato(potato)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, createdPotato)
}

func (h *PotatoHandler) GetPotato(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	potato, err := h.service.GetPotato(id)
	if err != nil {
		if err == storage.ErrNotFound {
			respondWithError(w, http.StatusNotFound, "Potato not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, potato)
}

func (h *PotatoHandler) GetAllPotatoes(w http.ResponseWriter, r *http.Request) {
	variety := r.URL.Query().Get("variety")

	var potatoes []models.Potato
	if variety != "" {
		potatoes = h.service.GetPotatoesByVariety(variety)
	} else {
		potatoes = h.service.GetAllPotatoes()
	}

	respondWithJSON(w, http.StatusOK, potatoes)
}

func (h *PotatoHandler) UpdatePotato(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var potato models.Potato
	if err := json.NewDecoder(r.Body).Decode(&potato); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	potato.ID = id
	updatedPotato, err := h.service.UpdatePotato(id, potato)
	if err != nil {
		if err == storage.ErrNotFound {
			respondWithError(w, http.StatusNotFound, "Potato not found")
			return
		}
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, updatedPotato)
}

func (h *PotatoHandler) DeletePotato(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.service.DeletePotato(id); err != nil {
		if err == storage.ErrNotFound {
			respondWithError(w, http.StatusNotFound, "Potato not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func (h *PotatoHandler) GetInventory(w http.ResponseWriter, r *http.Request) {
	summary := h.service.GetInventorySummary()
	respondWithJSON(w, http.StatusOK, summary)
}

func (h *PotatoHandler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	analytics := h.service.GetAnalytics()
	respondWithJSON(w, http.StatusOK, analytics)
}

func (h *PotatoHandler) CheckFreshness(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	potato, err := h.service.GetPotato(id)
	if err != nil {
		if err == storage.ErrNotFound {
			respondWithError(w, http.StatusNotFound, "Potato not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	freshness := h.service.CalculateFreshness(potato)
	respondWithJSON(w, http.StatusOK, map[string]string{
		"id":        potato.ID,
		"variety":   potato.Variety,
		"freshness": freshness,
	})
}

