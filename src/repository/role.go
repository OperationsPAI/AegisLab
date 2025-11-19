package repository

import (
	"fmt"

	"aegis/consts"
	"aegis/database"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BatchUpsertRoles performs batch upsert of roles
func BatchUpsertRoles(db *gorm.DB, roles []database.Role) error {
	if len(roles) == 0 {
		return fmt.Errorf("no roles to upsert")
	}

	if err := db.Omit(commonOmitFields).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoNothing: true,
	},
	).Create(&roles).Error; err != nil {
		return fmt.Errorf("failed to batch upsert roles: %v", err)
	}

	return nil
}

// CreateRole creates a role
func CreateRole(db *gorm.DB, role *database.Role) error {
	if err := db.Omit(commonOmitFields).Create(role).Error; err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}
	return nil
}

// DeleteRole soft deletes a role by setting its status to deleted
func DeleteRole(db *gorm.DB, roleID int) (int64, error) {
	result := db.Model(&database.Role{}).
		Where("id = ? AND status != ?", roleID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete role %d: %w", roleID, result.Error)
	}
	return result.RowsAffected, nil
}

// GetRoleByID gets role by ID
func GetRoleByID(db *gorm.DB, id int) (*database.Role, error) {
	var role database.Role
	if err := db.Where("id = ? and status != ?", id, consts.CommonDeleted).First(&role).Error; err != nil {
		return nil, fmt.Errorf("failed to find role with id %d: %w", id, err)
	}
	return &role, nil
}

// GetRoleByName gets role by name
func GetRoleByName(db *gorm.DB, name consts.RoleName) (*database.Role, error) {
	var role database.Role
	if err := db.
		Where("name = ? and status != ?", name, consts.CommonDeleted).
		First(&role).Error; err != nil {
		return nil, fmt.Errorf("failed to find role with name %s: %w", name, err)
	}
	return &role, nil
}

// GetRolePermissions gets role permissions
func GetRolePermissions(db *gorm.DB, roleID int) ([]database.Permission, error) {
	var permissions []database.Permission
	if err := db.Table("permissions").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role_id = ? AND permissions.status = ?", roleID, consts.CommonEnabled).
		Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %v", err)
	}
	return permissions, nil
}

// ListRoles gets role list
func ListRoles(db *gorm.DB, limit, offset int, isSystem *bool, status *consts.StatusType) ([]database.Role, int64, error) {
	var roles []database.Role
	var total int64

	query := db.Model(&database.Role{})
	if isSystem != nil {
		query = query.Where("is_system = ?", *isSystem)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count roles: %v", err)
	}

	if err := query.Limit(limit).Offset(offset).Order("updated_at DESC").Find(&roles).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list roles: %v", err)
	}

	return roles, total, nil
}

// ListSystemRoles gets system roles
func ListSystemRoles(db *gorm.DB) ([]database.Role, error) {
	var roles []database.Role
	if err := db.Where("is_system = ? AND status = ?", true, consts.CommonEnabled).
		Order("created_at ASC").Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("failed to get system roles: %v", err)
	}
	return roles, nil
}

// UpdateRole updates role information
func UpdateRole(db *gorm.DB, role *database.Role) error {
	if err := db.Omit(commonOmitFields).Save(role).Error; err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}
	return nil
}

// ===================== Role-Permission =====================

// BatchCreateRolePermissions creates multiple role-permission associations in a batch
func BatchCreateRolePermissions(db *gorm.DB, rolePermissions []database.RolePermission) error {
	if len(rolePermissions) == 0 {
		return nil
	}
	if err := db.Create(&rolePermissions).Error; err != nil {
		return fmt.Errorf("failed to batch create role permissions: %w", err)
	}
	return nil
}

// BatchDeleteRolePermisssions deletes multiple role-permission associations in a batch
func BatchDeleteRolePermisssions(db *gorm.DB, roleID int, permissionIDs []int) error {
	if len(permissionIDs) == 0 {
		return nil
	}
	if err := db.Where("role_id = ? AND permission_id IN (?)", roleID, permissionIDs).
		Delete(&database.RolePermission{}).Error; err != nil {
		return fmt.Errorf("failed to batch delete role permissions: %w", err)
	}
	return nil
}

// RemoveRolesFromPermission deletes all role-permission associations associated with a given permission
func RemoveRolesFromPermission(db *gorm.DB, permissionID int) error {
	if err := db.Where("permission_id = ?", permissionID).
		Delete(&database.RolePermission{}).Error; err != nil {
		return fmt.Errorf("failed to remove all roles from permission: %w", err)
	}
	return nil
}

// RemovePermissionsFromRole deletes all role-permission associations associated with a given role
func RemovePermissionsFromRole(db *gorm.DB, roleID int) error {
	if err := db.Where("role_id = ?", roleID).
		Delete(&database.RolePermission{}).Error; err != nil {
		return fmt.Errorf("failed to remove all permissions from role: %w", err)
	}
	return nil
}
