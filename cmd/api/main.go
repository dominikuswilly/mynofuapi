package main

import (
	"context"
	"log"
	delivery "mynofuapi/internal/delivery/http"
	"mynofuapi/internal/repository"
	"mynofuapi/internal/usecase"
	"net/http"

	_ "mynofuapi/docs"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/lib/pq"
	sqldblogger "github.com/simukti/sqldb-logger"
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

type sqlLogger struct{}

func (l *sqlLogger) Log(ctx context.Context, level sqldblogger.Level, msg string, data map[string]interface{}) {
	if msg == "QueryContext" || msg == "ExecContext" || msg == "Query" || msg == "Exec" || msg == "PrepareContext" {
		if err, ok := data["error"]; ok && err != nil {
			log.Printf("[SQL ERROR] %s | args: %v | duration: %v | error: %v", data["query"], data["args"], data["duration"], err)
		} else {
			log.Printf("[SQL] %s | args: %v | duration: %v", data["query"], data["args"], data["duration"])
		}
	}
}

func main() {
	// Database connection string
	connStr := "host=asus.netbird.cloud port=5432 user=user password=pass dbname=mynofudb sslmode=disable"
	db := sqldblogger.OpenDriver(connStr, &pq.Driver{}, &sqlLogger{})
	defer db.Close()

	// Check connection
	if err := db.Ping(); err != nil {
		log.Fatal("Could not connect to database:", err)
	}

	// Auto-initialize DB tables for Phase 1
	if err := repository.InitDBTables(db); err != nil {
		log.Fatal("Failed to initialize database tables:", err)
	}

	// Initialize Dependencies
	riderRepo := repository.NewRiderRepository(db, "rider_master")
	adminRepo := repository.NewAdminRepository(db, "admin_master")
	inventoryRepo := repository.NewInventoryRepository(db)

	transactionRepo := repository.NewTransactionRepository(db)
	productRepo := repository.NewProductRepository(db)
	restockRepo := repository.NewRestockRepository(db)
	wasteRepo := repository.NewWasteRepository(db)
	commissionRepo := repository.NewCommissionRepository(db)

	authUC := usecase.NewAuthUseCase(riderRepo)
	adminAuthUC := usecase.NewAuthUseCase(adminRepo)

	authHandler := delivery.NewAuthHandler(authUC)
	adminAuthHandler := delivery.NewAuthHandler(adminAuthUC)

	inventoryHandler := delivery.NewInventoryHandler(inventoryRepo)
	transactionHandler := delivery.NewTransactionHandler(transactionRepo)
	productHandler := delivery.NewProductHandler(productRepo)
	riderHandler := delivery.NewRiderHandler(riderRepo)
	restockHandler := delivery.NewRestockHandler(restockRepo)
	wasteHandler := delivery.NewWasteHandler(wasteRepo)
	commissionHandler := delivery.NewCommissionHandler(commissionRepo)

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
				r.Post("/confirm", inventoryHandler.ConfirmInventory)
				r.Post("/restock", restockHandler.CreateRequest) // Rider submits restock request
			})

			r.Get("/inventories", inventoryHandler.GetAllRiderInventories)
			r.Get("/inventories/{category}", inventoryHandler.GetRiderInventories)

			r.Route("/transaction", func(r chi.Router) {
				r.Post("/sales", transactionHandler.CreateSale)
				r.Post("/waste", wasteHandler.CreateReport) // Rider reports spilled/damaged stock
			})

			r.Route("/wallet", func(r chi.Router) {
				r.Get("/summary", commissionHandler.GetWalletSummary) // Rider balance summary
				r.Get("/history", commissionHandler.GetWalletHistory) // Rider balance history activity logs
			})
		})

		// Admin Routes
		r.Route("/admin", func(r chi.Router) {
			r.Use(delivery.AuthMiddleware(adminAuthUC))
			r.Get("/introspect", adminAuthHandler.Introspect)
			r.Get("/rider", riderHandler.GetRiders)
			r.Get("/rider/init-status", riderHandler.GetRidersWithStockStatus)

			r.Post("/rider", riderHandler.CreateRider)
			r.Patch("/rider/{id}", riderHandler.UpdateRiderStatus)

			r.Get("/product", productHandler.GetAdminProducts)
			r.Post("/product", productHandler.AddAdminProduct)
			r.Patch("/product/{id}", productHandler.PatchAdminProduct)
			r.Delete("/product/{id}", productHandler.DeleteAdminProduct)

			r.Get("/inventory/restock", restockHandler.ListRequests)                 // Admin lists restock requests
			r.Post("/inventory/restock/{id}/approve", restockHandler.ApproveRequest) // Admin approves restock
			r.Post("/inventory/restock/{id}/reject", restockHandler.RejectRequest)   // Admin rejects restock

			r.Get("/transaction/waste", wasteHandler.ListReports) // Admin lists waste/defect reports
		})

		r.Route("/admin/transaction", func(r chi.Router) {
			r.Use(delivery.AuthMiddleware(adminAuthUC))
			r.Get("/stock", transactionHandler.GetAdminStockReport)
			r.Post("/stock/init", transactionHandler.InitiateStock)
			r.Post("/close-session", transactionHandler.CloseSession) // Admin closes rider shift
		})
	})

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// Start Server
	log.Println("Starting server on :8080...")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
