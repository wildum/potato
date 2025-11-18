package main

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/williamdumont/potato-demo/background"
	"github.com/williamdumont/potato-demo/handlers"
	"github.com/williamdumont/potato-demo/seed"
	"github.com/williamdumont/potato-demo/service"
	"github.com/williamdumont/potato-demo/storage"
)

func main() {
	store := storage.NewInMemoryStorage()

	seedData(store)

	worker := background.NewWorker(store)
	worker.StartPotatoGenerator(3 * time.Second)
	worker.StartRecipeGenerator(8 * time.Second)
	worker.StartQualityDegradation(20 * time.Second)

	potatoService := service.NewPotatoService(store)
	recipeService := service.NewRecipeService(store)

	potatoHandler := handlers.NewPotatoHandler(potatoService)
	recipeHandler := handlers.NewRecipeHandler(recipeService)

	r := mux.NewRouter()

	api := r.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/potatoes", potatoHandler.GetAllPotatoes).Methods("GET")
	api.HandleFunc("/potatoes", potatoHandler.CreatePotato).Methods("POST")
	api.HandleFunc("/potatoes/{id}", potatoHandler.GetPotato).Methods("GET")
	api.HandleFunc("/potatoes/{id}", potatoHandler.UpdatePotato).Methods("PUT")
	api.HandleFunc("/potatoes/{id}", potatoHandler.DeletePotato).Methods("DELETE")
	api.HandleFunc("/potatoes/{id}/freshness", potatoHandler.CheckFreshness).Methods("GET")

	api.HandleFunc("/inventory", potatoHandler.GetInventory).Methods("GET")
	api.HandleFunc("/analytics", potatoHandler.GetAnalytics).Methods("GET")

	api.HandleFunc("/recipes", recipeHandler.GetAllRecipes).Methods("GET")
	api.HandleFunc("/recipes", recipeHandler.CreateRecipe).Methods("POST")
	api.HandleFunc("/recipes/{id}", recipeHandler.GetRecipe).Methods("GET")
	api.HandleFunc("/recipes/recommend", recipeHandler.RecommendRecipe).Methods("GET")

	api.HandleFunc("/health", healthCheck).Methods("GET")

	http.ListenAndServe(":8081", r)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"potato-service"}`))
}

func seedData(store storage.Storage) {
	seed.LoadSampleData(store)
}
