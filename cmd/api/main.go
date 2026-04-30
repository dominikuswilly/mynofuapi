package main

import (
	"log"
	delivery "mynofuapi/internal/delivery/http"
	"mynofuapi/internal/repository"
	"mynofuapi/internal/usecase"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// Initialize Dependencies
	userRepo := repository.NewMockUserRepository()
	authUC := usecase.NewAuthUseCase(userRepo)
	authHandler := delivery.NewAuthHandler(authUC)

	// Router setup
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Routes
	r.Route("/public", func(r chi.Router) {
		r.Post("/auth", authHandler.Login)
	})

	// Start Server
	log.Println("Starting server on :8080...")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
