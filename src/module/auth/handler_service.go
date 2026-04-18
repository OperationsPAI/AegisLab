package authmodule

import (
	"context"

	"aegis/utils"
)

// HandlerService captures the auth operations consumed by the HTTP handler.
type HandlerService interface {
	Login(context.Context, *LoginReq) (*LoginResp, error)
	Register(context.Context, *RegisterReq) (*UserInfo, error)
	RefreshToken(context.Context, *TokenRefreshReq) (*TokenRefreshResp, error)
	Logout(context.Context, *utils.Claims) error
	ChangePassword(context.Context, *ChangePasswordReq, int) error
	GetProfile(context.Context, int) (*UserProfileResp, error)
	CreateAPIKey(context.Context, int, *CreateAPIKeyReq) (*APIKeyWithSecretResp, error)
	ListAPIKeys(context.Context, int, *ListAPIKeyReq) (*ListAPIKeyResp, error)
	GetAPIKey(context.Context, int, int) (*APIKeyInfo, error)
	DeleteAPIKey(context.Context, int, int) error
	DisableAPIKey(context.Context, int, int) error
	EnableAPIKey(context.Context, int, int) error
	RevokeAPIKey(context.Context, int, int) error
	RotateAPIKey(context.Context, int, int) (*APIKeyWithSecretResp, error)
	ExchangeAPIKeyToken(context.Context, *APIKeyTokenReq, string, string) (*APIKeyTokenResp, error)
}

func AsHandlerService(service *Service) HandlerService {
	return service
}
