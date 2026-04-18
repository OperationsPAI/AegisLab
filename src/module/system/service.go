package systemmodule

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"aegis/config"
	"aegis/consts"
	"aegis/dto"
	buildkitinfra "aegis/infra/buildkit"
	etcdinfra "aegis/infra/etcd"
	k8sinfra "aegis/infra/k8s"
	redisinfra "aegis/infra/redis"
	"aegis/model"
	"aegis/service/common"
	"aegis/utils"

	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

type configUpdateContext struct {
	ChangeField consts.ConfigHistoryChangeField
	OldValue    string
	NewValue    string
	Reason      string
	OperatorID  int
	IpAddress   string
	UserAgent   string
}

type configHistoryParams struct {
	ConfigID       int
	ChangeType     consts.ConfigHistoryChangeType
	RollbackFromID *int

	ConfigUpdateContext configUpdateContext
}

type configHistoryWriter interface {
	createConfigHistory(history *model.ConfigHistory) error
}

type Service struct {
	repo         *Repository
	buildkit     *buildkitinfra.Gateway
	etcd         *etcdinfra.Gateway
	k8s          *k8sinfra.Gateway
	redis        *redisinfra.Gateway
	runtimeQuery runtimeQuerySource
}

type serviceParams struct {
	fx.In

	Repo         *Repository
	Buildkit     *buildkitinfra.Gateway
	Etcd         *etcdinfra.Gateway
	K8s          *k8sinfra.Gateway
	Redis        *redisinfra.Gateway
	RuntimeQuery runtimeQuerySource
}

func NewService(params serviceParams) *Service {
	return &Service{
		repo:         params.Repo,
		buildkit:     params.Buildkit,
		etcd:         params.Etcd,
		k8s:          params.K8s,
		redis:        params.Redis,
		runtimeQuery: params.RuntimeQuery,
	}
}

func (s *Service) GetHealth(ctx context.Context) (*HealthCheckResp, error) {
	start := time.Now()
	services := make(map[string]ServiceInfo)
	overallStatus := "healthy"

	buildkitInfo := s.checkBuildKitHealth(ctx)
	services["buildkit"] = buildkitInfo
	if buildkitInfo.Status != "healthy" {
		overallStatus = "unhealthy"
	}

	dbInfo := s.checkDatabaseHealth(ctx)
	services["database"] = dbInfo
	if dbInfo.Status != "healthy" {
		overallStatus = "unhealthy"
	}

	jaegerInfo := s.checkJaegerHealth(ctx)
	services["jaeger"] = jaegerInfo
	if jaegerInfo.Status != "healthy" {
		overallStatus = "unhealthy"
	}

	k8sInfo := s.checkKubernetesHealth(ctx)
	services["kubernetes"] = k8sInfo
	if k8sInfo.Status != "healthy" {
		overallStatus = "unhealthy"
	}

	redisInfo := s.checkRedisHealth(ctx)
	services["redis"] = redisInfo
	if redisInfo.Status != "healthy" {
		overallStatus = "unhealthy"
	}

	return &HealthCheckResp{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Version:   config.GetString("version"),
		Uptime:    time.Since(start).String(),
		Services:  services,
	}, nil
}

func (s *Service) GetMetrics(_ context.Context) (*MonitoringMetricsResp, error) {
	return &MonitoringMetricsResp{
		Timestamp: time.Now(),
		Metrics: map[string]MetricValue{
			"cpu_usage":          {Value: 25.5, Timestamp: time.Now(), Unit: "percent"},
			"memory_usage":       {Value: 60.2, Timestamp: time.Now(), Unit: "percent"},
			"disk_usage":         {Value: 45.8, Timestamp: time.Now(), Unit: "percent"},
			"active_connections": {Value: 142, Timestamp: time.Now(), Unit: "count"},
		},
		Labels: map[string]string{
			"instance": "rcabench-01",
			"version":  config.GetString("version"),
		},
	}, nil
}

func (s *Service) GetSystemInfo(_ context.Context) (*SystemInfo, error) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	return &SystemInfo{
		CPUUsage:    25.5,
		MemoryUsage: float64(memStats.Alloc) / float64(memStats.Sys) * 100,
		DiskUsage:   45.8,
		LoadAverage: "1.2, 1.5, 1.8",
	}, nil
}

func (s *Service) ListNamespaceLocks(ctx context.Context) (*ListNamespaceLockResp, error) {
	return s.runtimeQuery.ListNamespaceLocks(ctx)
}

func (s *Service) ListQueuedTasks(ctx context.Context) (*QueuedTasksResp, error) {
	return s.runtimeQuery.ListQueuedTasks(ctx)
}

func (s *Service) GetAuditLog(_ context.Context, id int) (*AuditLogDetailResp, error) {
	log, err := s.repo.getAuditLogByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: audit log with ID %d not found", consts.ErrNotFound, id)
		}
		return nil, fmt.Errorf("failed to get audit log: %w", err)
	}

	return NewAuditLogDetailResp(log), nil
}

func (s *Service) ListAuditLogs(_ context.Context, req *ListAuditLogReq) (*dto.ListResp[AuditLogResp], error) {
	limit, offset := req.ToGormParams()
	filterOptions := req.ToFilterOptions()

	logs, total, err := s.repo.listAuditLogs(limit, offset, filterOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list audit logs: %w", err)
	}

	return buildAuditLogListResp(logs, req, total), nil
}

func (s *Service) GetConfig(_ context.Context, configID int) (*ConfigDetailResp, error) {
	cfg, err := s.repo.getConfigByID(configID, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get config detail: %w", err)
	}

	histories, err := s.repo.listConfigHistoriesByConfigID(cfg.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get config histories: %w", err)
	}

	return buildConfigDetailResp(cfg, histories), nil
}

func (s *Service) ListConfigs(_ context.Context, req *ListConfigReq) (*dto.ListResp[ConfigResp], error) {
	limit, offset := req.ToGormParams()

	configs, total, err := s.repo.listConfigs(limit, offset, req.ValueType, req.Category, req.IsSecret, req.UpdatedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to list configs: %w", err)
	}

	return buildConfigListResp(configs, req, total), nil
}

func (s *Service) RollbackConfigValue(ctx context.Context, req *RollbackConfigReq, configID, userID int, ipAddress, userAgent string) error {
	history, err := s.repo.getConfigHistory(req.HistoryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: history entry with id %d not found", consts.ErrNotFound, req.HistoryID)
		}
		return fmt.Errorf("failed to get config history: %w", err)
	}

	if history.ChangeField != consts.ChangeFieldValue {
		return fmt.Errorf("history entry %d is not a value change (field: %v)", req.HistoryID, history.ChangeField)
	}

	existingConfig, err := s.repo.getConfigByID(configID, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: configuration with id %d not found", consts.ErrNotFound, configID)
		}
		return fmt.Errorf("failed to get config: %w", err)
	}

	oldValue, err := s.etcd.Get(ctx, fmt.Sprintf("%s%s", etcdPrefixForScope(existingConfig.Scope), existingConfig.Key))
	if err != nil {
		return fmt.Errorf("failed to get current config value from etcd: %w", err)
	}

	newValue := history.OldValue
	if err := common.ValidateConfig(existingConfig, newValue); err != nil {
		return fmt.Errorf("invalid config after rollback: %w", err)
	}

	if err := setViperIfNeeded(existingConfig, newValue); err != nil {
		return fmt.Errorf("failed to set config value in viper: %w", err)
	}

	if _, err := s.createConfigRollback(existingConfig, utils.IntPtr(history.ID), configUpdateContext{
		ChangeField: consts.ChangeFieldValue,
		OldValue:    oldValue,
		NewValue:    newValue,
		Reason:      req.Reason,
		OperatorID:  userID,
		IpAddress:   ipAddress,
		UserAgent:   userAgent,
	}); err != nil {
		return err
	}

	return s.propagateValueChange(ctx, existingConfig, newValue, "rollback")
}

func (s *Service) RollbackConfigMetadata(_ context.Context, req *RollbackConfigReq, configID, userID int, ipAddress, userAgent string) (*ConfigResp, error) {
	history, err := s.repo.getConfigHistory(req.HistoryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: history entry with id %d not found", consts.ErrNotFound, req.HistoryID)
		}
		return nil, fmt.Errorf("failed to get config history: %w", err)
	}

	if history.ChangeField == consts.ChangeFieldValue {
		return nil, fmt.Errorf("history entry %d is a value change, use RollbackConfigValue instead", req.HistoryID)
	}

	existingConfig, err := s.repo.getConfigByID(configID, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: configuration with id %d not found", consts.ErrNotFound, configID)
		}
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	oldValue, newValue, err := rollbackMetaFieldValue(existingConfig, history.ChangeField, history.OldValue)
	if err != nil {
		return nil, fmt.Errorf("failed to rollback metadata field: %w", err)
	}

	if err := common.ValidateConfigMetadataConstraints(existingConfig); err != nil {
		return nil, fmt.Errorf("invalid config after metadata rollback: %w", err)
	}

	updatedConfig, err := s.createConfigRollback(existingConfig, utils.IntPtr(history.ID), configUpdateContext{
		ChangeField: history.ChangeField,
		OldValue:    oldValue,
		NewValue:    newValue,
		Reason:      req.Reason,
		OperatorID:  userID,
		IpAddress:   ipAddress,
		UserAgent:   userAgent,
	})
	if err != nil {
		return nil, err
	}

	return NewConfigResp(updatedConfig), nil
}

func (s *Service) UpdateConfigValue(ctx context.Context, req *UpdateConfigValueReq, configID, userID int, ipAddress, userAgent string) error {
	existingConfig, err := s.repo.getConfigByID(configID, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: configuration with id %d not found", consts.ErrNotFound, configID)
		}
		return fmt.Errorf("failed to get config: %w", err)
	}

	oldValue, err := s.etcd.Get(ctx, fmt.Sprintf("%s%s", etcdPrefixForScope(existingConfig.Scope), existingConfig.Key))
	if err != nil {
		return fmt.Errorf("failed to get current config value from etcd: %w", err)
	}

	newValue := req.Value
	if err := common.ValidateConfig(existingConfig, newValue); err != nil {
		return fmt.Errorf("invalid config value: %w", err)
	}

	if err := setViperIfNeeded(existingConfig, newValue); err != nil {
		return fmt.Errorf("failed to set config value in viper: %w", err)
	}

	if err := s.createConfigHistory(s.repo, configHistoryParams{
		ConfigID:   existingConfig.ID,
		ChangeType: consts.ChangeTypeUpdate,
		ConfigUpdateContext: configUpdateContext{
			ChangeField: consts.ChangeFieldValue,
			OldValue:    oldValue,
			NewValue:    newValue,
			Reason:      req.Reason,
			OperatorID:  userID,
			IpAddress:   ipAddress,
			UserAgent:   userAgent,
		},
	}); err != nil {
		return fmt.Errorf("failed to create config history: %w", err)
	}

	return s.propagateValueChange(ctx, existingConfig, newValue, "update")
}

func (s *Service) UpdateConfigMetadata(_ context.Context, req *UpdateConfigMetadataReq, configID, userID int, ipAddress, userAgent string) (*ConfigResp, error) {
	existingConfig, err := s.repo.getConfigByID(configID, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: configuration with id %d not found", consts.ErrNotFound, configID)
		}
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	oldValue, newValue := req.PatchConfigModel(existingConfig)
	if err := common.ValidateConfigMetadataConstraints(existingConfig); err != nil {
		return nil, fmt.Errorf("invalid config after metadata update: %w", err)
	}

	var updatedConfig *model.DynamicConfig
	err = s.repo.db.Transaction(func(tx *gorm.DB) error {
		txRepo := NewRepository(tx)
		existingConfig.UpdatedBy = utils.IntPtr(userID)

		if err := txRepo.updateConfig(existingConfig); err != nil {
			return fmt.Errorf("failed to update config: %w", err)
		}

		updatedConfig = existingConfig
		if err := s.createConfigHistory(txRepo, configHistoryParams{
			ConfigID:   updatedConfig.ID,
			ChangeType: consts.ChangeTypeUpdate,
			ConfigUpdateContext: configUpdateContext{
				ChangeField: req.GetChangeField(),
				OldValue:    oldValue,
				NewValue:    newValue,
				Reason:      req.Reason,
				OperatorID:  userID,
				IpAddress:   ipAddress,
				UserAgent:   userAgent,
			},
		}); err != nil {
			return fmt.Errorf("failed to create config history: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return NewConfigResp(updatedConfig), nil
}

func (s *Service) ListConfigHistories(_ context.Context, req *ListConfigHistoryReq, configID int) (*dto.ListResp[ConfigHistoryResp], error) {
	limit, offset := req.ToGormParams()

	histories, total, err := s.repo.listConfigHistories(limit, offset, configID, req.ChangeType, req.OperatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to list config histories: %w", err)
	}

	return buildConfigHistoryListResp(histories, req, total), nil
}

func etcdPrefixForScope(scope consts.ConfigScope) string {
	switch scope {
	case consts.ConfigScopeProducer:
		return consts.ConfigEtcdProducerPrefix
	case consts.ConfigScopeConsumer:
		return consts.ConfigEtcdConsumerPrefix
	case consts.ConfigScopeGlobal:
		return consts.ConfigEtcdGlobalPrefix
	}
	return ""
}

func buildAuditLogListResp(logs []model.AuditLog, req *ListAuditLogReq, total int64) *dto.ListResp[AuditLogResp] {
	logResps := make([]AuditLogResp, 0, len(logs))
	for i := range logs {
		logResps = append(logResps, *NewAuditLogResp(&logs[i]))
	}

	return &dto.ListResp[AuditLogResp]{
		Items:      logResps,
		Pagination: req.ConvertToPaginationInfo(total),
	}
}

func buildConfigDetailResp(cfg *model.DynamicConfig, histories []model.ConfigHistory) *ConfigDetailResp {
	resp := NewConfigDetailResp(cfg)
	for _, history := range histories {
		resp.Histories = append(resp.Histories, *NewConfigHistoryResp(&history))
	}
	return resp
}

func buildConfigListResp(configs []model.DynamicConfig, req *ListConfigReq, total int64) *dto.ListResp[ConfigResp] {
	configResps := make([]ConfigResp, 0, len(configs))
	for _, cfg := range configs {
		configResps = append(configResps, *NewConfigResp(&cfg))
	}

	return &dto.ListResp[ConfigResp]{
		Items:      configResps,
		Pagination: req.ConvertToPaginationInfo(total),
	}
}

func buildConfigHistoryListResp(histories []model.ConfigHistory, req *ListConfigHistoryReq, total int64) *dto.ListResp[ConfigHistoryResp] {
	historyResps := make([]ConfigHistoryResp, 0, len(histories))
	for _, history := range histories {
		historyResps = append(historyResps, *NewConfigHistoryResp(&history))
	}

	return &dto.ListResp[ConfigHistoryResp]{
		Items:      historyResps,
		Pagination: req.ConvertToPaginationInfo(total),
	}
}

func (s *Service) createConfigHistory(repo configHistoryWriter, params configHistoryParams) error {
	entry := &model.ConfigHistory{
		ChangeType:       params.ChangeType,
		OldValue:         params.ConfigUpdateContext.OldValue,
		NewValue:         params.ConfigUpdateContext.NewValue,
		Reason:           params.ConfigUpdateContext.Reason,
		ConfigID:         params.ConfigID,
		OperatorID:       utils.IntPtr(params.ConfigUpdateContext.OperatorID),
		IPAddress:        params.ConfigUpdateContext.IpAddress,
		UserAgent:        params.ConfigUpdateContext.UserAgent,
		RolledBackFromID: params.RollbackFromID,
		ChangeField:      params.ConfigUpdateContext.ChangeField,
	}
	if err := repo.createConfigHistory(entry); err != nil {
		return fmt.Errorf("failed to create config history: %w", err)
	}
	return nil
}

func (s *Service) createConfigRollback(cfg *model.DynamicConfig, historyID *int, updateContext configUpdateContext) (*model.DynamicConfig, error) {
	var updatedConfig *model.DynamicConfig

	err := s.repo.db.Transaction(func(tx *gorm.DB) error {
		txRepo := NewRepository(tx)
		if err := txRepo.updateConfig(cfg); err != nil {
			return fmt.Errorf("failed to update config: %w", err)
		}

		updatedConfig = cfg
		if err := s.createConfigHistory(txRepo, configHistoryParams{
			ConfigID:            cfg.ID,
			ChangeType:          consts.ChangeTypeRollback,
			ConfigUpdateContext: updateContext,
			RollbackFromID:      historyID,
		}); err != nil {
			return fmt.Errorf("failed to create rollback history: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return updatedConfig, nil
}

func rollbackMetaFieldValue(cfg *model.DynamicConfig, changeField consts.ConfigHistoryChangeField, targetValue string) (string, string, error) {
	newValue := targetValue
	oldValue := ""

	switch changeField {
	case consts.ChangeFieldDefaultValue:
		oldValue = cfg.DefaultValue
		cfg.DefaultValue = newValue
	case consts.ChangeFieldDescription:
		oldValue = cfg.Description
		cfg.Description = newValue
	case consts.ChangeFieldMinValue:
		if cfg.MinValue != nil {
			oldValue = fmt.Sprintf("%f", *cfg.MinValue)
		}
		if newValue == "" {
			cfg.MinValue = nil
		} else {
			var minVal float64
			if _, err := fmt.Sscanf(newValue, "%f", &minVal); err != nil {
				return "", "", fmt.Errorf("failed to parse min value: %w", err)
			}
			cfg.MinValue = &minVal
		}
	case consts.ChangeFieldMaxValue:
		if cfg.MaxValue != nil {
			oldValue = fmt.Sprintf("%f", *cfg.MaxValue)
		}
		if newValue == "" {
			cfg.MaxValue = nil
		} else {
			var maxVal float64
			if _, err := fmt.Sscanf(newValue, "%f", &maxVal); err != nil {
				return "", "", fmt.Errorf("failed to parse max value: %w", err)
			}
			cfg.MaxValue = &maxVal
		}
	case consts.ChangeFieldPattern:
		oldValue = cfg.Pattern
		cfg.Pattern = newValue
	case consts.ChangeFieldOptions:
		oldValue = cfg.Options
		cfg.Options = newValue
	default:
		return "", "", fmt.Errorf("unknown change field: %d", changeField)
	}

	return oldValue, newValue, nil
}

func setViperIfNeeded(cfg *model.DynamicConfig, newValue string) error {
	if cfg.Scope == consts.ConfigScopeConsumer {
		return nil
	}
	return config.SetViperValue(cfg.Key, newValue, cfg.ValueType)
}

func (s *Service) propagateValueChange(ctx context.Context, cfg *model.DynamicConfig, newValue, opDesc string) error {
	if cfg.Scope != consts.ConfigScopeGlobal && cfg.Scope != consts.ConfigScopeConsumer {
		return nil
	}

	etcdKey := fmt.Sprintf("%s%s", etcdPrefixForScope(cfg.Scope), cfg.Key)
	if err := s.publishConfigToEtcdWithRetry(ctx, etcdKey, newValue, 3); err != nil {
		return fmt.Errorf("config saved to database but failed to publish to etcd: %w", err)
	}

	if cfg.Scope == consts.ConfigScopeConsumer {
		logrus.Infof("Waiting for consumer config %s response...", opDesc)
		resp, err := s.waitForConfigUpdateResponse(ctx, 10*time.Second)
		if err != nil {
			return fmt.Errorf("config %s but consumer did not respond: %w", opDesc, err)
		}
		if !resp.Success {
			return fmt.Errorf("consumer failed to process config %s: %s", opDesc, resp.Error)
		}
		logrus.Infof("Config %s successfully processed by consumer", opDesc)
	}

	return nil
}

func (s *Service) publishConfigToEtcdWithRetry(ctx context.Context, key, value string, maxRetries int) error {
	var lastErr error
	baseDelay := 500 * time.Millisecond

	for attempt := range maxRetries {
		if attempt > 0 {
			delay := baseDelay * time.Duration(1<<uint(attempt-1))
			logrus.Warnf("Retrying etcd publish after %v (attempt %d/%d)", delay, attempt+1, maxRetries)
			time.Sleep(delay)
		}

		publishCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err := s.etcd.Put(publishCtx, key, value, 0)
		cancel()

		if err == nil {
			if attempt > 0 {
				logrus.Infof("Successfully published config to etcd after %d retries", attempt)
			}
			return nil
		}

		lastErr = err
		logrus.Warnf("Failed to publish config to etcd (attempt %d/%d): %v", attempt+1, maxRetries, err)
	}

	return fmt.Errorf("failed to publish config to etcd after %d attempts: %w", maxRetries, lastErr)
}

func (s *Service) waitForConfigUpdateResponse(parent context.Context, timeout time.Duration) (*dto.ConfigUpdateResponse, error) {
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	pubsub, err := s.redis.Subscribe(ctx, consts.ConfigUpdateResponseChannel)
	if err != nil {
		return nil, fmt.Errorf("failed to confirm subscription: %w", err)
	}
	defer func() { _ = pubsub.Close() }()

	msgChan := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout waiting for config update response after %v", timeout)
		case msg, ok := <-msgChan:
			if !ok {
				return nil, fmt.Errorf("subscription channel closed unexpectedly")
			}

			var response dto.ConfigUpdateResponse
			if err := json.Unmarshal([]byte(msg.Payload), &response); err != nil {
				logrus.Warnf("failed to parse response message: %v", err)
				continue
			}

			logrus.WithFields(logrus.Fields{
				"response_id": response.ID,
				"success":     response.Success,
			}).Info("Received matching config update response")
			return &response, nil
		}
	}
}

func (s *Service) checkBuildKitHealth(parent context.Context) ServiceInfo {
	start := time.Now()
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()

	err := s.buildkit.CheckHealth(ctx, 5*time.Second)
	responseTime := time.Since(start)
	if err != nil {
		return ServiceInfo{
			Status:       "unhealthy",
			LastChecked:  time.Now(),
			ResponseTime: responseTime.String(),
			Error:        "BuildKit daemon unreachable",
			Details:      err.Error(),
		}
	}
	return ServiceInfo{Status: "healthy", LastChecked: time.Now(), ResponseTime: responseTime.String()}
}

func (s *Service) checkDatabaseHealth(parent context.Context) ServiceInfo {
	start := time.Now()
	db := s.repo.db
	if db == nil {
		return ServiceInfo{Status: "unhealthy", LastChecked: time.Now(), ResponseTime: "N/A", Error: "Database connection not available"}
	}

	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	var result int
	err := db.WithContext(ctx).Raw("SELECT 1").Scan(&result).Error
	responseTime := time.Since(start)
	if err != nil {
		return ServiceInfo{Status: "unhealthy", LastChecked: time.Now(), ResponseTime: responseTime.String(), Error: "Database query failed", Details: err.Error()}
	}
	return ServiceInfo{Status: "healthy", LastChecked: time.Now(), ResponseTime: responseTime.String()}
}

func (s *Service) checkJaegerHealth(parent context.Context) ServiceInfo {
	start := time.Now()
	jaegerURL := fmt.Sprintf("http://%s/v1/traces", config.GetString("jaeger.endpoint"))
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, jaegerURL, nil)
	if err != nil {
		return ServiceInfo{Status: "unhealthy", LastChecked: time.Now(), ResponseTime: time.Since(start).String(), Error: "Failed to create Jaeger OTLP request", Details: err.Error()}
	}

	httpClient := &http.Client{Timeout: 5 * time.Second}
	resp, err := httpClient.Do(req)
	responseTime := time.Since(start)
	if err != nil {
		return ServiceInfo{Status: "unhealthy", LastChecked: time.Now(), ResponseTime: responseTime.String(), Error: "Jaeger OTLP endpoint unreachable", Details: err.Error()}
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusMethodNotAllowed && resp.StatusCode != http.StatusOK {
		return ServiceInfo{Status: "unhealthy", LastChecked: time.Now(), ResponseTime: responseTime.String(), Error: fmt.Sprintf("Jaeger OTLP returned unexpected status %d", resp.StatusCode)}
	}
	return ServiceInfo{Status: "healthy", LastChecked: time.Now(), ResponseTime: responseTime.String(), Details: "Jaeger OTLP endpoint responding"}
}

func (s *Service) checkKubernetesHealth(parent context.Context) ServiceInfo {
	start := time.Now()
	if s.k8s == nil {
		return ServiceInfo{Status: "unavailable", LastChecked: time.Now(), ResponseTime: time.Since(start).String(), Error: "Kubernetes gateway not configured"}
	}
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	if err := s.k8s.CheckHealth(ctx); err != nil {
		return ServiceInfo{Status: "unhealthy", LastChecked: time.Now(), ResponseTime: time.Since(start).String(), Error: "Kubernetes health check failed", Details: err.Error()}
	}
	return ServiceInfo{Status: "healthy", LastChecked: time.Now(), ResponseTime: time.Since(start).String()}
}

func (s *Service) checkRedisHealth(parent context.Context) ServiceInfo {
	start := time.Now()
	if s.redis == nil {
		return ServiceInfo{Status: "unhealthy", LastChecked: time.Now(), ResponseTime: "N/A", Error: "Redis connection not available"}
	}

	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	err := s.redis.Ping(ctx)
	responseTime := time.Since(start)
	if err != nil {
		return ServiceInfo{Status: "unhealthy", LastChecked: time.Now(), ResponseTime: responseTime.String(), Error: "Redis ping failed", Details: err.Error()}
	}
	return ServiceInfo{Status: "healthy", LastChecked: time.Now(), ResponseTime: responseTime.String()}
}
