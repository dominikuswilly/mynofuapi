package http

import (
	"encoding/json"
	"mynofuapi/internal/domain"
	"net/http"

	"github.com/go-chi/chi/v5"
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
// @Param        category  query     string  false  "Category"
// @Success      200  {object}  domain.ProductResponse
// @Failure      500  {string}  string "Database error"
// @Router       /private/product [get]
func (h *ProductHandler) GetProducts(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")

	products, err := h.repo.GetAllProducts(r.Context(), category)
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

// PatchProduct godoc
// @Summary      Update a product
// @Description  Update product name and/or amount_sell
// @Tags         products
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id       path      string                      true  "Product ID"
// @Param        request  body      domain.PatchProductRequest  true  "Update request"
// @Success      200      {object}  map[string]string "{"status": "success"}"
// @Failure      400      {string}  string "Invalid request body"
// @Failure      500      {string}  string "Database error"
// @Router       /private/product/{id} [patch]
func (h *ProductHandler) PatchProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req domain.PatchProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.repo.UpdateProduct(r.Context(), id, req)
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"status": "success",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
