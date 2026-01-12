package producer

import (
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// CreateUser handles the business logic for creating a new user
func CreateUser(req *dto.CreateUserReq) (*dto.UserResp, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &database.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		FullName: req.FullName,
		Phone:    req.Phone,
		Avatar:   req.Avatar,
		Status:   consts.CommonEnabled,
		IsActive: true,
	}

	var createdUser *database.User
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if _, err := repository.GetUserByUsername(tx, user.Username); err == nil {
			return fmt.Errorf("%w: username %s already exists", consts.ErrAlreadyExists, user.Username)
		}

		if _, err := repository.GetUserByEmail(user.Email); err == nil {
			return fmt.Errorf("%w: email %s already exists", consts.ErrAlreadyExists, user.Email)
		}

		if err := repository.CreateUser(tx, user); err != nil {
			return err
		}

		createdUser = user
		return nil
	})
	if err != nil {
		return nil, err
	}

	return dto.NewUserResp(createdUser), nil
}

// DeleteUser deletes an existing user by marking their status as deleted
func DeleteUser(userID int) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		user, err := repository.GetUserByID(tx, userID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: user not found", consts.ErrNotFound)
			}
			return fmt.Errorf("failed to get user: %w", err)
		}

		// Remove all associations with containers, datasets, and projects
		if _, err := repository.RemoveContainersFromUser(tx, user.ID); err != nil {
			return fmt.Errorf("failed to remove containers from user: %w", err)
		}
		if _, err = repository.RemoveDatasetsFromUser(tx, user.ID); err != nil {
			return fmt.Errorf("failed to remove datasets from user: %w", err)
		}
		if _, err = repository.RemoveProjectsFromUser(tx, user.ID); err != nil {
			return fmt.Errorf("failed to remove projects from user: %w", err)
		}

		// Remove associated permissions and roles
		if err := repository.RemovePermissionsFromUser(tx, user.ID); err != nil {
			return fmt.Errorf("failed to remove projects from user: %w", err)
		}
		if err := repository.RemoveRolesFromUser(tx, user.ID); err != nil {
			return fmt.Errorf("failed to remove roles from user: %w", err)
		}

		rows, err := repository.DeleteUser(tx, userID)
		if err != nil {
			return fmt.Errorf("failed to delete user: %w", err)
		}
		if rows == 0 {
			return fmt.Errorf("%w: user id %d not found", consts.ErrNotFound, userID)
		}

		return nil
	})
}

// GetUserDetail retrieves detailed information about a user by their ID
func GetUserDetail(userID int) (*dto.UserDetailResp, error) {
	user, err := repository.GetUserByID(database.DB, userID)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("%w: user with ID %d not found", consts.ErrNotFound, userID)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	resp := dto.NewUserDetailResp(user)

	globalRoles, err := repository.ListRolesByUserID(database.DB, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user global roles: %w", err)
	}

	resp.GlobalRoles = make([]dto.RoleResp, len(globalRoles))
	for i, role := range globalRoles {
		roleResp := *dto.NewRoleResp(&role)
		resp.GlobalRoles[i] = roleResp
	}

	permissions, err := repository.ListPermissionsByUserID(database.DB, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	resp.Permissions = make([]dto.PermissionResp, len(permissions))
	for i, permission := range permissions {
		resp.Permissions[i] = *dto.NewPermissionResp(&permission)
	}

	userContainers, userDatasets, userProjects, err := getAllUserResourceRoles(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user resource roles: %w", err)
	}

	resp.ContainerRoles = userContainers
	resp.DatasetRoles = userDatasets
	resp.ProjectRoles = userProjects

	return resp, nil
}

// ListUsers lists users based on the provided filters
func ListUsers(req *dto.ListUserReq) (*dto.ListResp[dto.UserResp], error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	limit, offset := req.ToGormParams()

	users, total, err := repository.ListUsers(database.DB, limit, offset, req.IsActive, req.Status)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	userResps := make([]dto.UserResp, len(users))
	for i, u := range users {
		userResps[i] = *dto.NewUserResp(&u)
	}

	resp := dto.ListResp[dto.UserResp]{
		Items:      userResps,
		Pagination: req.ConvertToPaginationInfo(total),
	}
	return &resp, nil
}

// UpdateUser updates an existing user's details
func UpdateUser(req *dto.UpdateUserReq, userID int) (*dto.UserResp, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	var updatedUser *database.User

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		existingUser, err := repository.GetUserByID(tx, userID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: user not found", consts.ErrNotFound)
			}
			return fmt.Errorf("failed to get user: %w", err)
		}

		req.PatchUserModel(existingUser)

		if err := repository.UpdateUser(tx, existingUser); err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}

		updatedUser = existingUser
		return nil
	})
	if err != nil {
		return nil, err
	}

	return dto.NewUserResp(updatedUser), nil
}

func SearchUsers(req *dto.SearchReq) (*dto.SearchResp[dto.UserResp], error) {
	return nil, nil
}
