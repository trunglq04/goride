package auth

import (
	"crypto/rsa"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Context keys used to store authenticated user information.
const (
	ContextKeyUserID = "auth_user_id"
	ContextKeyEmail  = "auth_email"
	ContextKeyRole   = "auth_role"
)

// JWTAuthMiddleware returns a Gin middleware that validates JWT access tokens.
// The middleware:
//  1. Extracts the Bearer token from the Authorization header
//  2. Validates the token using the RSA public key
//  3. Sets userID, email, and role in the Gin context
//  4. Returns 401 if the token is missing, invalid, or expired
func JWTAuthMiddleware(publicKey *rsa.PublicKey) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header must be in the format: Bearer <token>",
			})
			return
		}

		tokenString := parts[1]

		claims, err := ValidateAccessToken(tokenString, publicKey)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			return
		}

		// Set authenticated user info in context
		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyEmail, claims.Email)
		c.Set(ContextKeyRole, claims.Role)

		c.Next()
	}
}

// RequireRole returns a Gin middleware that checks if the authenticated user
// has one of the allowed roles. Must be used AFTER JWTAuthMiddleware.
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	roleSet := make(map[string]bool, len(allowedRoles))
	for _, r := range allowedRoles {
		roleSet[r] = true
	}

	return func(c *gin.Context) {
		role, exists := c.Get(ContextKeyRole)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			return
		}

		roleStr, ok := role.(string)
		if !ok || !roleSet[roleStr] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Insufficient permissions",
			})
			return
		}

		c.Next()
	}
}

// GetUserIDFromContext extracts the authenticated user ID from the Gin context.
func GetUserIDFromContext(c *gin.Context) (string, bool) {
	userID, exists := c.Get(ContextKeyUserID)
	if !exists {
		return "", false
	}
	str, ok := userID.(string)
	return str, ok
}

// GetRoleFromContext extracts the authenticated user role from the Gin context.
func GetRoleFromContext(c *gin.Context) (string, bool) {
	role, exists := c.Get(ContextKeyRole)
	if !exists {
		return "", false
	}
	str, ok := role.(string)
	return str, ok
}
