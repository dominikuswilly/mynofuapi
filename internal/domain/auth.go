package domain

import "context"

type AuthRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type AuthUseCase interface {
	Authenticate(ctx context.Context, req AuthRequest) (AuthResponse, error)
}

type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (User, error)
}

type User struct {
	ID       string
	Username string
	Password string // Hashed
}
