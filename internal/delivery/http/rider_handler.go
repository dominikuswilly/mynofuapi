package http

import (
	"encoding/json"
	"mynofuapi/internal/domain"
	"net/http"
	"strings"
)

type RiderHandler struct {
	riderRepo domain.UserRepository
}

func NewRiderHandler(riderRepo domain.UserRepository) *RiderHandler {
	return &RiderHandler{
		riderRepo: riderRepo,
	}
}

func (h *RiderHandler) CreateRider(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name           string `json:"name"`
		Username       string `json:"username"`
		WhatsappNumber string `json:"whatsapp_number"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	rider := domain.Rider{
		Name:           req.Name,
		Username:       req.Username,
		WhatsappNumber: req.WhatsappNumber,
	}

	err := h.riderRepo.CreateRider(r.Context(), rider)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}


	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *RiderHandler) GetRiders(w http.ResponseWriter, r *http.Request) {
	riders, err := h.riderRepo.GetAllRiders(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(riders)
}

func (h *RiderHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		NewPassword string `json:"new_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.NewPassword == "" {
		http.Error(w, "New password is required", http.StatusBadRequest)
		return
	}

	// Get claims from context
	claims, ok := r.Context().Value(ClaimsKey).(domain.AuthResponse)
	if !ok {
		http.Error(w, "Unauthorized: invalid claims", http.StatusUnauthorized)
		return
	}

	err := h.riderRepo.UpdatePassword(r.Context(), claims.UserID, req.NewPassword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "password updated successfully"})
}

