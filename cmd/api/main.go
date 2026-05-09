package main

import (
	"database/sql"
	"log"
	delivery "mynofuapi/internal/delivery/http"
	"mynofuapi/internal/repository"
	"mynofuapi/internal/usecase"
	"net/http"

	_ "mynofuapi/docs"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title           MyNofu API
// @version         1.0
// @description     API for MyNofu application.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization

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
	riderRepo := repository.NewRiderRepository(db, "rider_master")
	adminRepo := repository.NewAdminRepository(db, "admin_master")
	inventoryRepo := repository.NewInventoryRepository(db)

	transactionRepo := repository.NewTransactionRepository(db)
	productRepo := repository.NewProductRepository(db)

	authUC := usecase.NewAuthUseCase(riderRepo)
	adminAuthUC := usecase.NewAuthUseCase(adminRepo)

	authHandler := delivery.NewAuthHandler(authUC)
	adminAuthHandler := delivery.NewAuthHandler(adminAuthUC)

	inventoryHandler := delivery.NewInventoryHandler(inventoryRepo)
	transactionHandler := delivery.NewTransactionHandler(transactionRepo)
	productHandler := delivery.NewProductHandler(productRepo)
	riderHandler := delivery.NewRiderHandler(riderRepo)

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
		// Rider Routes
		r.Group(func(r chi.Router) {
			r.Use(delivery.AuthMiddleware(authUC))
			r.Get("/introspect", authHandler.Introspect)
			r.Get("/product", productHandler.GetProducts)
			r.Patch("/change-password", riderHandler.ChangePassword)


			r.Route("/inventory", func(r chi.Router) {
				r.Get("/categories", inventoryHandler.GetCategories)
				r.Get("/check-confirmation", inventoryHandler.CheckConfirmation)
			})

			r.Get("/inventories", inventoryHandler.GetAllRiderInventories)
			r.Get("/inventories/{category}", inventoryHandler.GetRiderInventories)

			r.Route("/transaction", func(r chi.Router) {
				r.Post("/sales", transactionHandler.CreateSale)
			})
		})

		// Admin Routes
		r.Route("/admin", func(r chi.Router) {
			r.Use(delivery.AuthMiddleware(adminAuthUC))
			r.Get("/introspect", adminAuthHandler.Introspect)
			r.Get("/rider", riderHandler.GetRiders)
			r.Post("/rider", riderHandler.CreateRider)
			r.Patch("/rider/{id}", riderHandler.UpdateRiderStatus)

			r.Get("/product", productHandler.GetAdminProducts)
			r.Post("/product", productHandler.AddAdminProduct)
			r.Patch("/product/{id}", productHandler.PatchAdminProduct)
			r.Delete("/product/{id}", productHandler.DeleteAdminProduct)
		})

		r.Route("/admin/transaction", func(r chi.Router) {
			r.Use(delivery.AuthMiddleware(adminAuthUC))
			r.Post("/stock/init", transactionHandler.InitiateStock)
		})
	})


	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// Start Server
	log.Println("Starting server on :8080...")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
