package api

import (
	"context"
	"github.com/KalashnikovProjects/ZadachaGoYaLyceum/internal/auth"
	"net/http"
	"strings"
)

// AuthenticationMiddleware проверяет наличие и валидность Bearer токена в запросе
func AuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		userId, err := auth.LoadUserIdFromToken(token)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), userId, "userId")

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
