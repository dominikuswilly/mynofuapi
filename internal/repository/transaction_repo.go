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
