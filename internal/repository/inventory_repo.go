package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"mynofuapi/internal/domain"
)

type inventoryRepo struct {
	db *sql.DB
}

func NewInventoryRepository(db *sql.DB) domain.InventoryRepository {
	return &inventoryRepo{
		db: db,
	}
}

func (r *inventoryRepo) GetRiderInventory(ctx context.Context, riderID int, category string) ([]domain.RiderInventory, error) {
	query := `
		SELECT c_product_id, c_product_nm, i_qty_base, i_qty_current, i_amt_sell
		FROM rider_inventory
		WHERE i_rider_id = $1 AND LOWER(c_category) = LOWER($2)
		AND ts_created_at::date = CURRENT_DATE
		ORDER BY c_product_nm ASC
	`
	log.Printf("Executing query: %s with params: [%d, %s]", query, riderID, category)

	rows, err := r.db.QueryContext(ctx, query, riderID, category)
	if err != nil {
		log.Printf("Error querying inventory: %v", err)
		return nil, err
	}
	defer rows.Close()

	var inventories []domain.RiderInventory
	for rows.Next() {
		var inv domain.RiderInventory

		var amt float64
		err := rows.Scan(
			&inv.ProductID,
			&inv.ProductName,
			&inv.QtyBase,
			&inv.QtyCurrent,
			&amt,
		)
		if err != nil {
			log.Printf("Error scanning inventory row: %v", err)
			continue
		}
		inv.AmountSell = fmt.Sprintf("%.0f", amt)
		inventories = append(inventories, inv)
	}

	return inventories, nil
}

func (r *inventoryRepo) GetAllRiderInventory(ctx context.Context, riderID int) ([]domain.RiderInventory, error) {
	query := `
		SELECT c_product_id, c_product_nm, i_qty_base, i_qty_current, i_amt_sell
		FROM rider_inventory
		WHERE i_rider_id = $1
		AND ts_created_at::date = CURRENT_DATE
		ORDER BY c_product_nm ASC
	`
	log.Printf("Executing query: %s with params: [%d]", query, riderID)

	rows, err := r.db.QueryContext(ctx, query, riderID)
	if err != nil {
		log.Printf("Error querying inventory: %v", err)
		return nil, err
	}
	defer rows.Close()

	var inventories []domain.RiderInventory
	for rows.Next() {
		var inv domain.RiderInventory
		var amt float64
		err := rows.Scan(
			&inv.ProductID,
			&inv.ProductName,
			&inv.QtyBase,
			&inv.QtyCurrent,
			&amt,
		)
		if err != nil {
			log.Printf("Error scanning inventory row: %v", err)
			continue
		}
		inv.AmountSell = fmt.Sprintf("%.0f", amt)
		inventories = append(inventories, inv)
	}

	return inventories, nil
}

func (r *inventoryRepo) CheckInventoryConfirmation(ctx context.Context, riderID int) (bool, []domain.RiderInventory, error) {
	// 1. Check if any inventory is confirmed for today
	checkQuery := `
		SELECT 
			CASE 
				WHEN NOT EXISTS (SELECT 1 FROM rider_inventory WHERE i_rider_id = $1 AND ts_created_at::date = CURRENT_DATE) THEN false
				WHEN EXISTS (SELECT 1 FROM rider_inventory WHERE i_rider_id = $1 AND ts_created_at::date = CURRENT_DATE AND (ts_confirmed_at IS NULL OR i_confirmed = 0)) THEN false
				ELSE true
			END
	`

	var isConfirmed bool
	err := r.db.QueryRowContext(ctx, checkQuery, riderID).Scan(&isConfirmed)
	if err != nil {
		log.Printf("Error checking inventory confirmation: %v", err)
		return false, nil, err
	}

	// 2. If not confirmed, get the items that need confirmation
	var items []domain.RiderInventory
	if !isConfirmed {
		itemsQuery := `
			SELECT c_product_id, c_product_nm, i_qty_base, i_qty_current, i_amt_sell
			FROM rider_inventory
			WHERE i_rider_id = $1 
			AND ts_created_at::date = CURRENT_DATE
			AND (ts_confirmed_at IS NULL OR i_confirmed = 0)
			ORDER BY c_product_nm ASC
		`


		rows, err := r.db.QueryContext(ctx, itemsQuery, riderID)
		if err != nil {
			log.Printf("Error querying unconfirmed items: %v", err)
			return false, nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var inv domain.RiderInventory
			var amt float64
			err := rows.Scan(
				&inv.ProductID,
				&inv.ProductName,
				&inv.QtyBase,
				&inv.QtyCurrent,
				&amt,
			)
			if err != nil {
				continue
			}
			inv.AmountSell = fmt.Sprintf("%.0f", amt)
			items = append(items, inv)
		}
	}

	if items == nil {
		items = []domain.RiderInventory{}
	}

	return isConfirmed, items, nil
}

func (r *inventoryRepo) ConfirmInventory(ctx context.Context, riderID int, productID string, status string, confirmedBy string) error {
	confirmedFlag := 0
	if status == "accepted" {
		confirmedFlag = 1
	}

	query := `
		UPDATE rider_inventory 
		SET i_confirmed = $1,
		    ts_confirmed_at = NOW(), 
		    c_confirmed_by = $2,
		    c_updated_by = $3,
		    ts_updated_at = NOW()
		WHERE i_rider_id = $4 
		AND c_product_id = $5 
		AND ts_created_at::date = CURRENT_DATE
	`
	_, err := r.db.ExecContext(ctx, query, confirmedFlag, confirmedBy, confirmedBy, riderID, productID)
	if err != nil {
		log.Printf("Error confirming inventory: %v", err)
		return err
	}
	return nil
}



