package middleware

import (
	"net/http"
	"strconv"

	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/repository"
	"github.com/gin-gonic/gin"
)

// RequirePermission creates a middleware that requires specific permission
func RequirePermission(action, resourceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First ensure user is authenticated
		if !RequireAuth(c) {
			c.Abort()
			return
		}

		userID, _ := GetCurrentUserID(c)

		// Get project ID from URL parameters if available
		var projectID *int
		if projectIDStr := c.Param("project_id"); projectIDStr != "" {
			if id, err := strconv.Atoi(projectIDStr); err == nil {
				projectID = &id
			}
		}

		// Check user permission
		hasPermission, err := repository.CheckUserPermission(userID, action, resourceName, projectID)
		if err != nil {
			dto.ErrorResponse(c, http.StatusInternalServerError, "Permission check failed: "+err.Error())
			c.Abort()
			return
		}

		if !hasPermission {
			dto.ErrorResponse(c, http.StatusForbidden, "Insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyPermission creates a middleware that requires any of the specified permissions
func RequireAnyPermission(permissions []PermissionCheck) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First ensure user is authenticated
		if !RequireAuth(c) {
			c.Abort()
			return
		}

		userID, _ := GetCurrentUserID(c)

		// Get project ID from URL parameters if available
		var projectID *int
		if projectIDStr := c.Param("project_id"); projectIDStr != "" {
			if id, err := strconv.Atoi(projectIDStr); err == nil {
				projectID = &id
			}
		}

		// Check if user has any of the required permissions
		hasAnyPermission := false
		for _, perm := range permissions {
			hasPermission, err := repository.CheckUserPermission(userID, perm.Action, perm.ResourceName, projectID)
			if err != nil {
				continue // Log error in production
			}
			if hasPermission {
				hasAnyPermission = true
				break
			}
		}

		if !hasAnyPermission {
			dto.ErrorResponse(c, http.StatusForbidden, "Insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAllPermissions creates a middleware that requires all specified permissions
func RequireAllPermissions(permissions []PermissionCheck) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First ensure user is authenticated
		if !RequireAuth(c) {
			c.Abort()
			return
		}

		userID, _ := GetCurrentUserID(c)

		// Get project ID from URL parameters if available
		var projectID *int
		if projectIDStr := c.Param("project_id"); projectIDStr != "" {
			if id, err := strconv.Atoi(projectIDStr); err == nil {
				projectID = &id
			}
		}

		// Check if user has all required permissions
		for _, perm := range permissions {
			hasPermission, err := repository.CheckUserPermission(userID, perm.Action, perm.ResourceName, projectID)
			if err != nil {
				dto.ErrorResponse(c, http.StatusInternalServerError, "Permission check failed: "+err.Error())
				c.Abort()
				return
			}
			if !hasPermission {
				dto.ErrorResponse(c, http.StatusForbidden, "Insufficient permissions")
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// RequireOwnership creates a middleware that requires user to be owner of resource
func RequireOwnership(resourceType string, ownerIDParam string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First ensure user is authenticated
		if !RequireAuth(c) {
			c.Abort()
			return
		}

		currentUserID, _ := GetCurrentUserID(c)

		// Get owner ID from URL parameters
		ownerIDStr := c.Param(ownerIDParam)
		if ownerIDStr == "" {
			dto.ErrorResponse(c, http.StatusBadRequest, "Resource owner ID not specified")
			c.Abort()
			return
		}

		ownerID, err := strconv.Atoi(ownerIDStr)
		if err != nil {
			dto.ErrorResponse(c, http.StatusBadRequest, "Invalid owner ID format")
			c.Abort()
			return
		}

		// Check if current user is the owner
		if currentUserID != ownerID {
			dto.ErrorResponse(c, http.StatusForbidden, "Access denied: You can only access your own "+resourceType)
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdminOrOwnership creates a middleware that requires user to be admin or owner
func RequireAdminOrOwnership(ownerIDParam string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First ensure user is authenticated
		if !RequireAuth(c) {
			c.Abort()
			return
		}

		currentUserID, _ := GetCurrentUserID(c)

		// Check if user is admin
		hasAdminPermission, err := repository.CheckUserPermission(currentUserID, "admin", "system", nil)
		if err == nil && hasAdminPermission {
			// User is admin, allow access
			c.Next()
			return
		}

		// Check ownership
		ownerIDStr := c.Param(ownerIDParam)
		if ownerIDStr == "" {
			dto.ErrorResponse(c, http.StatusBadRequest, "Resource owner ID not specified")
			c.Abort()
			return
		}

		ownerID, err := strconv.Atoi(ownerIDStr)
		if err != nil {
			dto.ErrorResponse(c, http.StatusBadRequest, "Invalid owner ID format")
			c.Abort()
			return
		}

		if currentUserID != ownerID {
			dto.ErrorResponse(c, http.StatusForbidden, "Access denied: Admin privileges or ownership required")
			c.Abort()
			return
		}

		c.Next()
	}
}

// PermissionCheck represents a permission requirement
type PermissionCheck struct {
	Action       string
	ResourceName string
}

// Common permission middlewares
var (
	// User management permissions
	RequireUserRead   = RequirePermission("read", "users")
	RequireUserWrite  = RequirePermission("write", "users")
	RequireUserDelete = RequirePermission("delete", "users")

	// Role management permissions
	RequireRoleRead   = RequirePermission("read", "roles")
	RequireRoleWrite  = RequirePermission("write", "roles")
	RequireRoleDelete = RequirePermission("delete", "roles")

	// Permission management permissions
	RequirePermissionRead   = RequirePermission("read", "permissions")
	RequirePermissionWrite  = RequirePermission("write", "permissions")
	RequirePermissionDelete = RequirePermission("delete", "permissions")

	// Project management permissions
	RequireProjectRead   = RequirePermission("read", "projects")
	RequireProjectWrite  = RequirePermission("write", "projects")
	RequireProjectDelete = RequirePermission("delete", "projects")

	// Dataset management permissions
	RequireDatasetRead   = RequirePermission("read", "dataset")
	RequireDatasetWrite  = RequirePermission("write", "dataset")
	RequireDatasetDelete = RequirePermission("delete", "dataset")

	// System administration permissions
	RequireSystemAdmin = RequirePermission("admin", "system")
	RequireSystemRead  = RequirePermission("read", "system")

	// Audit permissions
	RequireAuditRead = RequirePermission("read", "audit")

	// Resource ownership middlewares
	RequireUserOwnership        = RequireOwnership("user", "id")
	RequireAdminOrUserOwnership = RequireAdminOrOwnership("id")

	// Container management permissions
	RequireContainerRead   = RequirePermission("read", "container")
	RequireContainerWrite  = RequirePermission("write", "container")
	RequireContainerDelete = RequirePermission("delete", "container")

	// Task management permissions
	RequireTaskRead   = RequirePermission("read", "task")
	RequireTaskWrite  = RequirePermission("write", "task")
	RequireTaskDelete = RequirePermission("delete", "task")

	// Fault injection management permissions
	RequireFaultInjectionRead   = RequirePermission("read", "fault_injection")
	RequireFaultInjectionWrite  = RequirePermission("write", "fault_injection")
	RequireFaultInjectionDelete = RequirePermission("delete", "fault_injection")
)
