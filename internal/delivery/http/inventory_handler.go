package http

import (
	"encoding/json"
	"mynofuapi/internal/domain"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type InventoryHandler struct {
	repo domain.InventoryRepository
}

func NewInventoryHandler(repo domain.InventoryRepository) *InventoryHandler {
	return &InventoryHandler{
		repo: repo,
	}
}

func (h *InventoryHandler) GetRiderInventories(w http.ResponseWriter, r *http.Request) {
	category := chi.URLParam(r, "category")

	// Get UserID and Token from context (stored by AuthMiddleware)
	claims, ok := r.Context().Value(ClaimsKey).(domain.AuthResponse)
	if !ok || claims.UserID == "" {
		http.Error(w, "Unauthorized: missing user info", http.StatusUnauthorized)
		return
	}

	riderID, err := strconv.Atoi(claims.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID format", http.StatusInternalServerError)
		return
	}

	inventories, err := h.repo.GetRiderInventory(r.Context(), riderID, category)
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := domain.RiderInventoryResponse{
		Status: "success",
		Data:   inventories,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *InventoryHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	// Get Token from context
	claims, _ := r.Context().Value(ClaimsKey).(domain.AuthResponse)

	var accessToken *string
	if claims.AccessToken != nil {
		accessToken = claims.AccessToken
	}

	// Mock data
	categories := []domain.Category{
		{ID: "cat_001", Name: "KOPI", Slug: "kopi", IconURL: nil, IsActive: true},
		{ID: "cat_002", Name: "COKELAT", Slug: "cokelat", IconURL: nil, IsActive: true},
		{ID: "cat_003", Name: "TEH", Slug: "teh", IconURL: nil, IsActive: true},
		{ID: "cat_004", Name: "SNACK", Slug: "snack", IconURL: nil, IsActive: true},
	}

	response := domain.CategoryResponse{
		Status:      "success",
		Data:        categories,
		AccessToken: accessToken,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
