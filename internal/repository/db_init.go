package repository

import (
	"database/sql"
	"fmt"
	"log"
)

func InitDBTables(db *sql.DB) error {
	log.Println("Initializing database tables for Phase 1...")

	// 1. rider_waste_reports
	wasteQuery := `
		CREATE TABLE IF NOT EXISTS rider_waste_reports (
			c_id VARCHAR(50) PRIMARY KEY,
			ts_created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
			c_created_by VARCHAR(100) NOT NULL,
			c_updated_by VARCHAR(100),
			ts_updated_at TIMESTAMP WITH TIME ZONE,
			i_rider_id BIGINT NOT NULL,
			c_product_id VARCHAR(50) NOT NULL,
			c_product_nm VARCHAR(255) NOT NULL,
			i_qty BIGINT NOT NULL,
			c_reason VARCHAR(255) NOT NULL
		);
	`
	_, err := db.Exec(wasteQuery)
	if err != nil {
		return fmt.Errorf("failed to create rider_waste_reports table: %w", err)
	}
	log.Println("rider_waste_reports table checked/created.")

	// 2. rider_restock_requests
	restockReqQuery := `
		CREATE TABLE IF NOT EXISTS rider_restock_requests (
			c_id VARCHAR(50) PRIMARY KEY,
			ts_created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
			c_created_by VARCHAR(100) NOT NULL,
			c_updated_by VARCHAR(100),
			ts_updated_at TIMESTAMP WITH TIME ZONE,
			i_rider_id BIGINT NOT NULL,
			c_status VARCHAR(20) DEFAULT 'pending' NOT NULL
		);
	`
	_, err = db.Exec(restockReqQuery)
	if err != nil {
		return fmt.Errorf("failed to create rider_restock_requests table: %w", err)
	}
	log.Println("rider_restock_requests table checked/created.")

	// 3. rider_restock_items
	restockItemsQuery := `
		CREATE TABLE IF NOT EXISTS rider_restock_items (
			c_id VARCHAR(50) PRIMARY KEY,
			c_request_id VARCHAR(50) REFERENCES rider_restock_requests(c_id) ON DELETE CASCADE NOT NULL,
			c_product_id VARCHAR(50) NOT NULL,
			c_product_nm VARCHAR(255) NOT NULL,
			i_qty BIGINT NOT NULL
		);
	`
	_, err = db.Exec(restockItemsQuery)
	if err != nil {
		return fmt.Errorf("failed to create rider_restock_items table: %w", err)
	}
	log.Println("rider_restock_items table checked/created.")

	// 4. rider_commissions
	commissionsQuery := `
		CREATE TABLE IF NOT EXISTS rider_commissions (
			c_id VARCHAR(50) PRIMARY KEY,
			ts_created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
			c_created_by VARCHAR(100) NOT NULL,
			ts_date DATE NOT NULL,
			i_rider_id BIGINT NOT NULL,
			i_amt_sales NUMERIC NOT NULL,
			i_amt_commission NUMERIC NOT NULL,
			UNIQUE (i_rider_id, ts_date)
		);
	`
	_, err = db.Exec(commissionsQuery)
	if err != nil {
		return fmt.Errorf("failed to create rider_commissions table: %w", err)
	}
	log.Println("rider_commissions table checked/created.")

	log.Println("Database tables initialized successfully!")
	return nil
}
