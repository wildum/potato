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

var recipeTracer = otel.Tracer("github.com/williamdumont/potato-demo/handlers/recipe")

type RecipeHandler struct {
	service   *service.RecipeService
	telemetry TelemetryRecorder
}

func NewRecipeHandler(service *service.RecipeService, telemetry TelemetryRecorder) *RecipeHandler {
	return &RecipeHandler{
		service:   service,
		telemetry: telemetry,
	}
}

func (h *RecipeHandler) CreateRecipe(w http.ResponseWriter, r *http.Request) {
	_, span := recipeTracer.Start(r.Context(), "RecipeHandler.CreateRecipe")
	defer span.End()

	var recipe models.Recipe
	if err := json.NewDecoder(r.Body).Decode(&recipe); err != nil {
		recordSpanError(span, err, "validation_error", "client_error", "invalid request payload")
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	span.SetAttributes(attribute.String("recipe.name", recipe.Name))
	defer r.Body.Close()

	createdRecipe, err := h.service.CreateRecipe(recipe)
	if err != nil {
		recordSpanError(span, err, "validation_error", "client_error", err.Error())
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	span.SetAttributes(attribute.String("recipe.id", createdRecipe.ID))
	span.SetStatus(codes.Ok, "recipe created")
	respondWithJSON(w, http.StatusCreated, createdRecipe)
}

func (h *RecipeHandler) GetRecipe(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	_, span := recipeTracer.Start(r.Context(), "RecipeHandler.GetRecipe")
	defer span.End()
	span.SetAttributes(attribute.String("recipe.id", id))

	recipe, err := h.service.GetRecipe(id)
	if err != nil {
		status := http.StatusInternalServerError
		msg := err.Error()
		errType := "storage_error"
		errCategory := "server_error"
		if err == storage.ErrRecipeNotFound {
			status = http.StatusNotFound
			msg = "Recipe not found"
			errType = "not_found"
			errCategory = "client_error"
		}
		recordSpanError(span, err, errType, errCategory, msg)
		respondWithError(w, status, msg)
		return
	}

	if h.telemetry != nil {
		h.telemetry.RecordRecipeView(r.Context(), recipe.ID, recipe.Name)
	}
	span.SetStatus(codes.Ok, "recipe retrieved")
	respondWithJSON(w, http.StatusOK, recipe)
}

func (h *RecipeHandler) GetAllRecipes(w http.ResponseWriter, r *http.Request) {
	variety := r.URL.Query().Get("variety")

	_, span := recipeTracer.Start(r.Context(), "RecipeHandler.GetAllRecipes")
	defer span.End()
	if variety != "" {
		span.SetAttributes(attribute.String("recipe.variety", variety))
	}

	var recipes []models.Recipe
	if variety != "" {
		recipes = h.service.GetRecipesByVariety(variety)
	} else {
		recipes = h.service.GetAllRecipes()
	}

	span.SetAttributes(attribute.Int("recipe.count", len(recipes)))
	span.SetStatus(codes.Ok, "recipe list retrieved")
	respondWithJSON(w, http.StatusOK, recipes)
}

func (h *RecipeHandler) RecommendRecipe(w http.ResponseWriter, r *http.Request) {
	variety := r.URL.Query().Get("variety")
	difficulty := r.URL.Query().Get("difficulty")

	_, span := recipeTracer.Start(r.Context(), "RecipeHandler.RecommendRecipe")
	defer span.End()
	span.SetAttributes(
		attribute.String("recipe.variety", variety),
		attribute.String("recipe.difficulty", difficulty),
	)

	if variety == "" {
		recordSpanError(span, nil, "validation_error", "client_error", "missing variety parameter")
		respondWithError(w, http.StatusBadRequest, "variety parameter is required")
		return
	}

	recipe, err := h.service.RecommendRecipe(variety, difficulty)
	if err != nil {
		errType := "not_found"
		errCategory := "client_error"
		if err != storage.ErrRecipeNotFound {
			errType = "storage_error"
			errCategory = "server_error"
		}
		recordSpanError(span, err, errType, errCategory, err.Error())
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	span.SetAttributes(attribute.String("recipe.id", recipe.ID))
	if h.telemetry != nil {
		h.telemetry.RecordRecipeView(r.Context(), recipe.ID, recipe.Name)
	}
	span.SetStatus(codes.Ok, "recipe recommendation ready")
	respondWithJSON(w, http.StatusOK, recipe)
}
