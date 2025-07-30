package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"runtime/debug"
	"time"

	chaos "github.com/LGU-SE-Internal/chaos-experiment/handler"
	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/executor"
	"github.com/LGU-SE-Internal/rcabench/middleware"
	"github.com/LGU-SE-Internal/rcabench/repository"
	"github.com/LGU-SE-Internal/rcabench/utils"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
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

	faultResourceMap, err := chaos.GetChaosResourceMap()
	if err != nil {
		logrus.Errorf("failed to get fault resource map: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get fault resource map")
		return
	}

	dto.SuccessResponse(c, dto.InjectionFieldMappingResp{
		StatusMap:        dto.DatasetStatusMap,
		FaultTypeMap:     chaos.ChaosTypeMap,
		FaultResourceMap: faultResourceMap,
	})
}

// GetNsResourcesMap
//
//	@Summary		获取命名空间资源映射
//	@Description	获取所有命名空间及其对应的资源信息映射，或查询指定命名空间的资源信息。返回命名空间到资源的映射表，用于故障注入配置和资源管理
//	@Tags			injection
//	@Produce		json
//	@Param			namespace	query		string	false	"命名空间名称，不指定时返回所有命名空间的资源映射"
//	@Success		200			{object}	dto.GenericResponse[dto.NsResourcesResp]	"成功返回命名空间资源映射表"
//	@Success		200			{object}	dto.GenericResponse[chaos.Resources]		"指定命名空间时返回该命名空间的资源信息"
//	@Failure		404			{object}	dto.GenericResponse[any]					"指定的命名空间不存在"
//	@Failure		500			{object}	dto.GenericResponse[any]					"服务器内部错误，无法获取资源映射"
//	@Router			/api/v1/injections/ns-resources [get]
func GetNsResourceMap(c *gin.Context) {
	namespace := c.Query("namespace")

	resourcesMap, err := chaos.GetNsResources()
	if err != nil {
		logrus.Errorf("failed to get all resources: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get all resources")
		return
	}

	if namespace != "" {
		resources, exists := resourcesMap[namespace]
		if !exists {
			logrus.Errorf("namespace %s not found in resources map", namespace)
			dto.ErrorResponse(c, http.StatusNotFound, fmt.Sprintf("Namespace %s not found", namespace))
			return
		}

		dto.SuccessResponse(c, resources)
		return
	}

	dto.SuccessResponse(c, dto.NsResourcesResp(resourcesMap))
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
//	@Param			project_name		query		string	false	"项目名称过滤"
//	@Param			env					query		string	false	"环境标签过滤"	Enums(dev, prod)	default(prod)
//	@Param			batch				query		string	false	"批次标签过滤"
//	@Param			tag					query		string	false	"分类标签过滤"	Enums(train, test)	default(train)
//	@Param			benchmark			query		string	false	"基准测试类型过滤"	Enums(clickhouse)	default(clickhouse)
//	@Param			status				query		int		false	"状态过滤，具体值参考字段映射接口(/mapping)"	default(0)
//	@Param			fault_type			query		int		false	"故障类型过滤，具体值参考字段映射接口(/mapping)"	default(0)
//	@Param			sort_field			query		string	false	"排序字段，默认created_at" default(created_at)
//	@Param			sort_order			query		string	false	"排序方式，默认desc"	Enums(asc, desc)	default(desc)
//	@Param			limit				query		int		false	"结果数量限制，用于控制返回记录数量"	minimum(0)	default(0)
//	@Param			page_num			query		int		false	"分页查询，页码"	minimum(0)	default(0)
//	@Param			page_size			query		int		false	"分页查询，每页数量"	minimum(0)	default(0)
//	@Param			lookback			query		string	false	"时间范围查询，支持自定义相对时间(1h/24h/7d)或custom 默认不设置"
//	@Param			custom_start_time	query		string	false	"自定义开始时间，RFC3339格式，当lookback=custom时必需"	Format(date-time)
//	@Param			custom_end_time		query		string	false	"自定义结束时间，RFC3339格式，当lookback=custom时必需"	Format(date-time)
//	@Success		200					{object}	dto.GenericResponse[dto.ListInjectionsResp]	"成功返回故障注入记录列表"
//	@Failure		400					{object}	dto.GenericResponse[any]	"请求参数错误，如参数格式不正确、验证失败等"
//	@Failure		500					{object}	dto.GenericResponse[any]	"服务器内部错误"
//	@Router			/api/v1/injections	[get]
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

	total, injections, err := repository.ListInjections(&req)
	if err != nil {
		logrus.Errorf("failed to list injections: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	var items []dto.InjectionItem
	if err := copier.Copy(&items, &injections); err != nil {
		logrus.Errorf("failed to copy injection records: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to copy injection records")
		return
	}

	if total == 0 {
		total = int64(len(items))
	}

	dto.SuccessResponse(c, dto.ListInjectionsResp{
		Total: total,
		Items: items,
	})
}

// QueryInjection
//
//	@Summary		查询单个故障注入记录
//	@Description	根据名称或任务ID查询故障注入记录详情，两个参数至少提供一个
//	@Tags			injection
//	@Produce		json
//	@Param			name		query		string	false	"故障注入名称"
//	@Param			task_id		query		string	false	"任务ID"
//	@Success		200			{object}	dto.GenericResponse[dto.QueryInjectionResp]	"成功返回故障注入记录详情"
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

	groundTruthMap, err := repository.GetGroundtruthMap([]string{item.InjectionName})
	if err != nil {
		logrus.Errorf("failed to get ground truth map: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get ground truth map")
		return
	}

	dto.SuccessResponse(c, dto.QueryInjectionResp{
		FaultInjectionSchedule: *item,
		GroundTruth:            groundTruthMap[item.InjectionName],
	})
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

	configs, err := req.ParseInjectionSpecs()
	if err != nil {
		handleError(err, err.Error(), http.StatusBadRequest)
		return
	}

	newConfigs, err := removeDuplicated(configs)
	if err != nil {
		logrus.Errorf("failed to remove duplicated configs: %v", err)
		span.SetStatus(codes.Error, "panic in SubmitFaultInjection")
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to remove duplicated configs")
		return
	}

	project, err := repository.GetProject("name", req.ProjectName)
	if err != nil {
		logrus.Errorf("failed to get project by name %s: %v", req.ProjectName, err)
		span.SetStatus(codes.Error, "panic in SubmitFaultInjection")
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get project by name")
		return
	}

	traces := make([]dto.Trace, 0, len(newConfigs))
	for _, config := range newConfigs {
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
			ProjectID:   &project.ID,
		}
		task.SetGroupCtx(spanCtx)

		taskID, traceID, err := executor.SubmitTask(spanCtx, task)
		if err != nil {
			handleError(err, "failed to submit injection task", http.StatusInternalServerError)
			return
		}

		traces = append(traces, dto.Trace{TraceID: traceID, HeadTaskID: taskID, Index: config.Index})
	}

	span.SetStatus(codes.Ok, fmt.Sprintf("Successfully submitted %d fault injections with groupID: %s", len(traces), groupID))

	duplicatedCount := len(req.Specs) - len(newConfigs)
	logrus.Infof("Duplicated %d configurations, original count: %d", duplicatedCount, len(req.Specs))

	if len(req.Algorithms) != 0 {
		if err := repository.SetAlgorithmItemsToRedis(spanCtx, consts.InjectionAlgorithmsKey, groupID, req.Algorithms); err != nil {
			logrus.Errorf("failed to cache algorithms items to Redis: %v", err)
			span.SetStatus(codes.Error, "panic in SubmitFaultInjection")
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to cache algorithm items to Redis")
			return
		}
	}

	dto.JSONResponse(c, http.StatusAccepted, "Fault injections submitted successfully", dto.SubmitInjectionResp{
		DuplicatedCount: duplicatedCount,
		OriginalCount:   len(req.Specs),
		SubmitResp: dto.SubmitResp{
			GroupID: groupID,
			Traces:  traces,
		},
	})
}

func removeDuplicated(configs []dto.InjectionConfig) ([]dto.InjectionConfig, error) {

	displayDatas := []string{}
	for _, config := range configs {
		displayDatas = append(displayDatas, config.DisplayData)
	}

	missingIndices, err := findMissingIndices(displayDatas, 10)
	if err != nil {
		return nil, err
	}

	newConfigs := make([]dto.InjectionConfig, 0)
	for _, idx := range missingIndices {
		conf := configs[idx]
		conf.ExecuteTime = time.Now().Add(time.Second * time.Duration(rand.Int()%120))
		newConfigs = append(newConfigs, conf)
	}

	return newConfigs, nil
}

// TODO 修复container的时候因为pod一定不同，可以重复注入
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
			EngineConfig:  conf,
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

	_, records, err := repository.GetAllFaultInjectionWithIssues(&req)
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
			DatasetID:           record.DatasetID,
			EngineConfig:        conf,
			InjectionName:       record.InjectionName,
			Issues:              record.Issues,
			AbnormalAvgDuration: record.AbnormalAvgDuration,
			NormalAvgDuration:   record.NormalAvgDuration,
			AbnormalSuccRate:    record.AbnormalSuccRate,
			NormalSuccRate:      record.NormalSuccRate,
			AbnormalP99:         record.AbnormalP99,
			NormalP99:           record.NormalP99,
		})
	}

	dto.SuccessResponse(c, items)
}

// GetInjectionStats
//
//	@Summary		获取故障注入统计信息
//	@Description	获取故障注入记录的统计信息，包括有问题、没有问题和总记录数量
//	@Tags			injection
//	@Produce		json
//	@Param			lookback			query	string	false	"时间范围查询，支持自定义相对时间(1h/24h/7d)或custom 默认不设置"
//	@Param			custom_start_time	query	string	false	"自定义开始时间，RFC3339格式，当lookback=custom时必需"	Format(date-time)
//	@Param			custom_end_time		query	string	false	"自定义结束时间，RFC3339格式，当lookback=custom时必需"	Format(date-time)
//	@Success		200					{object}	dto.GenericResponse[dto.InjectionStatsResp]	"成功返回故障注入统计信息"
//	@Failure		400					{object}	dto.GenericResponse[any]	"请求参数错误，如时间格式不正确或参数验证失败等"
//	@Failure		500					{object}	dto.GenericResponse[any]	"服务器内部错误"
//	@Router			/api/v1/injections/analysis/stats [get]
func GetInjectionStats(c *gin.Context) {
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

	stats, err := repository.GetInjectionStats(&req)
	if err != nil {
		logrus.Errorf("failed to get injection stats: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get injection stats")
		return
	}

	resp, err := utils.ConvertToType[dto.InjectionStatsResp](stats)
	if err != nil {
		logrus.Errorf("failed to convert to type: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to convert to resp type")
		return
	}

	dto.SuccessResponse(c, resp)
}
