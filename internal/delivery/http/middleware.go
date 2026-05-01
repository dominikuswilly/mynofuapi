package http

import (
	"bytes"
	"io"
	"log"
	"mynofuapi/internal/domain"
	"net/http"
	"strings"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get Request ID from header
		requestID := r.Header.Get("X-Request-ID")
		if requestID != "" {
			w.Header().Set("X-Request-ID", requestID)
		}

		// Log Request Headers
		log.Printf("Request: %s %s", r.Method, r.URL.Path)
		for name, values := range r.Header {
			log.Printf("Request Header: %s: %s", name, strings.Join(values, ", "))
		}

		// Log Request Body
		var bodyBytes []byte
		if r.Body != nil {
			bodyBytes, _ = io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Restore for handler
			log.Printf("Request Body: %s", string(bodyBytes))
		}

		// Wrap ResponseWriter
		rw := &responseWriter{
			ResponseWriter: w,
			body:           &bytes.Buffer{},
		}

		next.ServeHTTP(rw, r)

		// Log Response Headers
		for name, values := range rw.Header() {
			log.Printf("Response Header: %s: %s", name, strings.Join(values, ", "))
		}

		// Log Response Status and Body
		log.Printf("Response Status: %d", rw.status)
		log.Printf("Response Body: %s", rw.body.String())
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
	body   *bytes.Buffer
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.status == 0 {
		rw.status = http.StatusOK
	}
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}

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
