package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"mynofuapi/internal/domain"
	"mynofuapi/pkg/utils"
	"time"

	_ "github.com/lib/pq"
)

type userRepo struct {
	db        *sql.DB
	tableName string
}

func NewUserRepository(db *sql.DB, tableName string) domain.UserRepository {
	return &userRepo{
		db:        db,
		tableName: tableName,
	}
}

func (r *userRepo) FindByUsername(ctx context.Context, username string) (domain.User, error) {
	log.Printf("Searching for user: %s in %s", username, r.tableName)
	query := fmt.Sprintf(`
		SELECT i_id, c_username, c_password, c_nm, i_active 
		FROM %s 
		WHERE c_username = $1 AND i_active = 1
	`, r.tableName)

	var user domain.User
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Name,
		&user.IsActive,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("User not found: %s", username)
			return domain.User{}, errors.New("user not found")
		}
		log.Printf("Database error during Scan: %v", err)
		return domain.User{}, err
	}

	return user, nil
}

func (r *userRepo) FindByID(ctx context.Context, id string) (domain.User, error) {
	log.Printf("Searching for user by ID: %s in %s", id, r.tableName)
	query := fmt.Sprintf(`
		SELECT i_id, c_username, c_password, c_nm, i_active 
		FROM %s 
		WHERE i_id = $1 AND i_active = 1
	`, r.tableName)

	var user domain.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Name,
		&user.IsActive,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("User not found by ID: %s", id)
			return domain.User{}, errors.New("user not found")
		}
		log.Printf("Database error during Scan by ID: %v", err)
		return domain.User{}, err
	}

	return user, nil
}

func (r *userRepo) GetAllRiders(ctx context.Context) ([]domain.Rider, error) {
	query := `
		SELECT i_id, c_nm, c_username, ts_created_at, i_active, c_whatsapp_no 
		FROM rider_master
		ORDER BY i_id
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("Error querying all riders: %v", err)
		return nil, err
	}
	defer rows.Close()

	var riders []domain.Rider
	for rows.Next() {
		var rider domain.Rider
		var createdAt time.Time
		err := rows.Scan(
			&rider.ID,
			&rider.Name,
			&rider.Username,
			&createdAt,
			&rider.Active,
			&rider.WhatsappNumber,
		)
		if err != nil {
			log.Printf("Error scanning rider row: %v", err)
			continue
		}
		rider.CreatedAt = utils.FormatTime(createdAt)
		riders = append(riders, rider)
	}

	if riders == nil {
		riders = []domain.Rider{}
	}

	return riders, nil
}
