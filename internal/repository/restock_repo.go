package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"mynofuapi/internal/domain"
)

type restockRepo struct {
	db *sql.DB
}

func NewRestockRepository(db *sql.DB) domain.RestockRepository {
	return &restockRepo{db: db}
}

func (r *restockRepo) CreateRequest(ctx context.Context, riderID int, creatorName string, req domain.CreateRestockRequest) (string, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	// 1. Insert into rider_restock_requests
	requestID := ""
	insertParentQuery := `
		INSERT INTO rider_restock_requests (c_id, ts_created_at, c_created_by, i_rider_id, c_status)
		VALUES (gen_random_uuid(), NOW(), $1, $2, 'pending')
		RETURNING c_id
	`
	err = tx.QueryRowContext(ctx, insertParentQuery, creatorName, riderID).Scan(&requestID)
	if err != nil {
		log.Printf("Error inserting restock parent: %v", err)
		return "", fmt.Errorf("failed to create restock request: %w", err)
	}

	// 2. Insert into rider_restock_items
	for _, item := range req.Items {
		// Fetch product name first to keep data denormalized
		var productNm string
		productQuery := `SELECT c_nm FROM product_master WHERE c_id = $1`
		err = tx.QueryRowContext(ctx, productQuery, item.ProductID).Scan(&productNm)
		if err != nil {
			log.Printf("Error looking up product %s for restock: %v", item.ProductID, err)
			return "", fmt.Errorf("product %s not found: %w", item.ProductID, err)
		}

		insertItemQuery := `
			INSERT INTO rider_restock_items (c_id, c_request_id, c_product_id, c_product_nm, i_qty)
			VALUES (gen_random_uuid(), $1, $2, $3, $4)
		`
		_, err = tx.ExecContext(ctx, insertItemQuery, requestID, item.ProductID, productNm, item.Quantity)
		if err != nil {
			log.Printf("Error inserting restock item %s: %v", item.ProductID, err)
			return "", fmt.Errorf("failed to add restock item %s: %w", item.ProductID, err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return "", err
	}

	return requestID, nil
}

func (r *restockRepo) ListRequests(ctx context.Context, riderID int, status string) ([]domain.RestockRequest, error) {
	whereClause := "1=1"
	var params []interface{}
	paramCount := 0

	if riderID > 0 {
		paramCount++
		whereClause += fmt.Sprintf(" AND r.i_rider_id = $%d", paramCount)
		params = append(params, riderID)
	}

	if status != "" {
		paramCount++
		whereClause += fmt.Sprintf(" AND r.c_status = $%d", paramCount)
		params = append(params, status)
	}

	query := fmt.Sprintf(`
		SELECT 
			r.c_id, 
			to_char(r.ts_created_at AT TIME ZONE 'Asia/Jakarta', 'YYYY-MM-DD HH24:MI:SS'), 
			r.c_created_by, 
			r.i_rider_id, 
			COALESCE(rm.c_nm, '') as rider_name,
			r.c_status
		FROM rider_restock_requests r
		LEFT JOIN rider_master rm ON r.i_rider_id = rm.i_id
		WHERE %s
		ORDER BY r.ts_created_at DESC
	`, whereClause)

	rows, err := r.db.QueryContext(ctx, query, params...)
	if err != nil {
		log.Printf("Error listing restock requests: %v", err)
		return nil, err
	}
	defer rows.Close()

	var requests []domain.RestockRequest
	for rows.Next() {
		var req domain.RestockRequest
		err := rows.Scan(
			&req.ID,
			&req.CreatedAt,
			&req.CreatedBy,
			&req.RiderID,
			&req.RiderName,
			&req.Status,
		)
		if err != nil {
			log.Printf("Error scanning restock request row: %v", err)
			continue
		}

		// Fetch items for this request
		items, err := r.getRequestItems(ctx, req.ID)
		if err != nil {
			log.Printf("Error fetching items for request %s: %v", req.ID, err)
			continue
		}
		req.Items = items

		requests = append(requests, req)
	}

	if requests == nil {
		requests = []domain.RestockRequest{}
	}

	return requests, nil
}

func (r *restockRepo) GetRequestByID(ctx context.Context, id string) (domain.RestockRequest, error) {
	query := `
		SELECT 
			r.c_id, 
			to_char(r.ts_created_at AT TIME ZONE 'Asia/Jakarta', 'YYYY-MM-DD HH24:MI:SS'), 
			r.c_created_by, 
			r.i_rider_id, 
			COALESCE(rm.c_nm, '') as rider_name,
			r.c_status
		FROM rider_restock_requests r
		LEFT JOIN rider_master rm ON r.i_rider_id = rm.i_id
		WHERE r.c_id = $1
	`
	var req domain.RestockRequest
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&req.ID,
		&req.CreatedAt,
		&req.CreatedBy,
		&req.RiderID,
		&req.RiderName,
		&req.Status,
	)
	if err != nil {
		return domain.RestockRequest{}, err
	}

	items, err := r.getRequestItems(ctx, req.ID)
	if err != nil {
		return domain.RestockRequest{}, err
	}
	req.Items = items

	return req, nil
}

func (r *restockRepo) getRequestItems(ctx context.Context, requestID string) ([]domain.RestockItem, error) {
	query := `
		SELECT c_product_id, c_product_nm, i_qty
		FROM rider_restock_items
		WHERE c_request_id = $1
		ORDER BY c_product_nm ASC
	`
	rows, err := r.db.QueryContext(ctx, query, requestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.RestockItem
	for rows.Next() {
		var item domain.RestockItem
		if err := rows.Scan(&item.ProductID, &item.ProductName, &item.Quantity); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if items == nil {
		items = []domain.RestockItem{}
	}

	return items, nil
}

func (r *restockRepo) UpdateRequestStatus(ctx context.Context, id string, status string, adminName string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Get request details
	var riderID int
	var currentStatus string
	checkQuery := `
		SELECT i_rider_id, c_status 
		FROM rider_restock_requests 
		WHERE c_id = $1
	`
	err = tx.QueryRowContext(ctx, checkQuery, id).Scan(&riderID, &currentStatus)
	if err != nil {
		return fmt.Errorf("restock request not found: %w", err)
	}

	if currentStatus != "pending" {
		return fmt.Errorf("request has already been resolved as %s", currentStatus)
	}

	// 2. Update status in parent table
	updateParentQuery := `
		UPDATE rider_restock_requests
		SET c_status = $1, c_updated_by = $2, ts_updated_at = NOW()
		WHERE c_id = $3
	`
	_, err = tx.ExecContext(ctx, updateParentQuery, status, adminName, id)
	if err != nil {
		return err
	}

	// 3. If approved, add products to rider's inventory
	if status == "approved" {
		items, err := r.getRequestItems(ctx, id)
		if err != nil {
			return err
		}

		for _, item := range items {
			// Try to update existing daily stock
			updateStockQuery := `
				UPDATE rider_inventory
				SET i_qty_base = (i_qty_base::integer + $1::integer)::text,
				    i_qty_current = (i_qty_current::integer + $2::integer)::text,
				    c_updated_by = $3,
				    ts_updated_at = NOW()
				WHERE i_rider_id = $4
				AND c_product_id = $5
				AND (ts_created_at AT TIME ZONE 'Asia/Jakarta')::date = (CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Jakarta')::date
			`
			res, err := tx.ExecContext(ctx, updateStockQuery, item.Quantity, item.Quantity, adminName, riderID, item.ProductID)
			if err != nil {
				return fmt.Errorf("failed to increment stock for product %s: %w", item.ProductID, err)
			}

			rowsAffected, _ := res.RowsAffected()
			if rowsAffected == 0 {
				// Daily inventory entry does not exist yet. Let's create it automatically.
				var productNm, category string
				var amtSell float64
				productQuery := `
					SELECT c_nm, c_category, i_amt_sell 
					FROM product_master 
					WHERE c_id = $1
				`
				err = tx.QueryRowContext(ctx, productQuery, item.ProductID).Scan(&productNm, &category, &amtSell)
				if err != nil {
					return fmt.Errorf("product %s not found in catalog: %w", item.ProductID, err)
				}

				insertInventoryQuery := `
					INSERT INTO rider_inventory (
						c_id, c_created_by, ts_created_at, c_product_id, 
						i_qty_base, i_qty_current, i_rider_id, 
						c_category, c_product_nm, i_amt_sell
					)
					VALUES (gen_random_uuid(), $1, NOW(), $2, $3, $4, $5, $6, $7, $8)
				`
				_, err = tx.ExecContext(ctx, insertInventoryQuery,
					adminName,
					item.ProductID,
					fmt.Sprintf("%d", item.Quantity),
					fmt.Sprintf("%d", item.Quantity),
					riderID,
					category,
					productNm,
					amtSell,
				)
				if err != nil {
					return fmt.Errorf("failed to auto-create daily inventory row for product %s: %w", item.ProductID, err)
				}
			}
		}
	}

	return tx.Commit()
}
