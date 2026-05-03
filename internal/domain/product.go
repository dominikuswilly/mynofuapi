package domain

import "context"

type Product struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Category   string `json:"category"`
	AmountSell int    `json:"amount_sell"`
	Active     int    `json:"active"`
}

type ProductResponse struct {
	Status string    `json:"status"`
	Data   []Product `json:"data"`
}

type PatchProductRequest struct {
	Name       *string `json:"name"`
	AmountSell *int    `json:"amount_sell"`
}

type ProductRepository interface {
	GetAllProducts(ctx context.Context, category string) ([]Product, error)
	UpdateProduct(ctx context.Context, id string, req PatchProductRequest) error
	DeleteProduct(ctx context.Context, id string) error
}
