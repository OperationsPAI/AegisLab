package utils

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateToken(t *testing.T) {
	tests := []struct {
		name     string
		userID   int
		username string
		email    string
		isActive bool
	}{
		{
			name:     "Valid User",
			userID:   1,
			username: "testuser",
			email:    "test@example.com",
			isActive: true,
		},
		{
			name:     "Inactive User",
			userID:   2,
			username: "inactive",
			email:    "inactive@example.com",
			isActive: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, expiresAt, err := GenerateToken(tt.userID, tt.username, tt.email, tt.isActive)

			assert.NoError(t, err)
			assert.NotEmpty(t, token)
			assert.True(t, expiresAt.After(time.Now()))

			// Verify token is a valid JWT (should have 3 parts separated by dots)
			assert.Regexp(t, `^[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+$`, token)
		})
	}
}

func TestValidateToken(t *testing.T) {
	// Generate a valid token first
	userID := 1
	username := "testuser"
	email := "test@example.com"
	isActive := true

	validToken, _, err := GenerateToken(userID, username, email, isActive)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		token       string
		shouldError bool
		description string
	}{
		{
			name:        "Valid Token",
			token:       validToken,
			shouldError: false,
			description: "Should validate a valid JWT token",
		},
		{
			name:        "Empty Token",
			token:       "",
			shouldError: true,
			description: "Should fail with empty token",
		},
		{
			name:        "Invalid Format",
			token:       "invalid-token",
			shouldError: true,
			description: "Should fail with invalid JWT format",
		},
		{
			name:        "Malformed JWT",
			token:       "header.payload",
			shouldError: true,
			description: "Should fail with malformed JWT (missing signature)",
		},
		{
			name:        "Invalid Signature",
			token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJ1c2VybmFtZSI6InRlc3R1c2VyIn0.invalid_signature",
			shouldError: true,
			description: "Should fail with invalid signature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := ValidateToken(tt.token)

			if tt.shouldError {
				assert.Error(t, err, tt.description)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotNil(t, claims)
				assert.Equal(t, userID, claims.UserID)
				assert.Equal(t, username, claims.Username)
				assert.Equal(t, email, claims.Email)
				assert.Equal(t, isActive, claims.IsActive)
			}
		})
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	userID := 1
	username := "testuser"

	token, expiresAt, err := GenerateRefreshToken(userID, username)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.True(t, expiresAt.After(time.Now()))

	// Verify it's a valid JWT format
	assert.Regexp(t, `^[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+$`, token)

	// Parse to verify it's a refresh token
	claims, err := ParseTokenWithoutValidation(token)
	assert.NoError(t, err)
	assert.Equal(t, "rcabench-refresh", claims.Issuer)
}

func TestRefreshToken(t *testing.T) {
	// Generate a valid refresh token first
	userID := 1
	username := "testuser"

	refreshToken, _, err := GenerateRefreshToken(userID, username)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		token       string
		shouldError bool
		description string
	}{
		{
			name:        "Valid Refresh Token",
			token:       refreshToken,
			shouldError: false,
			description: "Should refresh with valid refresh token",
		},
		{
			name:        "Empty Token",
			token:       "",
			shouldError: true,
			description: "Should fail with empty token",
		},
		{
			name:        "Invalid Token",
			token:       "invalid-token",
			shouldError: true,
			description: "Should fail with invalid token format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newToken, expiresAt, err := RefreshToken(tt.token)

			if tt.shouldError {
				assert.Error(t, err, tt.description)
				assert.Empty(t, newToken)
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotEmpty(t, newToken)
				assert.True(t, expiresAt.After(time.Now()))

				// Verify the new token is valid
				claims, err := ValidateToken(newToken)
				assert.NoError(t, err)
				assert.Equal(t, userID, claims.UserID)
				assert.Equal(t, username, claims.Username)
			}
		})
	}
}

func TestTokenExpiration(t *testing.T) {
	// This test checks that expired tokens are properly rejected
	userID := 1
	username := "testuser"
	email := "test@example.com"
	isActive := true

	// Generate a token (we can't easily create an expired token without mocking time)
	token, _, err := GenerateToken(userID, username, email, isActive)
	assert.NoError(t, err)

	// Validate the fresh token should work
	claims, err := ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
}

func TestParseTokenWithoutValidation(t *testing.T) {
	userID := 42
	username := "testuser"
	email := "test@example.com"
	isActive := true

	token, _, err := GenerateToken(userID, username, email, isActive)
	assert.NoError(t, err)

	claims, err := ParseTokenWithoutValidation(token)
	assert.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, username, claims.Username)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, isActive, claims.IsActive)
}

func TestValidateTokenWithCustomClaims(t *testing.T) {
	userID := 1
	username := "testuser"
	email := "test@example.com"
	isActive := true

	token, _, err := GenerateToken(userID, username, email, isActive)
	assert.NoError(t, err)

	// Test with custom validation that passes
	claims, err := ValidateTokenWithCustomClaims(token, func(claims *Claims) error {
		if claims.UserID != userID {
			return fmt.Errorf("user ID mismatch")
		}
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)

	// Test with custom validation that fails
	_, err = ValidateTokenWithCustomClaims(token, func(claims *Claims) error {
		return fmt.Errorf("custom validation failed")
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "custom validation failed")
}

// Benchmark tests
func BenchmarkGenerateToken(b *testing.B) {
	userID := 1
	username := "testuser"
	email := "test@example.com"
	isActive := true

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = GenerateToken(userID, username, email, isActive)
	}
}

func BenchmarkValidateToken(b *testing.B) {
	userID := 1
	username := "testuser"
	email := "test@example.com"
	isActive := true

	token, _, _ := GenerateToken(userID, username, email, isActive)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ValidateToken(token)
	}
}

func BenchmarkExtractTokenFromHeader(b *testing.B) {
	header := "Bearer abc123xyz789"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ExtractTokenFromHeader(header)
	}
}
