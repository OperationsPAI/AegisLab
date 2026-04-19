package system

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"aegis/utils"

	"github.com/gin-gonic/gin"
)

func init() {
	utils.InitValidator()
}

func TestGetConfigRejectsInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(&Service{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/system/configs/abc", nil)
	c.Request = req
	c.Params = gin.Params{{Key: "config_id", Value: "abc"}}

	h.GetConfig(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestGetAuditLogRejectsInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(&Service{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/system/audit/abc", nil)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "abc"}}

	h.GetAuditLog(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestListConfigsRejectsInvalidQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(&Service{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/system/configs?size=999", nil)

	h.ListConfigs(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestListAuditLogsRejectsInvalidQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(&Service{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/system/audit?start_date=not-a-date", nil)

	h.ListAuditLogs(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestRollbackConfigValueRequiresAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(&Service{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/system/configs/1/rollback/value", bytes.NewBufferString(`{"history_id":1,"reason":"rollback"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "config_id", Value: "1"}}

	h.RollbackConfigValue(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestUpdateConfigMetadataRejectsInvalidPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(&Service{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/system/configs/1/metadata", bytes.NewBufferString(`{"reason":"update"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", 1)
	c.Params = gin.Params{{Key: "config_id", Value: "1"}}

	h.UpdateConfigMetadata(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}
