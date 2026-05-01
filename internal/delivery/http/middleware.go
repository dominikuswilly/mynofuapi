package http

import (
	"mynofuapi/internal/domain"
	"net/http"
	"strings"
)

func AuthMiddleware(authUseCase domain.AuthUseCase) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Unauthorized: missing token", http.StatusUnauthorized)
				return
			}

			accessToken := ""
			if len(authHeader) > 7 && strings.HasPrefix(authHeader, "Bearer ") {
				accessToken = authHeader[7:]
			} else {
				accessToken = authHeader
			}

			refreshToken := r.Header.Get("X-Refresh-Token")
			if refreshToken == "" {
				refreshToken = r.URL.Query().Get("refresh_token")
			}

			// We use Introspect logic to validate
			_, err := authUseCase.Introspect(r.Context(), accessToken, refreshToken)
			if err != nil {
				http.Error(w, "Unauthorized: " + err.Error(), http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
