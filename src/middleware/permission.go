package middleware

import (
	"net/http"
	"strconv"

	"aegis/database"
	"aegis/dto"
	"aegis/repository"

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

		userID, exists := GetCurrentUserID(c)
		if !exists {
			dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
			c.Abort()
			return
		}

		// Get project ID from URL parameters if available
		var projectID *int
		if projectIDStr := c.Param("project_id"); projectIDStr != "" {
			if id, err := strconv.Atoi(projectIDStr); err == nil {
				projectID = &id
			}
		}

		var containerID *int
		if containerIDStr := c.Param("container_id"); containerIDStr != "" {
			if id, err := strconv.Atoi(containerIDStr); err == nil {
				containerID = &id
			}
		}

		var datasetID *int
		if datasetIDStr := c.Param("dataset_id"); datasetIDStr != "" {
			if id, err := strconv.Atoi(datasetIDStr); err == nil {
				datasetID = &id
			}
		}

		// Check user permission
		hasPermission, err := repository.CheckUserPermission(database.DB, userID, action, resourceName, projectID, containerID, datasetID)
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

		userID, exists := GetCurrentUserID(c)
		if !exists {
			dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
			c.Abort()
			return
		}

		// Get project ID from URL parameters if available
		var projectID *int
		if projectIDStr := c.Param("project_id"); projectIDStr != "" {
			if id, err := strconv.Atoi(projectIDStr); err == nil {
				projectID = &id
			}
		}

		var containerID *int
		if containerIDStr := c.Param("container_id"); containerIDStr != "" {
			if id, err := strconv.Atoi(containerIDStr); err == nil {
				containerID = &id
			}
		}

		var datasetID *int
		if datasetIDStr := c.Param("dataset_id"); datasetIDStr != "" {
			if id, err := strconv.Atoi(datasetIDStr); err == nil {
				datasetID = &id
			}
		}

		// Check if user has any of the required permissions
		hasAnyPermission := false
		for _, perm := range permissions {
			hasPermission, err := repository.CheckUserPermission(database.DB, userID, perm.Action, perm.ResourceName, projectID, containerID, datasetID)
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

		userID, exists := GetCurrentUserID(c)
		if !exists {
			dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
			c.Abort()
			return
		}

		// Get project ID from URL parameters if available
		var projectID *int
		if projectIDStr := c.Param("project_id"); projectIDStr != "" {
			if id, err := strconv.Atoi(projectIDStr); err == nil {
				projectID = &id
			}
		}

		var containerID *int
		if containerIDStr := c.Param("container_id"); containerIDStr != "" {
			if id, err := strconv.Atoi(containerIDStr); err == nil {
				containerID = &id
			}
		}

		var datasetID *int
		if datasetIDStr := c.Param("dataset_id"); datasetIDStr != "" {
			if id, err := strconv.Atoi(datasetIDStr); err == nil {
				datasetID = &id
			}
		}

		// Check if user has all required permissions
		for _, perm := range permissions {
			hasPermission, err := repository.CheckUserPermission(database.DB, userID, perm.Action, perm.ResourceName, projectID, containerID, datasetID)
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

// TODO
// RequireOwnership creates a middleware that requires user to be owner of resource
func RequireOwnership(resourceType string, ownerIDParam string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First ensure user is authenticated
		if !RequireAuth(c) {
			c.Abort()
			return
		}

		userID, exists := GetCurrentUserID(c)
		if !exists {
			dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
			c.Abort()
			return
		}

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
		if userID != ownerID {
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

		userID, exists := GetCurrentUserID(c)
		if !exists {
			dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
			c.Abort()
			return
		}

		// Check if user is admin
		hasAdminPermission, err := repository.CheckUserPermission(database.DB, userID, "admin", "system", nil, nil, nil)
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

		if userID != ownerID {
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

	// Container Version management permissions
	RequireContainerVersionRead   = RequirePermission("read", "container_version")
	RequireContainerVersionWrite  = RequirePermission("write", "container_version")
	RequireContainerVersionDelete = RequirePermission("delete", "container_version")

	// Dataset management permissions
	RequireDatasetRead   = RequirePermission("read", "dataset")
	RequireDatasetWrite  = RequirePermission("write", "dataset")
	RequireDatasetDelete = RequirePermission("delete", "dataset")

	// Dataset Version management permissions
	RequireDatasetVersionRead   = RequirePermission("read", "dataset_version")
	RequireDatasetVersionWrite  = RequirePermission("write", "dataset_version")
	RequireDatasetVersionDelete = RequirePermission("delete", "dataset_version")

	// Project management permissions
	RequireProjectRead   = RequirePermission("read", "project")
	RequireProjectWrite  = RequirePermission("write", "project")
	RequireProjectDelete = RequirePermission("delete", "project")

	// Label management permissions
	RequireLabelRead   = RequirePermission("read", "label")
	RequireLabelWrite  = RequirePermission("write", "label")
	RequireLabelDelete = RequirePermission("delete", "label")

	// User management permissions
	RequireUserRead   = RequirePermission("read", "user")
	RequireUserWrite  = RequirePermission("write", "user")
	RequireUserDelete = RequirePermission("delete", "user")

	// Role management permissions
	RequireRoleRead   = RequirePermission("read", "role")
	RequireRoleWrite  = RequirePermission("write", "role")
	RequireRoleDelete = RequirePermission("delete", "role")

	// Permission management permissions
	RequirePermissionRead   = RequirePermission("read", "permission")
	RequirePermissionWrite  = RequirePermission("write", "permission")
	RequirePermissionDelete = RequirePermission("delete", "permission")

	// Task management permissions
	RequireTaskRead   = RequirePermission("read", "task")
	RequireTaskWrite  = RequirePermission("write", "task")
	RequireTaskDelete = RequirePermission("delete", "task")

	// Iinjection management permissions
	RequireInjectionRead   = RequirePermission("read", "injection")
	RequireInjectionWrite  = RequirePermission("write", "injection")
	RequireInjectionDelete = RequirePermission("delete", "injection")

	// Execution management permissions
	RequireExecutionRead   = RequirePermission("read", "execution")
	RequireExecutionWrite  = RequirePermission("write", "execution")
	RequireExecutionDelete = RequirePermission("delete", "execution")
)
