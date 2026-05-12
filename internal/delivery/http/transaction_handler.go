package http

import (
	"encoding/json"
	"mynofuapi/internal/domain"
	"net/http"
	"strconv"
)

type TransactionHandler struct {
	repo domain.TransactionRepository
}

func NewTransactionHandler(repo domain.TransactionRepository) *TransactionHandler {
	return &TransactionHandler{
		repo: repo,
	}
}

// CreateSale godoc
// @Summary      Create a sale
// @Description  Create a new sale transaction for the authenticated rider
// @Tags         transaction
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        request  body      domain.SaleRequest  true  "Sale request"
// @Success      200      {object}  map[string]string "{"status": "success"}"
// @Failure      401      {string}  string "Unauthorized"
// @Failure      500      {string}  string "Error message"
// @Router       /private/transaction/sales [post]
func (h *TransactionHandler) CreateSale(w http.ResponseWriter, r *http.Request) {
	// Get UserID from context
	claims, ok := r.Context().Value(ClaimsKey).(domain.AuthResponse)
	if !ok || claims.UserID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	riderID, err := strconv.Atoi(claims.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusInternalServerError)
		return
	}

	var req domain.SaleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.repo.CreateSale(r.Context(), riderID, claims.Name, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *TransactionHandler) InitiateStock(w http.ResponseWriter, r *http.Request) {
	// Get Admin info from context
	claims, ok := r.Context().Value(ClaimsKey).(domain.AuthResponse)
	if !ok || claims.UserID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req domain.StockInitiationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.repo.InitiateStock(r.Context(), claims.UserID, claims.Name, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *TransactionHandler) GetAdminStockReport(w http.ResponseWriter, r *http.Request) {
	report, err := h.repo.GetAdminStockReport(r.Context())
	if err != nil {
		http.Error(w, "Error generating report: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := domain.AdminStockResponse{
		Status: "success",
		Data:   report,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}


