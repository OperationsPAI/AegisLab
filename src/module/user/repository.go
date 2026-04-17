package usermodule

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

func (r *Repository) ensureUserUnique(username, email string) error {
	var existingByUsername model.User
	if err := r.db.Where("username = ?", username).First(&existingByUsername).Error; err == nil {
		return fmt.Errorf("%w: username %s already exists", consts.ErrAlreadyExists, username)
	}

	var existingByEmail model.User
	if err := r.db.Where("email = ?", email).First(&existingByEmail).Error; err == nil {
		return fmt.Errorf("%w: email %s already exists", consts.ErrAlreadyExists, email)
	}
	return nil
}

func (r *Repository) createUserIfUnique(user *model.User) error {
	if err := r.ensureUserUnique(user.Username, user.Email); err != nil {
		return err
	}
	if err := r.db.Omit("active_username").Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *Repository) getUserDetailBase(userID int) (*model.User, error) {
	var user model.User
	if err := r.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to find user with id %d: %w", userID, err)
	}
	return &user, nil
}

func (r *Repository) DeleteUserCascade(userID int) (int64, error) {
	if err := r.ensureActiveRecordExists(&model.User{}, userID, "user"); err != nil {
		return 0, err
	}

	if err := r.db.Model(&model.UserContainer{}).
		Where("user_id = ? AND status != ?", userID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted).Error; err != nil {
		return 0, fmt.Errorf("failed to remove containers from user: %w", err)
	}
	if err := r.db.Model(&model.UserDataset{}).
		Where("user_id = ? AND status != ?", userID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted).Error; err != nil {
		return 0, fmt.Errorf("failed to remove datasets from user: %w", err)
	}
	if err := r.db.Model(&model.UserProject{}).
		Where("user_id = ? AND status != ?", userID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted).Error; err != nil {
		return 0, fmt.Errorf("failed to remove projects from user: %w", err)
	}
	if err := r.db.Where("user_id = ?", userID).Delete(&model.UserPermission{}).Error; err != nil {
		return 0, fmt.Errorf("failed to remove permissions from user: %w", err)
	}
	if err := r.db.Where("user_id = ?", userID).Delete(&model.UserRole{}).Error; err != nil {
		return 0, fmt.Errorf("failed to remove roles from user: %w", err)
	}

	result := r.db.Model(&model.User{}).
		Where("id = ? AND status != ?", userID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete user %d: %w", userID, result.Error)
	}
	return result.RowsAffected, nil
}

func (r *Repository) listUserViews(limit, offset int, isActive *bool, status *consts.StatusType) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	query := r.db.Model(&model.User{}).Where("status != ?", consts.CommonDeleted)
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

func (r *Repository) updateMutableUser(userID int, patch func(*model.User)) (*model.User, error) {
	var user model.User
	if err := r.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to find user with id %d: %w", userID, err)
	}
	patch(&user)
	if err := r.db.Omit("active_username").Save(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	return &user, nil
}

func (r *Repository) loadUserDetailRelations(userID int) ([]model.Role, []model.Permission, []model.UserContainer, []model.UserDataset, []model.UserProject, error) {
	var roles []model.Role
	if err := r.db.Table("roles").
		Joins("JOIN user_roles ur ON ur.role_id = roles.id").
		Where("ur.user_id = ? AND roles.status = ?", userID, consts.CommonEnabled).
		Find(&roles).Error; err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to list roles by user id: %w", err)
	}

	var permissions []model.Permission
	if err := r.db.Table("permissions").
		Joins("JOIN user_permissions up ON up.permission_id = permissions.id").
		Where("up.user_id = ? AND permissions.status = ?", userID, consts.CommonEnabled).
		Find(&permissions).Error; err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to list permissions by user id: %w", err)
	}

	var userContainers []model.UserContainer
	if err := r.db.Where("user_id = ? AND status != ?", userID, consts.CommonDeleted).
		Find(&userContainers).Error; err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to list user containers: %w", err)
	}

	var userDatasets []model.UserDataset
	if err := r.db.Where("user_id = ? AND status != ?", userID, consts.CommonDeleted).
		Find(&userDatasets).Error; err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to list user datasets: %w", err)
	}

	var userProjects []model.UserProject
	if err := r.db.Where("user_id = ? AND status != ?", userID, consts.CommonDeleted).
		Find(&userProjects).Error; err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to list user projects: %w", err)
	}

	return roles, permissions, userContainers, userDatasets, userProjects, nil
}

func (r *Repository) AssignGlobalRole(userID, roleID int) error {
	if err := r.ensureActiveRecordExists(&model.User{}, userID, "user"); err != nil {
		return err
	}
	if err := r.ensureActiveRecordExists(&model.Role{}, roleID, "role"); err != nil {
		return err
	}
	if err := r.db.Create(&model.UserRole{UserID: userID, RoleID: roleID}).Error; err != nil {
		return fmt.Errorf("failed to create user-role association: %w", err)
	}
	return nil
}

func (r *Repository) RemoveGlobalRole(userID, roleID int) error {
	if err := r.ensureActiveRecordExists(&model.User{}, userID, "user"); err != nil {
		return err
	}
	if err := r.ensureActiveRecordExists(&model.Role{}, roleID, "role"); err != nil {
		return err
	}
	if err := r.db.Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&model.UserRole{}).Error; err != nil {
		return fmt.Errorf("failed to delete user-role association: %w", err)
	}
	return nil
}

func (r *Repository) BuildUserPermissions(userID int, items []AssignUserPermissionItem) ([]model.UserPermission, error) {
	if err := r.ensureActiveRecordExists(&model.User{}, userID, "user"); err != nil {
		return nil, err
	}

	permissionIDs := make([]int, 0, len(items))
	for _, item := range items {
		permissionIDs = append(permissionIDs, item.PermissionID)
	}
	permissions, err := r.listPermissionsByIDs(permissionIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions by ids: %w", err)
	}
	permissionMap := make(map[int]struct{}, len(permissions))
	for _, permission := range permissions {
		permissionMap[permission.ID] = struct{}{}
	}

	userPermissions := make([]model.UserPermission, 0, len(items))
	for _, item := range items {
		if _, exists := permissionMap[item.PermissionID]; !exists {
			return nil, fmt.Errorf("%w: permission id %d not found", consts.ErrNotFound, item.PermissionID)
		}
		if item.ContainerID != nil {
			if err := r.ensureActiveRecordExists(&model.Container{}, *item.ContainerID, "container"); err != nil {
				return nil, fmt.Errorf("%w: container id %d not found", consts.ErrNotFound, *item.ContainerID)
			}
		}
		if item.DatasetID != nil {
			if err := r.ensureActiveRecordExists(&model.Dataset{}, *item.DatasetID, "dataset"); err != nil {
				return nil, fmt.Errorf("%w: dataset id %d not found", consts.ErrNotFound, *item.DatasetID)
			}
		}
		if item.ProjectID != nil {
			if err := r.ensureActiveRecordExists(&model.Project{}, *item.ProjectID, "project"); err != nil {
				return nil, fmt.Errorf("%w: project id %d not found", consts.ErrNotFound, *item.ProjectID)
			}
		}

		userPermission := item.ConvertToUserPermission()
		userPermission.UserID = userID
		userPermissions = append(userPermissions, *userPermission)
	}
	return userPermissions, nil
}

func (r *Repository) BatchCreateUserPermissions(userPermissions []model.UserPermission) error {
	if len(userPermissions) == 0 {
		return nil
	}
	if err := r.db.Create(&userPermissions).Error; err != nil {
		return fmt.Errorf("failed to batch create user permissions: %w", err)
	}
	return nil
}

func (r *Repository) BatchDeleteUserPermissions(userID int, permissionIDs []int) error {
	if err := r.ensureActiveRecordExists(&model.User{}, userID, "user"); err != nil {
		return err
	}

	permissions, err := r.listPermissionsByIDs(permissionIDs)
	if err != nil {
		return fmt.Errorf("failed to list permissions by ids: %w", err)
	}
	permissionMap := make(map[int]struct{}, len(permissions))
	for _, permission := range permissions {
		permissionMap[permission.ID] = struct{}{}
	}
	for _, permissionID := range permissionIDs {
		if _, exists := permissionMap[permissionID]; !exists {
			return fmt.Errorf("%w: permission id %d not found", consts.ErrNotFound, permissionID)
		}
	}

	if err := r.db.Where("user_id = ? AND permission_id IN (?)", userID, permissionIDs).
		Delete(&model.UserPermission{}).Error; err != nil {
		return fmt.Errorf("failed to batch delete user permissions: %w", err)
	}
	return nil
}

func (r *Repository) AssignContainerRole(userID, containerID, roleID int) error {
	if err := r.ensureActiveRecordExists(&model.User{}, userID, "user"); err != nil {
		return err
	}
	if err := r.ensureActiveRecordExists(&model.Container{}, containerID, "container"); err != nil {
		return err
	}
	if err := r.ensureActiveRecordExists(&model.Role{}, roleID, "role"); err != nil {
		return err
	}

	if err := r.db.Create(&model.UserContainer{
		UserID:      userID,
		ContainerID: containerID,
		RoleID:      roleID,
	}).Error; err != nil {
		return fmt.Errorf("failed to create user-container association: %w", err)
	}
	return nil
}

func (r *Repository) RemoveContainerRole(userID, containerID int) (int64, error) {
	if err := r.ensureActiveRecordExists(&model.User{}, userID, "user"); err != nil {
		return 0, err
	}
	if err := r.ensureActiveRecordExists(&model.Container{}, containerID, "container"); err != nil {
		return 0, err
	}
	result := r.db.Model(&model.UserContainer{}).
		Where("user_id = ? AND container_id = ? AND status != ?", userID, containerID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete user-container association: %w", result.Error)
	}
	return result.RowsAffected, nil
}

func (r *Repository) AssignDatasetRole(userID, datasetID, roleID int) error {
	if err := r.ensureActiveRecordExists(&model.User{}, userID, "user"); err != nil {
		return err
	}
	if err := r.ensureActiveRecordExists(&model.Dataset{}, datasetID, "dataset"); err != nil {
		return err
	}
	if err := r.ensureActiveRecordExists(&model.Role{}, roleID, "role"); err != nil {
		return err
	}

	if err := r.db.Create(&model.UserDataset{
		UserID:    userID,
		DatasetID: datasetID,
		RoleID:    roleID,
	}).Error; err != nil {
		return fmt.Errorf("failed to create user-dataset association: %w", err)
	}
	return nil
}

func (r *Repository) RemoveDatasetRole(userID, datasetID int) (int64, error) {
	if err := r.ensureActiveRecordExists(&model.User{}, userID, "user"); err != nil {
		return 0, err
	}
	if err := r.ensureActiveRecordExists(&model.Dataset{}, datasetID, "dataset"); err != nil {
		return 0, err
	}
	result := r.db.Model(&model.UserDataset{}).
		Where("user_id = ? AND dataset_id = ? AND status != ?", userID, datasetID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete user-dataset association: %w", result.Error)
	}
	return result.RowsAffected, nil
}

func (r *Repository) AssignProjectRole(userID, projectID, roleID int) error {
	if err := r.ensureActiveRecordExists(&model.User{}, userID, "user"); err != nil {
		return err
	}
	if err := r.ensureActiveRecordExists(&model.Project{}, projectID, "project"); err != nil {
		return err
	}
	if err := r.ensureActiveRecordExists(&model.Role{}, roleID, "role"); err != nil {
		return err
	}

	if err := r.db.Create(&model.UserProject{
		UserID:    userID,
		ProjectID: projectID,
		RoleID:    roleID,
	}).Error; err != nil {
		return fmt.Errorf("failed to create user-project association: %w", err)
	}
	return nil
}

func (r *Repository) RemoveProjectRole(userID, projectID int) (int64, error) {
	if err := r.ensureActiveRecordExists(&model.User{}, userID, "user"); err != nil {
		return 0, err
	}
	if err := r.ensureActiveRecordExists(&model.Project{}, projectID, "project"); err != nil {
		return 0, err
	}
	result := r.db.Model(&model.UserProject{}).
		Where("user_id = ? AND project_id = ? AND status != ?", userID, projectID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete user-project association: %w", result.Error)
	}
	return result.RowsAffected, nil
}

func (r *Repository) ensureActiveRecordExists(model any, id int, entity string) error {
	if err := r.db.Where("id = ? AND status != ?", id, consts.CommonDeleted).First(model).Error; err != nil {
		return fmt.Errorf("failed to find %s with id %d: %w", entity, id, err)
	}
	return nil
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
