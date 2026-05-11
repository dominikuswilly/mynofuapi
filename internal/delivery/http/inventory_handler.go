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

// GetRiderInventories godoc
// @Summary      Get rider inventories
// @Description  Retrieve a list of inventories for the authenticated rider by category
// @Tags         inventory
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        category  path      string  true  "Category"
// @Success      200       {object}  domain.RiderInventoryResponse
// @Failure      401       {string}  string "Unauthorized"
// @Failure      500       {string}  string "Database error"
// @Router       /private/inventories/{category} [get]
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

// GetCategories godoc
// @Summary      Get inventory categories
// @Description  Retrieve a list of all inventory categories
// @Tags         inventory
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200       {object}  domain.CategoryResponse
// @Router       /private/inventory/categories [get]
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

func (h *InventoryHandler) GetAllRiderInventories(w http.ResponseWriter, r *http.Request) {
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

	inventories, err := h.repo.GetAllRiderInventory(r.Context(), riderID)
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

func (h *InventoryHandler) CheckConfirmation(w http.ResponseWriter, r *http.Request) {
	// Get UserID from context
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

	isConfirmed, items, err := h.repo.CheckInventoryConfirmation(r.Context(), riderID)
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := domain.InventoryConfirmationResponse{
		Status:      "success",
		IsConfirmed: isConfirmed,
		Items:       items,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

