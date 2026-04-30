package repository

import (
	"context"
	"errors"
	"mynofuapi/internal/domain"
)

type mockUserRepo struct {
	users map[string]domain.User
}

func NewMockUserRepository() domain.UserRepository {
	return &mockUserRepo{
		users: map[string]domain.User{
			"admin": {
				ID:       "1",
				Username: "admin",
				Password: "password123",
			},
		},
	}
}

func (m *mockUserRepo) FindByUsername(ctx context.Context, username string) (domain.User, error) {
	user, ok := m.users[username]
	if !ok {
		return domain.User{}, errors.New("user not found")
	}
	return user, nil
}
