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

type ProductRepository interface {
	GetAllProducts(ctx context.Context) ([]Product, error)
}
