package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/williamdumont/potato-demo/models"
	"github.com/williamdumont/potato-demo/service"
	"github.com/williamdumont/potato-demo/storage"
)

type RecipeHandler struct {
	service *service.RecipeService
}

func NewRecipeHandler(service *service.RecipeService) *RecipeHandler {
	return &RecipeHandler{
		service: service,
	}
}

func (h *RecipeHandler) CreateRecipe(w http.ResponseWriter, r *http.Request) {
	var recipe models.Recipe
	if err := json.NewDecoder(r.Body).Decode(&recipe); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	createdRecipe, err := h.service.CreateRecipe(recipe)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, createdRecipe)
}

func (h *RecipeHandler) GetRecipe(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	recipe, err := h.service.GetRecipe(id)
	if err != nil {
		if err == storage.ErrRecipeNotFound {
			respondWithError(w, http.StatusNotFound, "Recipe not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, recipe)
}

func (h *RecipeHandler) GetAllRecipes(w http.ResponseWriter, r *http.Request) {
	variety := r.URL.Query().Get("variety")

	var recipes []models.Recipe
	if variety != "" {
		recipes = h.service.GetRecipesByVariety(variety)
	} else {
		recipes = h.service.GetAllRecipes()
	}

	respondWithJSON(w, http.StatusOK, recipes)
}

func (h *RecipeHandler) RecommendRecipe(w http.ResponseWriter, r *http.Request) {
	variety := r.URL.Query().Get("variety")
	difficulty := r.URL.Query().Get("difficulty")

	if variety == "" {
		respondWithError(w, http.StatusBadRequest, "variety parameter is required")
		return
	}

	recipe, err := h.service.RecommendRecipe(variety, difficulty)
	if err != nil {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, recipe)
}

