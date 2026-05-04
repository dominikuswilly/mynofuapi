package http

import (
	"encoding/json"
	"mynofuapi/internal/domain"
	"net/http"
)

type RiderHandler struct {
	repo domain.UserRepository
}

func NewRiderHandler(repo domain.UserRepository) *RiderHandler {
	return &RiderHandler{
		repo: repo,
	}
}

// GetRiders godoc
// @Summary      Get all riders
// @Description  Retrieve a list of all riders from the rider_master table
// @Tags         riders
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200  {object}  domain.RiderResponse
// @Failure      500  {string}  string "Database error"
// @Router       /private/rider [get]
func (h *RiderHandler) GetRiders(w http.ResponseWriter, r *http.Request) {
	riders, err := h.repo.GetAllRiders(r.Context())
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := domain.RiderResponse{
		Status: "success",
		Data:   riders,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
