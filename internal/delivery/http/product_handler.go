package http

import (
	"encoding/json"
	"mynofuapi/internal/domain"
	"net/http"
)

type ProductHandler struct {
	repo domain.ProductRepository
}

func NewProductHandler(repo domain.ProductRepository) *ProductHandler {
	return &ProductHandler{
		repo: repo,
	}
}

// GetProducts godoc
// @Summary      Get all products
// @Description  Retrieve a list of all products from the product_master table
// @Tags         products
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200  {object}  domain.ProductResponse
// @Failure      500  {string}  string "Database error"
// @Router       /private/product [get]
func (h *ProductHandler) GetProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.repo.GetAllProducts(r.Context())
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := domain.ProductResponse{
		Status: "success",
		Data:   products,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
