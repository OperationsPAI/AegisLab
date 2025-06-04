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
//	@Router			/api/v1/injections [get]
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
		logrus.Error(err)
		span.SetStatus(codes.Error, "failed to bind JSON")
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	configs, err := ParseInjectionSpecs(&req)
	if err != nil {
		logrus.Error(err)
		span.SetStatus(codes.Error, "failed to parse injection specs")
		dto.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	traces := make([]dto.Trace, 0, len(configs))
	for _, config := range configs {
		payload := map[string]any{
			consts.RestartIntarval:      req.Interval,
			consts.RestartFaultDuration: config.FaultDuration,
			consts.RestartInjectPayload: map[string]any{
				consts.InjectBenchmark:   req.Benchmark,
				consts.InjectFaultType:   config.FaultType,
				consts.InjectPreDuration: req.PreDuration,
				consts.InjectDisplayData: config.DisplayData,
				consts.InjectConf:        config.Conf,
				consts.InjectNode:        config.Node,
			},
		}

		task := &dto.UnifiedTask{
			Type:        consts.TaskTypeRestartService,
			Payload:     payload,
			Immediate:   false,
			ExecuteTime: config.ExecuteTime.Unix(),
			GroupID:     groupID,
		}
		task.SetGroupCtx(spanCtx)

		// taskID, traceID := "debuging", "debugging"
		taskID, traceID, err := executor.SubmitTask(spanCtx, task)
		if err != nil {
			message := "failed to submit injection task"
			logrus.Errorf("%s: %v", message, err)
			dto.ErrorResponse(c, http.StatusInternalServerError, message)
			return
		}

		traces = append(traces, dto.Trace{TraceID: traceID, HeadTaskID: taskID, Index: config.Index})
	}

	span.SetStatus(codes.Ok, fmt.Sprintf("Successfully submitted %d fault injections with groupID: %s", len(traces), groupID))

	resp := dto.InjectionSubmitResp{SubmitResp: dto.SubmitResp{
		GroupID: groupID,
		Traces:  traces,
	}}
	if !conf.GetBool("injection.enable_duplicate") {
		logrus.Infof("Duplicated %d configurations, original count: %d", len(req.Specs)-len(configs), len(req.Specs))
	}

	dto.JSONResponse(c, http.StatusAccepted, "Fault injections submitted successfully", resp)
}

func ParseInjectionSpecs(r *dto.InjectionSubmitReq) ([]*dto.InjectionConfig, error) {
	if len(r.Specs) == 0 {
		return nil, fmt.Errorf("spec must not be blank")
	}

	configs := make([]*dto.InjectionConfig, 0, len(r.Specs))
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
		})
	}

	displayDatas := make([]string, 0, len(configs))
	for _, config := range configs {
		displayDatas = append(displayDatas, config.DisplayData)
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
	return configs, nil
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
//	@Description	查询所有没有问题的故障注入记录列表
//	@Tags			injection
//	@Produce		json
//	@Success		200			{object}	dto.GenericResponse[[]dto.FaultInjectionNoIssuesResp]
//	@Failure		400			{object}	dto.GenericResponse[any]
//	@Failure		500			{object}	dto.GenericResponse[any]
//	@Router			/api/v1/injections/analysis/no-issues [get]
func GetFaultInjectionNoIssues(c *gin.Context) {
	_, records, err := repository.GetAllFaultInjectionNoIssues()
	if err != nil {
		logrus.Errorf("failed to get fault injection no issues: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "failed to get fault injection records")
		return
	}

	items := make([]dto.FaultInjectionNoIssuesResp, 0, len(records))

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
//	@Description	查询所有有问题的故障注入记录列表
//	@Tags			injection
//	@Produce		json
//	@Success		200			{object}	dto.GenericResponse[[]dto.FaultInjectionWithIssuesResp]
//	@Failure		400			{object}	dto.GenericResponse[any]
//	@Failure		500			{object}	dto.GenericResponse[any]
//	@Router			/api/v1/injections/analysis/with-issues [get]
func GetFaultInjectionWithIssues(c *gin.Context) {
	_, records, err := repository.GetAllFaultInjectionWithIssues()
	if err != nil {
		logrus.Errorf("failed to get fault injection with issues: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "failed to get fault injection records")
		return
	}

	items := make([]dto.FaultInjectionWithIssuesResp, 0, len(records))
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
//	@Description	根据数据集ID查询故障注入记录详情（包括是否有问题）
//	@Tags			injection
//	@Produce		json
//	@Param			dataset_id	path		int	true	"数据集ID"
//	@Success		200			{object}	dto.GenericResponse[any]
//	@Failure		400			{object}	dto.GenericResponse[any]
//	@Failure		404			{object}	dto.GenericResponse[any]
//	@Failure		500			{object}	dto.GenericResponse[any]
//	@Router			/api/v1/injections/analysis/dataset/{dataset_id} [get]
func GetFaultInjectionByDatasetID(c *gin.Context) {
	datasetIDStr := c.Param("dataset_id")
	datasetID, err := strconv.Atoi(datasetIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid dataset ID")
		return
	}

	// 先尝试查询有问题的记录
	withIssues, err := repository.GetFaultInjectionWithIssuesByDatasetID(datasetID)
	if err == nil {
		conf := chaos.Node{}
		err := json.Unmarshal([]byte(withIssues.EngineConfig), &conf)
		if err != nil {
			logrus.Errorf("failed to unmarshal engine config: %v", err)
			dto.ErrorResponse(c, http.StatusInternalServerError, "failed to parse engine config")
			return
		}

		dto.SuccessResponse(c, dto.FaultInjectionWithIssuesResp{
			DatasetID:     withIssues.DatasetID,
			DisplayConfig: withIssues.DisplayConfig,
			EngineConfig:  conf,
			PreDuration:   withIssues.PreDuration,
			InjectionName: withIssues.InjectionName,
			Issues:        withIssues.Issues,
		})
		return
	}

	// 如果没有找到有问题的记录，尝试查询没有问题的记录
	noIssues, err := repository.GetFaultInjectionNoIssuesByDatasetID(datasetID)
	if err == nil {
		conf := chaos.Node{}
		err := json.Unmarshal([]byte(noIssues.EngineConfig), &conf)
		if err != nil {
			logrus.Errorf("failed to unmarshal engine config: %v", err)
			dto.ErrorResponse(c, http.StatusInternalServerError, "failed to parse engine config")
			return
		}

		dto.SuccessResponse(c, dto.FaultInjectionNoIssuesResp{
			DatasetID:     noIssues.DatasetID,
			DisplayConfig: noIssues.DisplayConfig,
			EngineConfig:  conf,
			PreDuration:   noIssues.PreDuration,
			InjectionName: noIssues.InjectionName,
		})
		return
	}

	// 如果都没有找到，返回404
	dto.ErrorResponse(c, http.StatusNotFound, "Fault injection record not found for the given dataset ID")
}
