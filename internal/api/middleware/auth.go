package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"pvz-service/internal/domain/interfaces"
	"pvz-service/internal/domain/models"
)

type contextKey string

const (
	UserContextKey = contextKey("user")
)

// AuthMiddleware проверяет валидность JWT токена и добавляет информацию о пользователе в контекст
func AuthMiddleware(authService interfaces.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header is required", http.StatusUnauthorized)
				return
			}

			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Invalid authorization format, Bearer token required", http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == "" {
				http.Error(w, "Empty token provided", http.StatusUnauthorized)
				return
			}

			user, err := authService.ValidateToken(token)
			if err != nil {
				http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole проверяет, что у пользователя есть требуемая роль
func RequireRole(role models.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := r.Context().Value(UserContextKey).(*models.User)
			if !ok {
				http.Error(w, "Unauthorized: user not found in context", http.StatusUnauthorized)
				return
			}

			if user.Role != role && (role == models.RoleModerator) {
				http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserFromContext извлекает пользователя из контекста запроса
func GetUserFromContext(ctx context.Context) (*models.User, error) {
	user, ok := ctx.Value(UserContextKey).(*models.User)
	if !ok {
		return nil, errors.New("no user found in context")
	}
	return user, nil
}
