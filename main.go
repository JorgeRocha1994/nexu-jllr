package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"nexu-jllr/pkg/db"
	"nexu-jllr/pkg/handler"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}

	db.InitDB()
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Rutas
	r.Get("/brands", handler.GetAllBrands)
	r.Get("/brands/{id}/models", handler.GetModelsByBrandID)
	r.Post("/brands", handler.CreateBrand)
	r.Post("/brands/{id}/models", handler.CreateModelByBrandID)
	r.Get("/models", handler.GetAllModels)
	r.Put("/models/{id}", handler.UpdateModel)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		log.Println("Server running on port 8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
}
