package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/williamdumont/potato-demo/background"
	"github.com/williamdumont/potato-demo/handlers"
	"github.com/williamdumont/potato-demo/seed"
	"github.com/williamdumont/potato-demo/service"
	"github.com/williamdumont/potato-demo/storage"
)

const httpAddr = ":8081"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	telemetry, err := initOpenTelemetry(ctx)
	if err != nil {
		log.Fatalf("failed to initialize OpenTelemetry: %v", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := telemetry.Shutdown(shutdownCtx); err != nil {
			log.Printf("failed to shutdown OpenTelemetry: %v", err)
		}
	}()

	store := storage.NewInMemoryStorage()
	seedData(store)

	worker := background.NewWorker(store, telemetry.Logger())
	worker.StartPotatoGenerator(3 * time.Second)
	worker.StartRecipeGenerator(8 * time.Second)
	worker.StartQualityDegradation(20 * time.Second)
	worker.StartPotatoRemover(10 * time.Second)

	potatoService := service.NewPotatoService(store)
	recipeService := service.NewRecipeService(store)

	potatoHandler := handlers.NewPotatoHandler(potatoService, telemetry)
	recipeHandler := handlers.NewRecipeHandler(recipeService, telemetry)

	r := mux.NewRouter()
	api := r.PathPrefix("/api/v1").Subrouter()

	api.Handle("/potatoes", telemetry.WrapHandler("GET /potatoes", potatoHandler.GetAllPotatoes)).Methods("GET")
	api.Handle("/potatoes", telemetry.WrapHandler("POST /potatoes", potatoHandler.CreatePotato)).Methods("POST")
	api.Handle("/potatoes/{id}", telemetry.WrapHandler("GET /potatoes/{id}", potatoHandler.GetPotato)).Methods("GET")
	api.Handle("/potatoes/{id}", telemetry.WrapHandler("PUT /potatoes/{id}", potatoHandler.UpdatePotato)).Methods("PUT")
	api.Handle("/potatoes/{id}", telemetry.WrapHandler("DELETE /potatoes/{id}", potatoHandler.DeletePotato)).Methods("DELETE")
	api.Handle("/potatoes/{id}/freshness", telemetry.WrapHandler("GET /potatoes/{id}/freshness", potatoHandler.CheckFreshness)).Methods("GET")

	api.Handle("/inventory", telemetry.WrapHandler("GET /inventory", potatoHandler.GetInventory)).Methods("GET")
	api.Handle("/analytics", telemetry.WrapHandler("GET /analytics", potatoHandler.GetAnalytics)).Methods("GET")

	api.Handle("/recipes", telemetry.WrapHandler("GET /recipes", recipeHandler.GetAllRecipes)).Methods("GET")
	api.Handle("/recipes", telemetry.WrapHandler("POST /recipes", recipeHandler.CreateRecipe)).Methods("POST")
	api.Handle("/recipes/{id}", telemetry.WrapHandler("GET /recipes/{id}", recipeHandler.GetRecipe)).Methods("GET")
	api.Handle("/recipes/recommend", telemetry.WrapHandler("GET /recipes/recommend", recipeHandler.RecommendRecipe)).Methods("GET")

	api.Handle("/health", telemetry.WrapHandler("GET /health", healthCheck)).Methods("GET")

	server := &http.Server{
		Addr:    httpAddr,
		Handler: r,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		}
	}()

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("HTTP server error: %v", err)
	}
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"potato-service"}`))
}

func seedData(store storage.Storage) {
	seed.LoadSampleData(store)
}
