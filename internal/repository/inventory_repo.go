package repository

import (
	"context"
	"database/sql"
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
		SELECT c_product_id, c_product_nm, i_qty_base, i_qty_current
		FROM rider_inventory
		WHERE i_rider_id = $1 AND LOWER(c_category) = LOWER($2)
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
		// amount_sell is not in the table, setting a placeholder or 0
		inv.AmountSell = "0" 
		
		err := rows.Scan(
			&inv.ProductID,
			&inv.ProductName,
			&inv.QtyBase,
			&inv.QtyCurrent,
		)
		if err != nil {
			log.Printf("Error scanning inventory row: %v", err)
			continue
		}
		inventories = append(inventories, inv)
	}

	return inventories, nil
}
