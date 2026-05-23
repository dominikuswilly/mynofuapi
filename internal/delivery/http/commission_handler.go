package http

import (
	"encoding/json"
	"mynofuapi/internal/domain"
	"net/http"
	"strconv"
)

type CommissionHandler struct {
	repo domain.CommissionRepository
}

func NewCommissionHandler(repo domain.CommissionRepository) *CommissionHandler {
	return &CommissionHandler{repo: repo}
}

func (h *CommissionHandler) GetWalletSummary(w http.ResponseWriter, r *http.Request) {
	// Get Rider info from context
	claims, ok := r.Context().Value(ClaimsKey).(domain.AuthResponse)
	if !ok || claims.UserID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	riderID, err := strconv.Atoi(claims.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID format", http.StatusInternalServerError)
		return
	}

	summary, err := h.repo.GetWalletSummary(r.Context(), riderID)
	if err != nil {
		http.Error(w, "Error fetching wallet summary: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":          "success",
		"total_earnings":  summary.TotalEarnings,
		"current_balance": summary.CurrentBalance,
	})
}

func (h *CommissionHandler) GetWalletHistory(w http.ResponseWriter, r *http.Request) {
	// Get Rider info from context
	claims, ok := r.Context().Value(ClaimsKey).(domain.AuthResponse)
	if !ok || claims.UserID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	riderID, err := strconv.Atoi(claims.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID format", http.StatusInternalServerError)
		return
	}

	history, err := h.repo.GetWalletHistory(r.Context(), riderID)
	if err != nil {
		http.Error(w, "Error fetching wallet history: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   history,
	})
}
