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
	"github.com/LGU-SE-Internal/rcabench/config"
	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/executor"
	"github.com/LGU-SE-Internal/rcabench/middleware"
	"github.com/LGU-SE-Internal/rcabench/repository"
	"github.com/LGU-SE-Internal/rcabench/utils"
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
//	@Description	获取指定命名空间的故障注入配置信息，支持不同显示模式的配置树结构
//	@Tags			injection
//	@Produce		json
//	@Param			namespace	query		string	true	"命名空间，指定要获取配置的命名空间"
//	@Param			mode		query		string	false	"显示模式"	Enums(display, engine) default(engine)
//	@Success		200			{object}	dto.GenericResponse[chaos.Node]	"成功返回配置树结构"
//	@Failure		400			{object}	dto.GenericResponse[any]	"请求参数错误，如命名空间或模式参数缺失"
//	@Failure		500			{object}	dto.GenericResponse[any]	"服务器内部错误"
//	@Router			/api/v1/injections/conf [get]
func GetInjectionConf(c *gin.Context) {
	var req dto.InjectionConfReq
	if err := c.BindQuery(&req); err != nil {
		logrus.Errorf("failed to bind injection conf request: %v", err)
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	root, err := chaos.StructToNode[chaos.InjectionConf](req.Namespace)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"namespace": req.Namespace,
			"mode":      req.Mode,
		}).Errorf("struct InjectionConf to node failed: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to read injection configuration")
		return
	}

	dto.SuccessResponse(c, root)
}

// GetInjectionFieldMapping
//
//	@Summary		获取字段映射关系
//	@Description	获取状态和故障类型的字符串与数字映射关系，用于前端显示和API参数验证
//	@Tags			injection
//	@Produce		json
//	@Success		200	{object}	dto.GenericResponse[dto.InjectionFieldMappingResp]	"成功返回字段映射关系"
//	@Failure		500	{object}	dto.GenericResponse[any]	"服务器内部错误"
//	@Router			/api/v1/injections/mapping [get]
func GetInjectionFieldMapping(c *gin.Context) {
	if dto.DatasetStatusMap == nil || chaos.ChaosTypeMap == nil {
		logrus.Error("field mapping data is not initialized")
		dto.ErrorResponse(c, http.StatusInternalServerError, "Field mapping not available")
		return
	}

	dto.SuccessResponse(c, dto.InjectionFieldMappingResp{
		StatusMap:    dto.DatasetStatusMap,
		FaultTypeMap: chaos.ChaosTypeMap,
	})
}

// GetNsResourceMap
//
//	@Summary		获取命名空间资源映射
//	@Description	获取所有命名空间及其对应的资源信息映射，或查询指定命名空间的资源信息。返回命名空间到资源的映射表，用于故障注入配置和资源管理
//	@Tags			injection
//	@Produce		json
//	@Param			namespace	query		string	false	"命名空间名称，不指定时返回所有命名空间的资源映射"
//	@Success		200			{object}	dto.GenericResponse[dto.NsResourceResp]	"成功返回命名空间资源映射表"
//	@Success		200			{object}	dto.GenericResponse[chaos.Resource]		"指定命名空间时返回该命名空间的资源信息"
//	@Failure		404			{object}	dto.GenericResponse[any]				"指定的命名空间不存在"
//	@Failure		500			{object}	dto.GenericResponse[any]				"服务器内部错误，无法获取资源映射"
//	@Router			/api/v1/injections/ns-resources [get]
func GetNsResourceMap(c *gin.Context) {
	namespace := c.Query("namespace")

	resourceMap, err := chaos.GetAllResources()
	if err != nil {
		logrus.Errorf("failed to get all resources: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get all resources")
		return
	}

	if namespace != "" {
		resource, exists := resourceMap[namespace]
		if !exists {
			logrus.Errorf("namespace %s not found in resource map", namespace)
			dto.ErrorResponse(c, http.StatusNotFound, fmt.Sprintf("Namespace %s not found", namespace))
			return
		}

		dto.SuccessResponse(c, resource)
		return
	}

	dto.SuccessResponse(c, dto.NsResourceResp(resourceMap))
}

// ListDisplayConfigs
//
//	@Summary		获取已注入故障配置列表
//	@Description	根据多个TraceID获取对应的故障注入配置信息，用于查看已提交的故障注入任务的配置详情
//	@Tags			injection
//	@Produce		json
//	@Param			trace_ids	query		[]string	false	"TraceID列表，支持多个值，用于查询对应的配置信息"	collectionFormat(multi)
//	@Success		200			{object}	dto.GenericResponse[any]	"成功返回配置列表"
//	@Failure		400			{object}	dto.GenericResponse[any]	"请求参数错误，如TraceID参数缺失或格式不正确"
//	@Failure		500			{object}	dto.GenericResponse[any]	"服务器内部错误"
//	@Router			/api/v1/injections/configs [get]
func ListDisplayConfigs(c *gin.Context) {
	var req dto.ListDisplayConfigsReq
	if err := c.BindQuery(&req); err != nil {
		logrus.Errorf("failed to bind query parameters: %v", err)
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	if err := req.Validate(); err != nil {
		logrus.Errorf("invalid query parameters: %v", err)
		dto.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	configs, err := repository.ListDisplayConfigsByTraceIDs(req.TraceIDs)
	if err != nil {
		logrus.Errorf("failed to get display configs by trace IDs: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get injection configuration list")
		return
	}

	dto.SuccessResponse(c, configs)
}

// ListInjections
//
//	@Summary		获取故障注入记录列表
//	@Description	支持排序、过滤的故障注入记录查询接口。返回数据库原始记录列表，不进行数据转换。
//	@Tags			injection
//	@Produce		json
//	@Param			env					query	string	false	"环境标签过滤"
//	@Param			batch				query	string	false	"批次标签过滤"
//	@Param			benchmark			query	string	false	"基准测试类型过滤"	Enums(clickhouse)
//	@Param			status				query	int		false	"状态过滤，具体值参考字段映射接口(/mapping)"	default(0)
//	@Param			fault_type			query	int		false	"故障类型过滤，具体值参考字段映射接口(/mapping)"	default(0)
//	@Param			sort				query	string	false	"排序方式，默认desc。按created_at字段排序"	Enums(asc, desc)	default(desc)
//	@Param			limit				query	int		false	"结果数量限制，用于控制返回记录数量"	minimum(1)
//	@Param			lookback			query	string	false	"时间范围查询，支持自定义相对时间(1h/24h/7d)或custom 默认不设置"
//	@Param			custom_start_time	query	string	false	"自定义开始时间，RFC3339格式，当lookback=custom时必需"	Format(date-time)
//	@Param			custom_end_time		query	string	false	"自定义结束时间，RFC3339格式，当lookback=custom时必需"	Format(date-time)
//	@Success		200					{object}	dto.GenericResponse[[]database.FaultInjectionSchedule]	"成功返回故障注入记录列表"
//	@Failure		400			{object}	dto.GenericResponse[any]	"请求参数错误，如参数格式不正确、验证失败等"
//	@Failure		500			{object}	dto.GenericResponse[any]	"服务器内部错误"
//	@Router			/api/v1/injections [get]
func ListInjections(c *gin.Context) {
	var req dto.ListInjectionsReq
	if err := c.BindQuery(&req); err != nil {
		logrus.Errorf("failed to bind query parameters: %v", err)
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	if err := req.Validate(); err != nil {
		logrus.Errorf("invalid query parameters: %v", err)
		dto.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	_, records, err := repository.ListInjections(&req)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	dto.SuccessResponse(c, records)
}

// QueryInjection
//
//	@Summary		查询单个故障注入记录
//	@Description	根据名称或任务ID查询故障注入记录详情，两个参数至少提供一个
//	@Tags			injection
//	@Produce		json
//	@Param			name		query		string	false	"故障注入名称"
//	@Param			task_id		query		string	false	"任务ID"
//	@Success		200			{object}	dto.GenericResponse[database.FaultInjectionSchedule]	"成功返回故障注入记录详情"
//	@Failure		400			{object}	dto.GenericResponse[any]	"请求参数错误，如参数缺失、格式不正确或验证失败等"
//	@Failure		500			{object}	dto.GenericResponse[any]	"服务器内部错误"
//	@Router			/api/v1/injections/query [get]
func QueryInjection(c *gin.Context) {
	var req dto.QueryInjectionReq
	if err := c.BindQuery(&req); err != nil {
		logrus.Errorf("failed to bind query parameters: %v", err)
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	if err := req.Validate(); err != nil {
		logrus.Errorf("invalid query parameters: %v", err)
		dto.ErrorResponse(c, http.StatusBadRequest, err.Error())
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
//	@Summary		提交故障注入任务
//	@Description	提交故障注入任务，支持批量提交多个故障配置，系统会自动去重并返回提交结果
//	@Tags			injection
//	@Produce		json
//	@Consumes		application/json
//	@Param			body	body		dto.SubmitInjectionReq	true	"故障注入请求体"
//	@Success		202		{object}	dto.GenericResponse[dto.SubmitInjectionResp]	"成功提交故障注入任务"
//	@Failure		400		{object}	dto.GenericResponse[any]	"请求参数错误，如JSON格式不正确、参数验证失败或算法无效等"
//	@Failure		500		{object}	dto.GenericResponse[any]	"服务器内部错误"
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

	var req dto.SubmitInjectionReq
	if err := c.BindJSON(&req); err != nil {
		handleError(err, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		handleError(err, "Invalid query parameters", http.StatusBadRequest)
		return
	}

	if err := validateAlgorithms(req.Algorithms); err != nil {
		handleError(err, "Invalid algorithm specified", http.StatusBadRequest)
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
	logrus.Infof("Duplicated %d configurations, original count: %d", len(req.Specs)-len(configs), len(req.Specs))

	dto.JSONResponse(c, http.StatusAccepted, "Fault injections submitted successfully", dto.SubmitInjectionResp{
		DuplicatedCount: duplicatedCount,
		OriginalCount:   len(req.Specs),
		SubmitResp: dto.SubmitResp{
			GroupID: groupID,
			Traces:  traces,
		},
	})
}

func validateAlgorithms(algorithms []string) error {
	validAlgorithms, err := repository.ListContainers(&dto.FilterContainerOptions{
		Status: utils.BoolPtr(true),
		Type:   consts.ContainerTypeAlgorithm,
	})
	if err != nil {
		return fmt.Errorf("Failed to list algorithms: %v", err)
	}

	validAlgorithmMap := make(map[string]struct{}, len(validAlgorithms))
	for _, algorithm := range validAlgorithms {
		validAlgorithmMap[algorithm.Name] = struct{}{}
	}

	for _, algorithm := range algorithms {
		if algorithm == "" {
			return fmt.Errorf("Algorithm must not be empty")
		}

		detector := config.GetString("algo.detector")
		if algorithm == detector {
			return fmt.Errorf("Algorithm %s is not allowed for fault injection", detector)
		}

		if _, exists := validAlgorithmMap[algorithm]; !exists {
			return fmt.Errorf("Invalid algorithm: %s", algorithm)
		}
	}

	return nil
}

func parseInjectionSpecs(r *dto.SubmitInjectionReq) ([]*dto.InjectionConfig, error) {
	configs := make([]*dto.InjectionConfig, 0, len(r.Specs))
	displayDatas := make([]string, 0, len(r.Specs))

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

		existingBatch, err := repository.ListExistingDisplayConfigs(batch)
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

func createInjectionTask(req *dto.SubmitInjectionReq, config *dto.InjectionConfig, groupID string, spanCtx context.Context) *dto.UnifiedTask {
	payload := map[string]any{
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

	task := &dto.UnifiedTask{
		Type:        consts.TaskTypeRestartService,
		Payload:     payload,
		Immediate:   false,
		ExecuteTime: config.ExecuteTime.Unix(),
		GroupID:     groupID,
	}
	task.SetGroupCtx(spanCtx)

	return task
}

// analysis

// GetFaultInjectionNoIssues
//
//	@Summary		查询没有问题的故障注入记录
//	@Description	根据时间范围查询所有没有问题的故障注入记录列表，返回包含配置信息的详细记录
//	@Tags			injection
//	@Produce		json
//	@Param			env					query	string	false	"环境标签过滤"
//	@Param			batch				query	string	false	"批次标签过滤"
//	@Param			lookback			query	string	false	"时间范围查询，支持自定义相对时间(1h/24h/7d)或custom 默认不设置"
//	@Param			custom_start_time	query	string	false	"自定义开始时间，RFC3339格式，当lookback=custom时必需"	Format(date-time)
//	@Param			custom_end_time		query	string	false	"自定义结束时间，RFC3339格式，当lookback=custom时必需"	Format(date-time)
//	@Success		200					{object}	dto.GenericResponse[[]dto.FaultInjectionNoIssuesResp]	"成功返回没有问题的故障注入记录列表"
//	@Failure		400					{object}	dto.GenericResponse[any]	"请求参数错误，如时间格式不正确或参数验证失败等"
//	@Failure		500					{object}	dto.GenericResponse[any]	"服务器内部错"
//	@Router			/api/v1/injections/analysis/no-issues [get]
func GetFaultInjectionNoIssues(c *gin.Context) {
	var req dto.FaultInjectionNoIssuesReq
	if err := c.BindQuery(&req); err != nil {
		logrus.Errorf("failed to bind query parameters: %v", err)
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	if err := req.Validate(); err != nil {
		logrus.Errorf("invalid query parameters: %v", err)
		dto.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	_, records, err := repository.GetAllFaultInjectionNoIssues(&req)
	if err != nil {
		logrus.Errorf("failed to get injection record with no issues: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get injection records")
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
//	@Param			env					query	string	false	"环境标签过滤"
//	@Param			batch				query	string	false	"批次标签过滤"
//	@Param			lookback			query	string	false	"时间范围查询，支持自定义相对时间(1h/24h/7d)或custom 默认不设置"
//	@Param			custom_start_time	query	string	false	"自定义开始时间，RFC3339格式，当lookback=custom时必需"	Format(date-time)
//	@Param			custom_end_time		query	string	false	"自定义结束时间，RFC3339格式，当lookback=custom时必需"	Format(date-time)
//	@Success		200					{object}	dto.GenericResponse[[]dto.FaultInjectionWithIssuesResp]
//	@Failure		400					{object}	dto.GenericResponse[any]	"请求参数错误，如时间格式不正确或参数验证失败等"
//	@Failure		500					{object}	dto.GenericResponse[any]	"服务器内部错误"
//	@Router			/api/v1/injections/analysis/with-issues [get]
func GetFaultInjectionWithIssues(c *gin.Context) {
	var req dto.FaultInjectionWithIssuesReq
	if err := c.BindQuery(&req); err != nil {
		logrus.Errorf("failed to bind query parameters: %v", err)
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	if err := req.Validate(); err != nil {
		logrus.Errorf("invalid query parameters: %v", err)
		dto.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	records, err := repository.GetAllFaultInjectionWithIssues(&req)
	if err != nil {
		logrus.Errorf("failed to get fault injection with issues: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get fault injection records")
		return
	}

	var items []dto.FaultInjectionWithIssuesResp
	for _, record := range records {
		conf := chaos.Node{}
		err := json.Unmarshal([]byte(record.EngineConfig), &conf)
		if err != nil {
			logrus.Errorf("failed to unmarshal engine config: %v", err)
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to parse engine config")
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
//	@Description	获取故障注入记录的统计信息，包括有问题、没有问题和总记录数量
//	@Tags			injection
//	@Produce		json
//	@Param			lookback			query	string	false	"时间范围查询，支持自定义相对时间(1h/24h/7d)或custom 默认不设置"
//	@Param			custom_start_time	query	string	false	"自定义开始时间，RFC3339格式，当lookback=custom时必需"	Format(date-time)
//	@Param			custom_end_time		query	string	false	"自定义结束时间，RFC3339格式，当lookback=custom时必需"	Format(date-time)
//	@Success		200					{object}	dto.GenericResponse[dto.FaultInjectionStatisticsResp]	"成功返回故障注入统计信息"
//	@Failure		400					{object}	dto.GenericResponse[any]	"请求参数错误，如时间格式不正确或参数验证失败等"
//	@Failure		500					{object}	dto.GenericResponse[any]	"服务器内部错误"
//	@Router			/api/v1/injections/analysis/statistics [get]
func GetFaultInjectionStatistics(c *gin.Context) {
	var req dto.TimeRangeQuery
	if err := c.BindQuery(&req); err != nil {
		logrus.Errorf("failed to bind query parameters: %v", err)
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	if err := req.Validate(); err != nil {
		logrus.Errorf("invalid query parameters: %v", err)
		dto.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	opts, err := req.Convert()
	if err != nil {
		logrus.Errorf("failed to convert request: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to convert request")
		return
	}

	stats, err := repository.GetFaultInjectionStatistics(*opts)
	if err != nil {
		logrus.Errorf("failed to get injection statistics: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get injection statistics")
		return
	}

	dto.SuccessResponse(c, dto.FaultInjectionStatisticsResp{
		NoIssuesCount:   stats["no_issues"],
		WithIssuesCount: stats["with_issues"],
		TotalCount:      stats["total"],
	})
}
