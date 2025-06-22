package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"

	chaos "github.com/LGU-SE-Internal/chaos-experiment/handler"
	"github.com/LGU-SE-Internal/rcabench/client/k8s"
	"github.com/LGU-SE-Internal/rcabench/config"
	conf "github.com/LGU-SE-Internal/rcabench/config"
	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/executor"
	"github.com/LGU-SE-Internal/rcabench/middleware"
	"github.com/LGU-SE-Internal/rcabench/repository"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// CancelInjection
//
//	@Summary		取消故障注入任务
//	@Description	取消指定的故障注入任务
//	@Tags			injection
//	@Produce		application/json
//	@Param			task_id	path		string	true	"任务ID"
//	@Success		200		{object}	dto.GenericResponse[dto.InjectCancelResp]
//	@Failure		400		{object}	dto.GenericResponse[any]
//	@Failure		500		{object}	dto.GenericResponse[any]
//	@Router			/api/v1/injections/{task_id}/cancel [put]
func CancelInjection(c *gin.Context) {
}

// GetInjectionConf
//
//	@Summary		获取故障注入配置
//	@Description	获取指定命名空间的故障注入配置信息
//	@Tags			injection
//	@Produce		json
//	@Param			namespace	query		string	true	"命名空间"
//	@Param			mode		query		string	true	"显示模式(display/engine)"
//	@Success		200			{object}	dto.GenericResponse[chaos.Node]
//	@Failure		400			{object}	dto.GenericResponse[any]
//	@Failure		500			{object}	dto.GenericResponse[any]
//
//	@Router			/api/v1/injections/conf [get]
func GetInjectionConf(c *gin.Context) {
	var req dto.InjectionConfReq
	if err := c.BindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid Parameters")
		return
	}

	root, err := chaos.StructToNode[chaos.InjectionConf](req.Namespace)
	if err != nil {
		logrus.Errorf("struct InjectionConf to node failed: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "failed to read injection conf")
		return
	}

	dto.SuccessResponse(c, root)
}

// GetConfigList
//
//	@Summary		获取故障注入配置列表
//	@Description	根据多个 TraceID 获取对应的故障注入配置信息
//	@Tags			injection
//	@Produce		json
//	@Param			trace_ids	query		[]string	true	"Trace ID 列表"
//	@Success		200			{object}	dto.GenericResponse[any]
//	@Failure		400			{object}	dto.GenericResponse[any]
//	@Failure		500			{object}	dto.GenericResponse[any]
//	@Router			/api/v1/injections/configs [get]
func GetDisplayConfigList(c *gin.Context) {
	var req dto.InjectionConfigListReq
	if err := c.BindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid Parameters")
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	configs, err := repository.GetDisplayConfigByTraceIDs(req.TraceIDs)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "failed to get injection config list")
		return
	}

	dto.SuccessResponse(c, configs)
}

// GetInjectionList
//
//	@Summary		分页查询注入记录列表
//	@Description	获取注入记录列表（支持分页参数）
//	@Tags			injection
//	@Produce		json
//	@Consumes		application/json
//	@Param			page_num	query		int	false	"页码"	default(1)
//	@Param			page_size	query		int	false	"每页大小"	default(10)
//	@Success		200			{object}	dto.GenericResponse[dto.PaginationResp[dto.InjectionItem]]
//	@Failure		400			{object}	dto.GenericResponse[any]
//	@Failure		500			{object}	dto.GenericResponse[any]
//	@Router			/api/v1/injections/detail [get]
func GetInjectionList(c *gin.Context) {
	var req dto.InjectionListReq
	if err := c.BindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, formatErrorMessage(err, map[string]string{}))
		return
	}

	total, records, err := repository.ListInjectionWithPagination(req.PageNum, req.PageSize)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	items := make([]dto.InjectionItem, 0, len(records))
	for _, record := range records {
		var item dto.InjectionItem
		if err := item.Convert(record); err != nil {
			logrus.WithField("injection", record.ID).Error(err)
			dto.ErrorResponse(c, http.StatusInternalServerError, "invalid injection configuration")
			return
		}

		items = append(items, item)
	}

	totalPages := (total + int64(req.PageSize) - 1) / int64(req.PageSize)
	dto.SuccessResponse(c, dto.PaginationResp[dto.InjectionItem]{
		Total:      total,
		TotalPages: totalPages,
		Items:      items,
	})
}

// GetNSLock
//
//	@Summary		获取命名空间锁状态
//	@Description	获取命名空间锁状态信息
//	@Tags			injection
//	@Produce		json
//	@Success		200	{object}	dto.GenericResponse[any]
//	@Failure		500	{object}	dto.GenericResponse[any]
//	@Router			/api/v1/injections/ns/status [get]
func GetNSLock(c *gin.Context) {
	cli := k8s.GetMonitor()
	items, err := cli.InspectLock()
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "failed to inspect lock")
		return
	}

	dto.SuccessResponse(c, items)
}

// QueryInjection
//
//	@Summary		查询故障注入记录
//	@Description	根据名称或任务ID查询故障注入记录详情
//	@Tags			injection
//	@Produce		json
//	@Param			name		query		string	false	"注入名称"
//	@Param			task_id		query		string	false	"任务ID"
//	@Success		200			{object}	dto.GenericResponse[dto.InjectionItem]
//	@Failure		400			{object}	dto.GenericResponse[any]
//	@Failure		500			{object}	dto.GenericResponse[any]
//	@Router			/api/v1/injections/query [get]
func QueryInjection(c *gin.Context) {
	var req dto.QueryInjectionReq
	if err := c.BindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, formatErrorMessage(err, map[string]string{}))
		return
	}

	if req.Name == "" && req.TaskID == "" {
		dto.ErrorResponse(c, http.StatusBadRequest, "At least one of the name or task_id parameters must be provided")
		return
	}

	queryColumn := "injection_name"
	queryParam := req.Name
	if queryParam == "" {
		queryColumn = "task_id"
		queryParam = req.TaskID
	}

	item, err := repository.GetInjection(queryColumn, queryParam)
	if err != nil {
		logrus.Errorf("failed to get injection record: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get injection record")
		return
	}

	dto.SuccessResponse(c, item)
}

// SubmitFaultInjection
//
//	@Summary		注入故障
//	@Description	注入故障
//	@Tags			injection
//	@Produce		json
//	@Consumes		application/json
//	@Param			body	body		dto.InjectionSubmitReq	true	"请求体"
//	@Success		202		{object}	dto.GenericResponse[dto.SubmitResp]
//	@Failure		400		{object}	dto.GenericResponse[any]
//	@Failure		500		{object}	dto.GenericResponse[any]
//	@Router			/api/v1/injections [post]
func SubmitFaultInjection(c *gin.Context) {
	groupID := c.GetString("groupID")
	logrus.Infof("SubmitFaultInjection called, groupID: %s", groupID)

	// Get the span context from gin.Context
	ctx, ok := c.Get(middleware.SpanContextKey)
	if !ok {
		logrus.Error("failed to get span context from gin.Context")
		dto.ErrorResponse(c, http.StatusInternalServerError, "failed to get span context")
		return
	}

	spanCtx := ctx.(context.Context)
	span := trace.SpanFromContext(spanCtx)

	handleError := func(err error, message string, statusCode int) {
		logrus.Error(err)
		span.SetStatus(codes.Error, message)
		dto.ErrorResponse(c, statusCode, message)
	}

	defer func() {
		if err := recover(); err != nil {
			logrus.Errorf("SubmitFaultInjection panic: %v", err)
			logrus.Errorf("Stack trace: %s", debug.Stack())
			span.SetStatus(codes.Error, "panic in SubmitFaultInjection")
			dto.ErrorResponse(c, http.StatusInternalServerError, "Internal Server Error")
		}
	}()

	var req dto.InjectionSubmitReq
	if err := c.BindJSON(&req); err != nil {
		handleError(err, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if err := validateAlgorithms(req.Algorithms); err != nil {
		handleError(err, err.Error(), http.StatusBadRequest)
		return
	}

	configs, err := parseInjectionSpecs(&req)
	if err != nil {
		handleError(err, err.Error(), http.StatusBadRequest)
		return
	}

	traces := make([]dto.Trace, 0, len(configs))
	for _, config := range configs {
		task := createInjectionTask(&req, config, groupID, spanCtx)

		taskID, traceID, err := executor.SubmitTask(spanCtx, task)
		if err != nil {
			handleError(err, "failed to submit injection task", http.StatusInternalServerError)
			return
		}

		traces = append(traces, dto.Trace{TraceID: traceID, HeadTaskID: taskID, Index: config.Index})
	}

	span.SetStatus(codes.Ok, fmt.Sprintf("Successfully submitted %d fault injections with groupID: %s", len(traces), groupID))

	duplicatedCount := len(req.Specs) - len(configs)
	resp := dto.InjectionSubmitResp{
		DuplicatedCount: duplicatedCount,
		OriginalCount:   len(req.Specs),
		SubmitResp: dto.SubmitResp{
			GroupID: groupID,
			Traces:  traces,
		},
	}
	if !conf.GetBool("injection.enable_duplicate") {
		logrus.Infof("Duplicated %d configurations, original count: %d", len(req.Specs)-len(configs), len(req.Specs))
	}

	dto.JSONResponse(c, http.StatusAccepted, "Fault injections submitted successfully", resp)
}

// validateAlgorithms validates the provided algorithms against the valid algorithm list
func validateAlgorithms(algorithms []string) error {
	validAlgorithms, err := repository.ListAlgorithms(true)
	if err != nil {
		return fmt.Errorf("failed to list algorithms: %v", err)
	}

	validAlgorithmMap := make(map[string]struct{}, len(validAlgorithms))
	for _, algorithm := range validAlgorithms {
		validAlgorithmMap[algorithm.Name] = struct{}{}
	}

	for _, algorithm := range algorithms {
		if algorithm == "" {
			return fmt.Errorf("algorithm must not be empty")
		}

		detector := config.GetString("algo.detector")
		if algorithm == detector {
			return fmt.Errorf("algorithm %s is not allowed for fault injection", detector)
		}

		if _, exists := validAlgorithmMap[algorithm]; !exists {
			return fmt.Errorf("invalid algorithm: %s", algorithm)
		}
	}

	return nil
}

// createInjectionTask creates a unified task for fault injection
func createInjectionTask(req *dto.InjectionSubmitReq, config *dto.InjectionConfig, groupID string, spanCtx context.Context) *dto.UnifiedTask {
	var payload map[string]any
	taskType := consts.TaskTypeRestartService
	if req.DirectInject {
		payload = map[string]any{
			consts.InjectAlgorithms:  req.Algorithms,
			consts.InjectBenchmark:   req.Benchmark,
			consts.InjectFaultType:   config.FaultType,
			consts.InjectPreDuration: req.PreDuration,
			consts.InjectDisplayData: config.DisplayData,
			consts.InjectConf:        config.Conf,
			consts.InjectNode:        config.Node,
			consts.InjectLabels:      config.Labels,
			consts.InjectNamespace:   "ts4",
		}
		taskType = consts.TaskTypeFaultInjection
	} else {
		payload = map[string]any{
			consts.RestartIntarval:      req.Interval,
			consts.RestartFaultDuration: config.FaultDuration,
			consts.RestartInjectPayload: map[string]any{
				consts.InjectAlgorithms:  req.Algorithms,
				consts.InjectBenchmark:   req.Benchmark,
				consts.InjectFaultType:   config.FaultType,
				consts.InjectPreDuration: req.PreDuration,
				consts.InjectDisplayData: config.DisplayData,
				consts.InjectConf:        config.Conf,
				consts.InjectNode:        config.Node,
				consts.InjectLabels:      config.Labels,
			},
		}
	}

	task := &dto.UnifiedTask{
		Type:        taskType,
		Payload:     payload,
		Immediate:   false,
		ExecuteTime: config.ExecuteTime.Unix(),
		GroupID:     groupID,
	}
	task.SetGroupCtx(spanCtx)

	return task
}

func parseInjectionSpecs(r *dto.InjectionSubmitReq) ([]*dto.InjectionConfig, error) {
	if len(r.Specs) == 0 {
		return nil, fmt.Errorf("spec must not be blank")
	}

	configs := make([]*dto.InjectionConfig, 0, len(r.Specs))
	displayDatas := make([]string, 0, len(r.Specs))
	if r.Labels == nil {
		r.Labels = make([]dto.LabelItem, 0)
	}

	for idx, spec := range r.Specs {
		childNode, exists := spec.Children[strconv.Itoa(spec.Value)]
		if !exists {
			return nil, fmt.Errorf("failed to find key %d in the children", spec.Value)
		}

		faultDuration := childNode.Children[consts.DurationNodeKey].Value

		conf, err := chaos.NodeToStruct[chaos.InjectionConf](&spec)
		if err != nil {
			return nil, fmt.Errorf("failed to convert node to injecton conf: %v", err)
		}

		displayConfig, err := conf.GetDisplayConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to get display config: %v", err)
		}

		displayData, err := json.Marshal(displayConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal injection spec to display config: %v", err)
		}

		configs = append(configs, &dto.InjectionConfig{
			Index:         idx,
			FaultType:     spec.Value,
			FaultDuration: faultDuration,
			DisplayData:   string(displayData),
			Conf:          conf,
			Node:          &spec,
			Labels:        r.Labels,
		})
		displayDatas = append(displayDatas, string(displayData))
	}

	missingIndices, err := findMissingIndices(displayDatas, 10)
	if err != nil {
		return nil, err
	}

	newConfigs := make([]*dto.InjectionConfig, 0)
	for _, idx := range missingIndices {
		conf := configs[idx]
		conf.ExecuteTime = time.Now().Add(time.Second * time.Duration(rand.Int()%120))
		newConfigs = append(newConfigs, conf)
	}

	return newConfigs, nil
}

func findMissingIndices(confs []string, batch_size int) ([]int, error) {
	var missingIndices []int
	existingMap := make(map[string]struct{})

	for i := 0; i < len(confs); i += batch_size {
		end := min(i+batch_size, len(confs))
		batch := confs[i:end]

		existingBatch, err := repository.FindExistingDisplayConfigs(batch)
		if err != nil {
			return nil, err
		}

		for _, s := range existingBatch {
			existingMap[s] = struct{}{}
		}
	}

	for idx, s := range confs {
		if _, exists := existingMap[s]; !exists {
			missingIndices = append(missingIndices, idx)
		}
	}

	return missingIndices, nil
}

// GetFaultInjectionNoIssues
//
//	@Summary		查询没有问题的故障注入记录
//	@Description	根据时间范围查询所有没有问题的故障注入记录列表
//	@Tags			injection
//	@Produce		json
//	@Param			lookback			query		string	false	"相对时间查询，如 1h, 24h, 7d或者是custom"
//	@Param			custom_start_time	query		string	false	"当lookback=custom时必需，自定义开始时间 (RFC3339格式)"
//	@Param			custom_end_time		query		string	false	"当lookback=custom时必需，自定义结束时间 (RFC3339格式)"
//	@Success		200					{object}	dto.GenericResponse[[]dto.FaultInjectionNoIssuesResp]
//	@Failure		400					{object}	dto.GenericResponse[any]	"参数错误或时间格式错误"
//	@Failure		500					{object}	dto.GenericResponse[any]	"服务器内部错误"
//	@Router			/api/v1/injections/analysis/no-issues [get]
func GetFaultInjectionNoIssues(c *gin.Context) {
	var req dto.FaultInjectionNoIssuesReq
	if err := c.BindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid Parameters")
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Invalid Parameters: %v", err))
		return
	}

	opts, err := req.Convert()
	if err != nil {
		logrus.Errorf("failed to convert request: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to convert request")
		return
	}

	_, records, err := repository.GetAllFaultInjectionNoIssues(*opts)
	if err != nil {
		logrus.Errorf("failed to get fault injection no issues: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "failed to get fault injection records")
		return
	}

	var items []dto.FaultInjectionNoIssuesResp
	for _, record := range records {
		conf := chaos.Node{}
		err := json.Unmarshal([]byte(record.EngineConfig), &conf)
		if err != nil {
			logrus.Errorf("failed to unmarshal engine config: %v", err)
			dto.ErrorResponse(c, http.StatusInternalServerError, "failed to parse engine config")
			return
		}

		items = append(items, dto.FaultInjectionNoIssuesResp{
			DatasetID:     record.DatasetID,
			DisplayConfig: record.DisplayConfig,
			EngineConfig:  conf,
			PreDuration:   record.PreDuration,
			InjectionName: record.InjectionName,
		})
	}

	dto.SuccessResponse(c, items)
}

// GetFaultInjectionWithIssues
//
//	@Summary		查询有问题的故障注入记录
//	@Description	根据时间范围查询所有有问题的故障注入记录列表
//	@Tags			injection
//	@Produce		json
//	@Param			lookback			query		string	false	"相对时间查询，如 1h, 24h, 7d或者是custom"
//	@Param			custom_start_time	query		string	false	"当lookback=custom时必需，自定义开始时间 (RFC3339格式)"
//	@Param			custom_end_time		query		string	false	"当lookback=custom时必需，自定义结束时间 (RFC3339格式)"
//	@Success		200					{object}	dto.GenericResponse[[]dto.FaultInjectionWithIssuesResp]
//	@Failure		400					{object}	dto.GenericResponse[any]	"参数错误或时间格式错误"
//	@Failure		500					{object}	dto.GenericResponse[any]	"服务器内部错误"
//	@Router			/api/v1/injections/analysis/with-issues [get]
func GetFaultInjectionWithIssues(c *gin.Context) {
	var req dto.FaultInjectionNoIssuesReq // 注意：这里应该重命名为更通用的名称
	if err := c.BindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid Parameters")
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Invalid Parameters: %v", err))
		return
	}

	opts, err := req.Convert()
	if err != nil {
		logrus.Errorf("failed to convert request: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to convert request")
		return
	}

	_, records, err := repository.GetAllFaultInjectionWithIssues(*opts)
	if err != nil {
		logrus.Errorf("failed to get fault injection with issues: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "failed to get fault injection records")
		return
	}

	var items []dto.FaultInjectionWithIssuesResp
	for _, record := range records {
		conf := chaos.Node{}
		err := json.Unmarshal([]byte(record.EngineConfig), &conf)
		if err != nil {
			logrus.Errorf("failed to unmarshal engine config: %v", err)
			dto.ErrorResponse(c, http.StatusInternalServerError, "failed to parse engine config")
			return
		}
		items = append(items, dto.FaultInjectionWithIssuesResp{
			DatasetID:     record.DatasetID,
			DisplayConfig: record.DisplayConfig,
			EngineConfig:  conf,
			PreDuration:   record.PreDuration,
			InjectionName: record.InjectionName,
			Issues:        record.Issues,
		})
	}

	dto.SuccessResponse(c, items)
}

// GetFaultInjectionStatistics
//
//	@Summary		获取故障注入统计信息
//	@Description	获取故障注入记录的统计信息，包括有问题和没有问题的记录数量
//	@Tags			injection
//	@Produce		json
//	@Success		200	{object}	dto.GenericResponse[dto.FaultInjectionStatisticsResp]
//	@Failure		500	{object}	dto.GenericResponse[any]
//	@Router			/api/v1/injections/analysis/statistics [get]
func GetFaultInjectionStatistics(c *gin.Context) {
	stats, err := repository.GetFaultInjectionStatistics()
	if err != nil {
		logrus.Errorf("failed to get fault injection statistics: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "failed to get statistics")
		return
	}

	dto.SuccessResponse(c, dto.FaultInjectionStatisticsResp{
		NoIssuesCount:   stats["no_issues"],
		WithIssuesCount: stats["with_issues"],
		TotalCount:      stats["total"],
	})
}

// GetFaultInjectionByDatasetID
//
//	@Summary		根据数据集ID查询故障注入记录
//	@Description	根据数据集ID查询故障注入记录
//	@Tags			injection
//	@Produce		json
//	@Param			dataset_name	query		string	true	"数据集名称"
//	@Success		200			{object}	dto.GenericResponse[dto.FaultInjectionInjectionResp]
//	@Failure		400			{object}	dto.GenericResponse[any]
//	@Failure		404			{object}	dto.GenericResponse[any]
//	@Failure		500			{object}	dto.GenericResponse[any]
//	@Router			/api/v1/injections/detail [get]
func GetFaultInjectionByDatasetName(c *gin.Context) {
	datasetName := c.Query("dataset_name")
	if datasetName == "" {
		dto.ErrorResponse(c, http.StatusBadRequest, "Dataset name is required")
		return
	}
	dataset, err := repository.GetFLByDatasetName(datasetName)
	if err != nil {
		logrus.Errorf("failed to get fault injection by dataset name: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "failed to get fault injection by dataset name")
		return
	}
	groundTruth, err := repository.GetGroundtruthMap([]string{datasetName})
	if err != nil {
		logrus.Errorf("failed to get ground truth map: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "failed to get ground truth map")
		return
	}
	resp := dto.FaultInjectionInjectionResp{
		FaultInjectionSchedule: *dataset,
		GroundTruth:            groundTruth[datasetName],
	}
	dto.SuccessResponse(c, resp)
}
