package domain

import "context"

type RiderCommission struct {
	ID         string  `json:"id"`
	CreatedAt  string  `json:"created_at"`
	CreatedBy  string  `json:"created_by"`
	Date       string  `json:"date"`
	RiderID    int     `json:"rider_id"`
	AmtSales   float64 `json:"amt_sales"`
	Commission float64 `json:"commission"`
}

type WalletSummary struct {
	TotalEarnings  float64 `json:"total_earnings"`
	CurrentBalance float64 `json:"current_balance"`
}

type WalletActivity struct {
	ID        string `json:"id"`
	Type      string `json:"type"`       // transaction, request, report, bonus
	SubType   string `json:"sub_type"`   // Penjualan, Isi Ulang, Kerusakan, Bonus
	Amount    string `json:"amount"`     // e.g. "Rp 45.000", "20 Barang", "1 Kerusakan", "Rp 120.500"
	Time      string `json:"time"`       // e.g. "10:45 AM"
	DateGroup string `json:"date_group"` // "Hari Ini", "Kemarin", or date string
	Status    string `json:"status"`     // "Berhasil", "Menunggu", "Kritis", "Gagal"
}

type CommissionRepository interface {
	GetWalletSummary(ctx context.Context, riderID int) (WalletSummary, error)
	GetWalletHistory(ctx context.Context, riderID int) ([]WalletActivity, error)
}
