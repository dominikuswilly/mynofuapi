package repository

import (
	"context"
	"database/sql"
	"errors"
	"mynofuapi/internal/domain"

	_ "github.com/lib/pq"
)

type userRepo struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) domain.UserRepository {
	return &userRepo{
		db: db,
	}
}

func (r *userRepo) FindByUsername(ctx context.Context, username string) (domain.User, error) {
	query := `
		SELECT c_id, c_username, c_password, c_nm, i_active 
		FROM rider_master 
		WHERE c_username = $1 AND i_active = 1
	`

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
			return domain.User{}, errors.New("user not found")
		}
		return domain.User{}, err
	}

	return user, nil
}
