package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAPIKeyScopeMatchesTarget(t *testing.T) {
	tests := []struct {
		name   string
		scope  string
		target string
		want   bool
	}{
		{name: "global wildcard", scope: "*", target: "sdk:evaluations:read", want: true},
		{name: "sdk wildcard", scope: "sdk:*", target: "sdk:evaluations:read", want: true},
		{name: "sdk evaluations wildcard", scope: "sdk:evaluations:*", target: "sdk:evaluations:read", want: true},
		{name: "exact match", scope: "sdk:datasets:read", target: "sdk:datasets:read", want: true},
		{name: "resource only", scope: "sdk", target: "sdk:datasets:read", want: true},
		{name: "different family", scope: "project:read", target: "sdk:evaluations:read", want: false},
		{name: "different action", scope: "sdk:evaluations:write", target: "sdk:evaluations:read", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := apiKeyScopeMatchesTarget(tt.scope, tt.target); got != tt.want {
				t.Fatalf("apiKeyScopeMatchesTarget(%q, %q) = %v, want %v", tt.scope, tt.target, got, tt.want)
			}
		})
	}
}

func TestRequireAPIKeyScopesAny(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		setup      func(*gin.Context)
		wantStatus int
	}{
		{
			name: "user token bypasses explicit sdk scope gate",
			setup: func(c *gin.Context) {
				c.Set("user_id", 1)
				c.Set("is_active", true)
				c.Set("auth_type", "user")
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "api key with matching sdk scope passes",
			setup: func(c *gin.Context) {
				c.Set("user_id", 1)
				c.Set("is_active", true)
				c.Set("auth_type", "api_key")
				c.Set("api_key_scopes", []string{"sdk:evaluations:read"})
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "api key with non matching sdk scope denied",
			setup: func(c *gin.Context) {
				c.Set("user_id", 1)
				c.Set("is_active", true)
				c.Set("auth_type", "api_key")
				c.Set("api_key_scopes", []string{"sdk:datasets:read"})
			},
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := runMiddlewareChain(func(r *gin.Engine) {
				r.GET("/", func(c *gin.Context) {
					tt.setup(c)
					c.Next()
				}, RequireAPIKeyScopesAny("sdk:*", "sdk:evaluations:*", "sdk:evaluations:read"), func(c *gin.Context) {
					c.Status(http.StatusNoContent)
				})
			})
			if status != tt.wantStatus {
				t.Fatalf("status = %d, want %d", status, tt.wantStatus)
			}
		})
	}
}

func TestRequireHumanUserAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		setup      func(*gin.Context)
		wantStatus int
	}{
		{
			name: "human user token passes",
			setup: func(c *gin.Context) {
				c.Set("user_id", 1)
				c.Set("is_active", true)
				c.Set("auth_type", "user")
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "api key token denied",
			setup: func(c *gin.Context) {
				c.Set("user_id", 1)
				c.Set("is_active", true)
				c.Set("auth_type", "api_key")
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name: "service token denied",
			setup: func(c *gin.Context) {
				c.Set("is_service_token", true)
				c.Set("task_id", "task-1")
			},
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := runMiddlewareChain(func(r *gin.Engine) {
				r.GET("/", func(c *gin.Context) {
					tt.setup(c)
					c.Next()
				}, RequireHumanUserAuth(), func(c *gin.Context) {
					c.Status(http.StatusNoContent)
				})
			})
			if status != tt.wantStatus {
				t.Fatalf("status = %d, want %d", status, tt.wantStatus)
			}
		})
	}
}

func runMiddlewareChain(register func(*gin.Engine)) int {
	engine := gin.New()
	register(engine)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	return w.Code
}
