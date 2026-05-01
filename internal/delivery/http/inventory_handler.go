package http

import (
	"encoding/json"
	"mynofuapi/internal/domain"
	"net/http"
)

type InventoryHandler struct{}

func NewInventoryHandler() *InventoryHandler {
	return &InventoryHandler{}
}

func (h *InventoryHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	// Mock data
	categories := []domain.Category{
		{ID: "cat_001", Name: "KOPI", Slug: "kopi", IconURL: nil, IsActive: true},
		{ID: "cat_002", Name: "COKELAT", Slug: "cokelat", IconURL: nil, IsActive: true},
		{ID: "cat_003", Name: "TEH", Slug: "teh", IconURL: nil, IsActive: true},
		{ID: "cat_004", Name: "SNACK", Slug: "snack", IconURL: nil, IsActive: true},
	}

	response := domain.CategoryResponse{
		Status: "success",
		Data:   categories,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
