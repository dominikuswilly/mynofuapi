package domain

import "context"

type AuthRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	UserID           string  `json:"user_id"`
	AccessToken      *string `json:"access_token"`
	RefreshToken     *string `json:"refresh_token"`
	TokenType        string  `json:"token_type"`
	ExpiresIn        int     `json:"expires_in"`
	RefreshExpiresIn int     `json:"refresh_expires_in"`
}

type AuthUseCase interface {
	Authenticate(ctx context.Context, req AuthRequest) (AuthResponse, error)
	Introspect(ctx context.Context, accessToken, refreshToken string) (AuthResponse, error)
}

type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (User, error)
	FindByID(ctx context.Context, id string) (User, error)
}

type User struct {
	ID       string
	Username string
	Password string // Hashed
	Name     string
	IsActive int
}
