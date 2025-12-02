package middleware

import (
	"net/http"

	"aegis/dto"
	"aegis/utils"

	"github.com/gin-gonic/gin"
)

// JWTAuth is the JWT authentication middleware
// Supports both user tokens and service tokens (for K8s jobs)
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

		// Try to validate as user token first
		claims, err := utils.ValidateToken(token)
		if err == nil {
			// Valid user token - store user information in context
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("email", claims.Email)
			c.Set("is_active", claims.IsActive)
			c.Set("token_expires_at", claims.ExpiresAt.Time)
			c.Set("token_type", "user")
			c.Next()
			return
		}

		// Try to validate as service token (for K8s jobs)
		serviceClaims, serviceErr := utils.ValidateServiceToken(token)
		if serviceErr == nil {
			// Valid service token - store service information in context
			c.Set("task_id", serviceClaims.TaskID)
			c.Set("token_expires_at", serviceClaims.ExpiresAt.Time)
			c.Set("token_type", "service")
			c.Set("is_service_token", true)
			c.Next()
			return
		}

		// Both validations failed, return the user token error (more common case)
		dto.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized: "+err.Error())
		c.Abort()
	}
}

// OptionalJWTAuth is an optional JWT authentication middleware
// If token is provided, it validates it and sets user/service info
// If no token is provided, it continues without authentication
// Supports both user tokens and service tokens (for K8s jobs)
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

		// Try to validate as user token first
		claims, err := utils.ValidateToken(token)
		if err == nil {
			// Valid user token, set user information
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("email", claims.Email)
			c.Set("is_active", claims.IsActive)
			c.Set("token_expires_at", claims.ExpiresAt.Time)
			c.Set("token_type", "user")
			c.Next()
			return
		}

		// Try to validate as service token (for K8s jobs)
		serviceClaims, serviceErr := utils.ValidateServiceToken(token)
		if serviceErr == nil {
			// Valid service token, set service information
			c.Set("task_id", serviceClaims.TaskID)
			c.Set("token_expires_at", serviceClaims.ExpiresAt.Time)
			c.Set("token_type", "service")
			c.Set("is_service_token", true)
			c.Next()
			return
		}

		// Invalid token, continue without auth
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

// IsServiceToken checks if the current request is authenticated with a service token
func IsServiceToken(c *gin.Context) bool {
	isService, exists := c.Get("is_service_token")
	if !exists {
		return false
	}

	service, ok := isService.(bool)
	return ok && service
}

// GetTokenType returns the type of token used for authentication ("user", "service", or "")
func GetTokenType(c *gin.Context) string {
	tokenType, exists := c.Get("token_type")
	if !exists {
		return ""
	}

	t, ok := tokenType.(string)
	if !ok {
		return ""
	}
	return t
}

// GetServiceTaskID extracts task ID from service token context
func GetServiceTaskID(c *gin.Context) (string, bool) {
	taskID, exists := c.Get("task_id")
	if !exists {
		return "", false
	}

	id, ok := taskID.(string)
	return id, ok
}

// RequireAuth is a helper that ensures user or service is authenticated
func RequireAuth(c *gin.Context) bool {
	// Check if authenticated with service token
	if IsServiceToken(c) {
		taskID, exists := GetServiceTaskID(c)
		return exists && taskID != ""
	}

	// Check if authenticated with user token
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

// RequireUserAuth is a helper that ensures user (not service) is authenticated
func RequireUserAuth(c *gin.Context) bool {
	if IsServiceToken(c) {
		dto.ErrorResponse(c, http.StatusForbidden, "User authentication required, service token not allowed")
		return false
	}

	return RequireAuth(c)
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
