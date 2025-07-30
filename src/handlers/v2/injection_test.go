package v2

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupInjectionTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestListInjections(t *testing.T) {
	router := setupInjectionTestRouter()
	router.GET("/injections", ListInjections)

	req, _ := http.NewRequest("GET", "/injections?page=1&size=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Note: This test will fail without proper database setup and authentication
	// It's mainly for checking if the handler compiles and runs without panic
	assert.NotEqual(t, http.StatusInternalServerError, w.Code)
}

func TestGetInjection(t *testing.T) {
	router := setupInjectionTestRouter()
	router.GET("/injections/:id", GetInjection)

	req, _ := http.NewRequest("GET", "/injections/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Note: This test will fail without proper database setup and authentication
	// It's mainly for checking if the handler compiles and runs without panic
	assert.NotEqual(t, http.StatusInternalServerError, w.Code)
}

func TestUpdateInjection(t *testing.T) {
	router := setupInjectionTestRouter()
	router.PUT("/injections/:id", UpdateInjection)

	updateReq := dto.InjectionV2UpdateReq{
		Description: stringPtr("Updated description"),
		Status:      intPtr(1),
	}
	jsonData, _ := json.Marshal(updateReq)

	req, _ := http.NewRequest("PUT", "/injections/1", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Note: This test will fail without proper database setup and authentication
	// It's mainly for checking if the handler compiles and runs without panic
	assert.NotEqual(t, http.StatusInternalServerError, w.Code)
}

func TestDeleteInjection(t *testing.T) {
	router := setupInjectionTestRouter()
	router.DELETE("/injections/:id", DeleteInjection)

	req, _ := http.NewRequest("DELETE", "/injections/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Note: This test will fail without proper database setup and authentication
	// It's mainly for checking if the handler compiles and runs without panic
	assert.NotEqual(t, http.StatusInternalServerError, w.Code)
}

func TestSearchInjections(t *testing.T) {
	router := setupInjectionTestRouter()
	router.POST("/injections/search", SearchInjections)

	searchReq := dto.InjectionV2SearchReq{
		Page:   1,
		Size:   10,
		Search: "test",
	}
	jsonData, _ := json.Marshal(searchReq)

	req, _ := http.NewRequest("POST", "/injections/search", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Note: This test will fail without proper database setup and authentication
	// It's mainly for checking if the handler compiles and runs without panic
	assert.NotEqual(t, http.StatusInternalServerError, w.Code)
}

// Helper functions for creating pointers
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}
