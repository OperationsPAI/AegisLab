package repository

import (
	"errors"
	"fmt"
	"time"

	"aegis/consts"
	"aegis/database"

	"gorm.io/gorm"
)

const (
	userOmitFields          = "active_username"
	userContainerOmitFields = "active_user_container"
	userDatasetOmitFields   = "active_user_dataset"
	userProjectOmitFields   = "active_user_project"
)

// CreateUser creates a user
func CreateUser(db *gorm.DB, user *database.User) error {
	if err := db.Omit(userOmitFields).Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// DeleteUser soft deletes a user by setting its status to deleted
func DeleteUser(db *gorm.DB, userID int) (int64, error) {
	result := db.Model(&database.User{}).
		Where("id = ? AND status != ?", userID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete user %d: %w", userID, result.Error)
	}
	return result.RowsAffected, nil
}

// GetUserByID gets a user by ID
func GetUserByID(db *gorm.DB, id int) (*database.User, error) {
	var user database.User
	if err := db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to find user with id %d: %w", id, err)
	}
	return &user, nil
}

// GetUserByUsername gets a user by username
func GetUserByUsername(db *gorm.DB, username string) (*database.User, error) {
	var user database.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to find user with username %s: %w", username, err)
	}
	return &user, nil
}

// GetUserByEmail gets a user by email
func GetUserByEmail(email string) (*database.User, error) {
	var user database.User
	if err := database.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to find user with email %s: %w", email, err)
	}
	return &user, nil
}

// ListUsers lists users with filters and pagination
func ListUsers(db *gorm.DB, limit, offset int, isActive *bool, status *consts.StatusType) ([]database.User, int64, error) {
	var users []database.User
	var total int64

	query := db.Model(&database.User{}).Where("status != ?", consts.CommonDeleted)
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	if err := query.Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	return users, total, nil
}

// UpdateUser updates user information
func UpdateUser(db *gorm.DB, user *database.User) error {
	if err := db.Omit(userOmitFields).Save(user).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

// UpdateUserLoginTime updates user's last login time
func UpdateUserLoginTime(db *gorm.DB, userID int) error {
	now := db.NowFunc()
	if err := db.Model(&database.User{}).
		Where("id = ? AND status != ?", userID, consts.CommonDeleted).
		Update("last_login_at", now).Error; err != nil {
		return fmt.Errorf("failed to update user login time: %w", err)
	}
	return nil
}

// ===================== User-Role =====================

// CreateUserRole creates a user-role association
func CreateUserRole(db *gorm.DB, userRole *database.UserRole) error {
	if err := db.Create(userRole).Error; err != nil {
		return fmt.Errorf("failed to create user-role association: %w", err)
	}
	return nil
}

// DeleteUserRole deletes a user-role association
func DeleteUserRole(db *gorm.DB, userID, roleID int) error {
	if err := db.Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&database.UserRole{}).Error; err != nil {
		return fmt.Errorf("failed to delete user-role association: %w", err)
	}
	return nil
}

// RemoveUsersFromRole deletes all user-role associations associated with a given role
func RemoveUsersFromRole(db *gorm.DB, roleID int) error {
	if err := db.Where("role_id = ?", roleID).
		Delete(&database.UserRole{}).Error; err != nil {
		return fmt.Errorf("failed to delete all users from role: %w", err)
	}
	return nil
}

// RemoveRolesFromUser deletes all user-role associations associated with a given user
func RemoveRolesFromUser(db *gorm.DB, userID int) error {
	if err := db.Where("user_id = ?", userID).
		Delete(&database.UserRole{}).Error; err != nil {
		return fmt.Errorf("failed to delete all roles from user: %w", err)
	}
	return nil
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

// ListUsersByRoleID gets users who have a specific role
func ListUsersByRoleID(db *gorm.DB, roleID int) ([]database.User, error) {
	var users []database.User
	if err := db.Table("users").
		Joins("JOIN user_roles ON users.id = user_roles.user_id").
		Where("user_roles.role_id = ? AND users.status = ?", roleID, consts.CommonEnabled).
		Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to get role users: %v", err)
	}
	return users, nil
}

// ListRolesByUserID gets roles the user has
func ListRolesByUserID(db *gorm.DB, userID int) ([]database.Role, error) {
	var roles []database.Role
	if err := db.Table("roles").
		Joins("JOIN user_roles ur ON ur.role_id = roles.id").
		Where("ur.user_id = ? AND roles.status = ?", userID, consts.CommonEnabled).
		Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("failed to get global roles of the specific user: %w", err)
	}
	return roles, nil
}

// ===================== User-Permission =====================

// BatchCreateUserPermissions creates multiple user-permission associations in a batch
func BatchCreateUserPermissions(db *gorm.DB, userPermissions []database.UserPermission) error {
	if len(userPermissions) == 0 {
		return nil
	}
	if err := db.Create(&userPermissions).Error; err != nil {
		return fmt.Errorf("failed to batch create user permissions: %w", err)
	}
	return nil
}

// BatchDeleteUserPermisssions deletes multiple user-permission associations in a batch
func BatchDeleteUserPermisssions(db *gorm.DB, userID int, permissionIDs []int) error {
	if len(permissionIDs) == 0 {
		return nil
	}
	if err := db.Where("user_id = ? AND permission_id IN (?)", userID, permissionIDs).
		Delete(&database.UserPermission{}).Error; err != nil {
		return fmt.Errorf("failed to batch delete user permissions: %w", err)
	}
	return nil
}

// CheckUserPermission checks if user has specific permission
func CheckUserPermission(db *gorm.DB, userID int, action string, resourceName string, projectID, containerID, datasetID *int) (bool, error) {
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

	if datasetID != nil {
		directQuery = directQuery.Where("up.dataset_id IS NULL OR up.dataset_id = ?", *datasetID)
	} else {
		directQuery = directQuery.Where("up.dataset_id IS NULL")
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

	// Dataset role permissions
	if datasetID != nil {
		datasetRoleQuery := db.
			Select("rp.permission_id").
			Table("role_permissions rp").
			Joins("JOIN user_datasets ud ON rp.role_id = ud.role_id").
			Where("ud.user_id = ? AND ud.dataset_id = ? AND rp.permission_id = ?",
				userID, *datasetID, permissionID).
			Where("ud.status = ?", consts.CommonEnabled)

		finalQuery = db.Table("(? UNION ALL ?) as extra", finalQuery, datasetRoleQuery)
	}

	var count int64
	if err := finalQuery.Limit(1).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check user permission: %w", err)
	}

	return count > 0, nil
}

// RemoveUsersFromPermission deletes all user-permission associations associated with a given permission
func RemoveUsersFromPermission(db *gorm.DB, permissionID int) error {
	if err := db.Where("permission_id = ?", permissionID).
		Delete(&database.UserPermission{}).Error; err != nil {
		return fmt.Errorf("failed to delete all users from permission: %w", err)
	}
	return nil
}

// RemovePermissionsFromUser deletes all user-permission associations associated with a given user
func RemovePermissionsFromUser(db *gorm.DB, userID int) error {
	if err := db.Where("user_id = ?", userID).
		Delete(&database.UserPermission{}).Error; err != nil {
		return fmt.Errorf("failed to delete all permissions from user: %w", err)
	}
	return nil
}

// ListPermissionsByUserID lists all permissions a user has, including direct and role-based permissions
func ListPermissionsByUserID(db *gorm.DB, userID int) ([]database.Permission, error) {
	var permissions []database.Permission

	// Subquery 1: Get permissions from user's global roles
	rolePermissionsQuery := db.
		Table("permissions p").
		Select("p.*").
		Joins("JOIN role_permissions rp ON p.id = rp.permission_id").
		Joins("JOIN user_roles ur ON rp.role_id = ur.role_id").
		Where("ur.user_id = ? AND p.status = ?", userID, consts.CommonEnabled)

	// Subquery 2: Get direct permissions assigned to user
	directPermissionsQuery := db.
		Table("permissions p").
		Select("p.*").
		Joins("JOIN user_permissions up ON p.id = up.permission_id").
		Where("up.user_id = ? AND p.status = ?", userID, consts.CommonEnabled).
		Where("up.grant_type = ?", consts.GrantTypeGrant).
		Where("up.expires_at IS NULL OR up.expires_at > ?", time.Now())

	// Union both queries and get distinct permissions
	if err := db.Table("(?) UNION (?)", rolePermissionsQuery, directPermissionsQuery).
		Scan(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	return permissions, nil
}

// ===================== User-Container =====================

// CreateUserContainer creates a user-container association
func CreateUserContainer(db *gorm.DB, userContainer *database.UserContainer) error {
	if err := db.Omit(userContainerOmitFields).Create(userContainer).Error; err != nil {
		return fmt.Errorf("failed to create user-container association: %w", err)
	}
	return nil
}

// DeleteUserContainer deletes a user-container association
func DeleteUserContainer(db *gorm.DB, userID, containerID int) (int64, error) {
	result := db.Model(&database.UserContainer{}).
		Where("user_id = ? AND container_id = ? AND status != ?", userID, containerID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete user-container association: %w", result.Error)
	}
	return result.RowsAffected, nil
}

// ListContainersByUserID gets containers the user participates
func ListContainersByUserID(db *gorm.DB, userID int) ([]database.Container, error) {
	var containers []database.Container
	if err := db.Table("containers").
		Joins("JOIN user_containers uc ON uc.container_id = containers.id").
		Where("uc.user_id = ? AND containers.status = ?", userID, consts.CommonEnabled).
		Find(&containers).Error; err != nil {
		return nil, fmt.Errorf("failed to get containers of the specific user: %w", err)
	}
	return containers, nil
}

// ListUserContainersByUserID gets user-container associations for a specific user
func ListUserContainersByUserID(db *gorm.DB, userID int) ([]database.UserContainer, error) {
	var userContainers []database.UserContainer
	if err := db.Preload("Container").
		Preload("Role").
		Where("user_id = ? AND status = ?", userID, consts.CommonEnabled).
		Find(&userContainers).Error; err != nil {
		return nil, fmt.Errorf("failed to get user-container associations of the specific user: %w", err)
	}
	return userContainers, nil
}

// RemoveUsersFromContainer deletes all user-container associations for a given container
func RemoveUsersFromContainer(db *gorm.DB, containerID int) (int64, error) {
	result := db.Model(&database.UserContainer{}).
		Where("container_id = ? AND status != ?", containerID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete all users from container: %v", result.Error)
	}
	return result.RowsAffected, nil
}

// RemoveContainersFromRole deletes all user-container associations for a given role
func RemoveContainersFromRole(db *gorm.DB, roleID int) (int64, error) {
	result := db.Model(&database.UserContainer{}).
		Where("role_id = ? AND status != ?", roleID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if err := result.Error; err != nil {
		return 0, fmt.Errorf("failed to delete all containers from role: %v", err)
	}
	return result.RowsAffected, nil
}

// RemoveContainersFromUser deletes all user-container associations for a given user
func RemoveContainersFromUser(db *gorm.DB, userID int) (int64, error) {
	result := db.Model(&database.UserContainer{}).
		Where("user_id = ? AND status != ?", userID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if err := result.Error; err != nil {
		return 0, fmt.Errorf("failed to delete all containers from user: %v", err)
	}
	return result.RowsAffected, nil
}

// ===================== User-Dataset =====================

// CreateUserDataset creates a user-dataset association
func CreateUserDataset(db *gorm.DB, userDataset *database.UserDataset) error {
	if err := db.Omit(userDatasetOmitFields).Create(userDataset).Error; err != nil {
		return fmt.Errorf("failed to create user-dataset association: %w", err)
	}
	return nil
}

// DeleteUserDataset deletes a user-dataset association
func DeleteUserDataset(db *gorm.DB, userID, datasetID int) (int64, error) {
	result := db.Model(&database.UserDataset{}).
		Where("user_id = ? AND dataset_id = ? AND status != ?", userID, datasetID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete user-dataset association: %w", result.Error)
	}
	return result.RowsAffected, nil
}

// ListDatasetsByUserID gets datasets the user participates
func ListDatasetsByUserID(db *gorm.DB, userID int) ([]database.Dataset, error) {
	var datasets []database.Dataset
	if err := db.Table("datasets").
		Joins("JOIN user_datasets ud ON ud.dataset_id = datasets.id").
		Where("ud.user_id = ? AND datasets.status = ?", userID, consts.CommonEnabled).
		Find(&datasets).Error; err != nil {
		return nil, fmt.Errorf("failed to get datasets of the specific user: %w", err)
	}
	return datasets, nil
}

// ListUserDatasetsByUserID gets user-dataset associations for a specific user
func ListUserDatasetsByUserID(db *gorm.DB, userID int) ([]database.UserDataset, error) {
	var userDatasets []database.UserDataset
	if err := db.Preload("Dataset").
		Preload("Role").
		Where("user_id = ? AND status = ?", userID, consts.CommonEnabled).
		Find(&userDatasets).Error; err != nil {
		return nil, fmt.Errorf("failed to get user-dataset associations of the specific user: %w", err)
	}
	return userDatasets, nil
}

// RemoveUsersFromDataset deletes all user-dataset associations for a given dataset
func RemoveUsersFromDataset(db *gorm.DB, datasetID int) (int64, error) {
	result := db.Model(&database.UserDataset{}).
		Where("dataset_id = ? AND status != ?", datasetID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete all users from dataset: %v", result.Error)
	}
	return result.RowsAffected, nil
}

// RemoveDatasetsFromRole deletes all user-dataset associations for a given role
func RemoveDatasetsFromRole(db *gorm.DB, roleID int) (int64, error) {
	result := db.Model(&database.UserDataset{}).
		Where("role_id = ? AND status != ?", roleID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if err := result.Error; err != nil {
		return 0, fmt.Errorf("failed to delete all datasets from role: %v", err)
	}
	return result.RowsAffected, nil
}

// RemoveDatasetsFromUser deletes all user-dataset associations for a given user
func RemoveDatasetsFromUser(db *gorm.DB, userID int) (int64, error) {
	result := db.Model(&database.UserDataset{}).
		Where("user_id = ? AND status != ?", userID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete all datasets from user: %v", result.Error)
	}
	return result.RowsAffected, nil
}

// ===================== User-Project =====================

// CreateUserProject creates a user-project association
func CreateUserProject(db *gorm.DB, userProject *database.UserProject) error {
	if err := db.Omit(userProjectOmitFields).Create(userProject).Error; err != nil {
		return fmt.Errorf("failed to create user-project association: %w", err)
	}
	return nil
}

// DeleteUserProject deletes a user-project association
func DeleteUserProject(db *gorm.DB, userID, projectID int) (int64, error) {
	result := db.Model(&database.UserProject{}).
		Where("user_id = ? AND project_id = ? AND status != ?", userID, projectID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete user-project association: %w", result.Error)
	}
	return result.RowsAffected, nil
}

// ListProjectsByUserID gets projects the user participates
func ListProjectsByUserID(db *gorm.DB, userID int) ([]database.Project, error) {
	var projects []database.Project
	if err := db.Table("projects").
		Joins("JOIN user_projects up ON up.project_id = projects.id").
		Where("up.user_id = ? AND projects.status = ?", userID, consts.CommonEnabled).
		Find(&projects).Error; err != nil {
		return nil, fmt.Errorf("failed to get projects of the specific user: %w", err)
	}
	return projects, nil
}

// ListUserProjectsByUserID gets user-project associations for a specific user
func ListUserProjectsByUserID(db *gorm.DB, userID int) ([]database.UserProject, error) {
	var userProjects []database.UserProject
	if err := db.Preload("Project").
		Preload("Role").
		Where("user_id = ? AND status = ?", userID, consts.CommonEnabled).
		Find(&userProjects).Error; err != nil {
		return nil, fmt.Errorf("failed to get user-project associations of the specific user: %w", err)
	}
	return userProjects, nil
}

// RemoveUsersFromProject deletes all user-project associations for a given project
func RemoveUsersFromProject(db *gorm.DB, projectID int) (int64, error) {
	result := db.Model(&database.UserProject{}).
		Where("project_id = ? AND status != ?", projectID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete all users from project: %v", result.Error)
	}
	return result.RowsAffected, nil
}

// RemoveProjectsFromRole deletes all user-project associations for a given role
func RemoveProjectsFromRole(db *gorm.DB, roleID int) (int64, error) {
	result := db.Model(&database.UserProject{}).
		Where("role_id = ? AND status != ?", roleID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if err := result.Error; err != nil {
		return 0, fmt.Errorf("failed to delete all projects from role: %v", err)
	}
	return result.RowsAffected, nil
}

// RemoveProjectsFromUser deletes all user-project associations for a given user
func RemoveProjectsFromUser(db *gorm.DB, userID int) (int64, error) {
	result := db.Model(&database.UserProject{}).
		Where("user_id = ? AND status != ?", userID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete all projects from user: %v", result.Error)
	}
	return result.RowsAffected, nil
}
