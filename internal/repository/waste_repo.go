package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"mynofuapi/internal/domain"
	"strconv"
)

type wasteRepo struct {
	db *sql.DB
}

func NewWasteRepository(db *sql.DB) domain.WasteRepository {
	return &wasteRepo{db: db}
}

func (r *wasteRepo) CreateReport(ctx context.Context, riderID int, creatorName string, req domain.CreateWasteRequest) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Verify and fetch the current inventory row for today
	var productNm, qtyCurrentStr string
	inventoryQuery := `
		SELECT c_product_nm, i_qty_current
		FROM rider_inventory
		WHERE i_rider_id = $1 
		AND c_product_id = $2 
		AND (ts_created_at AT TIME ZONE 'Asia/Jakarta')::date = (CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Jakarta')::date
	`
	err = tx.QueryRowContext(ctx, inventoryQuery, riderID, req.ProductID).Scan(&productNm, &qtyCurrentStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("product %s not initiated in your daily stock", req.ProductID)
		}
		return err
	}

	qtyCurrent, err := strconv.Atoi(qtyCurrentStr)
	if err != nil {
		return fmt.Errorf("corrupt stock format: %w", err)
	}

	if qtyCurrent < req.Quantity {
		return fmt.Errorf("insufficient stock: you have %d, but trying to report %d defect", qtyCurrent, req.Quantity)
	}

	// 2. Decrement current stock in rider_inventory
	decrementQuery := `
		UPDATE rider_inventory
		SET i_qty_current = (i_qty_current::integer - $1::integer)::text,
		    c_updated_by = $2,
		    ts_updated_at = NOW()
		WHERE i_rider_id = $3 
		AND c_product_id = $4 
		AND (ts_created_at AT TIME ZONE 'Asia/Jakarta')::date = (CURRENT_TIMESTAMP AT TIME ZONE 'Asia/Jakarta')::date
	`
	_, err = tx.ExecContext(ctx, decrementQuery, req.Quantity, creatorName, riderID, req.ProductID)
	if err != nil {
		return fmt.Errorf("failed to update current inventory: %w", err)
	}

	// 3. Insert into rider_waste_reports
	insertQuery := `
		INSERT INTO rider_waste_reports (
			c_id, ts_created_at, c_created_by, i_rider_id, 
			c_product_id, c_product_nm, i_qty, c_reason
		)
		VALUES (gen_random_uuid(), NOW(), $1, $2, $3, $4, $5, $6)
	`
	_, err = tx.ExecContext(ctx, insertQuery, creatorName, riderID, req.ProductID, productNm, req.Quantity, req.Reason)
	if err != nil {
		return fmt.Errorf("failed to log waste report: %w", err)
	}

	return tx.Commit()
}

func (r *wasteRepo) ListReports(ctx context.Context, riderID int) ([]domain.WasteReport, error) {
	whereClause := "1=1"
	var params []interface{}

	if riderID > 0 {
		whereClause += " AND w.i_rider_id = $1"
		params = append(params, riderID)
	}

	query := fmt.Sprintf(`
		SELECT 
			w.c_id, 
			to_char(w.ts_created_at AT TIME ZONE 'Asia/Jakarta', 'YYYY-MM-DD HH24:MI:SS'), 
			w.c_created_by, 
			w.i_rider_id, 
			COALESCE(rm.c_nm, '') as rider_name,
			w.c_product_id, 
			w.c_product_nm, 
			w.i_qty, 
			w.c_reason
		FROM rider_waste_reports w
		LEFT JOIN rider_master rm ON w.i_rider_id = rm.i_id
		WHERE %s
		ORDER BY w.ts_created_at DESC
	`, whereClause)

	rows, err := r.db.QueryContext(ctx, query, params...)
	if err != nil {
		log.Printf("Error listing waste reports: %v", err)
		return nil, err
	}
	defer rows.Close()

	var reports []domain.WasteReport
	for rows.Next() {
		var w domain.WasteReport
		err := rows.Scan(
			&w.ID,
			&w.CreatedAt,
			&w.CreatedBy,
			&w.RiderID,
			&w.RiderName,
			&w.ProductID,
			&w.ProductName,
			&w.Quantity,
			&w.Reason,
		)
		if err != nil {
			log.Printf("Error scanning waste report: %v", err)
			continue
		}
		reports = append(reports, w)
	}

	if reports == nil {
		reports = []domain.WasteReport{}
	}

	return reports, nil
}
