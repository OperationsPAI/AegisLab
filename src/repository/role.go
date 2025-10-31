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
	return createModel(db.Omit(commonOmitFields), role)
}

// GetRoleByID gets role by ID
func GetRoleByID(db *gorm.DB, id int) (*database.Role, error) {
	return findModel[database.Role](db, "id = ? and status != ?", id, consts.CommonDeleted)
}

// GetRoleByName gets role by name
func GetRoleByName(db *gorm.DB, name string) (*database.Role, error) {
	return findModel[database.Role](db, "name = ? and status != ?", name, consts.CommonDeleted)
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

// GetRoleUsers gets users who have this role
func GetRoleUsers(db *gorm.DB, roleID int) ([]database.User, error) {
	var users []database.User
	if err := db.Table("users").
		Joins("JOIN user_roles ON users.id = user_roles.user_id").
		Where("user_roles.role_id = ? AND users.status = ?", roleID, consts.CommonEnabled).
		Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to get role users: %v", err)
	}
	return users, nil
}

// GetRoleUserCount gets count of users who have this role
func GetRoleUserCount(db *gorm.DB, roleID int) (int64, error) {
	var count int64
	if err := db.Table("users").
		Joins("JOIN user_roles ON users.id = user_roles.user_id").
		Where("user_roles.role_id = ? AND users.status = ?", roleID, consts.CommonEnabled).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to get role users: %v", err)
	}
	return count, nil
}

// GetSystemRoles gets system roles
func GetSystemRoles(db *gorm.DB) ([]database.Role, error) {
	var roles []database.Role
	if err := db.Where("is_system = ? AND status = ?", true, consts.CommonEnabled).
		Order("created_at ASC").Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("failed to get system roles: %v", err)
	}
	return roles, nil
}

// ListRoles gets role list
func ListRoles(db *gorm.DB, limit, offset int, isSystem *bool, status *int) ([]database.Role, int64, error) {
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

// UpdateRole updates role information
func UpdateRole(db *gorm.DB, role *database.Role) error {
	return updateModel(db.Omit(commonOmitFields), role)
}
