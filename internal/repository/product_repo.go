package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"mynofuapi/internal/domain"
	"strings"
)

type productRepo struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) domain.ProductRepository {
	return &productRepo{
		db: db,
	}
}

func (r *productRepo) GetAllProducts(ctx context.Context, category string, includeInactive bool) ([]domain.Product, error) {
	query := `
		SELECT c_id, c_nm, c_category, i_amt_sell, i_active
		FROM product_master
		WHERE 1=1
	`
	var args []interface{}

	if !includeInactive {
		query += " AND i_active = 1"
	}

	if category != "" {
		query += fmt.Sprintf(" AND LOWER(c_category) = LOWER($%d)", len(args)+1)
		args = append(args, category)
	}

	query += " ORDER BY c_id"

	rows, err := r.db.QueryContext(ctx, query, args...)
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

func (r *productRepo) UpdateProduct(ctx context.Context, id string, req domain.PatchProductRequest) error {
	query := "UPDATE product_master SET "
	var args []interface{}
	var updates []string
	argIdx := 1

	if req.Name != nil {
		updates = append(updates, fmt.Sprintf("c_nm = $%d", argIdx))
		args = append(args, *req.Name)
		argIdx++
	}

	if req.AmountSell != nil {
		updates = append(updates, fmt.Sprintf("i_amt_sell = $%d", argIdx))
		args = append(args, *req.AmountSell)
		argIdx++
	}

	if req.Active != nil {
		updates = append(updates, fmt.Sprintf("i_active = $%d", argIdx))
		args = append(args, *req.Active)
		argIdx++
	}

	if len(updates) == 0 {
		return nil // Nothing to update
	}

	query += strings.Join(updates, ", ")
	query += fmt.Sprintf(" WHERE c_id = $%d", argIdx)
	args = append(args, id)

	log.Printf("Executing update: %s with args: %v", query, args)

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *productRepo) DeleteProduct(ctx context.Context, id string) error {
	query := "UPDATE product_master SET i_active = 0 WHERE c_id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *productRepo) CreateProduct(ctx context.Context, req domain.CreateProductRequest) error {
	id := generateProductID(req.Name)
	query := `
		INSERT INTO product_master (c_id, c_nm, c_category, i_amt_sell, i_active)
		VALUES ($1, $2, $3, $4, 1)
	`
	log.Printf("Creating product with ID: %s, Name: %s", id, req.Name)
	_, err := r.db.ExecContext(ctx, query, id, strings.ToUpper(req.Name), req.Category, req.AmountSell)
	return err
}

func generateProductID(name string) string {
	// Remove spaces and convert to uppercase
	id := strings.ReplaceAll(name, " ", "")
	id = strings.ToUpper(id)

	// Max length 15
	if len(id) > 15 {
		id = id[:15]
	}
	return id
}
