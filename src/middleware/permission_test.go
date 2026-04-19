package middleware

import (
	"context"
	"testing"

	"aegis/consts"
	"aegis/dto"
	"aegis/utils"
)

type permissionCheckerStub struct{}

func (permissionCheckerStub) VerifyToken(context.Context, string) (*utils.Claims, error) {
	return nil, nil
}
func (permissionCheckerStub) VerifyServiceToken(context.Context, string) (*utils.ServiceClaims, error) {
	return nil, nil
}
func (permissionCheckerStub) CheckUserPermission(context.Context, *dto.CheckPermissionParams) (bool, error) {
	return false, nil
}
func (permissionCheckerStub) IsUserTeamAdmin(context.Context, int, int) (bool, error) {
	return false, nil
}
func (permissionCheckerStub) IsUserInTeam(context.Context, int, int) (bool, error) {
	return false, nil
}
func (permissionCheckerStub) IsTeamPublic(context.Context, int) (bool, error) {
	return false, nil
}
func (permissionCheckerStub) IsUserProjectAdmin(context.Context, int, int) (bool, error) {
	return false, nil
}
func (permissionCheckerStub) IsUserInProject(context.Context, int, int) (bool, error) {
	return false, nil
}
func (permissionCheckerStub) LogFailedAction(string, string, string, string, int, int, consts.ResourceName) error {
	return nil
}
func (permissionCheckerStub) LogUserAction(string, string, string, string, int, int, consts.ResourceName) error {
	return nil
}

func TestAPIKeyScopeMatchesPermission(t *testing.T) {
	permission := consts.PermProjectReadAll

	tests := []struct {
		name  string
		scope string
		want  bool
	}{
		{name: "wildcard all", scope: "*", want: true},
		{name: "resource only", scope: "project", want: true},
		{name: "resource action", scope: "project:read", want: true},
		{name: "resource action scope", scope: "project:read:all", want: true},
		{name: "resource wildcard action", scope: "project:*", want: true},
		{name: "resource action wildcard scope", scope: "project:read:*", want: true},
		{name: "full wildcard segments", scope: "project:*:*", want: true},
		{name: "other action", scope: "project:update", want: false},
		{name: "other resource", scope: "dataset:read", want: false},
		{name: "too specific mismatch", scope: "project:read:team", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := apiKeyScopeMatchesPermission(tt.scope, permission); got != tt.want {
				t.Fatalf("apiKeyScopeMatchesPermission(%q, %q) = %v, want %v", tt.scope, permission.String(), got, tt.want)
			}
		})
	}
}

func TestPermissionContextScopeAllowsPermission(t *testing.T) {
	ctx := &permissionContext{
		authType:     "api_key",
		apiKeyScopes: []string{"project:read", "execution:execute:project"},
	}

	if !ctx.scopeAllowsPermission(consts.PermProjectReadAll) {
		t.Fatalf("scopeAllowsPermission(project read) = false, want true")
	}
	if ctx.scopeAllowsPermission(consts.PermProjectUpdateAll) {
		t.Fatalf("scopeAllowsPermission(project update) = true, want false")
	}
	if !ctx.scopeAllowsPermission(consts.PermExecutionExecuteProject) {
		t.Fatalf("scopeAllowsPermission(execution execute project) = false, want true")
	}
}

func TestPermissionContextScopeAllowsAnyPermission(t *testing.T) {
	ctx := &permissionContext{
		authType:     "api_key",
		apiKeyScopes: []string{"team:read"},
	}

	if !ctx.scopeAllowsAnyPermission([]consts.PermissionRule{consts.PermTeamManageAll, consts.PermTeamReadAll}) {
		t.Fatalf("scopeAllowsAnyPermission(team manage/read) = false, want true")
	}
	if ctx.scopeAllowsAnyPermission([]consts.PermissionRule{consts.PermProjectManageAll, consts.PermProjectReadAll}) {
		t.Fatalf("scopeAllowsAnyPermission(project manage/read) = true, want false")
	}
}

func TestTeamAccessCheckScopes(t *testing.T) {
	memberCheck := teamAccessCheck(false)
	adminCheck := teamAccessCheck(true)
	teamID := 9

	memberCtx := &permissionContext{
		authType:     "api_key",
		apiKeyScopes: []string{"team:read"},
		teamID:       &teamID,
		checker:      permissionCheckerStub{},
	}
	allowed, err := memberCheck(memberCtx)
	if err != nil {
		t.Fatalf("memberCheck() error = %v", err)
	}
	if allowed {
		t.Fatalf("memberCheck() = true, want false without membership backing")
	}

	allowed, err = adminCheck(memberCtx)
	if err != nil {
		t.Fatalf("adminCheck() error = %v", err)
	}
	if allowed {
		t.Fatalf("adminCheck() = true, want false for read-only scope")
	}

	adminCtx := &permissionContext{
		authType:     "api_key",
		apiKeyScopes: []string{"team:manage"},
		teamID:       &teamID,
		isAdmin:      true,
		checker:      permissionCheckerStub{},
	}
	allowed, err = adminCheck(adminCtx)
	if err != nil {
		t.Fatalf("adminCheck(manage) error = %v", err)
	}
	if !allowed {
		t.Fatalf("adminCheck(manage) = false, want true")
	}
}

func TestProjectAccessCheckScopes(t *testing.T) {
	memberCheck := projectAccessCheck(false)
	adminCheck := projectAccessCheck(true)
	projectID := 7

	memberCtx := &permissionContext{
		authType:     "api_key",
		apiKeyScopes: []string{"project:read"},
		projectID:    &projectID,
		checker:      permissionCheckerStub{},
	}
	allowed, err := memberCheck(memberCtx)
	if err != nil {
		t.Fatalf("memberCheck() error = %v", err)
	}
	if allowed {
		t.Fatalf("memberCheck() = true, want false without project membership")
	}

	allowed, err = adminCheck(memberCtx)
	if err != nil {
		t.Fatalf("adminCheck() error = %v", err)
	}
	if allowed {
		t.Fatalf("adminCheck() = true, want false for read-only scope")
	}

	adminCtx := &permissionContext{
		authType:     "api_key",
		apiKeyScopes: []string{"project:manage"},
		projectID:    &projectID,
		isAdmin:      true,
		checker:      permissionCheckerStub{},
	}
	allowed, err = adminCheck(adminCtx)
	if err != nil {
		t.Fatalf("adminCheck(manage) error = %v", err)
	}
	if !allowed {
		t.Fatalf("adminCheck(manage) = false, want true")
	}
}
