package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"mynofuapi/internal/domain"
)

type adminRepo struct {
	db        *sql.DB
	tableName string
}

func NewAdminRepository(db *sql.DB, tableName string) domain.UserRepository {
	return &adminRepo{
		db:        db,
		tableName: tableName,
	}
}

func (r *adminRepo) FindByUsername(ctx context.Context, username string) (domain.User, error) {
	log.Printf("Searching for admin: %s in %s", username, r.tableName)
	query := fmt.Sprintf(`
		SELECT i_id, c_username, c_password, c_nm, i_active 
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
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("Admin not found: %s", username)
			return domain.User{}, errors.New("user not found")
		}
		log.Printf("Database error during Scan: %v", err)
		return domain.User{}, err
	}

	// Default values for admin-specific fields that don't exist in admin_master
	user.ChangePassword = 0
	user.WhatsappNumber = ""

	return user, nil
}

func (r *adminRepo) FindByID(ctx context.Context, id string) (domain.User, error) {
	log.Printf("Searching for admin by ID: %s in %s", id, r.tableName)
	query := fmt.Sprintf(`
		SELECT i_id, c_username, c_password, c_nm, i_active 
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
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("Admin not found by ID: %s", id)
			return domain.User{}, errors.New("user not found")
		}
		log.Printf("Database error during Scan by ID: %v", err)
		return domain.User{}, err
	}

	user.ChangePassword = 0
	user.WhatsappNumber = ""

	return user, nil
}

func (r *adminRepo) GetAllRiders(ctx context.Context) ([]domain.Rider, error) {
	return nil, errors.New("method not allowed for admin repository")
}

func (r *adminRepo) CreateRider(ctx context.Context, rider domain.Rider) error {
	return errors.New("method not allowed for admin repository")
}

func (r *adminRepo) UpdatePassword(ctx context.Context, id string, newPassword string) error {
	return errors.New("method not allowed for admin repository")
}

func (r *adminRepo) UpdateRiderStatus(ctx context.Context, id string, active int) error {
	return errors.New("method not allowed for admin repository")
}

func (r *adminRepo) GetAllRidersWithStockStatus(ctx context.Context) ([]domain.Rider, error) {
	return nil, errors.New("method not allowed for admin repository")
}



