package repository

import (
	"errors"
	"fmt"

	"github.com/LGU-SE-Internal/rcabench/database"
	"gorm.io/gorm"
)

// CreatePermission creates a permission
func CreatePermission(permission *database.Permission) error {
	if err := database.DB.Create(permission).Error; err != nil {
		return fmt.Errorf("failed to create permission: %v", err)
	}
	return nil
}

// GetPermissionByID gets permission by ID
func GetPermissionByID(id int) (*database.Permission, error) {
	var permission database.Permission
	if err := database.DB.Preload("Resource").First(&permission, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("permission with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get permission: %v", err)
	}
	return &permission, nil
}

// GetPermissionByName gets permission by name
func GetPermissionByName(name string) (*database.Permission, error) {
	var permission database.Permission
	if err := database.DB.Preload("Resource").Where("name = ?", name).First(&permission).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("permission '%s' not found", name)
		}
		return nil, fmt.Errorf("failed to get permission: %v", err)
	}
	return &permission, nil
}

// UpdatePermission updates permission information
func UpdatePermission(permission *database.Permission) error {
	if err := database.DB.Save(permission).Error; err != nil {
		return fmt.Errorf("failed to update permission: %v", err)
	}
	return nil
}

// DeletePermission soft deletes permission (sets status to -1)
func DeletePermission(id int) error {
	if err := database.DB.Model(&database.Permission{}).Where("id = ?", id).Update("status", -1).Error; err != nil {
		return fmt.Errorf("failed to delete permission: %v", err)
	}
	return nil
}

// ListPermissions gets permission list
func ListPermissions(page, pageSize int, action string, resourceID *int, status *int) ([]database.Permission, int64, error) {
	var permissions []database.Permission
	var total int64

	query := database.DB.Model(&database.Permission{}).Preload("Resource")

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if action != "" {
		query = query.Where("action = ?", action)
	}

	if resourceID != nil {
		query = query.Where("resource_id = ?", *resourceID)
	}

	        // Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count permissions: %v", err)
	}

	        // Paginated query
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&permissions).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list permissions: %v", err)
	}

	return permissions, total, nil
}

// GetPermissionsByResource gets permissions by resource
func GetPermissionsByResource(resourceID int) ([]database.Permission, error) {
	var permissions []database.Permission
	if err := database.DB.Where("resource_id = ? AND status = 1", resourceID).
		Order("action").Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get permissions by resource: %v", err)
	}
	return permissions, nil
}

// GetPermissionsByAction gets permissions by action
func GetPermissionsByAction(action string) ([]database.Permission, error) {
	var permissions []database.Permission
	if err := database.DB.Preload("Resource").Where("action = ? AND status = 1", action).
		Order("name").Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get permissions by action: %v", err)
	}
	return permissions, nil
}

// GetSystemPermissions gets system permissions
func GetSystemPermissions() ([]database.Permission, error) {
	var permissions []database.Permission
	if err := database.DB.Preload("Resource").Where("is_system = true AND status = 1").
		Order("resource_id, action").Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get system permissions: %v", err)
	}
	return permissions, nil
}

// CheckUserPermission checks if user has specific permission
func CheckUserPermission(userID int, action string, resourceName string, projectID *int) (bool, error) {
	       // 1. First check direct permissions
	var directPermCount int64
	directQuery := database.DB.Table("user_permissions").
		Joins("JOIN permissions ON user_permissions.permission_id = permissions.id").
		Joins("JOIN resources ON permissions.resource_id = resources.id").
		Where("user_permissions.user_id = ? AND permissions.action = ? AND resources.name = ?", userID, action, resourceName).
		Where("user_permissions.grant_type = 'grant'").
		Where("user_permissions.expires_at IS NULL OR user_permissions.expires_at > NOW()")

	if projectID != nil {
		directQuery = directQuery.Where("(user_permissions.project_id = ? OR user_permissions.project_id IS NULL)", *projectID)
	} else {
		directQuery = directQuery.Where("user_permissions.project_id IS NULL")
	}

	if err := directQuery.Count(&directPermCount).Error; err != nil {
		return false, fmt.Errorf("failed to check direct permissions: %v", err)
	}

	if directPermCount > 0 {
		return true, nil
	}

	       // 2. Check role permissions (global roles)
	var globalRolePermCount int64
	globalRoleQuery := database.DB.Table("user_roles").
		Joins("JOIN role_permissions ON user_roles.role_id = role_permissions.role_id").
		Joins("JOIN permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN resources ON permissions.resource_id = resources.id").
		Where("user_roles.user_id = ? AND permissions.action = ? AND resources.name = ?", userID, action, resourceName)

	if err := globalRoleQuery.Count(&globalRolePermCount).Error; err != nil {
		return false, fmt.Errorf("failed to check global role permissions: %v", err)
	}

	if globalRolePermCount > 0 {
		return true, nil
	}

	       // 3. Check project role permissions
	if projectID != nil {
		var projectRolePermCount int64
		projectRoleQuery := database.DB.Table("user_projects").
			Joins("JOIN role_permissions ON user_projects.role_id = role_permissions.role_id").
			Joins("JOIN permissions ON role_permissions.permission_id = permissions.id").
			Joins("JOIN resources ON permissions.resource_id = resources.id").
			Where("user_projects.user_id = ? AND user_projects.project_id = ? AND permissions.action = ? AND resources.name = ?",
				userID, *projectID, action, resourceName).
			Where("user_projects.status = 1")

		if err := projectRoleQuery.Count(&projectRolePermCount).Error; err != nil {
			return false, fmt.Errorf("failed to check project role permissions: %v", err)
		}

		if projectRolePermCount > 0 {
			return true, nil
		}
	}

	return false, nil
}

// GetUserPermissions gets all permissions for a user
func GetUserPermissions(userID int, projectID *int) ([]database.Permission, error) {
	var permissions []database.Permission

	       // Build base query
	baseQuery := `
		SELECT DISTINCT p.* FROM permissions p
		WHERE p.status = 1 AND (
			                       -- Direct permissions
			p.id IN (
				SELECT up.permission_id FROM user_permissions up 
				WHERE up.user_id = ? AND up.grant_type = 'grant'
				AND (up.expires_at IS NULL OR up.expires_at > NOW())
				%s
			)
			                       -- Global role permissions
			OR p.id IN (
				SELECT rp.permission_id FROM user_roles ur
				JOIN role_permissions rp ON ur.role_id = rp.role_id
				WHERE ur.user_id = ?
			)
			%s
		)
	`

	var args []interface{}
	args = append(args, userID)

	       // Handle project-level permissions
	var projectCondition string
	var projectRoleCondition string

	if projectID != nil {
		projectCondition = "AND (up.project_id = ? OR up.project_id IS NULL)"
		projectRoleCondition = `
			                       -- Project role permissions
			OR p.id IN (
				SELECT rp.permission_id FROM user_projects up
				JOIN role_permissions rp ON up.role_id = rp.role_id
				WHERE up.user_id = ? AND up.project_id = ? AND up.status = 1
			)
		`
		args = append(args, *projectID, userID, userID, *projectID)
	} else {
		projectCondition = "AND up.project_id IS NULL"
		projectRoleCondition = ""
		args = append(args, userID)
	}

	finalQuery := fmt.Sprintf(baseQuery, projectCondition, projectRoleCondition)

	if err := database.DB.Raw(finalQuery, args...).Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %v", err)
	}

	return permissions, nil
}

// GrantPermissionToUser grants direct permission to user
func GrantPermissionToUser(userID, permissionID int, projectID *int) error {
	userPermission := &database.UserPermission{
		UserID:       userID,
		PermissionID: permissionID,
		ProjectID:    projectID,
		GrantType:    "grant",
	}

	if err := database.DB.Create(userPermission).Error; err != nil {
		return fmt.Errorf("failed to grant permission to user: %v", err)
	}
	return nil
}

// RevokePermissionFromUser revokes direct permission from user
func RevokePermissionFromUser(userID, permissionID int, projectID *int) error {
	query := database.DB.Where("user_id = ? AND permission_id = ?", userID, permissionID)

	if projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	} else {
		query = query.Where("project_id IS NULL")
	}

	if err := query.Delete(&database.UserPermission{}).Error; err != nil {
		return fmt.Errorf("failed to revoke permission from user: %v", err)
	}
	return nil
}

// GetPermissionRoles retrieves all roles that have a specific permission
func GetPermissionRoles(permissionID int) ([]database.Role, error) {
	var roles []database.Role

	err := database.DB.Table("roles").
		Joins("JOIN role_permissions ON roles.id = role_permissions.role_id").
		Where("role_permissions.permission_id = ? AND roles.status != -1", permissionID).
		Find(&roles).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get roles for permission %d: %v", permissionID, err)
	}

	return roles, nil
}

// GetPermissionsByResourcePaginated retrieves permissions filtered by resource with pagination
func GetPermissionsByResourcePaginated(resourceID int, page, pageSize int) ([]database.Permission, int64, error) {
	var permissions []database.Permission
	var total int64

	query := database.DB.Model(&database.Permission{}).Where("resource_id = ? AND status != -1", resourceID)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count permissions for resource %d: %v", resourceID, err)
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&permissions).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get permissions for resource %d: %v", resourceID, err)
	}

	return permissions, total, nil
}

// CountPermissionsByAction returns count of permissions grouped by action
func CountPermissionsByAction() (map[string]int64, error) {
	type ActionCount struct {
		Action string `json:"action"`
		Count  int64  `json:"count"`
	}

	var results []ActionCount
	err := database.DB.Model(&database.Permission{}).
		Select("action, COUNT(*) as count").
		Where("status != -1").
		Group("action").
		Find(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to count permissions by action: %v", err)
	}

	actionCounts := make(map[string]int64)
	for _, result := range results {
		actionCounts[result.Action] = result.Count
	}

	return actionCounts, nil
}
