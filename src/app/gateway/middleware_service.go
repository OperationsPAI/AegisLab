package gateway

import (
	"context"

	"aegis/consts"
	"aegis/dto"
	"aegis/internalclient/iamclient"
	"aegis/middleware"
	"aegis/utils"
)

type remoteAwareMiddlewareService struct {
	base middleware.Service
	iam  *iamclient.Client
}

func (s remoteAwareMiddlewareService) VerifyToken(ctx context.Context, token string) (*utils.Claims, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.VerifyToken(ctx, token)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareMiddlewareService) VerifyServiceToken(ctx context.Context, token string) (*utils.ServiceClaims, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.VerifyServiceToken(ctx, token)
	}
	return nil, missingRemoteDependency("iam-service")
}

func (s remoteAwareMiddlewareService) CheckUserPermission(ctx context.Context, params *dto.CheckPermissionParams) (bool, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.CheckUserPermission(ctx, params)
	}
	return false, missingRemoteDependency("iam-service")
}

func (s remoteAwareMiddlewareService) IsUserTeamAdmin(ctx context.Context, userID, teamID int) (bool, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.IsUserTeamAdmin(ctx, userID, teamID)
	}
	return false, missingRemoteDependency("iam-service")
}

func (s remoteAwareMiddlewareService) IsUserInTeam(ctx context.Context, userID, teamID int) (bool, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.IsUserInTeam(ctx, userID, teamID)
	}
	return false, missingRemoteDependency("iam-service")
}

func (s remoteAwareMiddlewareService) IsTeamPublic(ctx context.Context, teamID int) (bool, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.IsTeamPublic(ctx, teamID)
	}
	return false, missingRemoteDependency("iam-service")
}

func (s remoteAwareMiddlewareService) IsUserProjectAdmin(ctx context.Context, userID, projectID int) (bool, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.IsUserProjectAdmin(ctx, userID, projectID)
	}
	return false, missingRemoteDependency("iam-service")
}

func (s remoteAwareMiddlewareService) IsUserInProject(ctx context.Context, userID, projectID int) (bool, error) {
	if s.iam != nil && s.iam.Enabled() {
		return s.iam.IsUserInProject(ctx, userID, projectID)
	}
	return false, missingRemoteDependency("iam-service")
}

func (s remoteAwareMiddlewareService) LogFailedAction(ipAddress, userAgent, action, errorMsg string, duration, userID int, resourceName consts.ResourceName) error {
	return s.base.LogFailedAction(ipAddress, userAgent, action, errorMsg, duration, userID, resourceName)
}

func (s remoteAwareMiddlewareService) LogUserAction(ipAddress, userAgent, action, details string, duration, userID int, resourceName consts.ResourceName) error {
	return s.base.LogUserAction(ipAddress, userAgent, action, details, duration, userID, resourceName)
}
