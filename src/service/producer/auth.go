package producer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"
	"aegis/utils"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Register handles user registration business logic
func Register(req *dto.RegisterReq) (*dto.UserInfo, error) {
	if req == nil {
		return nil, fmt.Errorf("register request is nil")
	}

	var createdUser *database.User

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// Check if user already exists
		if _, err := repository.GetUserByUsername(tx, req.Username); err == nil {
			return fmt.Errorf("%w: username is already taken", consts.ErrAlreadyExists)
		}

		if _, err := repository.GetUserByEmail(tx, req.Email); err == nil {
			return fmt.Errorf("%w: email is already registered", consts.ErrAlreadyExists)
		}

		// Hash password
		hashedPassword, err := utils.HashPassword(req.Password)
		if err != nil {
			return fmt.Errorf("password hashing failed: %w", err)
		}

		user := &database.User{
			Username: req.Username,
			Email:    req.Email,
			Password: hashedPassword,
			IsActive: true,
			Status:   consts.CommonEnabled,
		}

		if err := repository.CreateUser(tx, user); err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		createdUser = user
		return nil
	})
	if err != nil {
		return nil, err
	}

	return dto.NewUserInfo(createdUser), nil
}

// Login handles user authentication business logic
func Login(req *dto.LoginReq) (*dto.LoginResp, error) {
	if req == nil {
		return nil, fmt.Errorf("login request is nil")
	}

	var loginedUser *database.User
	var token string
	var expiresAt time.Time

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		user, err := repository.GetUserByUsername(tx, req.Username)
		if err != nil {
			return fmt.Errorf("%w: invalid username or password", consts.ErrAuthenticationFailed)
		}

		if !utils.VerifyPassword(req.Password, user.Password) {
			return fmt.Errorf("%w: invalid username or password", consts.ErrAuthenticationFailed)
		}

		// Generate token with user roles
		token, expiresAt, err = generateTokenWithRoles(tx, user)
		if err != nil {
			return err
		}

		if err := repository.UpdateUserLoginTime(tx, user.ID); err != nil {
			logrus.Errorf("failed to update last login time for user %d: %v", user.ID, err)
		}

		loginedUser = user
		return nil
	})
	if err != nil {
		return nil, err
	}

	roles, err := repository.ListRolesByUserID(database.DB, loginedUser.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user role: %w", err)
	}

	if len(roles) == 0 {
		return nil, fmt.Errorf("%w: user has no assigned role", consts.ErrPermissionDenied)
	}

	info := dto.NewUserInfo(loginedUser)
	info.Role = roles[0].Name

	resp := &dto.LoginResp{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      *info,
	}
	return resp, nil
}

// Logout handles user logout business logic
func Logout(ctx context.Context, claims *utils.Claims) error {
	metaData := map[string]any{
		"user_id": claims.UserID,
		"reason":  "User logout",
	}
	if err := repository.AddTokenToBlacklist(ctx, claims.ID, claims.ExpiresAt.Time, metaData); err != nil {
		logrus.Errorf("failed to add token to blacklist: %v", err)
		return fmt.Errorf("failed to blacklist token: %w", err)
	}
	return nil
}

// RefreshToken handles JWT token refresh business logic
func RefreshToken(req *dto.TokenRefreshReq) (*dto.TokenRefreshResp, error) {
	if req == nil {
		return nil, fmt.Errorf("token refresh request is nil")
	}

	// Validate refresh token and get user info
	refreshClaims, err := utils.ValidateToken(req.Token)
	if err != nil {
		return nil, fmt.Errorf("token refresh failed: %w", err)
	}

	// Fetch fresh user data from database
	user, err := repository.GetUserByID(database.DB, refreshClaims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Generate new access token with fresh user data
	newToken, expiresAt, err := generateTokenWithRoles(database.DB, user)
	if err != nil {
		return nil, err
	}

	response := &dto.TokenRefreshResp{
		Token:     newToken,
		ExpiresAt: expiresAt,
	}

	return response, nil
}

// ChangePassword handles password change business logic
func ChangePassword(req *dto.ChangePasswordReq, userID int) error {
	if req == nil {
		return fmt.Errorf("change password request is nil")
	}

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		user, err := repository.GetUserByID(tx, userID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: user not found", consts.ErrNotFound)
			}
			return fmt.Errorf("failed to get user: %w", err)
		}

		if !utils.VerifyPassword(req.OldPassword, user.Password) {
			return fmt.Errorf("invalid old password")
		}

		hashedPassword, err := utils.HashPassword(req.NewPassword)
		if err != nil {
			return fmt.Errorf("password hashing failed: %w", err)
		}
		user.Password = hashedPassword

		if err := repository.UpdateUser(tx, user); err != nil {
			return fmt.Errorf("failed to update password: %w", err)
		}

		return nil
	})

	return err
}

// GetProfile handles getting current user profile business logic
func GetProfile(userID int) (*dto.UserProfileResp, error) {
	user, err := repository.GetUserByID(database.DB, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: user not found", consts.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	resp := dto.NewUserProfileResp(user)
	userContainers, userDatasets, userProjects, err := getAllUserResourceRoles(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user resource roles: %w", err)
	}

	resp.ContainerRoles = userContainers
	resp.DatasetRoles = userDatasets
	resp.ProjectRoles = userProjects

	return resp, nil
}

// getAllUserResourceRoles fetches all container, dataset, project roles assigned to the user
func getAllUserResourceRoles(userID int) ([]dto.UserContainerInfo, []dto.UserDatasetInfo, []dto.UserProjectInfo, error) {
	userContainers, err := repository.ListUserContainersByUserID(database.DB, userID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to list user-container roles: %w", err)
	}
	var containerRoles []dto.UserContainerInfo
	for _, uc := range userContainers {
		containerRoles = append(containerRoles, *dto.NewUserContainerInfo(&uc))
	}

	userDatasets, err := repository.ListUserDatasetsByUserID(database.DB, userID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to list user-dataset roles: %w", err)
	}
	var datasetRoles []dto.UserDatasetInfo
	for _, ud := range userDatasets {
		datasetRoles = append(datasetRoles, *dto.NewUserDatasetInfo(&ud))
	}

	userProjects, err := repository.ListUserProjectsByUserID(database.DB, userID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to list user-project roles: %w", err)
	}
	var projectRoles []dto.UserProjectInfo
	for _, up := range userProjects {
		projectRoles = append(projectRoles, *dto.NewUserProjectInfo(&up))
	}

	return containerRoles, datasetRoles, projectRoles, nil
}

// ============================================================================
// Helper Functions
// ============================================================================

// generateTokenWithRoles fetches user roles and generates a JWT token with role information
func generateTokenWithRoles(db *gorm.DB, user *database.User) (string, time.Time, error) {
	// Get user's global roles
	roles, err := repository.ListRolesByUserID(db, user.ID)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to get user roles: %w", err)
	}

	// Check if user is system admin and build role names list
	isAdmin := false
	roleNames := make([]string, 0, len(roles))
	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
		if role.Name == string(consts.RoleSuperAdmin) || role.Name == string(consts.RoleAdmin) {
			isAdmin = true
		}
	}

	// Generate token with role information
	token, expiresAt, err := utils.GenerateToken(user.ID, user.Username, user.Email, user.IsActive, isAdmin, roleNames)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to generate token: %w", err)
	}

	return token, expiresAt, nil
}
