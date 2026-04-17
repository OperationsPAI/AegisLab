package authmodule

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"aegis/consts"
	"aegis/model"
	usermodule "aegis/module/user"
	"aegis/utils"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const accessKeySignatureTTL = 5 * time.Minute

type Service struct {
	userRepo      *UserRepository
	roleRepo      *RoleRepository
	accessKeyRepo *AccessKeyRepository
	tokenStore    *TokenStore
}

func NewService(userRepo *UserRepository, roleRepo *RoleRepository, accessKeyRepo *AccessKeyRepository, tokenStore *TokenStore) *Service {
	return &Service{
		userRepo:      userRepo,
		roleRepo:      roleRepo,
		accessKeyRepo: accessKeyRepo,
		tokenStore:    tokenStore,
	}
}

func (s *Service) Register(ctx context.Context, req *RegisterReq) (*UserInfo, error) {
	if req == nil {
		return nil, fmt.Errorf("register request is nil")
	}

	var createdUser *model.User
	err := s.userRepo.Transaction(func(tx *gorm.DB) error {
		userRepo := s.userRepo.withDB(tx)

		if _, err := userRepo.GetByUsername(req.Username); err == nil {
			return fmt.Errorf("%w: username is already taken", consts.ErrAlreadyExists)
		}

		if _, err := userRepo.GetByEmail(req.Email); err == nil {
			return fmt.Errorf("%w: email is already registered", consts.ErrAlreadyExists)
		}

		user := &model.User{
			Username: req.Username,
			Email:    req.Email,
			Password: req.Password,
			IsActive: true,
			Status:   consts.CommonEnabled,
		}

		if err := userRepo.Create(user); err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		createdUser = user
		return nil
	})
	if err != nil {
		return nil, err
	}

	return NewUserInfo(createdUser), nil
}

func (s *Service) Login(ctx context.Context, req *LoginReq) (*LoginResp, error) {
	if req == nil {
		return nil, fmt.Errorf("login request is nil")
	}

	var loginedUser *model.User
	var token string
	var expiresAt time.Time

	err := s.userRepo.Transaction(func(tx *gorm.DB) error {
		userRepo := s.userRepo.withDB(tx)
		roleRepo := s.roleRepo.withDB(tx)

		user, err := userRepo.GetByUsername(req.Username)
		if err != nil {
			return fmt.Errorf("%w: invalid username or password", consts.ErrAuthenticationFailed)
		}

		if !utils.VerifyPassword(req.Password, user.Password) {
			return fmt.Errorf("%w: invalid username or password", consts.ErrAuthenticationFailed)
		}

		token, expiresAt, err = s.generateTokenWithRoles(roleRepo, user)
		if err != nil {
			return err
		}

		if err := userRepo.UpdateLoginTime(user.ID); err != nil {
			logrus.Errorf("failed to update last login time for user %d: %v", user.ID, err)
		}

		loginedUser = user
		return nil
	})
	if err != nil {
		return nil, err
	}

	roles, err := s.roleRepo.ListByUserID(loginedUser.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user role: %w", err)
	}

	if len(roles) == 0 {
		return nil, fmt.Errorf("%w: user has no assigned role", consts.ErrPermissionDenied)
	}

	info := NewUserInfo(loginedUser)
	info.Role = roles[0].Name

	return &LoginResp{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      *info,
	}, nil
}

func (s *Service) RefreshToken(ctx context.Context, req *TokenRefreshReq) (*TokenRefreshResp, error) {
	if req == nil {
		return nil, fmt.Errorf("token refresh request is nil")
	}

	refreshClaims, err := utils.ValidateToken(req.Token)
	if err != nil {
		return nil, fmt.Errorf("token refresh failed: %w", err)
	}

	user, err := s.userRepo.GetByID(refreshClaims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	newToken, expiresAt, err := s.generateTokenWithRoles(s.roleRepo, user)
	if err != nil {
		return nil, err
	}

	return &TokenRefreshResp{
		Token:     newToken,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *Service) Logout(ctx context.Context, claims *utils.Claims) error {
	metaData := map[string]any{
		"user_id": claims.UserID,
		"reason":  "User logout",
	}
	if err := s.tokenStore.AddTokenToBlacklist(ctx, claims.ID, claims.ExpiresAt.Time, metaData); err != nil {
		logrus.Errorf("failed to add token to blacklist: %v", err)
		return fmt.Errorf("failed to blacklist token: %w", err)
	}
	return nil
}

func (s *Service) ChangePassword(ctx context.Context, req *ChangePasswordReq, userID int) error {
	if req == nil {
		return fmt.Errorf("change password request is nil")
	}

	return s.userRepo.Transaction(func(tx *gorm.DB) error {
		userRepo := s.userRepo.withDB(tx)

		user, err := userRepo.GetByID(userID)
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

		if err := userRepo.Update(user); err != nil {
			return fmt.Errorf("failed to update password: %w", err)
		}

		return nil
	})
}

func (s *Service) GetProfile(ctx context.Context, userID int) (*UserProfileResp, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: user not found", consts.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	resp := NewUserProfileResp(user)
	userContainers, userDatasets, userProjects, err := s.getAllUserResourceRoles(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user resource roles: %w", err)
	}

	resp.ContainerRoles = userContainers
	resp.DatasetRoles = userDatasets
	resp.ProjectRoles = userProjects

	return resp, nil
}

func (s *Service) CreateAccessKey(ctx context.Context, userID int, req *CreateAccessKeyReq) (*AccessKeyWithSecretResp, error) {
	if req == nil {
		return nil, fmt.Errorf("access key create request is nil")
	}

	accessKeyValue, err := generateCredentialValue("ak_", 16)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access key: %w", err)
	}
	secretKeyValue, err := generateCredentialValue("sk_", 24)
	if err != nil {
		return nil, fmt.Errorf("failed to generate secret key: %w", err)
	}
	secretHash, err := utils.HashPassword(secretKeyValue)
	if err != nil {
		return nil, fmt.Errorf("failed to hash secret key: %w", err)
	}
	secretCiphertext, err := utils.EncryptAccessKeySecret(secretKeyValue)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt secret key: %w", err)
	}

	key := &model.UserAccessKey{
		UserID:           userID,
		Name:             req.Name,
		Description:      req.Description,
		AccessKey:        accessKeyValue,
		SecretHash:       secretHash,
		SecretCiphertext: secretCiphertext,
		ExpiresAt:        req.ExpiresAt,
		Status:           consts.CommonEnabled,
	}
	if err := s.accessKeyRepo.Create(key); err != nil {
		return nil, err
	}

	resp := &AccessKeyWithSecretResp{
		AccessKeyInfo: *NewAccessKeyInfo(key),
		SecretKey:     secretKeyValue,
	}
	return resp, nil
}

func (s *Service) ListAccessKeys(ctx context.Context, userID int, req *ListAccessKeyReq) (*ListAccessKeyResp, error) {
	if req == nil {
		return nil, fmt.Errorf("access key list request is nil")
	}

	limit, offset := req.ToGormParams()
	keys, total, err := s.accessKeyRepo.ListByUserID(userID, limit, offset)
	if err != nil {
		return nil, err
	}

	items := make([]AccessKeyInfo, 0, len(keys))
	for i := range keys {
		items = append(items, *NewAccessKeyInfo(&keys[i]))
	}

	return &ListAccessKeyResp{
		Items:      items,
		Pagination: *req.ConvertToPaginationInfo(total),
	}, nil
}

func (s *Service) GetAccessKey(ctx context.Context, userID, accessKeyID int) (*AccessKeyInfo, error) {
	key, err := s.accessKeyRepo.GetByIDForUser(accessKeyID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: access key not found", consts.ErrNotFound)
		}
		return nil, err
	}
	return NewAccessKeyInfo(key), nil
}

func (s *Service) DeleteAccessKey(ctx context.Context, userID, accessKeyID int) error {
	key, err := s.accessKeyRepo.GetByIDForUser(accessKeyID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: access key not found", consts.ErrNotFound)
		}
		return err
	}

	key.Status = consts.CommonDeleted
	return s.accessKeyRepo.Update(key)
}

func (s *Service) DisableAccessKey(ctx context.Context, userID, accessKeyID int) error {
	return s.setAccessKeyStatus(userID, accessKeyID, consts.CommonDisabled)
}

func (s *Service) EnableAccessKey(ctx context.Context, userID, accessKeyID int) error {
	return s.setAccessKeyStatus(userID, accessKeyID, consts.CommonEnabled)
}

func (s *Service) RotateAccessKey(ctx context.Context, userID, accessKeyID int) (*AccessKeyWithSecretResp, error) {
	key, err := s.accessKeyRepo.GetByIDForUser(accessKeyID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: access key not found", consts.ErrNotFound)
		}
		return nil, err
	}

	secretKeyValue, err := generateCredentialValue("sk_", 24)
	if err != nil {
		return nil, fmt.Errorf("failed to generate secret key: %w", err)
	}
	secretHash, err := utils.HashPassword(secretKeyValue)
	if err != nil {
		return nil, fmt.Errorf("failed to hash secret key: %w", err)
	}
	secretCiphertext, err := utils.EncryptAccessKeySecret(secretKeyValue)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt secret key: %w", err)
	}

	key.SecretHash = secretHash
	key.SecretCiphertext = secretCiphertext
	if err := s.accessKeyRepo.Update(key); err != nil {
		return nil, err
	}

	return &AccessKeyWithSecretResp{
		AccessKeyInfo: *NewAccessKeyInfo(key),
		SecretKey:     secretKeyValue,
	}, nil
}

func (s *Service) ExchangeAccessKeyToken(ctx context.Context, req *AccessKeyTokenReq, method, path string) (*AccessKeyTokenResp, error) {
	if req == nil {
		return nil, fmt.Errorf("access key token request is nil")
	}

	key, err := s.accessKeyRepo.GetByAccessKey(req.AccessKey)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: invalid access key or secret key", consts.ErrAuthenticationFailed)
		}
		return nil, err
	}

	if key.Status != consts.CommonEnabled {
		return nil, fmt.Errorf("%w: access key is disabled", consts.ErrAuthenticationFailed)
	}
	if key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("%w: access key is expired", consts.ErrAuthenticationFailed)
	}
	timestampUnix, err := req.TimestampUnix()
	if err != nil {
		return nil, fmt.Errorf("%w: invalid request timestamp", consts.ErrAuthenticationFailed)
	}
	now := time.Now()
	requestTime := time.Unix(timestampUnix, 0)
	if requestTime.Before(now.Add(-accessKeySignatureTTL)) || requestTime.After(now.Add(accessKeySignatureTTL)) {
		return nil, fmt.Errorf("%w: request timestamp is outside the allowed window", consts.ErrAuthenticationFailed)
	}

	secretKey, err := utils.DecryptAccessKeySecret(key.SecretCiphertext)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt access key secret: %w", err)
	}
	if !utils.VerifyAccessKeyRequestSignature(secretKey, req.CanonicalString(method, path), req.Signature) {
		return nil, fmt.Errorf("%w: invalid access key signature", consts.ErrAuthenticationFailed)
	}
	if err := s.tokenStore.ReserveAccessKeyNonce(ctx, key.AccessKey, req.Nonce, accessKeySignatureTTL); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(key.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: access key owner not found", consts.ErrAuthenticationFailed)
		}
		return nil, err
	}
	if !user.IsActive || user.Status != consts.CommonEnabled {
		return nil, fmt.Errorf("%w: access key owner is inactive", consts.ErrAuthenticationFailed)
	}

	token, expiresAt, err := s.generateAccessKeyTokenWithRoles(s.roleRepo, user, key.ID)
	if err != nil {
		return nil, err
	}

	if err := s.accessKeyRepo.UpdateLastUsedAt(key.ID, time.Now()); err != nil {
		logrus.WithError(err).Warn("failed to update access key last used time")
	}

	return &AccessKeyTokenResp{
		Token:     token,
		TokenType: "Bearer",
		ExpiresAt: expiresAt,
		AuthType:  "access_key",
		AccessKey: key.AccessKey,
	}, nil
}

func (s *Service) generateTokenWithRoles(roleRepo *RoleRepository, user *model.User) (string, time.Time, error) {
	roles, err := roleRepo.ListByUserID(user.ID)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to get user roles: %w", err)
	}

	isAdmin := false
	roleNames := make([]string, 0, len(roles))
	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
		if role.Name == string(consts.RoleSuperAdmin) || role.Name == string(consts.RoleAdmin) {
			isAdmin = true
		}
	}

	token, expiresAt, err := utils.GenerateToken(user.ID, user.Username, user.Email, user.IsActive, isAdmin, roleNames)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to generate token: %w", err)
	}

	return token, expiresAt, nil
}

func (s *Service) generateAccessKeyTokenWithRoles(roleRepo *RoleRepository, user *model.User, accessKeyID int) (string, time.Time, error) {
	roles, err := roleRepo.ListByUserID(user.ID)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to get user roles: %w", err)
	}

	isAdmin := false
	roleNames := make([]string, 0, len(roles))
	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
		if role.Name == string(consts.RoleSuperAdmin) || role.Name == string(consts.RoleAdmin) {
			isAdmin = true
		}
	}

	token, expiresAt, err := utils.GenerateAccessKeyToken(user.ID, user.Username, user.Email, user.IsActive, isAdmin, roleNames, accessKeyID)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to generate access key token: %w", err)
	}

	return token, expiresAt, nil
}

func (s *Service) getAllUserResourceRoles(userID int) ([]usermodule.UserContainerInfo, []usermodule.UserDatasetInfo, []usermodule.UserProjectInfo, error) {
	userContainers, err := s.userRepo.ListContainerRoles(userID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to list user-container roles: %w", err)
	}
	containerRoles := make([]usermodule.UserContainerInfo, 0, len(userContainers))
	for _, uc := range userContainers {
		containerRoles = append(containerRoles, *usermodule.NewUserContainerInfo(&uc))
	}

	userDatasets, err := s.userRepo.ListDatasetRoles(userID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to list user-dataset roles: %w", err)
	}
	datasetRoles := make([]usermodule.UserDatasetInfo, 0, len(userDatasets))
	for _, ud := range userDatasets {
		datasetRoles = append(datasetRoles, *usermodule.NewUserDatasetInfo(&ud))
	}

	userProjects, err := s.userRepo.ListProjectRoles(userID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to list user-project roles: %w", err)
	}
	projectRoles := make([]usermodule.UserProjectInfo, 0, len(userProjects))
	for _, up := range userProjects {
		projectRoles = append(projectRoles, *usermodule.NewUserProjectInfo(&up))
	}

	return containerRoles, datasetRoles, projectRoles, nil
}

func (s *Service) setAccessKeyStatus(userID, accessKeyID int, status consts.StatusType) error {
	key, err := s.accessKeyRepo.GetByIDForUser(accessKeyID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: access key not found", consts.ErrNotFound)
		}
		return err
	}

	key.Status = status
	return s.accessKeyRepo.Update(key)
}

func generateCredentialValue(prefix string, randomBytes int) (string, error) {
	buf := make([]byte, randomBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(buf), nil
}
