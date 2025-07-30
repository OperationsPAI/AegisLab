package v2

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/LGU-SE-Internal/rcabench/database"
	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/utils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() error {
	// Use in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		return err
	}

	database.DB = db

	// Auto-migrate tables
	err = db.AutoMigrate(
		&database.User{},
		&database.Role{},
		&database.Permission{},
		&database.UserRole{},
		&database.RolePermission{},
	)
	if err != nil {
		return err
	}

	// Create test data
	hashedPassword, _ := utils.HashPassword("testpassword123")
	testUser := database.User{
		Username: "testuser",
		Email:    "test@example.com",
		FullName: "Test User",
		IsActive: true,
		Password: hashedPassword,
	}
	db.Create(&testUser)

	return nil
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestMain(m *testing.M) {
	// Setup test database
	err := setupTestDB()
	if err != nil {
		panic("Failed to setup test database: " + err.Error())
	}

	// Run tests
	code := m.Run()

	// Cleanup
	os.Exit(code)
}

func TestLogin(t *testing.T) {
	router := setupTestRouter()
	router.POST("/auth/login", Login)

	// Test password hashing first
	testPassword := "testpassword123"
	hashedPassword, err := utils.HashPassword(testPassword)
	assert.NoError(t, err)

	// Test verification
	isValid := utils.VerifyPassword(testPassword, hashedPassword)
	assert.True(t, isValid, "Password verification should work")

	tests := []struct {
		name         string
		requestBody  dto.LoginRequest
		expectedCode int
		description  string
	}{
		{
			name: "Valid Login",
			requestBody: dto.LoginRequest{
				Username: "testuser",
				Password: "testpassword123",
			},
			expectedCode: http.StatusOK,
			description:  "Should succeed with valid credentials",
		},
		{
			name: "Missing Username",
			requestBody: dto.LoginRequest{
				Password: "testpassword123",
			},
			expectedCode: http.StatusBadRequest,
			description:  "Should fail when username is missing",
		},
		{
			name: "Missing Password",
			requestBody: dto.LoginRequest{
				Username: "testuser",
			},
			expectedCode: http.StatusBadRequest,
			description:  "Should fail when password is missing",
		},
		{
			name:         "Empty Request",
			requestBody:  dto.LoginRequest{},
			expectedCode: http.StatusBadRequest,
			description:  "Should fail with empty request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBytes, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code, tt.description)

			if w.Code == http.StatusOK {
				var response dto.GenericResponse[dto.LoginResponse]
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.Data.Token, "Token should not be empty")
				assert.NotZero(t, response.Data.ExpiresAt, "ExpiresAt should not be zero")
			}
		})
	}
}

func TestRegister(t *testing.T) {
	router := setupTestRouter()
	router.POST("/auth/register", Register)

	tests := []struct {
		name         string
		requestBody  dto.RegisterRequest
		expectedCode int
		description  string
	}{
		{
			name: "Valid Registration",
			requestBody: dto.RegisterRequest{
				Username: "newuser",
				Email:    "newuser@example.com",
				Password: "password123",
				FullName: "New User",
			},
			expectedCode: http.StatusCreated,
			description:  "Should succeed with valid registration data",
		},
		{
			name: "Missing Username",
			requestBody: dto.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				FullName: "Test User",
			},
			expectedCode: http.StatusBadRequest,
			description:  "Should fail when username is missing",
		},
		{
			name: "Invalid Email",
			requestBody: dto.RegisterRequest{
				Username: "testuser",
				Email:    "invalid-email",
				Password: "password123",
				FullName: "Test User",
			},
			expectedCode: http.StatusBadRequest,
			description:  "Should fail with invalid email format",
		},
		{
			name: "Short Password",
			requestBody: dto.RegisterRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "123",
				FullName: "Test User",
			},
			expectedCode: http.StatusBadRequest,
			description:  "Should fail with password too short",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBytes, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code, tt.description)

			if w.Code == http.StatusCreated {
				var response dto.GenericResponse[dto.UserInfo]
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.requestBody.Username, response.Data.Username)
				assert.Equal(t, tt.requestBody.Email, response.Data.Email)
				assert.Equal(t, tt.requestBody.FullName, response.Data.FullName)
			}
		})
	}
}

func TestRefreshToken(t *testing.T) {
	router := setupTestRouter()
	router.POST("/auth/refresh", RefreshToken)

	tests := []struct {
		name         string
		requestBody  dto.TokenRefreshRequest
		expectedCode int
		description  string
	}{
		{
			name:         "Missing Token",
			requestBody:  dto.TokenRefreshRequest{},
			expectedCode: http.StatusBadRequest,
			description:  "Should fail when token is missing",
		},
		{
			name: "Invalid Token Format",
			requestBody: dto.TokenRefreshRequest{
				Token: "invalid-token",
			},
			expectedCode: http.StatusUnauthorized,
			description:  "Should fail with invalid token format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBytes, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/auth/refresh", bytes.NewBuffer(jsonBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code, tt.description)
		})
	}
}

func TestLogout(t *testing.T) {
	router := setupTestRouter()
	router.POST("/auth/logout", Logout)

	// Generate a valid token for testing
	validToken, _, _ := utils.GenerateToken(1, "testuser", "test@example.com", true)

	tests := []struct {
		name         string
		token        string
		expectedCode int
		description  string
	}{
		{
			name:         "Valid Logout",
			token:        validToken,
			expectedCode: http.StatusOK,
			description:  "Should succeed with valid token",
		},
		{
			name:         "Missing Token",
			token:        "",
			expectedCode: http.StatusBadRequest,
			description:  "Should fail when token is missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/auth/logout", nil)
			req.Header.Set("Content-Type", "application/json")

			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Debug: print response body if test fails
			if w.Code != tt.expectedCode {
				t.Logf("Response body: %s", w.Body.String())
			}

			assert.Equal(t, tt.expectedCode, w.Code, tt.description)
		})
	}
}

func TestChangePassword(t *testing.T) {
	router := setupTestRouter()
	router.POST("/auth/change-password", ChangePassword)

	tests := []struct {
		name         string
		requestBody  dto.ChangePasswordRequest
		expectedCode int
		description  string
	}{
		{
			name: "Missing Old Password",
			requestBody: dto.ChangePasswordRequest{
				NewPassword: "newpassword123",
			},
			expectedCode: http.StatusBadRequest,
			description:  "Should fail when old password is missing",
		},
		{
			name: "Missing New Password",
			requestBody: dto.ChangePasswordRequest{
				OldPassword: "oldpassword123",
			},
			expectedCode: http.StatusBadRequest,
			description:  "Should fail when new password is missing",
		},
		{
			name: "Short New Password",
			requestBody: dto.ChangePasswordRequest{
				OldPassword: "oldpassword123",
				NewPassword: "123",
			},
			expectedCode: http.StatusBadRequest,
			description:  "Should fail when new password is too short",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBytes, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/auth/change-password", bytes.NewBuffer(jsonBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code, tt.description)
		})
	}
}

// Integration test helpers
func createTestUser(t *testing.T) dto.UserInfo {
	// This would create a test user in the database
	// For now, return a mock user
	return dto.UserInfo{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		FullName: "Test User",
		IsActive: true,
	}
}

func generateTestToken(userID int, username, email string) (string, error) {
	// This would generate a real test token
	// For now, return a mock token
	return "test-token-" + username, nil
}

// Benchmark tests
func BenchmarkLogin(b *testing.B) {
	router := setupTestRouter()
	router.POST("/auth/login", Login)

	requestBody := dto.LoginRequest{
		Username: "testuser",
		Password: "testpassword123",
	}
	jsonBytes, _ := json.Marshal(requestBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBytes))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkRegister(b *testing.B) {
	router := setupTestRouter()
	router.POST("/auth/register", Register)

	requestBody := dto.RegisterRequest{
		Username: "newuser",
		Email:    "newuser@example.com",
		Password: "password123",
		FullName: "New User",
	}
	jsonBytes, _ := json.Marshal(requestBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBytes))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
