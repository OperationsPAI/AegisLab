package repository

import (
	"errors"
	"fmt"

	"aegis/database"

	"gorm.io/gorm"
)

// CreateRole creates a role
func CreateRole(role *database.Role) error {
	if err := database.DB.Create(role).Error; err != nil {
		return fmt.Errorf("failed to create role: %v", err)
	}
	return nil
}

// GetRoleByID gets role by ID
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

// GetRoleByName gets role by name
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

// GetRolePermissionsMap gets all permissions for multiple roles in batch (optimized)
func GetRolePermissionsMap(roleIDs []int) (map[int][]database.Permission, error) {
	if len(roleIDs) == 0 {
		return make(map[int][]database.Permission), nil
	}

	var relations []database.RolePermission
	if err := database.DB.Preload("Permission").
		Where("role_id IN ?", roleIDs).
		Find(&relations).Error; err != nil {
		return nil, fmt.Errorf("failed to get role permission relations: %v", err)
	}

	permissionsMap := make(map[int][]database.Permission)
	for _, relation := range relations {
		if relation.Permission != nil {
			permissionsMap[relation.RoleID] = append(permissionsMap[relation.RoleID], *relation.Permission)
		}
	}

	for _, id := range roleIDs {
		if _, exists := permissionsMap[id]; !exists {
			permissionsMap[id] = []database.Permission{}
		}
	}

	return permissionsMap, nil
}

// UpdateRole updates role information
func UpdateRole(role *database.Role) error {
	if err := database.DB.Save(role).Error; err != nil {
		return fmt.Errorf("failed to update role: %v", err)
	}
	return nil
}

// DeleteRole soft deletes role (sets status to -1)
func DeleteRole(id int) error {
	if err := database.DB.Model(&database.Role{}).Where("id = ?", id).Update("status", -1).Error; err != nil {
		return fmt.Errorf("failed to delete role: %v", err)
	}
	return nil
}

// ListRoles gets role list
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

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count roles: %v", err)
	}

	// Paginated query
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&roles).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list roles: %v", err)
	}

	return roles, total, nil
}

// GetRolePermissions gets role permissions
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

// AssignPermissionToRole assigns permission to role
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

// RemovePermissionFromRole removes permission from role
func RemovePermissionFromRole(roleID, permissionID int) error {
	if err := database.DB.Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Delete(&database.RolePermission{}).Error; err != nil {
		return fmt.Errorf("failed to remove permission from role: %v", err)
	}
	return nil
}

// GetRoleUsers gets users who have this role
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

// GetSystemRoles gets system roles
func GetSystemRoles() ([]database.Role, error) {
	var roles []database.Role
	if err := database.DB.Where("is_system = true AND status = 1").
		Order("created_at ASC").Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("failed to get system roles: %v", err)
	}
	return roles, nil
}
