package utils

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWT configuration
const (
	// Should be loaded from environment variables in production
	JWTSecret              = "your-secret-key-change-this-in-production"
	TokenExpiration        = 24 * time.Hour
	RefreshTokenExpiration = 7 * 24 * time.Hour
	ServiceTokenExpiration = 24 * time.Hour // Service token for K8s jobs
)

// Claims represents JWT claims structure
type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	IsActive bool   `json:"is_active"`
	jwt.RegisteredClaims
}

// RefreshClaims represents refresh token claims
type RefreshClaims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// ServiceClaims represents service token claims for K8s jobs
type ServiceClaims struct {
	TaskID string `json:"task_id"` // Associated task ID
	jwt.RegisteredClaims
}

// GenerateToken generates a new JWT token for the given user
func GenerateToken(userID int, username, email string, isActive bool) (string, time.Time, error) {
	expirationTime := time.Now().Add(TokenExpiration)

	claims := &Claims{
		UserID:   userID,
		Username: username,
		Email:    email,
		IsActive: isActive,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        fmt.Sprintf("jwt_%d_%d", userID, time.Now().Unix()), // JWT ID (jti)
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "rcabench",
			Subject:   strconv.Itoa(userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(JWTSecret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to generate token: %v", err)
	}

	return tokenString, expirationTime, nil
}

// GenerateRefreshToken generates a refresh token with longer expiration
func GenerateRefreshToken(userID int, username string) (string, time.Time, error) {
	expirationTime := time.Now().Add(RefreshTokenExpiration)

	claims := &RefreshClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "rcabench-refresh",
			Subject:   strconv.Itoa(userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(JWTSecret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to generate refresh token: %v", err)
	}

	return tokenString, expirationTime, nil
}

// GenerateServiceToken generates a service token for K8s jobs
// This token is used for job-to-service authentication without exposing user credentials
func GenerateServiceToken(taskID string) (string, time.Time, error) {
	expirationTime := time.Now().Add(ServiceTokenExpiration)

	claims := &ServiceClaims{
		TaskID: taskID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        fmt.Sprintf("svc_%s_%d", taskID, time.Now().Unix()), // Service token ID
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "rcabench-service",
			Subject:   taskID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(JWTSecret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to generate service token: %v", err)
	}

	return tokenString, expirationTime, nil
}

// ValidateToken validates and parses a JWT token
func ValidateToken(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, errors.New("token is required")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(JWTSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %v", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Note: No need to manually check ExpiresAt - jwt.ParseWithClaims already validates it
	// If token is expired, ParseWithClaims will return an error above

	// Check if user is active
	if !claims.IsActive {
		return nil, errors.New("user account is inactive")
	}

	return claims, nil
}

// ValidateTokenWithCustomClaims validates token with custom claims validation
func ValidateTokenWithCustomClaims(tokenString string, validateFunc func(*Claims) error) (*Claims, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if validateFunc != nil {
		if err := validateFunc(claims); err != nil {
			return nil, fmt.Errorf("custom validation failed: %v", err)
		}
	}

	return claims, nil
}

// ValidateServiceToken validates and parses a service token
func ValidateServiceToken(tokenString string) (*ServiceClaims, error) {
	if tokenString == "" {
		return nil, errors.New("service token is required")
	}

	token, err := jwt.ParseWithClaims(tokenString, &ServiceClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(JWTSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse service token: %v", err)
	}

	claims, ok := token.Claims.(*ServiceClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid service token")
	}

	// Verify it's a service token
	if claims.Issuer != "rcabench-service" {
		return nil, errors.New("not a valid service token")
	}

	return claims, nil
}

// RefreshToken refreshes an existing token if it's a valid refresh token
func RefreshToken(refreshTokenString string) (string, time.Time, error) {
	// Parse refresh token
	token, err := jwt.ParseWithClaims(refreshTokenString, &RefreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(JWTSecret), nil
	})

	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to parse refresh token: %v", err)
	}

	refreshClaims, ok := token.Claims.(*RefreshClaims)
	if !ok || !token.Valid {
		return "", time.Time{}, errors.New("invalid refresh token")
	}

	// Note: No need to manually check ExpiresAt - jwt.ParseWithClaims already validates it

	// Check if it's a refresh token (issued by refresh issuer)
	if refreshClaims.Issuer != "rcabench-refresh" {
		return "", time.Time{}, errors.New("not a valid refresh token")
	}

	// Generate new access token
	// Note: In production, you should fetch fresh user data from database
	return GenerateToken(refreshClaims.UserID, refreshClaims.Username, "", true)
}

// GetUserIDFromToken extracts user ID from a valid token
func GetUserIDFromToken(tokenString string) (int, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}

// GetUsernameFromToken extracts username from a valid token
func GetUsernameFromToken(tokenString string) (string, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return "", err
	}
	return claims.Username, nil
}

// ParseTokenWithoutValidation parses a token without validating signature or expiration
// Use this only for extracting claims from expired tokens for logging purposes
func ParseTokenWithoutValidation(tokenString string) (*Claims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %v", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// ExtractTokenFromHeader extracts the JWT token from the Authorization header
func ExtractTokenFromHeader(header string) (string, error) {
	if header == "" {
		return "", errors.New("authorization header is empty")
	}

	parts := strings.Split(header, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("invalid authorization header format")
	}

	return parts[1], nil
}
