package http

import (
	"encoding/json"
	"mynofuapi/internal/domain"
	"net/http"
	"strconv"
)

type WasteHandler struct {
	repo domain.WasteRepository
}

func NewWasteHandler(repo domain.WasteRepository) *WasteHandler {
	return &WasteHandler{repo: repo}
}

func (h *WasteHandler) CreateReport(w http.ResponseWriter, r *http.Request) {
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

	var req domain.CreateWasteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ProductID == "" || req.Quantity <= 0 || req.Reason == "" {
		http.Error(w, "product_id, positive quantity, and reason are required", http.StatusBadRequest)
		return
	}

	err = h.repo.CreateReport(r.Context(), riderID, claims.Name, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *WasteHandler) ListReports(w http.ResponseWriter, r *http.Request) {
	// Query params (optional filtering by rider)
	riderIDStr := r.URL.Query().Get("rider_id")

	riderID := 0
	if riderIDStr != "" {
		var err error
		riderID, err = strconv.Atoi(riderIDStr)
		if err != nil {
			http.Error(w, "Invalid rider_id format", http.StatusBadRequest)
			return
		}
	}

	reports, err := h.repo.ListReports(r.Context(), riderID)
	if err != nil {
		http.Error(w, "Error fetching waste reports: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   reports,
	})
}
