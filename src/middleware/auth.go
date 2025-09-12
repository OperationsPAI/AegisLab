package middleware

import (
	"net/http"

	"rcabench/dto"
	"rcabench/utils"
	"github.com/gin-gonic/gin"
)

// JWTAuth is the JWT authentication middleware
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		token, err := utils.ExtractTokenFromHeader(authHeader)
		if err != nil {
			dto.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized: "+err.Error())
			c.Abort()
			return
		}

		// Validate token
		claims, err := utils.ValidateToken(token)
		if err != nil {
			dto.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized: "+err.Error())
			c.Abort()
			return
		}

		// Store user information in context for use by handlers
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("is_active", claims.IsActive)
		c.Set("token_expires_at", claims.ExpiresAt.Time)

		c.Next()
	}
}

// OptionalJWTAuth is an optional JWT authentication middleware
// If token is provided, it validates it and sets user info
// If no token is provided, it continues without authentication
func OptionalJWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No authentication provided, continue
			c.Next()
			return
		}

		token, err := utils.ExtractTokenFromHeader(authHeader)
		if err != nil {
			// Invalid header format, continue without auth
			c.Next()
			return
		}

		claims, err := utils.ValidateToken(token)
		if err != nil {
			// Invalid token, continue without auth
			c.Next()
			return
		}

		// Valid token, set user information
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("is_active", claims.IsActive)
		c.Set("token_expires_at", claims.ExpiresAt.Time)

		c.Next()
	}
}

// GetCurrentUserID extracts current user ID from context
func GetCurrentUserID(c *gin.Context) (int, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	id, ok := userID.(int)
	return id, ok
}

// GetCurrentUsername extracts current username from context
func GetCurrentUsername(c *gin.Context) (string, bool) {
	username, exists := c.Get("username")
	if !exists {
		return "", false
	}

	name, ok := username.(string)
	return name, ok
}

// GetCurrentUserEmail extracts current user email from context
func GetCurrentUserEmail(c *gin.Context) (string, bool) {
	email, exists := c.Get("email")
	if !exists {
		return "", false
	}

	userEmail, ok := email.(string)
	return userEmail, ok
}

// IsCurrentUserActive checks if current user is active
func IsCurrentUserActive(c *gin.Context) bool {
	isActive, exists := c.Get("is_active")
	if !exists {
		return false
	}

	active, ok := isActive.(bool)
	return ok && active
}

// RequireAuth is a helper that ensures user is authenticated
func RequireAuth(c *gin.Context) bool {
	userID, exists := GetCurrentUserID(c)
	if !exists || userID <= 0 {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return false
	}

	if !IsCurrentUserActive(c) {
		dto.ErrorResponse(c, http.StatusForbidden, "User account is inactive")
		return false
	}

	return true
}

// RequireActiveUser ensures the current user exists and is active
func RequireActiveUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !RequireAuth(c) {
			c.Abort()
			return
		}
		c.Next()
	}
}
