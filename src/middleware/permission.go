package middleware

import (
	"net/http"
	"strconv"

	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"

	"github.com/gin-gonic/gin"
)

type permissionContext struct {
	userID      int
	containerID *int
	datasetID   *int
	projectID   *int
}

// permissionCheckFunc is a function that checks permission given the context
type permissionCheckFunc func(ctx *permissionContext) (bool, error)

// extractPermissionContext extracts common permission context from request
// Returns nil if service token (which bypasses permission checks)
// Returns error message if authentication fails
func extractPermissionContext(c *gin.Context) (*permissionContext, string) {
	// First ensure user is authenticated
	if !RequireAuth(c) {
		return nil, "" // RequireAuth already sent error response
	}

	// Service tokens bypass permission checks
	if IsServiceToken(c) {
		return nil, ""
	}

	userID, exists := GetCurrentUserID(c)
	if !exists {
		return nil, "Authentication required"
	}

	ctx := &permissionContext{userID: userID}

	// Extract optional IDs from URL parameters
	if projectIDStr := c.Param("project_id"); projectIDStr != "" {
		if id, err := strconv.Atoi(projectIDStr); err == nil {
			ctx.projectID = &id
		}
	}

	if containerIDStr := c.Param("container_id"); containerIDStr != "" {
		if id, err := strconv.Atoi(containerIDStr); err == nil {
			ctx.containerID = &id
		}
	}

	if datasetIDStr := c.Param("dataset_id"); datasetIDStr != "" {
		if id, err := strconv.Atoi(datasetIDStr); err == nil {
			ctx.datasetID = &id
		}
	}

	return ctx, ""
}

// withPermissionCheck creates a middleware decorator that wraps permission check logic
// This is similar to Python's decorator pattern
func withPermissionCheck(checkFunc permissionCheckFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		permCtx, errMsg := extractPermissionContext(c)

		// Auth failed (error already sent)
		if permCtx == nil && errMsg == "" && !IsServiceToken(c) {
			c.Abort()
			return
		}

		// Service token - bypass permission check
		if IsServiceToken(c) {
			c.Next()
			return
		}

		// Auth error
		if errMsg != "" {
			dto.ErrorResponse(c, http.StatusUnauthorized, errMsg)
			c.Abort()
			return
		}

		// Execute permission check
		hasPermission, err := checkFunc(permCtx)
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

// ============================================================================
// Permission Check Builders - create PermissionCheckFunc easily
// ============================================================================

// singlePermission creates a check for a single permission
func singlePermission(action, resourceName string) permissionCheckFunc {
	return func(ctx *permissionContext) (bool, error) {
		return repository.CheckUserPermission(
			database.DB, ctx.userID, action, resourceName,
			ctx.projectID, ctx.containerID, ctx.datasetID,
		)
	}
}

// anyPermission creates a check that passes if any permission is satisfied
func anyPermission(permissions []PermissionCheck) permissionCheckFunc {
	return func(ctx *permissionContext) (bool, error) {
		for _, perm := range permissions {
			hasPermission, err := repository.CheckUserPermission(
				database.DB, ctx.userID, perm.Action, perm.ResourceName,
				ctx.projectID, ctx.containerID, ctx.datasetID,
			)
			if err != nil {
				continue // Log error in production
			}
			if hasPermission {
				return true, nil
			}
		}
		return false, nil
	}
}

// allPermissions creates a check that passes only if all permissions are satisfied
func allPermissions(permissions []PermissionCheck) permissionCheckFunc {
	return func(ctx *permissionContext) (bool, error) {
		for _, perm := range permissions {
			hasPermission, err := repository.CheckUserPermission(
				database.DB, ctx.userID, perm.Action, perm.ResourceName,
				ctx.projectID, ctx.containerID, ctx.datasetID,
			)
			if err != nil {
				return false, err
			}
			if !hasPermission {
				return false, nil
			}
		}
		return true, nil
	}
}

// ownershipCheck creates a check for resource ownership
func ownershipCheck(ownerIDParam string) permissionCheckFunc {
	return func(ctx *permissionContext) (bool, error) {
		// This is a placeholder - actual implementation needs gin.Context
		// For ownership, we use a specialized middleware below
		return false, nil
	}
}

// adminOrOwnership creates a check for admin permission or ownership
func adminOrOwnership(ownerID int) permissionCheckFunc {
	return func(ctx *permissionContext) (bool, error) {
		// Check admin first
		hasAdmin, err := repository.CheckUserPermission(
			database.DB, ctx.userID, "admin", "system", nil, nil, nil,
		)
		if err == nil && hasAdmin {
			return true, nil
		}
		// Check ownership
		return ctx.userID == ownerID, nil
	}
}

// ============================================================================
// Simplified Middleware Constructors using the decorator pattern
// ============================================================================

// RequirePermission creates a middleware that requires specific permission
func RequirePermission(action, resourceName string) gin.HandlerFunc {
	return withPermissionCheck(singlePermission(action, resourceName))
}

// RequireAnyPermission creates a middleware that requires any of the specified permissions
func RequireAnyPermission(permissions []PermissionCheck) gin.HandlerFunc {
	return withPermissionCheck(anyPermission(permissions))
}

// RequireAllPermissions creates a middleware that requires all specified permissions
func RequireAllPermissions(permissions []PermissionCheck) gin.HandlerFunc {
	return withPermissionCheck(allPermissions(permissions))
}

// RequireOwnership creates a middleware that requires user to be owner of resource
func RequireOwnership(resourceType string, ownerIDParam string) gin.HandlerFunc {
	return withPermissionCheck(ownershipCheck(ownerIDParam))
}

// RequireAdminOrOwnership creates a middleware that requires user to be admin or owner
func RequireAdminOrOwnership(ownerIDParam string) gin.HandlerFunc {
	return withPermissionCheck(adminOrOwnership(0))
}

// ============================================================================
// Types and Common Permission Variables
// ============================================================================

// PermissionCheck represents a permission requirement
type PermissionCheck struct {
	Action       string
	ResourceName string
}

// Common permission middlewares
var (
	// System administration permissions
	RequireSystemAdmin = RequirePermission("admin", consts.ResourceSystem.String())
	RequireSystemRead  = RequirePermission("read", consts.ResourceSystem.String())

	// Audit permissions
	RequireAuditRead = RequirePermission("read", consts.ResourceAudit.String())

	// Configuration management permissions
	RequireConfigurationRead  = RequirePermission("read", consts.ResourceConfigruation.String())
	RequireConfigurationWrite = RequirePermission("write", consts.ResourceConfigruation.String())

	// Resource ownership middlewares
	RequireUserOwnership        = RequireOwnership("user", "id")
	RequireAdminOrUserOwnership = RequireAdminOrOwnership("id")

	// Container management permissions
	RequireContainerRead   = RequirePermission("read", consts.ResourceContainer.String())
	RequireContainerWrite  = RequirePermission("write", consts.ResourceContainer.String())
	RequireContainerDelete = RequirePermission("delete", consts.ResourceContainer.String())

	// Container Version management permissions
	RequireContainerVersionRead   = RequirePermission("read", consts.ResourceContainerVersion.String())
	RequireContainerVersionWrite  = RequirePermission("write", consts.ResourceContainerVersion.String())
	RequireContainerVersionDelete = RequirePermission("delete", consts.ResourceContainerVersion.String())

	// Dataset management permissions
	RequireDatasetRead   = RequirePermission("read", consts.ResourceDataset.String())
	RequireDatasetWrite  = RequirePermission("write", consts.ResourceDataset.String())
	RequireDatasetDelete = RequirePermission("delete", consts.ResourceDataset.String())

	// Dataset Version management permissions
	RequireDatasetVersionRead   = RequirePermission("read", consts.ResourceDatasetVersion.String())
	RequireDatasetVersionWrite  = RequirePermission("write", consts.ResourceDatasetVersion.String())
	RequireDatasetVersionDelete = RequirePermission("delete", consts.ResourceDatasetVersion.String())

	// Project management permissions
	RequireProjectRead   = RequirePermission("read", consts.ResourceProject.String())
	RequireProjectWrite  = RequirePermission("write", consts.ResourceProject.String())
	RequireProjectDelete = RequirePermission("delete", consts.ResourceProject.String())

	// Label management permissions
	RequireLabelRead   = RequirePermission("read", consts.ResourceLabel.String())
	RequireLabelWrite  = RequirePermission("write", consts.ResourceLabel.String())
	RequireLabelDelete = RequirePermission("delete", consts.ResourceLabel.String())

	// User management permissions
	RequireUserRead   = RequirePermission("read", consts.ResourceUser.String())
	RequireUserWrite  = RequirePermission("write", consts.ResourceUser.String())
	RequireUserDelete = RequirePermission("delete", consts.ResourceUser.String())

	// Role management permissions
	RequireRoleRead   = RequirePermission("read", consts.ResourceRole.String())
	RequireRoleWrite  = RequirePermission("write", consts.ResourceRole.String())
	RequireRoleDelete = RequirePermission("delete", consts.ResourceRole.String())

	// Permission management permissions
	RequirePermissionRead   = RequirePermission("read", consts.ResourcePermission.String())
	RequirePermissionWrite  = RequirePermission("write", consts.ResourcePermission.String())
	RequirePermissionDelete = RequirePermission("delete", consts.ResourcePermission.String())

	// Task management permissions
	RequireTaskRead   = RequirePermission("read", consts.ResourceTask.String())
	RequireTaskWrite  = RequirePermission("write", consts.ResourceTask.String())
	RequireTaskDelete = RequirePermission("delete", consts.ResourceTask.String())

	// Injection management permissions
	RequireInjectionRead   = RequirePermission("read", consts.ResourceInjection.String())
	RequireInjectionWrite  = RequirePermission("write", consts.ResourceInjection.String())
	RequireInjectionDelete = RequirePermission("delete", consts.ResourceInjection.String())

	// Execution management permissions
	RequireExecutionRead   = RequirePermission("read", consts.ResourceExecution.String())
	RequireExecutionWrite  = RequirePermission("write", consts.ResourceExecution.String())
	RequireExecutionDelete = RequirePermission("delete", consts.ResourceExecution.String())
)
