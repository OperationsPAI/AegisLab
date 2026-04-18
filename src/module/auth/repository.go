package authmodule

import (
	"aegis/consts"
	"aegis/model"
	"fmt"
	"time"

	"gorm.io/gorm"
)

const (
	userOmitFields          = "active_username"
	userContainerOmitFields = "active_user_container"
	userDatasetOmitFields   = "active_user_dataset"
	userProjectOmitFields   = "active_user_project"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *model.User) error {
	if err := r.db.Omit(userOmitFields).Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *UserRepository) GetByID(id int) (*model.User, error) {
	var user model.User
	if err := r.db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to find user with id %d: %w", id, err)
	}
	return &user, nil
}

func (r *UserRepository) GetByUsername(username string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to find user with username %s: %w", username, err)
	}
	return &user, nil
}

func (r *UserRepository) GetByEmail(email string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to find user with email %s: %w", email, err)
	}
	return &user, nil
}

func (r *UserRepository) Update(user *model.User) error {
	if err := r.db.Omit(userOmitFields).Save(user).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

func (r *UserRepository) UpdateLoginTime(userID int) error {
	now := r.db.NowFunc()
	if err := r.db.Model(&model.User{}).
		Where("id = ? AND status != ?", userID, consts.CommonDeleted).
		Update("last_login_at", now).Error; err != nil {
		return fmt.Errorf("failed to update user login time: %w", err)
	}
	return nil
}

func (r *UserRepository) ListContainerRoles(userID int) ([]model.UserContainer, error) {
	var userContainers []model.UserContainer
	if err := r.db.Preload("Container").
		Preload("Role").
		Where("user_id = ? AND status = ?", userID, consts.CommonEnabled).
		Find(&userContainers).Error; err != nil {
		return nil, fmt.Errorf("failed to get user-container associations of the specific user: %w", err)
	}
	return userContainers, nil
}

func (r *UserRepository) ListDatasetRoles(userID int) ([]model.UserDataset, error) {
	var userDatasets []model.UserDataset
	if err := r.db.Preload("Dataset").
		Preload("Role").
		Where("user_id = ? AND status = ?", userID, consts.CommonEnabled).
		Find(&userDatasets).Error; err != nil {
		return nil, fmt.Errorf("failed to get user-dataset associations of the specific user: %w", err)
	}
	return userDatasets, nil
}

func (r *UserRepository) ListProjectRoles(userID int) ([]model.UserProject, error) {
	var userProjects []model.UserProject
	if err := r.db.Preload("Project").
		Preload("Role").
		Where("user_id = ? AND status = ?", userID, consts.CommonEnabled).
		Find(&userProjects).Error; err != nil {
		return nil, fmt.Errorf("failed to get user-project associations of the specific user: %w", err)
	}
	return userProjects, nil
}

type RoleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) ListByUserID(userID int) ([]model.Role, error) {
	var roles []model.Role
	if err := r.db.Table("roles").
		Joins("JOIN user_roles ur ON ur.role_id = roles.id").
		Where("ur.user_id = ? AND roles.status = ?", userID, consts.CommonEnabled).
		Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("failed to get global roles of the specific user: %w", err)
	}
	return roles, nil
}

type AccessKeyRepository struct {
	db *gorm.DB
}

func NewAccessKeyRepository(db *gorm.DB) *AccessKeyRepository {
	return &AccessKeyRepository{db: db}
}

func (r *AccessKeyRepository) Create(key *model.UserAccessKey) error {
	if err := r.db.Create(key).Error; err != nil {
		return fmt.Errorf("failed to create access key: %w", err)
	}
	return nil
}

func (r *AccessKeyRepository) ListByUserID(userID, limit, offset int) ([]model.UserAccessKey, int64, error) {
	query := r.db.Model(&model.UserAccessKey{}).
		Where("user_id = ? AND status != ?", userID, consts.CommonDeleted)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count access keys: %w", err)
	}

	var keys []model.UserAccessKey
	if err := query.Order("id DESC").Limit(limit).Offset(offset).Find(&keys).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list access keys: %w", err)
	}
	return keys, total, nil
}

func (r *AccessKeyRepository) GetByIDForUser(id, userID int) (*model.UserAccessKey, error) {
	var key model.UserAccessKey
	if err := r.db.Where("id = ? AND user_id = ? AND status != ?", id, userID, consts.CommonDeleted).First(&key).Error; err != nil {
		return nil, fmt.Errorf("failed to find access key: %w", err)
	}
	return &key, nil
}

func (r *AccessKeyRepository) GetByAccessKey(accessKey string) (*model.UserAccessKey, error) {
	var key model.UserAccessKey
	if err := r.db.Where("access_key = ? AND status != ?", accessKey, consts.CommonDeleted).First(&key).Error; err != nil {
		return nil, fmt.Errorf("failed to find access key: %w", err)
	}
	return &key, nil
}

func (r *AccessKeyRepository) Update(key *model.UserAccessKey) error {
	if err := r.db.Save(key).Error; err != nil {
		return fmt.Errorf("failed to update access key: %w", err)
	}
	return nil
}

func (r *AccessKeyRepository) UpdateLastUsedAt(id int, usedAt time.Time) error {
	if err := r.db.Model(&model.UserAccessKey{}).Where("id = ?", id).Update("last_used_at", usedAt).Error; err != nil {
		return fmt.Errorf("failed to update access key last used time: %w", err)
	}
	return nil
}
