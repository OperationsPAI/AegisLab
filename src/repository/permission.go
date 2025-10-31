package repository

import (
	"errors"
	"fmt"
	"time"

	"aegis/consts"
	"aegis/database"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BatchUpsertPermissions performs batch upsert of permissions
func BatchUpsertPermissions(db *gorm.DB, perimissons []database.Permission) error {
	if len(perimissons) == 0 {
		return fmt.Errorf("no permissions to upsert")
	}

	if err := db.Omit(commonOmitFields).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoNothing: true,
	}).Create(&perimissons).Error; err != nil {
		return fmt.Errorf("failed to batch upsert permissions: %v", err)
	}

	return nil
}

// CreatePermission creates a permission
func CreatePermission(db *gorm.DB, permission *database.Permission) error {
	return createModel(db.Omit(commonOmitFields), permission)
}

// GetPermissionByID gets permission by ID
func GetPermissionByID(db *gorm.DB, id int) (*database.Permission, error) {
	return findModel[database.Permission](db.Preload("Resource"), "id = ? and status != ?", id, consts.CommonDeleted)
}

// GetPermissionByName gets permission by name
func GetPermissionByName(db *gorm.DB, name string) (*database.Permission, error) {
	return findModel[database.Permission](db.Preload("Resource"), "name = ? and status != ?", name, consts.CommonDeleted)
}

// GetPermissionsByAction gets permissions by action
func GetPermissionsByAction(db *gorm.DB, action string) ([]database.Permission, error) {
	var permissions []database.Permission
	if err := db.Preload("Resource").
		Where("action = ? AND status = ?", action, consts.CommonEnabled).
		Order("name").
		Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get permissions by action: %v", err)
	}
	return permissions, nil
}

// GetPermissionsByResource gets permissions by resource
func GetPermissionsByResource(db *gorm.DB, resourceID int) ([]database.Permission, error) {
	var permissions []database.Permission
	if err := db.
		Where("resource_id = ? AND status = ?", resourceID, consts.CommonEnabled).
		Order("action").
		Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get permissions by resource: %v", err)
	}
	return permissions, nil
}

// GetSystemPermissions gets system permissions
func GetSystemPermissions(db *gorm.DB) ([]database.Permission, error) {
	var permissions []database.Permission
	if err := db.Preload("Resource").
		Where("is_system = ? AND status = ?", true, consts.CommonEnabled).
		Order("resource_id, action").
		Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get system permissions: %v", err)
	}
	return permissions, nil
}

// UpdatePermission updates permission information
func UpdatePermission(db *gorm.DB, permission *database.Permission) error {
	return updateModel(db.Omit(commonOmitFields), permission)
}

// ListPermissions gets permission list
func ListPermissions(db *gorm.DB, limit, offset int, action string, isSystem *bool, status *int) ([]database.Permission, int64, error) {
	var permissions []database.Permission
	var total int64

	query := db.Model(&database.Permission{})
	if action != "" {
		query = query.Where("action = ?", action)
	}
	if isSystem != nil {
		query = query.Where("is_system = ?", *isSystem)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count permissions: %v", err)
	}

	if err := query.Limit(limit).Offset(offset).Order("updated_at DESC").Find(&permissions).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list permissions: %v", err)
	}

	return permissions, total, nil
}

// CheckUserPermission checks if user has specific permission
func CheckUserPermission(db *gorm.DB, userID int, action string, resourceName string, projectID, containerID *int) (bool, error) {
	// Find the target permission
	var permission database.Permission
	if err := database.DB.
		Select("permissions.*").
		Joins("JOIN resources ON permissions.resource_id = resources.id").
		Where("permissions.action = ? AND resources.name = ?", action, resourceName).
		First(&permission).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, fmt.Errorf("failed to find target permission: %w", err)
	}
	permissionID := permission.ID

	// Build queries to check direct and role-based permissions
	directQuery := db.
		Select("up.permission_id").
		Table("user_permissions up").
		Where("up.user_id = ? AND up.permission_id = ?", userID, permissionID).
		Where("up.grant_type = 'grant'").
		Where("up.expires_at IS NULL OR up.expires_at > ?", time.Now())

	if projectID != nil {
		directQuery = directQuery.Where("up.project_id IS NULL OR up.project_id = ?", *projectID)
	} else {
		directQuery = directQuery.Where("up.project_id IS NULL")
	}

	if containerID != nil {
		directQuery = directQuery.Where("up.container_id IS NULL OR up.container_id = ?", *containerID)
	} else {
		directQuery = directQuery.Where("up.container_id IS NULL")
	}

	globalRoleQuery := db.
		Select("rp.permission_id").
		Table("role_permissions rp").
		Joins("JOIN user_roles ur ON rp.role_id = ur.role_id").
		Where("ur.user_id = ? AND rp.permission_id = ?", userID, permissionID)

	finalQuery := db.Table("(? UNION ALL ?) as fixed", directQuery, globalRoleQuery)

	// Project role permissions
	if projectID != nil {
		projectRoleQuery := db.
			Select("rp.permission_id").
			Table("role_permissions rp").
			Joins("JOIN user_projects upr ON rp.role_id = upr.role_id").
			Where("upr.user_id = ? AND upr.project_id = ? AND rp.permission_id = ?",
				userID, *projectID, permissionID).
			Where("upr.status = ?", consts.CommonEnabled)

		finalQuery = db.Table("(? UNION ALL ?) as extra", finalQuery, projectRoleQuery)
	}

	// Container role permissions
	if containerID != nil {
		containerRoleQuery := db.
			Select("rp.permission_id").
			Table("role_permissions rp").
			Joins("JOIN user_containers uc ON rp.role_id = uc.role_id").
			Where("uc.user_id = ? AND uc.container_id = ? AND rp.permission_id = ?",
				userID, *containerID, permissionID).
			Where("uc.status = ?", consts.CommonEnabled)

		finalQuery = db.Table("(? UNION ALL ?) as extra", finalQuery, containerRoleQuery)
	}

	var count int64
	if err := finalQuery.Limit(1).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check user permission: %w", err)
	}

	return count > 0, nil
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

// GetPermissionRoles retrieves all roles that have a specific permission
func GetPermissionRoles(db *gorm.DB, permissionID int) ([]database.Role, error) {
	var roles []database.Role

	if err := db.Table("roles").
		Joins("JOIN role_permissions ON roles.id = role_permissions.role_id").
		Where("role_permissions.permission_id = ? AND roles.status != ?", permissionID, consts.CommonDeleted).
		Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("failed to get roles for permission %d: %v", permissionID, err)
	}

	return roles, nil
}
