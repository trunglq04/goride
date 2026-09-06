package auth

import (
	"context"
	"crypto/rsa"
	"net/http"
	"strings"

	"github.com/trunglq04/goride/shared/util"
)

// Context keys used to store authenticated user information.
type contextKey string

const (
	ContextKeyUserID contextKey = "auth_user_id"
	ContextKeyEmail  contextKey = "auth_email"
	ContextKeyRole   contextKey = "auth_role"
)

// JWTAuthMiddleware returns an HTTP middleware that validates JWT access tokens.
// The middleware:
//  1. Extracts the Bearer token from the Authorization header
//  2. Validates the token using the RSA public key
//  3. Sets userID, email, and role in the request context
//  4. Returns 401 if the token is missing, invalid, or expired
func JWTAuthMiddleware(publicKey *rsa.PublicKey) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				util.WriteError(w, http.StatusUnauthorized, "Authorization header is required")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				util.WriteError(w, http.StatusUnauthorized, "Authorization header must be in the format: Bearer <token>")
				return
			}

			tokenString := parts[1]

			claims, err := ValidateAccessToken(tokenString, publicKey)
			if err != nil {
				util.WriteError(w, http.StatusUnauthorized, "Invalid or expired token")
				return
			}

			// Set authenticated user info in context
			ctx := r.Context()
			ctx = context.WithValue(ctx, ContextKeyUserID, claims.UserID)
			ctx = context.WithValue(ctx, ContextKeyEmail, claims.Email)
			ctx = context.WithValue(ctx, ContextKeyRole, claims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole returns an HTTP middleware that checks if the authenticated user
// has one of the allowed roles. Must be used AFTER JWTAuthMiddleware.
func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	roleSet := make(map[string]bool, len(allowedRoles))
	for _, r := range allowedRoles {
		roleSet[r] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(ContextKeyRole).(string)
			if !ok || role == "" {
				util.WriteError(w, http.StatusUnauthorized, "Authentication required")
				return
			}

			if !roleSet[role] {
				util.WriteError(w, http.StatusForbidden, "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserIDFromContext extracts the authenticated user ID from the request context.
func GetUserIDFromContext(r *http.Request) (string, bool) {
	v, ok := r.Context().Value(ContextKeyUserID).(string)
	return v, ok
}

// GetRoleFromContext extracts the authenticated user role from the request context.
func GetRoleFromContext(r *http.Request) (string, bool) {
	v, ok := r.Context().Value(ContextKeyRole).(string)
	return v, ok
}
