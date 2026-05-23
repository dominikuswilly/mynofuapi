package http

import (
	"encoding/json"
	"mynofuapi/internal/domain"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type RestockHandler struct {
	repo domain.RestockRepository
}

func NewRestockHandler(repo domain.RestockRepository) *RestockHandler {
	return &RestockHandler{repo: repo}
}

func (h *RestockHandler) CreateRequest(w http.ResponseWriter, r *http.Request) {
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

	var req domain.CreateRestockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Items) == 0 {
		http.Error(w, "No items in restock request", http.StatusBadRequest)
		return
	}

	requestID, err := h.repo.CreateRequest(r.Context(), riderID, claims.Name, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":     "success",
		"request_id": requestID,
	})
}

func (h *RestockHandler) ListRequests(w http.ResponseWriter, r *http.Request) {
	// Query params
	status := r.URL.Query().Get("status")
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

	requests, err := h.repo.ListRequests(r.Context(), riderID, status)
	if err != nil {
		http.Error(w, "Error fetching requests: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   requests,
	})
}

func (h *RestockHandler) ApproveRequest(w http.ResponseWriter, r *http.Request) {
	// Get Admin info from context
	claims, ok := r.Context().Value(ClaimsKey).(domain.AuthResponse)
	if !ok || claims.UserID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	requestID := chi.URLParam(r, "id")
	if requestID == "" {
		http.Error(w, "Missing request ID", http.StatusBadRequest)
		return
	}

	err := h.repo.UpdateRequestStatus(r.Context(), requestID, "approved", claims.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *RestockHandler) RejectRequest(w http.ResponseWriter, r *http.Request) {
	// Get Admin info from context
	claims, ok := r.Context().Value(ClaimsKey).(domain.AuthResponse)
	if !ok || claims.UserID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	requestID := chi.URLParam(r, "id")
	if requestID == "" {
		http.Error(w, "Missing request ID", http.StatusBadRequest)
		return
	}

	err := h.repo.UpdateRequestStatus(r.Context(), requestID, "rejected", claims.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
