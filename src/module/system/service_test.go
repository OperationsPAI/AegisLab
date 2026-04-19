package system

import (
	"context"
	"testing"
	"time"

	"aegis/config"
	"aegis/consts"
	"aegis/dto"
	"aegis/model"
)

type fakeConfigHistoryWriter struct {
	history *model.ConfigHistory
	err     error
}

func (f *fakeConfigHistoryWriter) createConfigHistory(history *model.ConfigHistory) error {
	f.history = history
	return f.err
}

func TestGetMetricsReturnsExpectedLabels(t *testing.T) {
	svc := &Service{}

	resp, err := svc.GetMetrics(context.Background())
	if err != nil {
		t.Fatalf("GetMetrics() error = %v", err)
	}
	if resp == nil {
		t.Fatal("expected metrics response")
	}
	if _, ok := resp.Metrics["cpu_usage"]; !ok {
		t.Fatal("expected cpu_usage metric")
	}
	if _, ok := resp.Labels["instance"]; !ok {
		t.Fatal("expected instance label")
	}
}

func TestGetSystemInfoReturnsLoadAverage(t *testing.T) {
	svc := &Service{}

	resp, err := svc.GetSystemInfo(context.Background())
	if err != nil {
		t.Fatalf("GetSystemInfo() error = %v", err)
	}
	if resp == nil {
		t.Fatal("expected system info response")
	}
	if resp.LoadAverage == "" {
		t.Fatal("expected load average")
	}
}

func TestRollbackMetaFieldValueUpdatesDescription(t *testing.T) {
	cfg := &model.DynamicConfig{Description: "current"}

	oldValue, newValue, err := rollbackMetaFieldValue(cfg, consts.ChangeFieldDescription, "restored")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if oldValue != "current" {
		t.Fatalf("expected old value to be current, got %q", oldValue)
	}
	if newValue != "restored" {
		t.Fatalf("expected new value to be restored, got %q", newValue)
	}
	if cfg.Description != "restored" {
		t.Fatalf("expected config description to be restored, got %q", cfg.Description)
	}
}

func TestRollbackMetaFieldValueClearsMinValue(t *testing.T) {
	minValue := 10.5
	cfg := &model.DynamicConfig{MinValue: &minValue}

	oldValue, newValue, err := rollbackMetaFieldValue(cfg, consts.ChangeFieldMinValue, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if oldValue == "" {
		t.Fatal("expected old value to be populated")
	}
	if newValue != "" {
		t.Fatalf("expected new value to be empty, got %q", newValue)
	}
	if cfg.MinValue != nil {
		t.Fatal("expected min value to be cleared")
	}
}

func TestSetViperIfNeededSetsProducerScopeValue(t *testing.T) {
	cfg := &model.DynamicConfig{
		Key:       "system.test.int",
		Scope:     consts.ConfigScopeProducer,
		ValueType: consts.ConfigValueTypeInt,
	}

	if err := setViperIfNeeded(cfg, "42"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got := config.GetInt(cfg.Key); got != 42 {
		t.Fatalf("expected viper value 42, got %d", got)
	}
}

func TestSetViperIfNeededSkipsConsumerScope(t *testing.T) {
	cfg := &model.DynamicConfig{
		Key:       "system.test.consumer",
		Scope:     consts.ConfigScopeConsumer,
		ValueType: consts.ConfigValueTypeString,
	}

	if err := setViperIfNeeded(cfg, "remote-only"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got := config.GetString(cfg.Key); got != "" {
		t.Fatalf("expected consumer scope to skip local viper update, got %q", got)
	}
}

func TestCreateConfigHistoryBuildsExpectedEntry(t *testing.T) {
	writer := &fakeConfigHistoryWriter{}
	svc := &Service{}
	rollbackFromID := 9

	err := svc.createConfigHistory(writer, configHistoryParams{
		ConfigID:       12,
		ChangeType:     consts.ChangeTypeRollback,
		RollbackFromID: &rollbackFromID,
		ConfigUpdateContext: configUpdateContext{
			ChangeField: consts.ChangeFieldPattern,
			OldValue:    "old",
			NewValue:    "new",
			Reason:      "test reason",
			OperatorID:  3,
			IpAddress:   "127.0.0.1",
			UserAgent:   "unit-test",
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if writer.history == nil {
		t.Fatal("expected history to be written")
	}
	if writer.history.ConfigID != 12 {
		t.Fatalf("expected config id 12, got %d", writer.history.ConfigID)
	}
	if writer.history.ChangeType != consts.ChangeTypeRollback {
		t.Fatalf("expected rollback change type, got %v", writer.history.ChangeType)
	}
	if writer.history.ChangeField != consts.ChangeFieldPattern {
		t.Fatalf("expected pattern change field, got %v", writer.history.ChangeField)
	}
	if writer.history.OperatorID == nil || *writer.history.OperatorID != 3 {
		t.Fatalf("expected operator id 3, got %+v", writer.history.OperatorID)
	}
	if writer.history.RolledBackFromID == nil || *writer.history.RolledBackFromID != rollbackFromID {
		t.Fatalf("expected rollback from id %d, got %+v", rollbackFromID, writer.history.RolledBackFromID)
	}
}

func TestBuildConfigDetailRespIncludesHistories(t *testing.T) {
	operatorID := 7
	cfg := &model.DynamicConfig{
		ID:        1,
		Key:       "feature.flag",
		ValueType: consts.ConfigValueTypeString,
		Category:  "system",
	}
	histories := []model.ConfigHistory{
		{
			ID:          11,
			ConfigID:    1,
			ChangeType:  consts.ChangeTypeUpdate,
			ChangeField: consts.ChangeFieldValue,
			OldValue:    "off",
			NewValue:    "on",
			OperatorID:  &operatorID,
		},
	}

	resp := buildConfigDetailResp(cfg, histories)
	if resp == nil {
		t.Fatal("expected response")
	}
	if resp.ID != cfg.ID {
		t.Fatalf("expected config id %d, got %d", cfg.ID, resp.ID)
	}
	if len(resp.Histories) != 1 {
		t.Fatalf("expected 1 history, got %d", len(resp.Histories))
	}
	if resp.Histories[0].ID != 11 {
		t.Fatalf("expected history id 11, got %d", resp.Histories[0].ID)
	}
}

func TestBuildAuditLogListRespIncludesPaginationAndItems(t *testing.T) {
	state := consts.AuditLogStateSuccess
	status := consts.CommonEnabled
	req := &ListAuditLogReq{
		PaginationReq: dto.PaginationReq{Page: 2, Size: consts.PageSizeSmall},
		State:         &state,
		Status:        &status,
	}
	logs := []model.AuditLog{
		{ID: 1, Action: "deploy", IPAddress: "127.0.0.1", State: consts.AuditLogStateSuccess, Status: consts.CommonEnabled, CreatedAt: time.Now()},
	}

	resp := buildAuditLogListResp(logs, req, 21)
	if resp == nil {
		t.Fatal("expected response")
	}
	if len(resp.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(resp.Items))
	}
	if resp.Items[0].ID != 1 {
		t.Fatalf("expected item id 1, got %d", resp.Items[0].ID)
	}
	if resp.Pagination == nil || resp.Pagination.Page != 2 {
		t.Fatalf("expected page 2, got %+v", resp.Pagination)
	}
}

func TestBuildConfigHistoryListRespIncludesOperatorName(t *testing.T) {
	req := &ListConfigHistoryReq{
		PaginationReq: dto.PaginationReq{Page: 1, Size: consts.PageSizeMedium},
	}
	operatorID := 5
	histories := []model.ConfigHistory{
		{
			ID:          2,
			ConfigID:    10,
			ChangeType:  consts.ChangeTypeUpdate,
			ChangeField: consts.ChangeFieldDescription,
			OldValue:    "old desc",
			NewValue:    "new desc",
			OperatorID:  &operatorID,
			Operator:    &model.User{Username: "tester"},
		},
	}

	resp := buildConfigHistoryListResp(histories, req, 1)
	if resp == nil {
		t.Fatal("expected response")
	}
	if len(resp.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(resp.Items))
	}
	if resp.Items[0].OperatorName != "tester" {
		t.Fatalf("expected operator name tester, got %q", resp.Items[0].OperatorName)
	}
}

func TestEtcdPrefixForScopeReturnsExpectedPrefix(t *testing.T) {
	cases := map[consts.ConfigScope]string{
		consts.ConfigScopeProducer: consts.ConfigEtcdProducerPrefix,
		consts.ConfigScopeConsumer: consts.ConfigEtcdConsumerPrefix,
		consts.ConfigScopeGlobal:   consts.ConfigEtcdGlobalPrefix,
	}

	for scope, want := range cases {
		if got := etcdPrefixForScope(scope); got != want {
			t.Fatalf("expected prefix %q for scope %v, got %q", want, scope, got)
		}
	}
}
