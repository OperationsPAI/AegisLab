package repository

import (
	"fmt"

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
	if err := db.Omit(commonOmitFields).Create(permission).Error; err != nil {
		return fmt.Errorf("failed to create permission: %w", err)
	}
	return nil
}

// DeletePermission soft deletes a permission by setting its status to deleted
func DeletePermission(db *gorm.DB, permissionID int) (int64, error) {
	result := db.Model(&database.Permission{}).
		Where("id = ? AND status != ?", permissionID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete permission %d: %w", permissionID, result.Error)
	}
	return result.RowsAffected, nil
}

// GetPermissionByID gets permission by ID
func GetPermissionByID(db *gorm.DB, id int) (*database.Permission, error) {
	var permission database.Permission
	if err := db.Preload("Resource").Where("id = ? and status != ?", id, consts.CommonDeleted).First(&permission).Error; err != nil {
		return nil, fmt.Errorf("failed to find permission with id %d: %w", id, err)
	}
	return &permission, nil
}

// GetPermissionByName gets permission by name
func GetPermissionByName(db *gorm.DB, name string) (*database.Permission, error) {
	var permission database.Permission
	if err := db.Preload("Resource").Where("name = ? and status != ?", name, consts.CommonDeleted).First(&permission).Error; err != nil {
		return nil, fmt.Errorf("failed to find permission with name %s: %w", name, err)
	}
	return &permission, nil
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

// ListPermissions gets permission list
func ListPermissions(db *gorm.DB, limit, offset int, action consts.ActionName, isSystem *bool, status *consts.StatusType) ([]database.Permission, int64, error) {
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

// ListPermissionsByID lists permissions by their IDs
func ListPermissionsByID(db *gorm.DB, permissionIDs []int) ([]database.Permission, error) {
	if len(permissionIDs) == 0 {
		return []database.Permission{}, nil
	}

	var permissions []database.Permission
	if err := db.
		Where("id IN (?) AND status = ?", permissionIDs, consts.CommonEnabled).
		Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to query permissions: %w", err)
	}

	return permissions, nil
}

// ListPermissionsByNames lists permissions by their names
func ListPermissionsByNames(db *gorm.DB, permissionNames []string) ([]database.Permission, error) {
	if len(permissionNames) == 0 {
		return []database.Permission{}, nil
	}

	var permissions []database.Permission
	if err := db.
		Where("name IN (?) AND status = ?", permissionNames, consts.CommonEnabled).
		Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to query permissions: %w", err)
	}

	return permissions, nil
}

// ListSystemPermissions gets system permissions
func ListSystemPermissions(db *gorm.DB) ([]database.Permission, error) {
	var permissions []database.Permission
	if err := db.Where("is_system = ? AND status = ?", true, consts.CommonEnabled).
		Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get system permissions: %v", err)
	}
	return permissions, nil
}

// UpdatePermission updates permission information
func UpdatePermission(db *gorm.DB, permission *database.Permission) error {
	if err := db.Omit(commonOmitFields).Save(permission).Error; err != nil {
		return fmt.Errorf("failed to update permission: %w", err)
	}
	return nil
}

// GetPermissionRoles retrieves all roles that have a specific permission
func ListRolesByPermissionID(db *gorm.DB, permissionID int) ([]database.Role, error) {
	var roles []database.Role

	if err := db.Table("roles").
		Joins("JOIN role_permissions ON roles.id = role_permissions.role_id").
		Where("role_permissions.permission_id = ? AND roles.status != ?", permissionID, consts.CommonDeleted).
		Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("failed to get roles for permission %d: %v", permissionID, err)
	}

	return roles, nil
}
