package repository

import (
	"errors"
	"fmt"

	"github.com/LGU-SE-Internal/rcabench/database"
	"gorm.io/gorm"
)

// CreateRole 创建角色
func CreateRole(role *database.Role) error {
	if err := database.DB.Create(role).Error; err != nil {
		return fmt.Errorf("failed to create role: %v", err)
	}
	return nil
}

// GetRoleByID 根据ID获取角色
func GetRoleByID(id int) (*database.Role, error) {
	var role database.Role
	if err := database.DB.First(&role, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("role with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get role: %v", err)
	}
	return &role, nil
}

// GetRoleByName 根据名称获取角色
func GetRoleByName(name string) (*database.Role, error) {
	var role database.Role
	if err := database.DB.Where("name = ?", name).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("role '%s' not found", name)
		}
		return nil, fmt.Errorf("failed to get role: %v", err)
	}
	return &role, nil
}

// UpdateRole 更新角色信息
func UpdateRole(role *database.Role) error {
	if err := database.DB.Save(role).Error; err != nil {
		return fmt.Errorf("failed to update role: %v", err)
	}
	return nil
}

// DeleteRole 软删除角色（设置状态为-1）
func DeleteRole(id int) error {
	if err := database.DB.Model(&database.Role{}).Where("id = ?", id).Update("status", -1).Error; err != nil {
		return fmt.Errorf("failed to delete role: %v", err)
	}
	return nil
}

// ListRoles 获取角色列表
func ListRoles(page, pageSize int, roleType string, status *int) ([]database.Role, int64, error) {
	var roles []database.Role
	var total int64

	query := database.DB.Model(&database.Role{})

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if roleType != "" {
		query = query.Where("type = ?", roleType)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count roles: %v", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&roles).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list roles: %v", err)
	}

	return roles, total, nil
}

// GetRolePermissions 获取角色的权限
func GetRolePermissions(roleID int) ([]database.Permission, error) {
	var permissions []database.Permission
	if err := database.DB.Table("permissions").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role_id = ? AND permissions.status = 1", roleID).
		Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %v", err)
	}
	return permissions, nil
}

// AssignPermissionToRole 给角色分配权限
func AssignPermissionToRole(roleID, permissionID int) error {
	rolePermission := &database.RolePermission{
		RoleID:       roleID,
		PermissionID: permissionID,
	}

	if err := database.DB.Create(rolePermission).Error; err != nil {
		return fmt.Errorf("failed to assign permission to role: %v", err)
	}
	return nil
}

// RemovePermissionFromRole 移除角色的权限
func RemovePermissionFromRole(roleID, permissionID int) error {
	if err := database.DB.Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Delete(&database.RolePermission{}).Error; err != nil {
		return fmt.Errorf("failed to remove permission from role: %v", err)
	}
	return nil
}

// GetRoleUsers 获取拥有该角色的用户
func GetRoleUsers(roleID int) ([]database.User, error) {
	var users []database.User
	if err := database.DB.Table("users").
		Joins("JOIN user_roles ON users.id = user_roles.user_id").
		Where("user_roles.role_id = ? AND users.status = 1", roleID).
		Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to get role users: %v", err)
	}
	return users, nil
}

// GetSystemRoles 获取系统角色
func GetSystemRoles() ([]database.Role, error) {
	var roles []database.Role
	if err := database.DB.Where("is_system = true AND status = 1").
		Order("created_at ASC").Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("failed to get system roles: %v", err)
	}
	return roles, nil
}
