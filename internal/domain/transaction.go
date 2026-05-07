package domain

import "context"

type SaleItem struct {
	ProductID   string `json:"product_id"`
	ProductName string `json:"product_name"`
	QtySell     int    `json:"qty_sell"`
	AmtSell     int    `json:"amt_sell"`
}

type SaleRequest struct {
	Items         []SaleItem `json:"items"`
	PaymentMethod string     `json:"payment_method"`
	AmtPayTotal   int        `json:"amt_pay_total"`
}

type StockInitiationItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type StockInitiationRequest struct {
	RiderID string                `json:"rider_id"`
	Items   []StockInitiationItem `json:"items"`
}

type TransactionRepository interface {
	CreateSale(ctx context.Context, riderID int, creatorName string, req SaleRequest) error
	InitiateStock(ctx context.Context, adminID string, adminName string, req StockInitiationRequest) error
}

