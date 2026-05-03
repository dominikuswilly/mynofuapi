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
		if err.Error() == "user not found" {
			return domain.AuthResponse{}, errors.New("invalid credentials")
		}
		return domain.AuthResponse{}, err
	}

	// Simple password check (In a real app, use bcrypt)
	if user.Password != req.Password {
		return domain.AuthResponse{}, errors.New("invalid credentials")
	}

	return a.generateTokens(user)
}

func (a *authUseCase) Introspect(ctx context.Context, accessToken, refreshToken string) (domain.AuthResponse, error) {
	// Parse Access Token
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	})

	// If access token is valid, return nil for AccessToken (meaning no refresh happened)
	if err == nil && token.Valid {
		claims, ok := token.Claims.(jwt.MapClaims)
		userID := ""
		name := ""
		if ok {
			userID, _ = claims["sub"].(string)
			name, _ = claims["name"].(string)
		}

		return domain.AuthResponse{
			UserID:      userID,
			Name:        name,
			AccessToken: nil,
			TokenType:   "Bearer",
			ExpiresIn:   3600,
		}, nil
	}

	// If access token is expired or invalid, check refresh token
	if refreshToken == "" {
		return domain.AuthResponse{}, errors.New("unauthorized: access token expired and no refresh token provided")
	}

	refreshTokenParsed, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	})

	if err != nil || !refreshTokenParsed.Valid {
		return domain.AuthResponse{}, errors.New("unauthorized: refresh token invalid or expired")
	}

	claims, ok := refreshTokenParsed.Claims.(jwt.MapClaims)
	if !ok || claims["type"] != "refresh" {
		return domain.AuthResponse{}, errors.New("unauthorized: invalid token type")
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		return domain.AuthResponse{}, errors.New("unauthorized: invalid claims")
	}

	// Fetch user to get up-to-date claims
	user, err := a.userRepo.FindByID(ctx, userID)
	if err != nil {
		return domain.AuthResponse{}, errors.New("unauthorized: user no longer exists or is inactive")
	}

	// Generate new tokens
	return a.generateTokens(user)
}

func (a *authUseCase) generateTokens(user domain.User) (domain.AuthResponse, error) {
	// Access Token generation
	expiresIn := 3600 // Short for testing/demo as seen in user diff
	expirationTime := time.Now().Add(time.Duration(expiresIn) * time.Second)

	claims := jwt.MapClaims{
		"sub":      user.ID,
		"id":       user.ID,
		"username": user.Username,
		"name":     user.Name,
		"exp":      expirationTime.Unix(),
		"iat":      time.Now().Unix(),
		"type":     "access",
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("secret"))
	if err != nil {
		return domain.AuthResponse{}, errors.New("failed to generate access token")
	}

	// Refresh Token generation (3 years)
	refreshExpiresIn := 3 * 365 * 24 * 60 * 60 // 3 years in seconds
	refreshExpirationTime := time.Now().Add(time.Duration(refreshExpiresIn) * time.Second)

	refreshClaims := jwt.MapClaims{
		"sub":      user.ID,
		"id":       user.ID,
		"username": user.Username,
		"name":     user.Name,
		"exp":      refreshExpirationTime.Unix(),
		"iat":      time.Now().Unix(),
		"type":     "refresh",
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte("secret"))
	if err != nil {
		return domain.AuthResponse{}, errors.New("failed to generate refresh token")
	}

	return domain.AuthResponse{
		UserID:           user.ID,
		Name:             user.Name,
		AccessToken:      &accessToken,
		RefreshToken:     &refreshToken,
		TokenType:        "Bearer",
		ExpiresIn:        expiresIn,
		RefreshExpiresIn: refreshExpiresIn,
	}, nil
}
