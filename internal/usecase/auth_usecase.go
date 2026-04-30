package usecase

import (
	"context"
	"errors"
	"mynofuapi/internal/domain"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type authUseCase struct {
	userRepo domain.UserRepository
}

func NewAuthUseCase(userRepo domain.UserRepository) domain.AuthUseCase {
	return &authUseCase{
		userRepo: userRepo,
	}
}

func (a *authUseCase) Authenticate(ctx context.Context, req domain.AuthRequest) (domain.AuthResponse, error) {
	user, err := a.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return domain.AuthResponse{}, errors.New("invalid credentials")
	}

	// Simple password check (In a real app, use bcrypt)
	if user.Password != req.Password {
		return domain.AuthResponse{}, errors.New("invalid credentials")
	}

	// JWT Token generation
	expiresIn := 3600
	expirationTime := time.Now().Add(time.Duration(expiresIn) * time.Second)

	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": expirationTime.Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("secret")) // In production, use env variable
	if err != nil {
		return domain.AuthResponse{}, errors.New("failed to generate token")
	}

	return domain.AuthResponse{
		AccessToken: tokenString,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
	}, nil
}
