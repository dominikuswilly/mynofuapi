package repository

import (
	"context"
	"database/sql"
	"log"
	"mynofuapi/internal/domain"
)

type productRepo struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) domain.ProductRepository {
	return &productRepo{
		db: db,
	}
}

func (r *productRepo) GetAllProducts(ctx context.Context) ([]domain.Product, error) {
	query := `
		SELECT c_id, c_nm, c_category, i_amt_sell, i_active
		FROM product_master
		ORDER BY c_id
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("Error querying products: %v", err)
		return nil, err
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		var p domain.Product
		var amt float64
		err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Category,
			&amt,
			&p.Active,
		)
		if err != nil {
			log.Printf("Error scanning product row: %v", err)
			continue
		}
		p.AmountSell = int(amt)
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating product rows: %v", err)
		return nil, err
	}

	if products == nil {
		products = []domain.Product{}
	}

	return products, nil
}
