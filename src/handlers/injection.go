package handlers

import (
	"context"
	"encoding/json"
	"fmt"
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
//	@Summary		Cancel Fault Injection Task
//	@Description	Cancel the specified fault injection task
//	@Tags			injection
//	@Produce		application/json
//	@Param			task_id	path		string	true	"Task ID"
//	@Success		200		{object}	dto.GenericResponse[dto.InjectCancelResp]
//	@Failure		400		{object}	dto.GenericResponse[any]
//	@Failure		500		{object}	dto.GenericResponse[any]
//	@Router			/api/v1/injections/{task_id}/cancel [put]
func CancelInjection(c *gin.Context) {
}

// GetInjectionConf
//
//	@Summary		Get Fault Injection Configuration
//	@Description	Get fault injection configuration for the specified namespace, supporting different display modes for configuration tree structure
//	@Tags			injection
//	@Produce		json
//	@Param			namespace	query		string	true	"Namespace, specifies the namespace to get configuration for"
//	@Param			mode		query		string	false	"Display mode"	Enums(display, engine) default(engine)
//	@Success		200			{object}	dto.GenericResponse[chaos.Node]	"Successfully returned configuration tree structure"
//	@Failure		400			{object}	dto.GenericResponse[any]	"Request parameter error, such as missing namespace or mode parameter"
//	@Failure		500			{object}	dto.GenericResponse[any]	"Internal server error"
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
//	@Summary		Get Field Mapping
//	@Description	Get string-to-number mapping relationships for status and fault types, used for frontend display and API parameter validation
//	@Tags			injection
//	@Produce		json
//	@Success		200	{object}	dto.GenericResponse[dto.InjectionFieldMappingResp]	"Successfully returned field mapping relationships"
//	@Failure		500	{object}	dto.GenericResponse[any]	"Internal server error"
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
//	@Summary		Get Namespace Resource Mapping
//	@Description	Get mapping of all namespaces and their corresponding resource information, or query resource information for a specific namespace. Returns a mapping table from namespace to resources, used for fault injection configuration and resource management
//	@Tags			injection
//	@Produce		json
//	@Param			namespace	query		string	false	"Namespace name, returns resource mappings for all namespaces if not specified"
//	@Success		200			{object}	dto.GenericResponse[dto.NsResourcesResp]	"Successfully returned namespace resource mapping table"
//	@Success		200			{object}	dto.GenericResponse[chaos.Resources]		"Returns resource information for the specified namespace when a namespace is provided"
//	@Failure		404			{object}	dto.GenericResponse[any]					"The specified namespace does not exist"
//	@Failure		500			{object}	dto.GenericResponse[any]					"Internal server error, unable to get resource mapping"
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
//	@Summary		Get Injected Fault Configuration List
//	@Description	Get fault injection configuration information based on multiple TraceIDs, used to view configuration details of submitted fault injection tasks
//	@Tags			injection
//	@Produce		json
//	@Param			trace_ids	query		[]string	false	"TraceID list, supports multiple values, used to query corresponding configuration information"	collectionFormat(multi)
//	@Success		200			{object}	dto.GenericResponse[any]	"Successfully returned configuration list"
//	@Failure		400			{object}	dto.GenericResponse[any]	"Request parameter error, such as missing TraceID parameter or incorrect format"
//	@Failure		500			{object}	dto.GenericResponse[any]	"Internal server error"
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
//	@Summary		Get Fault Injection Record List
//	@Description	Fault injection record query interface supporting sorting and filtering. Returns the original database record list without data conversion.
//	@Tags			injection
//	@Produce		json
//	@Param			project_name		query		string	false	"Project name filter"
//	@Param			env					query		string	false	"Environment label filter"	Enums(dev, prod)	default(prod)
//	@Param			batch				query		string	false	"Batch label filter"
//	@Param			tag					query		string	false	"Category label filter"	Enums(train, test)	default(train)
//	@Param			benchmark			query		string	false	"Benchmark type filter"	Enums(clickhouse)	default(clickhouse)
//	@Param			status				query		int		false	"Status filter, refer to field mapping interface (/mapping) for specific values"	default(0)
//	@Param			fault_type			query		int		false	"Fault type filter, refer to field mapping interface (/mapping) for specific values"	default(0)
//	@Param			sort_field			query		string	false	"Sort field, default created_at" default(created_at)
//	@Param			sort_order			query		string	false	"Sort order, default desc"	Enums(asc, desc)	default(desc)
//	@Param			limit				query		int		false	"Result quantity limit, used to control the number of returned records"	minimum(0)	default(0)
//	@Param			page_num			query		int		false	"Pagination query, page number"	minimum(0)	default(0)
//	@Param			page_size			query		int		false	"Pagination query, records per page"	minimum(0)	default(0)
//	@Param			lookback			query		string	false	"Time range query, supports custom relative time (1h/24h/7d) or custom, default not set"
//	@Param			custom_start_time	query		string	false	"Custom start time, RFC3339 format, required when lookback=custom"	Format(date-time)
//	@Param			custom_end_time		query		string	false	"Custom end time, RFC3339 format, required when lookback=custom"	Format(date-time)
//	@Success		200					{object}	dto.GenericResponse[dto.ListInjectionsResp]	"Successfully returned fault injection record list"
//	@Failure		400					{object}	dto.GenericResponse[any]	"Request parameter error, such as incorrect parameter format, validation failure, etc."
//	@Failure		500					{object}	dto.GenericResponse[any]	"Internal server error"
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
//	@Summary		Query Single Fault Injection Record
//	@Description	Query fault injection record details by name or task ID, at least one of the two parameters must be provided
//	@Tags			injection
//	@Produce		json
//	@Param			name		query		string	false	"Fault injection name"
//	@Param			task_id		query		string	false	"Task ID"
//	@Success		200			{object}	dto.GenericResponse[dto.QueryInjectionResp]	"Successfully returned fault injection record details"
//	@Failure		400			{object}	dto.GenericResponse[any]	"Request parameter error, such as missing parameters, incorrect format, or validation failure, etc."
//	@Failure		500			{object}	dto.GenericResponse[any]	"Internal server error"
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
//	@Summary		Submit Fault Injection Task
//	@Description	Submit fault injection task, supporting batch submission of multiple fault configurations, the system will automatically deduplicate and return submission results
//	@Tags			injection
//	@Produce		json
//	@Consumes		application/json
//	@Param			body	body		dto.SubmitInjectionReq	true	"Fault injection request body"
//	@Success		202		{object}	dto.GenericResponse[dto.SubmitInjectionResp]	"Successfully submitted fault injection task"
//	@Failure		400		{object}	dto.GenericResponse[any]	"Request parameter error, such as incorrect JSON format, parameter validation failure, or invalid algorithm, etc."
//	@Failure		500		{object}	dto.GenericResponse[any]	"Internal server error"
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
		span.SetStatus(codes.Error, "failed in SubmitFaultInjection - removeDuplicated")
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
	// Goal: filter out items that already exist in DB, using engine_config as uniqueness key,
	// and drop duplicates within the incoming request while preserving order.
	if len(configs) == 0 {
		return nil, nil
	}

	// Build engine_config JSON string for each item from cfg.Node, because DB stores EngineConfig as JSON string.
	engineStrings := make([]string, len(configs))
	for i, cfg := range configs {
		if cfg.Node == nil {
			engineStrings[i] = "" // cannot dedup without engine config; keep it later
			continue
		}
		b, err := json.Marshal(cfg.Node)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal engine config at index %d: %w", i, err)
		}
		engineStrings[i] = string(b)
	}

	orderedUniqueIdx := make([]int, 0, len(configs))
	seen := make(map[string]struct{}, len(configs))
	for i, key := range engineStrings {
		if key == "" {
			orderedUniqueIdx = append(orderedUniqueIdx, i)
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		orderedUniqueIdx = append(orderedUniqueIdx, i)
	}

	// Query DB for existing engine_config values in batches
	const batchSize = 100
	existed := make(map[string]struct{})
	// Collect unique keys to query to avoid excessive IN list
	keys := make([]string, 0, len(seen))
	for k := range seen {
		if k != "" {
			keys = append(keys, k)
		}
	}
	for start := 0; start < len(keys); start += batchSize {
		end := start + batchSize
		if end > len(keys) {
			end = len(keys)
		}
		batch := keys[start:end]
		existing, err := repository.ListExistingEngineConfigs(batch)
		if err != nil {
			return nil, err
		}
		for _, v := range existing {
			existed[v] = struct{}{}
		}
	}

	out := make([]dto.InjectionConfig, 0, len(orderedUniqueIdx))
	for _, idx := range orderedUniqueIdx {
		key := engineStrings[idx]
		if key == "" {
			out = append(out, configs[idx])
			continue
		}
		if _, ok := existed[key]; ok {
			continue
		}
		configs[idx].ExecuteTime = time.Now().Add(time.Duration(idx*2) * time.Second) // ensure unique execute_time
		out = append(out, configs[idx])
	}
	return out, nil
}

// GetFaultInjectionNoIssues
//
//	@Summary		Query Fault Injection Records Without Issues
//	@Description	Query all fault injection records without issues based on time range, returning detailed records including configuration information
//	@Tags			injection
//	@Produce		json
//	@Param			env					query	string	false	"Environment label filter"
//	@Param			batch				query	string	false	"Batch label filter"
//	@Param			lookback			query	string	false	"Time range query, supports custom relative time (1h/24h/7d) or custom, default not set"
//	@Param			custom_start_time	query	string	false	"Custom start time, RFC3339 format, required when lookback=custom"	Format(date-time)
//	@Param			custom_end_time		query	string	false	"Custom end time, RFC3339 format, required when lookback=custom"	Format(date-time)
//	@Success		200					{object}	dto.GenericResponse[[]dto.FaultInjectionNoIssuesResp]	"Successfully returned fault injection records without issues"
//	@Failure		400					{object}	dto.GenericResponse[any]	"Request parameter error, such as incorrect time format or parameter validation failure, etc."
//	@Failure		500					{object}	dto.GenericResponse[any]	"Internal server error"
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
//	@Summary		Query Fault Injection Records With Issues
//	@Description	Query all fault injection records with issues based on time range
//	@Tags			injection
//	@Produce		json
//	@Param			env					query	string	false	"Environment label filter"
//	@Param			batch				query	string	false	"Batch label filter"
//	@Param			lookback			query	string	false	"Time range query, supports custom relative time (1h/24h/7d) or custom, default not set"
//	@Param			custom_start_time	query	string	false	"Custom start time, RFC3339 format, required when lookback=custom"	Format(date-time)
//	@Param			custom_end_time		query	string	false	"Custom end time, RFC3339 format, required when lookback=custom"	Format(date-time)
//	@Success		200					{object}	dto.GenericResponse[[]dto.FaultInjectionWithIssuesResp]
//	@Failure		400					{object}	dto.GenericResponse[any]	"Request parameter error, such as incorrect time format or parameter validation failure, etc."
//	@Failure		500					{object}	dto.GenericResponse[any]	"Internal server error"
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
//	@Summary		Get Fault Injection Statistics
//	@Description	Get statistical information of fault injection records, including counts of records with issues, without issues, and total records
//	@Tags			injection
//	@Produce		json
//	@Param			lookback			query	string	false	"Time range query, supports custom relative time (1h/24h/7d) or custom, default not set"
//	@Param			custom_start_time	query	string	false	"Custom start time, RFC3339 format, required when lookback=custom"	Format(date-time)
//	@Param			custom_end_time		query	string	false	"Custom end time, RFC3339 format, required when lookback=custom"	Format(date-time)
//	@Success		200					{object}	dto.GenericResponse[dto.InjectionStatsResp]	"Successfully returned fault injection statistics"
//	@Failure		400					{object}	dto.GenericResponse[any]	"Request parameter error, such as incorrect time format or parameter validation failure, etc."
//	@Failure		500					{object}	dto.GenericResponse[any]	"Internal server error"
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
