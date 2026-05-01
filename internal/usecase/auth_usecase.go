package usecase

import (
	"context"
	"errors"
	"mynofuapi/internal/domain"
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

	if user.IsActive != 1 {
		return domain.AuthResponse{}, errors.New("user is inactive")
	}

	// Mock token generation
	return domain.AuthResponse{
		AccessToken: "mock-jwt-token",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
	}, nil
}
