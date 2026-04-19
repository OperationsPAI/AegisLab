package middleware

import (
	"net/http"
	"strings"

	"aegis/dto"

	"github.com/gin-gonic/gin"
)

func apiKeyScopeMatchesTarget(scope, target string) bool {
	scope = strings.TrimSpace(scope)
	target = strings.TrimSpace(target)
	if scope == "" || target == "" {
		return false
	}
	if scope == "*" {
		return true
	}

	targetParts := strings.Split(target, ":")
	scopeParts := strings.Split(scope, ":")
	if len(scopeParts) > len(targetParts) {
		return false
	}
	for len(scopeParts) < len(targetParts) {
		scopeParts = append(scopeParts, "*")
	}
	for i := range targetParts {
		part := strings.TrimSpace(scopeParts[i])
		if part == "*" {
			continue
		}
		if part != targetParts[i] {
			return false
		}
	}
	return true
}

func apiKeyScopesAllowAnyTarget(scopes, targets []string) bool {
	if len(scopes) == 0 || len(targets) == 0 {
		return false
	}
	for _, scope := range scopes {
		for _, target := range targets {
			if apiKeyScopeMatchesTarget(scope, target) {
				return true
			}
		}
	}
	return false
}

// RequireHumanUserAuth rejects service tokens and API key bearer tokens.
// It is intended for self-service user/account endpoints.
func RequireHumanUserAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !RequireUserAuth(c) {
			c.Abort()
			return
		}
		if GetAuthType(c) == "api_key" {
			dto.ErrorResponse(c, http.StatusForbidden, "User session required, API key token not allowed")
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireAPIKeyScopesAny applies explicit scope checks only to API key bearer tokens.
// Human user tokens continue through unchanged.
func RequireAPIKeyScopesAny(targets ...string) gin.HandlerFunc {
	trimmed := make([]string, 0, len(targets))
	for _, target := range targets {
		target = strings.TrimSpace(target)
		if target != "" {
			trimmed = append(trimmed, target)
		}
	}

	return func(c *gin.Context) {
		if GetAuthType(c) != "api_key" {
			c.Next()
			return
		}
		scopes, ok := GetCurrentAPIKeyScopes(c)
		if !ok || !apiKeyScopesAllowAnyTarget(scopes, trimmed) {
			dto.ErrorResponse(c, http.StatusForbidden, "API key scope does not allow this endpoint")
			c.Abort()
			return
		}
		c.Next()
	}
}
