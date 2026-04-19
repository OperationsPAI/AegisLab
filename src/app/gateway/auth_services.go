package gateway

import (
	"context"

	auth "aegis/module/auth"
	"aegis/utils"
)

type authIAMClient interface {
	Enabled() bool
	Login(context.Context, *auth.LoginReq) (*auth.LoginResp, error)
	Register(context.Context, *auth.RegisterReq) (*auth.UserInfo, error)
	RefreshToken(context.Context, *auth.TokenRefreshReq) (*auth.TokenRefreshResp, error)
	Logout(context.Context, *utils.Claims) error
	ChangePassword(context.Context, *auth.ChangePasswordReq, int) error
	GetProfile(context.Context, int) (*auth.UserProfileResp, error)
	CreateAPIKey(context.Context, int, *auth.CreateAPIKeyReq) (*auth.APIKeyWithSecretResp, error)
	ListAPIKeys(context.Context, int, *auth.ListAPIKeyReq) (*auth.ListAPIKeyResp, error)
	GetAPIKey(context.Context, int, int) (*auth.APIKeyInfo, error)
	DeleteAPIKey(context.Context, int, int) error
	DisableAPIKey(context.Context, int, int) error
	EnableAPIKey(context.Context, int, int) error
	RevokeAPIKey(context.Context, int, int) error
	RotateAPIKey(context.Context, int, int) (*auth.APIKeyWithSecretResp, error)
	ExchangeAPIKeyToken(context.Context, *auth.APIKeyTokenReq, string, string) (*auth.APIKeyTokenResp, error)
}

type remoteAwareAuthService struct {
	auth.HandlerService
	iam authIAMClient
}

func (s remoteAwareAuthService) Login(ctx context.Context, req *auth.LoginReq) (*auth.LoginResp, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.Login(ctx, req)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareAuthService) Register(ctx context.Context, req *auth.RegisterReq) (*auth.UserInfo, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.Register(ctx, req)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareAuthService) RefreshToken(ctx context.Context, req *auth.TokenRefreshReq) (*auth.TokenRefreshResp, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.RefreshToken(ctx, req)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareAuthService) Logout(ctx context.Context, claims *utils.Claims) error {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.Logout(ctx, claims)
	}
	return missingRemoteDependency("iam-service")
}

func (s remoteAwareAuthService) ChangePassword(ctx context.Context, req *auth.ChangePasswordReq, userID int) error {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.ChangePassword(ctx, req, userID)
	}
	return missingRemoteDependency("iam-service")
}

func (s remoteAwareAuthService) GetProfile(ctx context.Context, userID int) (*auth.UserProfileResp, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.GetProfile(ctx, userID)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareAuthService) CreateAPIKey(ctx context.Context, userID int, req *auth.CreateAPIKeyReq) (*auth.APIKeyWithSecretResp, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.CreateAPIKey(ctx, userID, req)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareAuthService) ListAPIKeys(ctx context.Context, userID int, req *auth.ListAPIKeyReq) (*auth.ListAPIKeyResp, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.ListAPIKeys(ctx, userID, req)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareAuthService) GetAPIKey(ctx context.Context, userID, accessKeyID int) (*auth.APIKeyInfo, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.GetAPIKey(ctx, userID, accessKeyID)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareAuthService) DeleteAPIKey(ctx context.Context, userID, accessKeyID int) error {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.DeleteAPIKey(ctx, userID, accessKeyID)
	}
	return missingRemoteDependency("iam-service")
}

func (s remoteAwareAuthService) DisableAPIKey(ctx context.Context, userID, accessKeyID int) error {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.DisableAPIKey(ctx, userID, accessKeyID)
	}
	return missingRemoteDependency("iam-service")
}

func (s remoteAwareAuthService) EnableAPIKey(ctx context.Context, userID, accessKeyID int) error {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.EnableAPIKey(ctx, userID, accessKeyID)
	}
	return missingRemoteDependency("iam-service")
}

func (s remoteAwareAuthService) RevokeAPIKey(ctx context.Context, userID, accessKeyID int) error {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.RevokeAPIKey(ctx, userID, accessKeyID)
	}
	return missingRemoteDependency("iam-service")
}

func (s remoteAwareAuthService) RotateAPIKey(ctx context.Context, userID, accessKeyID int) (*auth.APIKeyWithSecretResp, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.RotateAPIKey(ctx, userID, accessKeyID)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareAuthService) ExchangeAPIKeyToken(ctx context.Context, req *auth.APIKeyTokenReq, method, path string) (*auth.APIKeyTokenResp, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.ExchangeAPIKeyToken(ctx, req, method, path)
	}
	return nil, missingRemoteDependency("iam-service")
}
