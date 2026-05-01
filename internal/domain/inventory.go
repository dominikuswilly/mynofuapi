package domain

import "context"

type Category struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	IconURL  *string `json:"icon_url"`
	IsActive bool   `json:"is_active"`
}

type CategoryResponse struct {
	Status      string     `json:"status"`
	Data        []Category `json:"data"`
	AccessToken *string    `json:"access_token"`
}

type RiderInventory struct {
	ProductID   string `json:"product_id"`
	ProductName string `json:"product_name"`
	AmountSell  string `json:"amount_sell"`
	QtyBase     int    `json:"qty_base"`
	QtyCurrent  int    `json:"qty_current"`
}

type RiderInventoryResponse struct {
	Status      string           `json:"status"`
	Data        []RiderInventory `json:"data"`
	AccessToken *string          `json:"access_token"`
}

type InventoryRepository interface {
	GetRiderInventory(ctx context.Context, riderID int, category string) ([]RiderInventory, error)
}
