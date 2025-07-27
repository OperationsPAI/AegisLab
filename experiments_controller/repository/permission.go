package repository

import (
	"errors"
	"fmt"

	"github.com/LGU-SE-Internal/rcabench/database"
	"gorm.io/gorm"
)

// CreatePermission 创建权限
func CreatePermission(permission *database.Permission) error {
	if err := database.DB.Create(permission).Error; err != nil {
		return fmt.Errorf("failed to create permission: %v", err)
	}
	return nil
}

// GetPermissionByID 根据ID获取权限
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

// GetPermissionByName 根据名称获取权限
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

// UpdatePermission 更新权限信息
func UpdatePermission(permission *database.Permission) error {
	if err := database.DB.Save(permission).Error; err != nil {
		return fmt.Errorf("failed to update permission: %v", err)
	}
	return nil
}

// DeletePermission 软删除权限（设置状态为-1）
func DeletePermission(id int) error {
	if err := database.DB.Model(&database.Permission{}).Where("id = ?", id).Update("status", -1).Error; err != nil {
		return fmt.Errorf("failed to delete permission: %v", err)
	}
	return nil
}

// ListPermissions 获取权限列表
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

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count permissions: %v", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&permissions).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list permissions: %v", err)
	}

	return permissions, total, nil
}

// GetPermissionsByResource 根据资源获取权限
func GetPermissionsByResource(resourceID int) ([]database.Permission, error) {
	var permissions []database.Permission
	if err := database.DB.Where("resource_id = ? AND status = 1", resourceID).
		Order("action").Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get permissions by resource: %v", err)
	}
	return permissions, nil
}

// GetPermissionsByAction 根据动作获取权限
func GetPermissionsByAction(action string) ([]database.Permission, error) {
	var permissions []database.Permission
	if err := database.DB.Preload("Resource").Where("action = ? AND status = 1", action).
		Order("name").Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get permissions by action: %v", err)
	}
	return permissions, nil
}

// GetSystemPermissions 获取系统权限
func GetSystemPermissions() ([]database.Permission, error) {
	var permissions []database.Permission
	if err := database.DB.Preload("Resource").Where("is_system = true AND status = 1").
		Order("resource_id, action").Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get system permissions: %v", err)
	}
	return permissions, nil
}

// CheckUserPermission 检查用户是否有特定权限
func CheckUserPermission(userID int, action string, resourceName string, projectID *int) (bool, error) {
	// 1. 先检查直接权限
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

	// 2. 检查角色权限（全局角色）
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

	// 3. 检查项目角色权限
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

// GetUserPermissions 获取用户的所有权限
func GetUserPermissions(userID int, projectID *int) ([]database.Permission, error) {
	var permissions []database.Permission

	// 构建基础查询
	baseQuery := `
		SELECT DISTINCT p.* FROM permissions p
		WHERE p.status = 1 AND (
			-- 直接权限
			p.id IN (
				SELECT up.permission_id FROM user_permissions up 
				WHERE up.user_id = ? AND up.grant_type = 'grant'
				AND (up.expires_at IS NULL OR up.expires_at > NOW())
				%s
			)
			-- 全局角色权限
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

	// 处理项目级权限
	var projectCondition string
	var projectRoleCondition string

	if projectID != nil {
		projectCondition = "AND (up.project_id = ? OR up.project_id IS NULL)"
		projectRoleCondition = `
			-- 项目角色权限
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

// GrantPermissionToUser 给用户授予直接权限
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

// RevokePermissionFromUser 撤销用户的直接权限
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
