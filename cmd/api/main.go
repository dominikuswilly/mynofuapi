package main

import (
	"database/sql"
	"log"
	delivery "mynofuapi/internal/delivery/http"
	"mynofuapi/internal/repository"
	"mynofuapi/internal/usecase"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
)

func main() {
	// Database connection string
	connStr := "host=db.netbird.cloud port=5440 user=mynofu password=nofu2025 dbname=mynofudb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Check connection
	if err := db.Ping(); err != nil {
		log.Fatal("Could not connect to database:", err)
	}

	// Initialize Dependencies
	riderRepo := repository.NewUserRepository(db, "rider_master")
	adminRepo := repository.NewUserRepository(db, "admin_master")
	inventoryRepo := repository.NewInventoryRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)

	authUC := usecase.NewAuthUseCase(riderRepo)
	adminAuthUC := usecase.NewAuthUseCase(adminRepo)

	authHandler := delivery.NewAuthHandler(authUC)
	adminAuthHandler := delivery.NewAuthHandler(adminAuthUC)

	inventoryHandler := delivery.NewInventoryHandler(inventoryRepo)
	transactionHandler := delivery.NewTransactionHandler(transactionRepo)

	// Router setup
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(delivery.LoggingMiddleware)

	// Routes
	r.Route("/public", func(r chi.Router) {
		r.Post("/auth", authHandler.Login)
		r.Post("/admin/auth", adminAuthHandler.Login)
	})

	r.Route("/private", func(r chi.Router) {
		r.Use(delivery.AuthMiddleware(authUC))

		r.Get("/introspect", authHandler.Introspect)
		r.Get("/admin/introspect", adminAuthHandler.Introspect)

		r.Route("/inventory", func(r chi.Router) {
			r.Get("/categories", inventoryHandler.GetCategories)
		})

		r.Get("/inventories/{category}", inventoryHandler.GetRiderInventories)

		r.Route("/transaction", func(r chi.Router) {
			r.Post("/sales", transactionHandler.CreateSale)
		})
	})

	// Start Server
	log.Println("Starting server on :8080...")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
