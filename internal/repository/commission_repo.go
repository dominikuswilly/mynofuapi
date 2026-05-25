package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"mynofuapi/internal/domain"
	"sort"
	"time"
)

type commissionRepo struct {
	db *sql.DB
}

func NewCommissionRepository(db *sql.DB) domain.CommissionRepository {
	return &commissionRepo{db: db}
}

func (r *commissionRepo) GetWalletSummary(ctx context.Context, riderID int) (domain.WalletSummary, error) {
	query := `
		SELECT COALESCE(SUM(i_amt_commission), 0)
		FROM rider_commissions
		WHERE i_rider_id = $1
	`
	var total float64
	err := r.db.QueryRowContext(ctx, query, riderID).Scan(&total)
	if err != nil {
		log.Printf("Error fetching commission sum: %v", err)
		return domain.WalletSummary{}, err
	}

	return domain.WalletSummary{
		TotalEarnings:  total,
		CurrentBalance: total, // Matches earnings since no withdrawals are implemented yet
	}, nil
}

func (r *commissionRepo) GetWalletHistory(ctx context.Context, riderID int) ([]domain.WalletActivity, error) {
	var activities []domain.WalletActivity

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.FixedZone("WIB", 7*3600)
	}
	todayStr := time.Now().In(loc).Format("2006-01-02")
	yesterdayStr := time.Now().In(loc).AddDate(0, 0, -1).Format("2006-01-02")

	// Helper to format date groups
	getDateGroup := func(dateStr string) string {
		if dateStr == todayStr {
			return "Hari Ini"
		} else if dateStr == yesterdayStr {
			return "Kemarin"
		}
		// Parse and format to e.g. "22 Mei 2026"
		t, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return dateStr
		}
		months := []string{"", "Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}
		return fmt.Sprintf("%d %s %d", t.Day(), months[t.Month()], t.Year())
	}

	// Helper to format transaction time
	formatTimeStr := func(t time.Time) string {
		return t.In(loc).Format("03:04 PM")
	}

	// Helper to sort activities by timestamp descending
	type sortableActivity struct {
		activity  domain.WalletActivity
		timestamp time.Time
	}
	var list []sortableActivity

	// 1. Fetch Sales Transactions
	salesQuery := `
		SELECT 
			c_id, 
			i_amt_pay_total, 
			ts_created_at
		FROM sales_master
		WHERE i_rider_id = $1
	`
	rows, err := r.db.QueryContext(ctx, salesQuery, riderID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id string
			var amt float64
			var ts time.Time
			if err := rows.Scan(&id, &amt, &ts); err == nil {
				dateStr := ts.In(loc).Format("2006-01-02")
				list = append(list, sortableActivity{
					timestamp: ts,
					activity: domain.WalletActivity{
						ID:        id,
						Type:      "transaction",
						SubType:   "Penjualan",
						Amount:    fmt.Sprintf("Rp %.0f", amt),
						Time:      formatTimeStr(ts),
						DateGroup: getDateGroup(dateStr),
						Status:    "Berhasil",
					},
				})
			}
		}
	}

	// 2. Fetch Restock Requests
	restocksQuery := `
		SELECT 
			r.c_id, 
			r.c_status, 
			r.ts_created_at,
			COALESCE(SUM(ri.i_qty), 0) as total_qty
		FROM rider_restock_requests r
		LEFT JOIN rider_restock_items ri ON r.c_id = ri.c_request_id
		WHERE r.i_rider_id = $1
		GROUP BY r.c_id, r.c_status, r.ts_created_at
	`
	rows, err = r.db.QueryContext(ctx, restocksQuery, riderID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id, status string
			var ts time.Time
			var qty int
			if err := rows.Scan(&id, &status, &ts, &qty); err == nil {
				statusMapped := "Menunggu"
				if status == "approved" {
					statusMapped = "Berhasil"
				} else if status == "rejected" {
					statusMapped = "Gagal"
				}

				dateStr := ts.In(loc).Format("2006-01-02")
				list = append(list, sortableActivity{
					timestamp: ts,
					activity: domain.WalletActivity{
						ID:        id,
						Type:      "request",
						SubType:   "Isi Ulang",
						Amount:    fmt.Sprintf("%d Barang", qty),
						Time:      formatTimeStr(ts),
						DateGroup: getDateGroup(dateStr),
						Status:    statusMapped,
					},
				})
			}
		}
	}

	// 3. Fetch Defects (Waste reports)
	defectsQuery := `
		SELECT 
			c_id, 
			i_qty, 
			ts_created_at
		FROM rider_waste_reports
		WHERE i_rider_id = $1
	`
	rows, err = r.db.QueryContext(ctx, defectsQuery, riderID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id string
			var qty int
			var ts time.Time
			if err := rows.Scan(&id, &qty, &ts); err == nil {
				dateStr := ts.In(loc).Format("2006-01-02")
				list = append(list, sortableActivity{
					timestamp: ts,
					activity: domain.WalletActivity{
						ID:        id,
						Type:      "report",
						SubType:   "Kerusakan",
						Amount:    fmt.Sprintf("%d Kerusakan", qty),
						Time:      formatTimeStr(ts),
						DateGroup: getDateGroup(dateStr),
						Status:    "Kritis",
					},
				})
			}
		}
	}

	// 4. Fetch Commissions (Bonus)
	commissionsQuery := `
		SELECT 
			c_id, 
			i_amt_commission, 
			ts_created_at
		FROM rider_commissions
		WHERE i_rider_id = $1
	`
	rows, err = r.db.QueryContext(ctx, commissionsQuery, riderID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id string
			var commission float64
			var ts time.Time
			if err := rows.Scan(&id, &commission, &ts); err == nil {
				dateStr := ts.In(loc).Format("2006-01-02")
				list = append(list, sortableActivity{
					timestamp: ts,
					activity: domain.WalletActivity{
						ID:        id,
						Type:      "transaction",
						SubType:   "Bonus",
						Amount:    fmt.Sprintf("Rp %.0f", commission),
						Time:      formatTimeStr(ts),
						DateGroup: getDateGroup(dateStr),
						Status:    "Berhasil",
					},
				})
			}
		}
	}

	// Sort activities by timestamp descending
	sort.Slice(list, func(i, j int) bool {
		return list[i].timestamp.After(list[j].timestamp)
	})

	for _, item := range list {
		activities = append(activities, item.activity)
	}

	if activities == nil {
		activities = []domain.WalletActivity{}
	}

	return activities, nil
}
