package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"mynofuapi/internal/domain"
	"mynofuapi/pkg/utils"
	"strings"
	"time"

	"github.com/lib/pq"
)

type riderRepo struct {
	db        *sql.DB
	tableName string
}

func NewRiderRepository(db *sql.DB, tableName string) domain.UserRepository {
	return &riderRepo{
		db:        db,
		tableName: tableName,
	}
}

func (r *riderRepo) FindByUsername(ctx context.Context, username string) (domain.User, error) {
	log.Printf("Searching for rider: %s in %s", username, r.tableName)
	query := fmt.Sprintf(`
		SELECT i_id, c_username, c_password, c_nm, i_active, i_change_password, c_whatsapp_no
		FROM %s 
		WHERE c_username = $1
	`, r.tableName)

	var user domain.User
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Name,
		&user.IsActive,
		&user.ChangePassword,
		&user.WhatsappNumber,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("Rider not found: %s", username)
			return domain.User{}, errors.New("user not found")
		}
		log.Printf("Database error during Scan: %v", err)
		return domain.User{}, err
	}

	return user, nil
}

func (r *riderRepo) FindByID(ctx context.Context, id string) (domain.User, error) {
	log.Printf("Searching for rider by ID: %s in %s", id, r.tableName)
	query := fmt.Sprintf(`
		SELECT i_id, c_username, c_password, c_nm, i_active, i_change_password, c_whatsapp_no
		FROM %s 
		WHERE i_id = $1
	`, r.tableName)

	var user domain.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Name,
		&user.IsActive,
		&user.ChangePassword,
		&user.WhatsappNumber,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("Rider not found by ID: %s", id)
			return domain.User{}, errors.New("user not found")
		}
		log.Printf("Database error during Scan by ID: %v", err)
		return domain.User{}, err
	}

	return user, nil
}

func (r *riderRepo) GetAllRiders(ctx context.Context) ([]domain.Rider, error) {
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

func (r *riderRepo) CreateRider(ctx context.Context, rider domain.Rider) error {
	query := `
		INSERT INTO rider_master (c_nm, c_username, c_whatsapp_no, c_password, i_change_password, i_active, ts_created_at, c_created_by)
		VALUES ($1, $2, $3, $4, 1, 1, NOW(), 'SYSTEM')
	`
	_, err := r.db.ExecContext(ctx, query, rider.Name, rider.Username, rider.WhatsappNumber, "nofuoke")
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			if strings.Contains(pqErr.Detail, "c_username") || strings.Contains(pqErr.Detail, "username") {
				return errors.New("username already exists")
			}
			if strings.Contains(pqErr.Detail, "c_whatsapp_no") || strings.Contains(pqErr.Detail, "whatsapp") {
				return errors.New("whatsapp number already exists")
			}
			return errors.New("rider with this username or whatsapp number already exists")
		}
		log.Printf("Error creating rider: %v", err)
		return err
	}
	return nil
}
