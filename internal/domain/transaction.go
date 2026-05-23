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

type RiderStockSummary struct {
	RiderID   int                `json:"rider_id"`
	RiderName string             `json:"rider_name"`
	StockList []RiderStockByDate `json:"stock_list"`
}

type RiderStockByDate struct {
	CreatedAt string           `json:"created_at"`
	ItemList  []RiderStockItem `json:"item_list"`
}



type RiderStockItem struct {
	ProductID       string  `json:"product_id"`
	ProductName     string  `json:"product_name"`
	ProductCategory string  `json:"product_category"`
	QtyBase         int     `json:"qty_base"`
	QtyCurrent      int     `json:"qty_current"`
	Confirmed       int     `json:"confirmed"`
	ConfirmedAt     *string `json:"confirmed_at"`
	ConfirmedBy     *string `json:"confirmed_by"`
	Closed          int     `json:"closed"`
	ClosedAt        *string `json:"closed_at"`
	ClosedBy        *string `json:"closed_by"`
	CreatedAt       string  `json:"-"`
	CreatedBy       string  `json:"created_by"`
	Status          string  `json:"status"`
}

type AdminStockResponse struct {
	Status string              `json:"status"`
	Data   []RiderStockSummary `json:"data"`
}

type ActualStockItem struct {
	ProductID   string `json:"product_id"`
	PhysicalQty int    `json:"physical_qty"`
}

type CloseSessionRequest struct {
	RiderID      int               `json:"rider_id"`
	ActualStocks []ActualStockItem `json:"actual_stocks"`
}

type TransactionRepository interface {
	CreateSale(ctx context.Context, riderID int, creatorName string, req SaleRequest) error
	InitiateStock(ctx context.Context, adminID string, adminName string, req StockInitiationRequest) error
	GetAdminStockReport(ctx context.Context, riderID string, riderName string, dateStart string, dateEnd string, status string) ([]RiderStockSummary, error)
	CloseSession(ctx context.Context, req CloseSessionRequest, adminName string) (float64, error)
}





