package domain

import "context"

type AuthRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	UserID           string  `json:"user_id"`
	Name             string  `json:"name"`
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
	GetAllRiders(ctx context.Context, active string) ([]Rider, error)
	CreateRider(ctx context.Context, rider Rider) error
	UpdatePassword(ctx context.Context, id string, newPassword string) error
	UpdateRiderStatus(ctx context.Context, id string, active int) error
	GetAllRidersWithStockStatus(ctx context.Context) ([]Rider, error)
}


type Rider struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Username       string `json:"username"`
	WhatsappNumber string `json:"whatsapp_number"`
	Active         int    `json:"active"`
	CreatedAt      string `json:"created_at"`
	CanInit        bool   `json:"can_init"`
}



type User struct {
	ID             string
	Username       string
	Password       string // Hashed
	Name           string
	IsActive       int
	ChangePassword int
	WhatsappNumber string
}
