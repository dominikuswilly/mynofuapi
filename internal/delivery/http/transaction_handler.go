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
