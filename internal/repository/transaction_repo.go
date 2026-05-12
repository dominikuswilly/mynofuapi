package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"mynofuapi/internal/domain"
)

type transactionRepo struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) domain.TransactionRepository {
	return &transactionRepo{
		db: db,
	}
}

func (r *transactionRepo) CreateSale(ctx context.Context, riderID int, creatorName string, req domain.SaleRequest) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Insert into sales_master
	var salesID string
	masterQuery := `
		INSERT INTO sales_master (c_id, i_amt_pay_total, i_item_total, c_payment_method, i_rider_id, c_created_by, ts_created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, NOW())
		RETURNING c_id
	`
	err = tx.QueryRowContext(ctx, masterQuery, req.AmtPayTotal, len(req.Items), req.PaymentMethod, riderID, creatorName).Scan(&salesID)
	if err != nil {
		log.Printf("Error inserting sales_master: %v", err)
		return fmt.Errorf("failed to create sales master: %w", err)
	}

	// 2. Insert into sales_detail and update inventory
	for _, item := range req.Items {
		subtotal := item.QtySell * item.AmtSell
		detailQuery := `
			INSERT INTO sales_detail (c_id, c_sales_id, c_product_id, i_qty, i_amt_sell, i_amt_subtotal)
			VALUES (gen_random_uuid(), $1, $2, $3, $4, $5)
		`
		_, err = tx.ExecContext(ctx, detailQuery, salesID, item.ProductID, item.QtySell, item.AmtSell, subtotal)
		if err != nil {
			log.Printf("Error inserting sales_detail: %v", err)
			return fmt.Errorf("failed to create sales detail for product %s: %w", item.ProductID, err)
		}

		// Update stock
		inventoryQuery := `
			UPDATE rider_inventory
			SET i_qty_current = (i_qty_current::integer - $1::integer)::text
			WHERE i_rider_id = $2 
			AND c_product_id = $3 
			AND ts_created_at::date = CURRENT_DATE
		`
		result, err := tx.ExecContext(ctx, inventoryQuery, item.QtySell, riderID, item.ProductID)
		if err != nil {
			log.Printf("Error updating inventory: %v", err)
			return fmt.Errorf("failed to update inventory for product %s: %w", item.ProductID, err)
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return fmt.Errorf("insufficient stock or product %s not found in inventory", item.ProductID)
		}
	}

	return tx.Commit()
}

func (r *transactionRepo) InitiateStock(ctx context.Context, adminID string, adminName string, req domain.StockInitiationRequest) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Process each item and insert into rider_inventory
	for _, item := range req.Items {
		// Fetch product details
		var productNm, category string
		var amtSell float64
		productQuery := `
			SELECT c_nm, c_category, i_amt_sell 
			FROM product_master 
			WHERE c_id = $1
		`
		err = tx.QueryRowContext(ctx, productQuery, item.ProductID).Scan(&productNm, &category, &amtSell)
		if err != nil {
			log.Printf("Error fetching product details for %s: %v", item.ProductID, err)
			return fmt.Errorf("product %s not found: %w", item.ProductID, err)
		}

		// Insert into rider_inventory
		inventoryQuery := `
			INSERT INTO rider_inventory (
				c_id, c_created_by, ts_created_at, c_product_id, 
				i_qty_base, i_qty_current, i_rider_id, 
				c_category, c_product_nm, i_amt_sell
			)
			VALUES (gen_random_uuid(), $1, NOW(), $2, $3, $4, $5, $6, $7, $8)
		`
		// Convert quantity to string (text) and riderID to int as per table schema
		_, err = tx.ExecContext(ctx, inventoryQuery,
			adminName,
			item.ProductID,
			fmt.Sprintf("%d", item.Quantity),
			fmt.Sprintf("%d", item.Quantity),
			req.RiderID,
			category,
			productNm,
			amtSell,
		)

		if err != nil {
			log.Printf("Error inserting into rider_inventory for %s: %v", item.ProductID, err)
			return fmt.Errorf("failed to initiate stock for product %s: %w", item.ProductID, err)
		}
	}

	return tx.Commit()
}

func (r *transactionRepo) GetAdminStockReport(ctx context.Context) ([]domain.RiderStockSummary, error) {
	query := `
		SELECT 
			inv.i_rider_id, 
			rm.c_nm as rider_name,
			inv.c_product_id, 
			inv.c_product_nm, 
			inv.c_category, 
			inv.i_qty_base, 
			inv.i_qty_current, 
			COALESCE(inv.i_confirmed, 0), 
			to_char(inv.ts_confirmed_at AT TIME ZONE 'Asia/Jakarta', 'YYYY-MM-DD HH24:MI:SS'), 
			inv.c_confirmed_by, 
			to_char(inv.ts_created_at AT TIME ZONE 'Asia/Jakarta', 'YYYY-MM-DD HH24:MI:SS'), 
			inv.c_created_by
		FROM rider_inventory inv
		LEFT JOIN rider_master rm ON inv.i_rider_id = rm.i_id
		WHERE (inv.ts_created_at AT TIME ZONE 'Asia/Jakarta')::date = (CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Jakarta')::date
		ORDER BY inv.i_rider_id, inv.c_product_nm
	`


	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("Error querying admin stock report: %v", err)
		return nil, err
	}
	defer rows.Close()

	summaryMap := make(map[int]*domain.RiderStockSummary)
	var riderIDs []int

	for rows.Next() {
		var riderID int
		var riderName string
		var item domain.RiderStockItem
		var confirmedAt, confirmedBy, createdAt, createdBy sql.NullString

		err := rows.Scan(
			&riderID,
			&riderName,
			&item.ProductID,
			&item.ProductName,
			&item.ProductCategory,
			&item.QtyBase,
			&item.QtyCurrent,
			&item.Confirmed,
			&confirmedAt,
			&confirmedBy,
			&createdAt,
			&createdBy,
		)
		if err != nil {
			log.Printf("Error scanning stock report row: %v", err)
			continue
		}

		if confirmedAt.Valid {
			item.ConfirmedAt = &confirmedAt.String
		}
		if confirmedBy.Valid {
			item.ConfirmedBy = &confirmedBy.String
		}
		item.CreatedAt = createdAt.String
		item.CreatedBy = createdBy.String

		// Default values for closed (not yet in DB)
		item.Closed = 0
		item.ClosedAt = nil
		item.ClosedBy = nil

		if summary, ok := summaryMap[riderID]; ok {
			summary.StockList = append(summary.StockList, item)
		} else {
			summary := &domain.RiderStockSummary{
				RiderID:   riderID,
				RiderName: riderName,
				StockList: []domain.RiderStockItem{item},
			}
			summaryMap[riderID] = summary
			riderIDs = append(riderIDs, riderID)
		}
	}

	result := make([]domain.RiderStockSummary, 0, len(riderIDs))
	for _, id := range riderIDs {
		result = append(result, *summaryMap[id])
	}

	return result, nil
}


