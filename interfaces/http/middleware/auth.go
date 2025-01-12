package middleware

import (
	"context"
	db "github.com/goodfoodcesi/auth-api/infrastructure/database/sqlc"
	"net/http"
	"strings"

	"github.com/goodfoodcesi/auth-api/interfaces/http/response"

	"github.com/goodfoodcesi/auth-api/infrastructure/jwt"
	"go.uber.org/zap"
)

func AuthMiddleware(logger *zap.Logger, tm *jwt.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Error(w, http.StatusUnauthorized, "Authorization header is missing", nil)
				return
			}

			bearerToken := strings.Split(authHeader, " ")
			if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
				response.Error(w, http.StatusUnauthorized, "Invalid authorization header format", nil)
				return
			}

			claims, err := tm.ValidateToken(bearerToken[1], false)
			if err != nil {
				logger.Error("invalid token", zap.Error(err))
				response.Error(w, http.StatusBadRequest, "Invalid token", nil)
				return
			}

			ctx := context.WithValue(r.Context(), "userID", claims.UserID)
			ctx = context.WithValue(ctx, "userRole", claims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(requiredRole db.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := r.Context().Value("userRole")
			if role == nil {
				response.Error(w, http.StatusUnauthorized, "Unauthorized", nil)
				return
			}

			if role.(db.UserRole) != requiredRole {
				response.Error(w, http.StatusForbidden, "Forbidden", nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
