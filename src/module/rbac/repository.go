package rbacmodule

import (
	"aegis/consts"
	"aegis/model"
	"fmt"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) withDB(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Transaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}

func (r *Repository) createRoleRecord(role *model.Role) error {
	if err := r.db.Create(role).Error; err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}
	return nil
}

func (r *Repository) deleteRoleCascade(roleID int) (int64, error) {
	role, err := r.loadRole(roleID)
	if err != nil {
		return 0, err
	}
	if role.IsSystem {
		return 0, fmt.Errorf("%w: cannot delete system role", consts.ErrPermissionDenied)
	}

	if err := r.db.Model(&model.UserContainer{}).
		Where("role_id = ? AND status != ?", role.ID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted).Error; err != nil {
		return 0, fmt.Errorf("failed to remove containers with role: %w", err)
	}
	if err := r.db.Model(&model.UserDataset{}).
		Where("role_id = ? AND status != ?", role.ID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted).Error; err != nil {
		return 0, fmt.Errorf("failed to remove datasets with role: %w", err)
	}
	if err := r.db.Model(&model.UserProject{}).
		Where("role_id = ? AND status != ?", role.ID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted).Error; err != nil {
		return 0, fmt.Errorf("failed to remove projects with role: %w", err)
	}
	if err := r.db.Where("role_id = ?", role.ID).Delete(&model.RolePermission{}).Error; err != nil {
		return 0, fmt.Errorf("failed to remove permissions with role: %w", err)
	}
	if err := r.db.Where("role_id = ?", role.ID).Delete(&model.UserRole{}).Error; err != nil {
		return 0, fmt.Errorf("failed to remove users with role: %w", err)
	}

	result := r.db.Model(&model.Role{}).
		Where("id = ? AND status != ?", role.ID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete role %d: %w", role.ID, result.Error)
	}
	return result.RowsAffected, nil
}

func (r *Repository) loadRoleDetail(roleID int) (*model.Role, int64, []model.Permission, error) {
	role, err := r.loadRole(roleID)
	if err != nil {
		return nil, 0, nil, err
	}

	var userCount int64
	if err := r.db.Table("users").
		Joins("JOIN user_roles ON users.id = user_roles.user_id").
		Where("user_roles.role_id = ? AND users.status = ?", role.ID, consts.CommonEnabled).
		Count(&userCount).Error; err != nil {
		return nil, 0, nil, fmt.Errorf("failed to get role user count: %w", err)
	}

	var permissions []model.Permission
	if err := r.db.Table("permissions").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role_id = ? AND permissions.status = ?", role.ID, consts.CommonEnabled).
		Find(&permissions).Error; err != nil {
		return nil, 0, nil, fmt.Errorf("failed to get role permissions: %w", err)
	}

	return role, userCount, permissions, nil
}

func (r *Repository) listRoleViews(limit, offset int, isSystem *bool, status *consts.StatusType) ([]model.Role, int64, error) {
	var roles []model.Role
	var total int64

	query := r.db.Model(&model.Role{})
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

func (r *Repository) updateMutableRole(roleID int, patch func(*model.Role)) (*model.Role, error) {
	role, err := r.loadRole(roleID)
	if err != nil {
		return nil, err
	}
	if role.IsSystem {
		return nil, fmt.Errorf("%w: cannot update system role", consts.ErrPermissionDenied)
	}

	patch(role)
	if err := r.db.Omit("ActiveName").Save(role).Error; err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}
	return role, nil
}

func (r *Repository) loadAssignablePermissions(permissionIDs []int) (map[int]model.Permission, error) {
	if len(permissionIDs) == 0 {
		return map[int]model.Permission{}, nil
	}

	unique := make(map[int]struct{}, len(permissionIDs))
	for _, id := range permissionIDs {
		unique[id] = struct{}{}
	}

	deduplicatedIDs := make([]int, 0, len(unique))
	for id := range unique {
		deduplicatedIDs = append(deduplicatedIDs, id)
	}

	permissions, err := r.listPermissionsByIDs(deduplicatedIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions by ids: %w", err)
	}

	result := make(map[int]model.Permission, len(permissions))
	for _, permission := range permissions {
		result[permission.ID] = permission
	}
	return result, nil
}

func (r *Repository) AssignRolePermissions(roleID int, permissionIDs []int) error {
	role, err := r.loadRole(roleID)
	if err != nil {
		return err
	}
	if role.IsSystem {
		return fmt.Errorf("%w: cannot assign permissions to system role", consts.ErrPermissionDenied)
	}

	permissionMap, err := r.loadAssignablePermissions(permissionIDs)
	if err != nil {
		return err
	}

	rolePermissions := make([]model.RolePermission, 0, len(permissionIDs))
	for _, permissionID := range permissionIDs {
		if _, exists := permissionMap[permissionID]; !exists {
			return fmt.Errorf("%w: permission id %d not found", consts.ErrNotFound, permissionID)
		}
		rolePermissions = append(rolePermissions, model.RolePermission{
			RoleID:       role.ID,
			PermissionID: permissionID,
		})
	}

	if len(rolePermissions) == 0 {
		return nil
	}
	if err := r.db.Create(&rolePermissions).Error; err != nil {
		return fmt.Errorf("failed to batch create role permissions: %w", err)
	}
	return nil
}

func (r *Repository) RemoveRolePermissions(roleID int, permissionIDs []int) error {
	role, err := r.loadRole(roleID)
	if err != nil {
		return err
	}
	if role.IsSystem {
		return fmt.Errorf("%w: cannot remove permissions of system role", consts.ErrPermissionDenied)
	}

	permissionMap, err := r.loadAssignablePermissions(permissionIDs)
	if err != nil {
		return err
	}
	for _, permissionID := range permissionIDs {
		if _, exists := permissionMap[permissionID]; !exists {
			return fmt.Errorf("%w: permission id %d not found", consts.ErrNotFound, permissionID)
		}
	}

	if len(permissionIDs) == 0 {
		return nil
	}
	if err := r.db.Where("role_id = ? AND permission_id IN (?)", role.ID, permissionIDs).
		Delete(&model.RolePermission{}).Error; err != nil {
		return fmt.Errorf("failed to batch delete role permissions: %w", err)
	}
	return nil
}

func (r *Repository) listUsersFromRole(roleID int) (*model.Role, []model.User, error) {
	role, err := r.loadRole(roleID)
	if err != nil {
		return nil, nil, err
	}

	var users []model.User
	if err := r.db.Table("users").
		Joins("JOIN user_roles ON users.id = user_roles.user_id").
		Where("user_roles.role_id = ? AND users.status = ?", role.ID, consts.CommonEnabled).
		Find(&users).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to get role users: %w", err)
	}
	return role, users, nil
}

func (r *Repository) getPermissionDetail(permissionID int) (*model.Permission, error) {
	var permission model.Permission
	if err := r.db.Preload("Resource").
		Where("id = ? and status != ?", permissionID, consts.CommonDeleted).
		First(&permission).Error; err != nil {
		return nil, fmt.Errorf("failed to find permission with id %d: %w", permissionID, err)
	}
	return &permission, nil
}

func (r *Repository) listPermissionViews(limit, offset int, action consts.ActionName, isSystem *bool, status *consts.StatusType) ([]model.Permission, int64, error) {
	var permissions []model.Permission
	var total int64

	query := r.db.Model(&model.Permission{})
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

func (r *Repository) listRolesFromPermission(permissionID int) (*model.Permission, []model.Role, error) {
	permission, err := r.getPermissionDetail(permissionID)
	if err != nil {
		return nil, nil, err
	}

	var roles []model.Role
	if err := r.db.Table("roles").
		Joins("JOIN role_permissions ON roles.id = role_permissions.role_id").
		Where("role_permissions.permission_id = ? AND roles.status != ?", permission.ID, consts.CommonDeleted).
		Find(&roles).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to get permission roles: %w", err)
	}
	return permission, roles, nil
}

func (r *Repository) getResourceDetail(resourceID int) (*model.Resource, error) {
	var resource model.Resource
	if err := r.db.
		Where("id = ? and status != ?", resourceID, consts.CommonDeleted).
		First(&resource).Error; err != nil {
		return nil, fmt.Errorf("failed to find resource with id %d: %w", resourceID, err)
	}
	return &resource, nil
}

func (r *Repository) listResourceViews(limit, offset int, resourceType *consts.ResourceType, category *consts.ResourceCategory) ([]model.Resource, int64, error) {
	var resources []model.Resource
	var total int64

	query := r.db.Model(&model.Resource{}).Preload("Parent")
	if resourceType != nil {
		query = query.Where("type = ?", *resourceType)
	}
	if category != nil {
		query = query.Where("category = ?", *category)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count resources: %v", err)
	}
	if err := query.Limit(limit).Offset(offset).Find(&resources).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list resources: %v", err)
	}
	return resources, total, nil
}

func (r *Repository) listResourcePermissions(resourceID int) (*model.Resource, []model.Permission, error) {
	resource, err := r.getResourceDetail(resourceID)
	if err != nil {
		return nil, nil, err
	}

	var permissions []model.Permission
	if err := r.db.
		Where("resource_id = ? AND status = ?", resource.ID, consts.CommonEnabled).
		Order("action").
		Find(&permissions).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to get permissions by resource: %w", err)
	}
	return resource, permissions, nil
}

func (r *Repository) loadRole(roleID int) (*model.Role, error) {
	var role model.Role
	if err := r.db.Where("id = ? and status != ?", roleID, consts.CommonDeleted).First(&role).Error; err != nil {
		return nil, fmt.Errorf("failed to find role with id %d: %w", roleID, err)
	}
	return &role, nil
}

func (r *Repository) listPermissionsByIDs(permissionIDs []int) ([]model.Permission, error) {
	if len(permissionIDs) == 0 {
		return []model.Permission{}, nil
	}

	var permissions []model.Permission
	if err := r.db.Where("id IN (?) AND status = ?", permissionIDs, consts.CommonEnabled).
		Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to query permissions: %w", err)
	}
	return permissions, nil
}
