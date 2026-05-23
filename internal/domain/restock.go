package domain

import "context"

type RestockItem struct {
	ProductID   string `json:"product_id"`
	ProductName string `json:"product_name"`
	Quantity    int    `json:"quantity"`
}

type RestockRequest struct {
	ID        string        `json:"id"`
	CreatedAt string        `json:"created_at"`
	CreatedBy string        `json:"created_by"`
	RiderID   int           `json:"rider_id"`
	RiderName string        `json:"rider_name,omitempty"`
	Status    string        `json:"status"` // pending, approved, rejected
	Items     []RestockItem `json:"items"`
}

type CreateRestockRequest struct {
	Items []struct {
		ProductID string `json:"product_id"`
		Quantity  int    `json:"quantity"`
	} `json:"items"`
}

type RestockRepository interface {
	CreateRequest(ctx context.Context, riderID int, creatorName string, req CreateRestockRequest) (string, error)
	ListRequests(ctx context.Context, riderID int, status string) ([]RestockRequest, error)
	UpdateRequestStatus(ctx context.Context, id string, status string, adminName string) error
	GetRequestByID(ctx context.Context, id string) (RestockRequest, error)
}
