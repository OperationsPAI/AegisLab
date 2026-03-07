package repository

import (
	"fmt"
	"time"

	"aegis/consts"
	"aegis/database"
	"aegis/dto"

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
		DoUpdates: clause.AssignmentColumns([]string{}),
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

// GetPermissionByActionAndResource gets permission by action and resource name
func GetPermissionByActionAndResource(db *gorm.DB, action consts.ActionName, scope consts.ResourceScope, resourceName consts.ResourceName) (*database.Permission, error) {
	var permission database.Permission
	if err := db.
		Select("permissions.*").
		Joins("JOIN resources ON permissions.resource_id = resources.id").
		Where("permissions.action = ? AND permissions.scope= ? AND resources.name = ?", action, scope, resourceName).
		Where("permissions.status != ?", consts.CommonDeleted).
		First(&permission).Error; err != nil {
		return nil, fmt.Errorf("failed to find permission with action %s and resource %s: %w", action, resourceName, err)
	}
	return &permission, nil
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

// CheckUserHasPermission checks if user has specific permission through various sources
func CheckUserHasPermission(db *gorm.DB, params *dto.CheckPermissionParams, permissionID int) (bool, error) {
	// Build direct permission query
	directQuery := buildDirectPermissionQuery(db, params.UserID, permissionID, params.ProjectID, params.ContainerID, params.DatasetID)

	// Build global role permission query
	globalRoleQuery := buildGlobalRolePermissionQuery(db, params.UserID, permissionID)

	// Combine direct and global role permissions
	finalQuery := db.Table("(? UNION ALL ?) as base", directQuery, globalRoleQuery)

	// Add team role permissions if teamID is provided
	if params.TeamID != nil {
		teamRoleQuery := buildTeamRolePermissionQuery(db, params.UserID, permissionID, *params.TeamID)
		finalQuery = db.Table("(? UNION ALL ?) as combined", finalQuery, teamRoleQuery)
	}

	// Add project role permissions if projectID is provided
	if params.ProjectID != nil {
		projectRoleQuery := buildProjectRolePermissionQuery(db, params.UserID, permissionID, *params.ProjectID)
		finalQuery = db.Table("(? UNION ALL ?) as combined", finalQuery, projectRoleQuery)
	}

	// Add container role permissions if containerID is provided
	if params.ContainerID != nil {
		containerRoleQuery := buildContainerRolePermissionQuery(db, params.UserID, permissionID, *params.ContainerID)
		finalQuery = db.Table("(? UNION ALL ?) as combined", finalQuery, containerRoleQuery)
	}

	// Add dataset role permissions if datasetID is provided
	if params.DatasetID != nil {
		datasetRoleQuery := buildDatasetRolePermissionQuery(db, params.UserID, permissionID, *params.DatasetID)
		finalQuery = db.Table("(? UNION ALL ?) as combined", finalQuery, datasetRoleQuery)
	}

	var count int64
	if err := finalQuery.Limit(1).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check user permission: %w", err)
	}

	return count > 0, nil
}

// buildDirectPermissionQuery builds query for direct user permissions
func buildDirectPermissionQuery(db *gorm.DB, userID int, permissionID int, projectID, containerID, datasetID *int) *gorm.DB {
	query := db.
		Select("up.permission_id").
		Table("user_permissions up").
		Where("up.user_id = ? AND up.permission_id = ?", userID, permissionID).
		Where("up.grant_type = ?", consts.GrantTypeGrant).
		Where("up.expires_at IS NULL OR up.expires_at > ?", time.Now())

	if projectID != nil {
		query = query.Where("up.project_id IS NULL OR up.project_id = ?", *projectID)
	} else {
		query = query.Where("up.project_id IS NULL")
	}

	if containerID != nil {
		query = query.Where("up.container_id IS NULL OR up.container_id = ?", *containerID)
	} else {
		query = query.Where("up.container_id IS NULL")
	}

	if datasetID != nil {
		query = query.Where("up.dataset_id IS NULL OR up.dataset_id = ?", *datasetID)
	} else {
		query = query.Where("up.dataset_id IS NULL")
	}

	return query
}

// buildGlobalRolePermissionQuery builds query for global role permissions
func buildGlobalRolePermissionQuery(db *gorm.DB, userID int, permissionID int) *gorm.DB {
	return db.
		Select("rp.permission_id").
		Table("role_permissions rp").
		Joins("JOIN user_roles ur ON rp.role_id = ur.role_id").
		Where("ur.user_id = ? AND rp.permission_id = ?", userID, permissionID)
}

// buildTeamRolePermissionQuery builds query for team-specific role permissions
func buildTeamRolePermissionQuery(db *gorm.DB, userID int, permissionID int, teamID int) *gorm.DB {
	return db.
		Select("rp.permission_id").
		Table("role_permissions rp").
		Joins("JOIN user_teams ut ON rp.role_id = ut.role_id").
		Where("ut.user_id = ? AND ut.team_id = ? AND rp.permission_id = ?", userID, teamID, permissionID).
		Where("ut.status = ?", consts.CommonEnabled)
}

// buildProjectRolePermissionQuery builds query for project-specific role permissions
func buildProjectRolePermissionQuery(db *gorm.DB, userID int, permissionID int, projectID int) *gorm.DB {
	return db.
		Select("rp.permission_id").
		Table("role_permissions rp").
		Joins("JOIN user_projects upr ON rp.role_id = upr.role_id").
		Where("upr.user_id = ? AND upr.project_id = ? AND rp.permission_id = ?", userID, projectID, permissionID).
		Where("upr.status = ?", consts.CommonEnabled)
}

// buildContainerRolePermissionQuery builds query for container-specific role permissions
func buildContainerRolePermissionQuery(db *gorm.DB, userID int, permissionID int, containerID int) *gorm.DB {
	return db.
		Select("rp.permission_id").
		Table("role_permissions rp").
		Joins("JOIN user_containers uc ON rp.role_id = uc.role_id").
		Where("uc.user_id = ? AND uc.container_id = ? AND rp.permission_id = ?", userID, containerID, permissionID).
		Where("uc.status = ?", consts.CommonEnabled)
}

// buildDatasetRolePermissionQuery builds query for dataset-specific role permissions
func buildDatasetRolePermissionQuery(db *gorm.DB, userID int, permissionID int, datasetID int) *gorm.DB {
	return db.
		Select("rp.permission_id").
		Table("role_permissions rp").
		Joins("JOIN user_datasets ud ON rp.role_id = ud.role_id").
		Where("ud.user_id = ? AND ud.dataset_id = ? AND rp.permission_id = ?", userID, datasetID, permissionID).
		Where("ud.status = ?", consts.CommonEnabled)
}
