package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"aegis/consts"
	"aegis/dto"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type permissionContext struct {
	userID       int
	isAdmin      bool
	roles        []string
	authType     string
	apiKeyScopes []string
	checker      permissionChecker
	ctx          context.Context
	teamID       *int
	projectID    *int
	containerID  *int
	datasetID    *int
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

	// Get admin status and roles from context (already validated in JWT)
	isAdmin := false
	if val, exists := c.Get("is_admin"); exists {
		if v, ok := val.(bool); ok {
			isAdmin = v
		}
	}
	roles := []string{}
	if val, exists := c.Get("user_roles"); exists {
		if v, ok := val.([]string); ok {
			roles = v
		}
	}

	ctx := &permissionContext{
		userID:   userID,
		isAdmin:  isAdmin,
		roles:    roles,
		checker:  permissionCheckerFromContext(c),
		ctx:      c.Request.Context(),
		authType: GetAuthType(c),
	}
	if scopes, ok := GetCurrentAPIKeyScopes(c); ok {
		ctx.apiKeyScopes = append([]string(nil), scopes...)
	}

	// Extract optional IDs from URL parameters
	if teamIDstr := c.Param(consts.URLPathTeamID); teamIDstr != "" {
		if id, err := strconv.Atoi(teamIDstr); err == nil {
			ctx.teamID = &id
		}
	}

	if projectIDStr := c.Param(consts.URLPathProjectID); projectIDStr != "" {
		if id, err := strconv.Atoi(projectIDStr); err == nil {
			ctx.projectID = &id
		}
	}

	if containerIDStr := c.Param(consts.URLPathContainerID); containerIDStr != "" {
		if id, err := strconv.Atoi(containerIDStr); err == nil {
			ctx.containerID = &id
		}
	}

	if datasetIDStr := c.Param(consts.URLPathDatasetID); datasetIDStr != "" {
		if id, err := strconv.Atoi(datasetIDStr); err == nil {
			ctx.datasetID = &id
		}
	}

	return ctx, ""
}

func (ctx *permissionContext) isAPIKeyAuth() bool {
	return ctx != nil && ctx.authType == "api_key"
}

func (ctx *permissionContext) scopeAllowsPermission(permission consts.PermissionRule) bool {
	if !ctx.isAPIKeyAuth() {
		return true
	}
	if len(ctx.apiKeyScopes) == 0 {
		return false
	}
	for _, scope := range ctx.apiKeyScopes {
		if apiKeyScopeMatchesPermission(scope, permission) {
			return true
		}
	}
	return false
}

func (ctx *permissionContext) scopeAllowsAnyPermission(permissions []consts.PermissionRule) bool {
	for _, permission := range permissions {
		if ctx.scopeAllowsPermission(permission) {
			return true
		}
	}
	return false
}

func apiKeyScopeMatchesPermission(scope string, permission consts.PermissionRule) bool {
	return apiKeyScopeMatchesTarget(scope, permission.String())
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
func singlePermission(permission consts.PermissionRule) permissionCheckFunc {
	return func(ctx *permissionContext) (bool, error) {
		if !ctx.scopeAllowsPermission(permission) {
			return false, nil
		}
		return ctx.checker.CheckUserPermission(ctx.ctx, &dto.CheckPermissionParams{
			UserID:       ctx.userID,
			Action:       permission.Action,
			Scope:        permission.Scope,
			ResourceName: permission.Resource,
			TeamID:       ctx.teamID,
			ProjectID:    ctx.projectID,
			ContainerID:  ctx.containerID,
			DatasetID:    ctx.datasetID,
		},
		)
	}
}

// anyPermission creates a check that passes if any permission is satisfied
func anyPermission(permissions []consts.PermissionRule) permissionCheckFunc {
	return func(ctx *permissionContext) (bool, error) {
		for _, perm := range permissions {
			if !ctx.scopeAllowsPermission(perm) {
				continue
			}
			hasPermission, err := ctx.checker.CheckUserPermission(
				ctx.ctx,
				&dto.CheckPermissionParams{
					UserID:       ctx.userID,
					Action:       perm.Action,
					Scope:        perm.Scope,
					ResourceName: perm.Resource,
					TeamID:       ctx.teamID,
					ProjectID:    ctx.projectID,
					ContainerID:  ctx.containerID,
					DatasetID:    ctx.datasetID,
				},
			)
			if err != nil {
				logrus.Warnf("Permission check error: %f", err)
				continue
			}
			if hasPermission {
				return true, nil
			}
		}
		return false, nil
	}
}

// allPermissions creates a check that passes only if all permissions are satisfied
func allPermissions(permissions []consts.PermissionRule) permissionCheckFunc {
	return func(ctx *permissionContext) (bool, error) {
		for _, perm := range permissions {
			if !ctx.scopeAllowsPermission(perm) {
				return false, nil
			}
			hasPermission, err := ctx.checker.CheckUserPermission(
				ctx.ctx,
				&dto.CheckPermissionParams{
					UserID:       ctx.userID,
					Action:       perm.Action,
					Scope:        perm.Scope,
					ResourceName: perm.Resource,
					TeamID:       ctx.teamID,
					ProjectID:    ctx.projectID,
					ContainerID:  ctx.containerID,
					DatasetID:    ctx.datasetID,
				},
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
// Supports checking ownership for different resource types by extracting IDs from context
func ownershipCheck(resourceType string, ownerIDGetter func(*permissionContext) (*int, error)) permissionCheckFunc {
	return func(ctx *permissionContext) (bool, error) {
		// Get the owner ID dynamically
		ownerID, err := ownerIDGetter(ctx)
		if err != nil {
			return false, err
		}
		if ownerID == nil {
			return false, fmt.Errorf("owner ID not found for resource type: %s", resourceType)
		}

		// Check ownership
		return ctx.userID == *ownerID, nil
	}
}

// adminOrOwnership creates a check for admin permission or ownership
// Supports checking admin status OR ownership for different resource types
func adminOrOwnership(resourceType string, ownerIDGetter func(*permissionContext) (*int, error)) permissionCheckFunc {
	return func(ctx *permissionContext) (bool, error) {
		// Check admin status from JWT context (fast path - no DB query)
		if ctx.isAdmin {
			return true, nil
		}

		// Get the owner ID dynamically
		ownerID, err := ownerIDGetter(ctx)
		if err != nil {
			return false, err
		}
		if ownerID == nil {
			return false, fmt.Errorf("owner ID not found for resource type: %s", resourceType)
		}

		// Check ownership
		return ctx.userID == *ownerID, nil
	}
}

// teamAccessCheck creates a unified check for team access (member/admin/public)
func teamAccessCheck(requireAdmin bool) permissionCheckFunc {
	return func(ctx *permissionContext) (bool, error) {
		if ctx.teamID == nil {
			return false, fmt.Errorf("team_id is required")
		}

		requiredScopes := []consts.PermissionRule{consts.PermTeamReadAll, consts.PermTeamManageAll}
		if requireAdmin {
			requiredScopes = []consts.PermissionRule{consts.PermTeamManageAll}
		}
		if !ctx.scopeAllowsAnyPermission(requiredScopes) {
			return false, nil
		}

		// Check if system admin (from JWT token, no DB query)
		if ctx.isAdmin {
			return true, nil
		}

		// If admin access required, check team admin status
		if requireAdmin {
			isTeamAdmin, err := ctx.checker.IsUserTeamAdmin(ctx.ctx, ctx.userID, *ctx.teamID)
			if err != nil {
				return false, err
			}
			return isTeamAdmin, nil
		}

		// For member access: check if member OR team is public
		isMember, err := ctx.checker.IsUserInTeam(ctx.ctx, ctx.userID, *ctx.teamID)
		if err == nil && isMember {
			return true, nil
		}

		// Check if team is public
		isPublic, err := ctx.checker.IsTeamPublic(ctx.ctx, *ctx.teamID)
		if err == nil && isPublic {
			return true, nil
		}

		return false, nil
	}
}

// projectAccessCheck creates a check for project access (member/admin/public)
func projectAccessCheck(requireAdmin bool) permissionCheckFunc {
	return func(ctx *permissionContext) (bool, error) {
		if ctx.projectID == nil {
			return false, fmt.Errorf("project_id is required")
		}

		requiredScopes := []consts.PermissionRule{consts.PermProjectReadAll, consts.PermProjectManageAll}
		if requireAdmin {
			requiredScopes = []consts.PermissionRule{consts.PermProjectManageAll}
		}
		if !ctx.scopeAllowsAnyPermission(requiredScopes) {
			return false, nil
		}

		// Check if system admin (from JWT token, no DB query)
		if ctx.isAdmin {
			return true, nil
		}

		// Check project admin status if required
		if requireAdmin {
			isProjectAdmin, err := ctx.checker.IsUserProjectAdmin(ctx.ctx, ctx.userID, *ctx.projectID)
			if err != nil {
				return false, err
			}
			return isProjectAdmin, nil
		}

		// Check if user is project member
		isMember, err := ctx.checker.IsUserInProject(ctx.ctx, ctx.userID, *ctx.projectID)
		if err != nil {
			return false, err
		}
		return isMember, nil
	}
}

// ============================================================================
// Simplified Middleware Constructors using the decorator pattern
// ============================================================================

// RequirePermission creates a middleware that requires specific permission
func RequirePermission(permission consts.PermissionRule) gin.HandlerFunc {
	return withPermissionCheck(singlePermission(permission))
}

// RequireAnyPermission creates a middleware that requires any of the specified permissions
func RequireAnyPermission(permissions []consts.PermissionRule) gin.HandlerFunc {
	return withPermissionCheck(anyPermission(permissions))
}

// RequireAllPermissions creates a middleware that requires all specified permissions
func RequireAllPermissions(permissions []consts.PermissionRule) gin.HandlerFunc {
	return withPermissionCheck(allPermissions(permissions))
}

// RequireOwnership creates a middleware that requires user to be owner of resource
// ownerIDGetter is a function that extracts the owner ID from the permission context
func RequireOwnership(resourceType string, ownerIDGetter func(*permissionContext) (*int, error)) gin.HandlerFunc {
	return withPermissionCheck(ownershipCheck(resourceType, ownerIDGetter))
}

// RequireAdminOrOwnership creates a middleware that requires user to be system admin or owner
// ownerIDGetter is a function that extracts the owner ID from the permission context
func RequireAdminOrOwnership(resourceType string, ownerIDGetter func(*permissionContext) (*int, error)) gin.HandlerFunc {
	return withPermissionCheck(adminOrOwnership(resourceType, ownerIDGetter))
}

// RequireSystemAdmin creates a middleware that requires user to be system admin (from JWT claim)
func RequireSystemAdmin() gin.HandlerFunc {
	return withPermissionCheck(func(ctx *permissionContext) (bool, error) {
		return ctx.isAdmin, nil
	})
}

// RequireTeamAccess creates a middleware for team access control
// If requireAdmin is true, checks for team admin OR system admin
// If requireAdmin is false, checks for team member OR public team OR system admin
func RequireTeamAccess(requireAdmin bool) gin.HandlerFunc {
	return withPermissionCheck(teamAccessCheck(requireAdmin))
}

// RequireProjectAccess creates a middleware for project access control
// If requireAdmin is true, checks for project admin OR system admin
// If requireAdmin is false, checks for project member OR system admin
func RequireProjectAccess(requireAdmin bool) gin.HandlerFunc {
	return withPermissionCheck(projectAccessCheck(requireAdmin))
}

// ============================================================================
// Common Permission Variables
// ============================================================================

// Common permission middlewares with inheritance support
// Higher level permissions (manage) inherit lower level permissions (read, update, delete, etc.)
var (
	// System administration permissions
	// manage inherits all other permissions
	RequireSystemRead      = RequireAnyPermission([]consts.PermissionRule{consts.PermSystemRead, consts.PermSystemConfigure, consts.PermSystemManage})
	RequireSystemConfigure = RequireAnyPermission([]consts.PermissionRule{consts.PermSystemConfigure, consts.PermSystemManage})

	// Audit permissions
	RequireAuditRead  = RequireAnyPermission([]consts.PermissionRule{consts.PermAuditRead, consts.PermAuditAudit})
	RequireAuditAudit = RequirePermission(consts.PermAuditAudit)

	// Configuration management permissions
	RequireConfigurationRead      = RequireAnyPermission([]consts.PermissionRule{consts.PermConfigurationRead, consts.PermConfigurationUpdate, consts.PermConfigurationConfigure})
	RequireConfigurationUpdate    = RequireAnyPermission([]consts.PermissionRule{consts.PermConfigurationUpdate, consts.PermConfigurationConfigure})
	RequireConfigurationConfigure = RequirePermission(consts.PermConfigurationConfigure)

	// Resource ownership middlewares - using owner ID getters
	// Example: Get user ID from URL parameter
	getUserIDFromURL = func(ctx *permissionContext) (*int, error) {
		// This is a simple example - in production, extract from gin.Context
		// For now, we just check if the current user is accessing their own resource
		return &ctx.userID, nil
	}
	RequireUserOwnership        = RequireOwnership("user", getUserIDFromURL)
	RequireAdminOrUserOwnership = RequireAdminOrOwnership("user", getUserIDFromURL)

	// Team-specific access control (unified with ownership pattern)
	// These check membership/admin status for the specific team in the URL
	RequireTeamAdminAccess  = RequireTeamAccess(true)  // Team admin OR system admin
	RequireTeamMemberAccess = RequireTeamAccess(false) // Member OR public OR system admin

	// Project-specific access control
	RequireProjectAdminAccess  = RequireProjectAccess(true)  // Project admin OR system admin
	RequireProjectMemberAccess = RequireProjectAccess(false) // Member OR system admin

	// User management permissions (manage inherits all others)
	RequireUserRead   = RequireAnyPermission([]consts.PermissionRule{consts.PermUserReadAll})
	RequireUserCreate = RequireAnyPermission([]consts.PermissionRule{consts.PermUserCreateAll})
	RequireUserUpdate = RequireAnyPermission([]consts.PermissionRule{consts.PermUserUpdateAll})
	RequireUserDelete = RequireAnyPermission([]consts.PermissionRule{consts.PermUserDeleteAll})
	RequireUserAssign = RequireAnyPermission([]consts.PermissionRule{consts.PermUserAssignAll})

	// Role management permissions
	RequireRoleRead   = RequireAnyPermission([]consts.PermissionRule{consts.PermRoleReadAll})
	RequireRoleCreate = RequireAnyPermission([]consts.PermissionRule{consts.PermRoleCreateAll})
	RequireRoleUpdate = RequireAnyPermission([]consts.PermissionRule{consts.PermRoleUpdateAll})
	RequireRoleDelete = RequireAnyPermission([]consts.PermissionRule{consts.PermRoleDeleteAll})
	RequireRoleGrant  = RequireAnyPermission([]consts.PermissionRule{consts.PermRoleGrantAll})
	RequireRoleRevoke = RequireAnyPermission([]consts.PermissionRule{consts.PermRoleRevokeAll})

	// Permission management permissions
	RequirePermissionRead = RequireAnyPermission([]consts.PermissionRule{consts.PermPermissionReadAll, consts.PermPermissionManageAll})

	// Team management permissions (manage > all CRUD; all scope > team scope)
	RequireTeamRead   = RequireAnyPermission([]consts.PermissionRule{consts.PermTeamReadAll, consts.PermTeamManageAll})
	RequireTeamCreate = RequireAnyPermission([]consts.PermissionRule{consts.PermTeamCreateAll, consts.PermTeamManageAll})
	RequireTeamUpdate = RequireAnyPermission([]consts.PermissionRule{consts.PermTeamUpdateAll, consts.PermTeamManageAll})
	RequireTeamDelete = RequireAnyPermission([]consts.PermissionRule{consts.PermTeamDeleteAll, consts.PermTeamManageAll})
	RequireTeamManage = RequirePermission(consts.PermTeamManageAll)

	// Project management permissions (manage > all CRUD; all scope > team scope > project scope)
	RequireProjectRead             = RequireAnyPermission([]consts.PermissionRule{consts.PermProjectReadAll, consts.PermProjectReadOwn})
	RequireProjectCreate           = RequireAnyPermission([]consts.PermissionRule{consts.PermProjectCreateOwn, consts.PermProjectCreateTeam})
	RequireProjectUpdate           = RequireAnyPermission([]consts.PermissionRule{consts.PermProjectUpdateAll, consts.PermProjectUpdateOwn})
	RequireProjectDelete           = RequireAnyPermission([]consts.PermissionRule{consts.PermProjectDeleteAll, consts.PermProjectDeleteOwn})
	RequireProjectManage           = RequireAnyPermission([]consts.PermissionRule{consts.PermProjectManageAll, consts.PermProjectManageOwn})
	RequireProjectInjectionExecute = RequirePermission(consts.PermInjectionExecuteProject)
	RequireProjectExecutionExecute = RequirePermission(consts.PermExecutionExecuteProject)

	// Container management permissions (manage > execute > update/delete > read; all scope > team scope > own)
	RequireContainerRead    = RequireAnyPermission([]consts.PermissionRule{consts.PermContainerReadAll, consts.PermContainerReadTeam, consts.PermContainerManageAll})
	RequireContainerCreate  = RequireAnyPermission([]consts.PermissionRule{consts.PermContainerCreateAll, consts.PermContainerCreateTeam, consts.PermContainerCreateOwn, consts.PermContainerManageAll})
	RequireContainerUpdate  = RequireAnyPermission([]consts.PermissionRule{consts.PermContainerUpdateAll, consts.PermContainerUpdateTeam, consts.PermContainerManageAll, consts.PermContainerExecuteAll, consts.PermContainerExecuteTeam})
	RequireContainerDelete  = RequireAnyPermission([]consts.PermissionRule{consts.PermContainerDeleteAll, consts.PermContainerManageAll})
	RequireContainerManage  = RequirePermission(consts.PermContainerManageAll)
	RequireContainerExecute = RequireAnyPermission([]consts.PermissionRule{consts.PermContainerExecuteAll, consts.PermContainerExecuteTeam, consts.PermContainerManageAll})

	// Container Version management permissions
	RequireContainerVersionRead   = RequireAnyPermission([]consts.PermissionRule{consts.PermContainerVersionReadAll, consts.PermContainerVersionReadTeam, consts.PermContainerVersionManageAll})
	RequireContainerVersionCreate = RequireAnyPermission([]consts.PermissionRule{consts.PermContainerVersionCreateAll, consts.PermContainerVersionCreateTeam, consts.PermContainerVersionManageAll})
	RequireContainerVersionUpdate = RequireAnyPermission([]consts.PermissionRule{consts.PermContainerVersionUpdateAll, consts.PermContainerVersionUpdateTeam, consts.PermContainerVersionManageAll})
	RequireContainerVersionDelete = RequireAnyPermission([]consts.PermissionRule{consts.PermContainerVersionDeleteAll, consts.PermContainerVersionManageAll})
	RequireContainerVersionManage = RequirePermission(consts.PermContainerVersionManageAll)
	RequireContainerVersionUpload = RequireAnyPermission([]consts.PermissionRule{consts.PermContainerVersionUploadAll, consts.PermContainerVersionUploadTeam, consts.PermContainerVersionManageAll})

	// Dataset management permissions (manage > update/delete > read; all scope > team scope > own)
	RequireDatasetRead   = RequireAnyPermission([]consts.PermissionRule{consts.PermDatasetReadAll, consts.PermDatasetReadTeam, consts.PermDatasetManageAll})
	RequireDatasetCreate = RequireAnyPermission([]consts.PermissionRule{consts.PermDatasetCreateAll, consts.PermDatasetCreateTeam, consts.PermDatasetCreateOwn, consts.PermDatasetManageAll})
	RequireDatasetUpdate = RequireAnyPermission([]consts.PermissionRule{consts.PermDatasetUpdateAll, consts.PermDatasetUpdateTeam, consts.PermDatasetManageAll})
	RequireDatasetDelete = RequireAnyPermission([]consts.PermissionRule{consts.PermDatasetDeleteAll, consts.PermDatasetManageAll})
	RequireDatasetManage = RequirePermission(consts.PermDatasetManageAll)

	// Dataset Version management permissions
	RequireDatasetVersionRead     = RequireAnyPermission([]consts.PermissionRule{consts.PermDatasetVersionReadAll, consts.PermDatasetVersionReadTeam, consts.PermDatasetVersionManageAll})
	RequireDatasetVersionCreate   = RequireAnyPermission([]consts.PermissionRule{consts.PermDatasetVersionCreateAll, consts.PermDatasetVersionCreateTeam, consts.PermDatasetVersionManageAll})
	RequireDatasetVersionUpdate   = RequireAnyPermission([]consts.PermissionRule{consts.PermDatasetVersionUpdateAll, consts.PermDatasetVersionUpdateTeam, consts.PermDatasetVersionManageAll})
	RequireDatasetVersionDelete   = RequireAnyPermission([]consts.PermissionRule{consts.PermDatasetVersionDeleteAll, consts.PermDatasetVersionManageAll})
	RequireDatasetVersionManage   = RequirePermission(consts.PermDatasetVersionManageAll)
	RequireDatasetVersionDownload = RequireAnyPermission([]consts.PermissionRule{consts.PermDatasetVersionDownloadAll, consts.PermDatasetVersionDownloadTeam, consts.PermDatasetVersionManageAll})

	// Label management permissions
	RequireLabelRead   = RequireAnyPermission([]consts.PermissionRule{consts.PermLabelReadAll})
	RequireLabelCreate = RequireAnyPermission([]consts.PermissionRule{consts.PermLabelCreateAll, consts.PermLabelCreateOwn})
	RequireLabelUpdate = RequireAnyPermission([]consts.PermissionRule{consts.PermLabelUpdateAll})
	RequireLabelDelete = RequireAnyPermission([]consts.PermissionRule{consts.PermLabelDeleteAll})

	// Task management permissions (execute/stop > update/delete > read)
	RequireTaskRead    = RequireAnyPermission([]consts.PermissionRule{consts.PermTaskReadAll})
	RequireTaskCreate  = RequireAnyPermission([]consts.PermissionRule{consts.PermTaskCreateAll, consts.PermTaskExecuteAll})
	RequireTaskUpdate  = RequireAnyPermission([]consts.PermissionRule{consts.PermTaskUpdateAll, consts.PermTaskExecuteAll})
	RequireTaskDelete  = RequireAnyPermission([]consts.PermissionRule{consts.PermTaskDeleteAll, consts.PermTaskExecuteAll})
	RequireTaskExecute = RequireAnyPermission([]consts.PermissionRule{consts.PermTaskExecuteAll})
	RequireTaskStop    = RequireAnyPermission([]consts.PermissionRule{consts.PermTaskStopAll, consts.PermTaskExecuteAll})

	// Trace management permissions (monitor > read)
	RequireTraceRead    = RequireAnyPermission([]consts.PermissionRule{consts.PermTraceReadAll, consts.PermTraceMonitorAll})
	RequireTraceMonitor = RequireAnyPermission([]consts.PermissionRule{consts.PermTraceMonitorAll})
)
