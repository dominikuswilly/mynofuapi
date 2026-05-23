package domain

import "context"

type WasteReport struct {
	ID          string `json:"id"`
	CreatedAt   string `json:"created_at"`
	CreatedBy   string `json:"created_by"`
	RiderID     int    `json:"rider_id"`
	RiderName   string `json:"rider_name,omitempty"`
	ProductID   string `json:"product_id"`
	ProductName string `json:"product_name"`
	Quantity    int    `json:"quantity"`
	Reason      string `json:"reason"`
}

type CreateWasteRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
	Reason    string `json:"reason"`
}

type WasteRepository interface {
	CreateReport(ctx context.Context, riderID int, creatorName string, req CreateWasteRequest) error
	ListReports(ctx context.Context, riderID int) ([]WasteReport, error)
}
