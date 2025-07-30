package v2

import (
	"context"
	"fmt"
	"net/http"

	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/database"
	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/executor"
	"github.com/LGU-SE-Internal/rcabench/repository"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// SearchAlgorithms handles complex algorithm search with advanced filtering
//
//	@Summary Search algorithms
//	@Description Search algorithms with complex filtering, sorting and pagination. Algorithms are containers with type 'algorithm'
//	@Tags Algorithms
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param request body dto.AlgorithmSearchRequest true "Algorithm search request"
//	@Success 200 {object} dto.GenericResponse[dto.SearchResponse[dto.AlgorithmResponse]] "Algorithms retrieved successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/algorithms/search [post]
func SearchAlgorithms(c *gin.Context) {
	// Check permission first
	userID, exists := c.Get("user_id")
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	checker := repository.NewPermissionChecker(userID.(int), nil)
	canRead, err := checker.CanReadResource(consts.ResourceContainer)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Permission check failed: "+err.Error())
		return
	}

	if !canRead {
		dto.ErrorResponse(c, http.StatusForbidden, "No permission to read algorithms")
		return
	}

	var req dto.AlgorithmSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	// Convert to SearchRequest and ensure algorithm type filter
	searchReq := req.ConvertToSearchRequest()
	searchReq.AddFilter("type", dto.OpEqual, string(consts.ContainerTypeAlgorithm))

	// Validate search request
	if err := searchReq.ValidateSearchRequest(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid search parameters: "+err.Error())
		return
	}

	// Execute search using query builder
	searchResult, err := repository.ExecuteSearch(database.DB, searchReq, database.Container{})
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to search algorithms: "+err.Error())
		return
	}

	// Convert database containers to response DTOs
	var algorithmResponses []dto.AlgorithmResponse
	for _, container := range searchResult.Items {
		algorithmResponse := dto.AlgorithmResponse{
			ID:        container.ID,
			Name:      container.Name,
			Type:      container.Type,
			Image:     container.Image,
			Tag:       container.Tag,
			Command:   container.Command,
			EnvVars:   container.EnvVars,
			Status:    container.Status,
			CreatedAt: container.CreatedAt,
			UpdatedAt: container.UpdatedAt,
		}

		algorithmResponses = append(algorithmResponses, algorithmResponse)
	}

	// Build final response
	response := dto.SearchResponse[dto.AlgorithmResponse]{
		Items:      algorithmResponses,
		Pagination: searchResult.Pagination,
		Filters:    searchResult.Filters,
		Sort:       searchResult.Sort,
	}

	dto.SuccessResponse(c, response)
}

// ListAlgorithms handles simple algorithm listing
//
//	@Summary List algorithms
//	@Description Get a simple list of all active algorithms without complex filtering
//	@Tags Algorithms
//	@Produce json
//	@Security BearerAuth
//	@Param page query int false "Page number" default(1)
//	@Param size query int false "Page size" default(20)
//	@Success 200 {object} dto.GenericResponse[dto.SearchResponse[dto.AlgorithmResponse]] "Algorithms retrieved successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/algorithms [get]
func ListAlgorithms(c *gin.Context) {
	// Check permission first
	userID, exists := c.Get("user_id")
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	checker := repository.NewPermissionChecker(userID.(int), nil)
	canRead, err := checker.CanReadResource(consts.ResourceContainer)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Permission check failed: "+err.Error())
		return
	}

	if !canRead {
		dto.ErrorResponse(c, http.StatusForbidden, "No permission to read algorithms")
		return
	}

	// Create a basic search request from query parameters
	req := dto.AlgorithmSearchRequest{
		AdvancedSearchRequest: dto.AdvancedSearchRequest{
			SearchRequest: dto.SearchRequest{
				Page: 1,
				Size: 20,
			},
		},
	}

	// Parse pagination from query parameters
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := parseIntParam(pageStr); err == nil && page > 0 {
			req.Page = page
		}
	}
	if sizeStr := c.Query("size"); sizeStr != "" {
		if size, err := parseIntParam(sizeStr); err == nil && size > 0 && size <= 1000 {
			req.Size = size
		}
	}

	// Convert to SearchRequest and ensure algorithm type filter and active status
	searchReq := req.ConvertToSearchRequest()
	searchReq.AddFilter("type", dto.OpEqual, string(consts.ContainerTypeAlgorithm))
	searchReq.AddFilter("status", dto.OpEqual, true)

	// Add default sorting by name
	searchReq.AddSort("name", dto.SortASC)

	// Validate search request
	if err := searchReq.ValidateSearchRequest(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid search parameters: "+err.Error())
		return
	}

	// Execute search using query builder
	searchResult, err := repository.ExecuteSearch(database.DB, searchReq, database.Container{})
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get algorithm list: "+err.Error())
		return
	}

	// Convert database containers to response DTOs
	var algorithmResponses []dto.AlgorithmResponse
	for _, container := range searchResult.Items {
		algorithmResponse := dto.AlgorithmResponse{
			ID:        container.ID,
			Name:      container.Name,
			Type:      container.Type,
			Image:     container.Image,
			Tag:       container.Tag,
			Command:   container.Command,
			EnvVars:   container.EnvVars,
			Status:    container.Status,
			CreatedAt: container.CreatedAt,
			UpdatedAt: container.UpdatedAt,
		}

		algorithmResponses = append(algorithmResponses, algorithmResponse)
	}

	// Build final response
	response := dto.SearchResponse[dto.AlgorithmResponse]{
		Items:      algorithmResponses,
		Pagination: searchResult.Pagination,
		Filters:    searchResult.Filters,
		Sort:       searchResult.Sort,
	}

	dto.SuccessResponse(c, response)
}

// UploadDetectorResults uploads detector algorithm results
//
//	@Summary Upload detector algorithm results
//	@Description Upload detection results for detector algorithms via API instead of file collection
//	@Tags Algorithms
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param algorithm_id path int true "Algorithm ID"
//	@Param execution_id path int true "Execution ID"
//	@Param request body dto.DetectorResultRequest true "Detector results"
//	@Success 200 {object} dto.GenericResponse[dto.AlgorithmResultUploadResponse] "Results uploaded successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 404 {object} dto.GenericResponse[any] "Algorithm or execution not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/algorithms/{algorithm_id}/executions/{execution_id}/detectors [post]
func UploadDetectorResults(c *gin.Context) {
	// Parse path parameters
	algorithmID, err := parseIntParam(c.Param("algorithm_id"))
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid algorithm ID: "+err.Error())
		return
	}

	executionID, err := parseIntParam(c.Param("execution_id"))
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid execution ID: "+err.Error())
		return
	}

	// Check permissions
	userID, exists := c.Get("user_id")
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	checker := repository.NewPermissionChecker(userID.(int), nil)
	canWrite, err := checker.CanWriteResource(consts.ResourceContainer)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Permission check failed: "+err.Error())
		return
	}

	if !canWrite {
		dto.ErrorResponse(c, http.StatusForbidden, "No permission to upload algorithm results")
		return
	}

	// Parse request body
	var req dto.DetectorResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	// Validate request data
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Data validation failed: "+err.Error())
		return
	}

	// Verify algorithm and execution record exist
	var algorithm database.Container
	if err := database.DB.Where("id = ? AND type = ?", algorithmID, consts.ContainerTypeAlgorithm).First(&algorithm).Error; err != nil {
		dto.ErrorResponse(c, http.StatusNotFound, "Algorithm not found")
		return
	}

	var execution database.ExecutionResult
	if err := database.DB.Where("id = ? AND algorithm_id = ?", executionID, algorithmID).First(&execution).Error; err != nil {
		dto.ErrorResponse(c, http.StatusNotFound, "Execution record not found")
		return
	}

	// Check if results already exist
	var existingCount int64
	database.DB.Model(&database.Detector{}).Where("execution_id = ?", executionID).Count(&existingCount)
	if existingCount > 0 {
		dto.ErrorResponse(c, http.StatusConflict, "Detector results already exist for this execution")
		return
	}

	// Convert to database entities
	var detectorResults []database.Detector
	for _, item := range req.Results {
		detectorResult := database.Detector{
			ExecutionID:         executionID,
			SpanName:            item.SpanName,
			Issues:              item.Issues,
			AbnormalAvgDuration: item.AbnormalAvgDuration,
			NormalAvgDuration:   item.NormalAvgDuration,
			AbnormalSuccRate:    item.AbnormalSuccRate,
			NormalSuccRate:      item.NormalSuccRate,
			AbnormalP90:         item.AbnormalP90,
			NormalP90:           item.NormalP90,
			AbnormalP95:         item.AbnormalP95,
			NormalP95:           item.NormalP95,
			AbnormalP99:         item.AbnormalP99,
			NormalP99:           item.NormalP99,
		}
		detectorResults = append(detectorResults, detectorResult)
	}

	// Save to database
	if err := database.DB.Create(&detectorResults).Error; err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to save detector results: "+err.Error())
		return
	}

	// Check for anomalies
	hasAnomalies := req.HasAnomalies()

	// Build response
	response := dto.AlgorithmResultUploadResponse{
		ExecutionID:  executionID,
		AlgorithmID:  algorithmID,
		ResultCount:  len(detectorResults),
		UploadedAt:   detectorResults[0].CreatedAt,
		HasAnomalies: hasAnomalies,
		Message:      fmt.Sprintf("Successfully uploaded %d detector results", len(detectorResults)),
	}

	dto.SuccessResponse(c, response)
}

// UploadGranularityResults uploads granularity algorithm results with dual creation modes
//
//	@Summary Upload granularity algorithm results
//	@Description Upload granularity results for regular algorithms. Supports two modes: 1) Use existing execution_id, 2) Auto-create execution using algorithm_id and datapack_id
//	@Tags Algorithms
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param algorithm_id path int true "Algorithm ID"
//	@Param execution_id path int false "Execution ID (optional - will create new if not provided)"
//	@Param request body dto.GranularityResultEnhancedRequest true "Granularity results with optional execution creation"
//	@Success 200 {object} dto.GenericResponse[dto.AlgorithmResultUploadResponse] "Results uploaded successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 404 {object} dto.GenericResponse[any] "Algorithm or datapack not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/algorithms/{algorithm_id}/results [post]
//	@Router /api/v2/algorithms/{algorithm_id}/executions/{execution_id}/results [post]
func UploadGranularityResults(c *gin.Context) {
	// Parse algorithm_id parameter
	algorithmID, err := parseIntParam(c.Param("algorithm_id"))
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid algorithm ID: "+err.Error())
		return
	}

	// Parse optional execution_id parameter
	var executionID int
	executionIDParam := c.Param("execution_id")
	hasExecutionID := executionIDParam != ""

	if hasExecutionID {
		executionID, err = parseIntParam(executionIDParam)
		if err != nil {
			dto.ErrorResponse(c, http.StatusBadRequest, "Invalid execution ID: "+err.Error())
			return
		}
	}

	// Check permissions
	userID, exists := c.Get("user_id")
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	checker := repository.NewPermissionChecker(userID.(int), nil)
	canWrite, err := checker.CanWriteResource(consts.ResourceContainer)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Permission check failed: "+err.Error())
		return
	}

	if !canWrite {
		dto.ErrorResponse(c, http.StatusForbidden, "No permission to upload algorithm results")
		return
	}

	// Parse request body
	var req dto.GranularityResultEnhancedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	// Validate request data
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Data validation failed: "+err.Error())
		return
	}

	// Verify algorithm exists
	var algorithm database.Container
	if err := database.DB.Where("id = ? AND type = ?", algorithmID, consts.ContainerTypeAlgorithm).First(&algorithm).Error; err != nil {
		dto.ErrorResponse(c, http.StatusNotFound, "Algorithm not found")
		return
	}

	var execution database.ExecutionResult
	var isNewExecution bool

	if hasExecutionID {
		// Use existing execution
		if err := database.DB.Where("id = ? AND algorithm_id = ?", executionID, algorithmID).First(&execution).Error; err != nil {
			dto.ErrorResponse(c, http.StatusNotFound, "Execution record not found")
			return
		}
	} else {
		// Create new execution record
		if req.DatapackID == 0 {
			dto.ErrorResponse(c, http.StatusBadRequest, "datapack_id is required when execution_id is not provided")
			return
		}

		// Verify datapack exists (datapack is actually a FaultInjectionSchedule)
		var datapack database.FaultInjectionSchedule
		if err := database.DB.Where("id = ?", req.DatapackID).First(&datapack).Error; err != nil {
			dto.ErrorResponse(c, http.StatusNotFound, "Datapack not found")
			return
		}

		// Create new execution record
		execution = database.ExecutionResult{
			TaskID:      nil, // TaskID can be null
			AlgorithmID: algorithmID,
			DatapackID:  req.DatapackID,
			Status:      1,
		}

		if err := database.DB.Create(&execution).Error; err != nil {
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to create execution record: "+err.Error())
			return
		}

		executionID = execution.ID
		isNewExecution = true

		// Add label to indicate this is a manual upload (TaskID is null)
		labelDescription := "User manually uploaded execution result via API"
		if err := repository.AddExecutionResultLabel(executionID, consts.ExecutionLabelSource, consts.ExecutionSourceManual, labelDescription); err != nil {
			logrus.Warnf("Warning: Failed to create execution result label: %v\n", err)
		}
	}

	// Check if results already exist (only if using existing execution)
	if !isNewExecution {
		var existingCount int64
		database.DB.Model(&database.GranularityResult{}).Where("execution_id = ?", executionID).Count(&existingCount)
		if existingCount > 0 {
			dto.ErrorResponse(c, http.StatusConflict, "Granularity results already exist for this execution")
			return
		}
	}

	// Convert to database entities
	var granularityResults []database.GranularityResult
	for _, item := range req.Results {
		granularityResult := database.GranularityResult{
			ExecutionID: executionID,
			Level:       item.Level,
			Result:      item.Result,
			Rank:        item.Rank,
			Confidence:  item.Confidence,
		}
		granularityResults = append(granularityResults, granularityResult)
	}

	// Save to database
	if err := database.DB.Create(&granularityResults).Error; err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to save granularity results: "+err.Error())
		return
	}

	// Build response message
	message := fmt.Sprintf("Successfully uploaded %d granularity results", len(granularityResults))
	if isNewExecution {
		message += fmt.Sprintf(" (created new execution record with ID: %d)", executionID)
	}

	// Build response
	response := dto.AlgorithmResultUploadResponse{
		ExecutionID: executionID,
		AlgorithmID: algorithmID,
		ResultCount: len(granularityResults),
		UploadedAt:  granularityResults[0].CreatedAt,
		Message:     message,
	}

	dto.SuccessResponse(c, response)
}

// SubmitAlgorithmExecution submits algorithm execution for single datapack or dataset
//
//	@Summary Submit algorithm execution
//	@Description Submit algorithm execution task for a single datapack (v1 compatible) or dataset (v2 feature). The system will create execution tasks and return tracking information.
//	@Tags Algorithms
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param request body dto.AlgorithmExecutionRequest true "Algorithm execution request"
//	@Success 202 {object} dto.GenericResponse[dto.AlgorithmExecutionResponse] "Algorithm execution submitted successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 404 {object} dto.GenericResponse[any] "Project, algorithm, datapack or dataset not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/algorithms/execute [post]
func SubmitAlgorithmExecution(c *gin.Context) {
	// Check permissions
	userID, exists := c.Get("user_id")
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	checker := repository.NewPermissionChecker(userID.(int), nil)
	canWrite, err := checker.CanWriteResource(consts.ResourceContainer)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Permission check failed: "+err.Error())
		return
	}

	if !canWrite {
		dto.ErrorResponse(c, http.StatusForbidden, "No permission to execute algorithms")
		return
	}

	// Parse request body
	var req dto.AlgorithmExecutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	// Validate request data
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Data validation failed: "+err.Error())
		return
	}

	// Get project
	project, err := repository.GetProject("name", req.ProjectName)
	if err != nil {
		dto.ErrorResponse(c, http.StatusNotFound, "Project not found: "+req.ProjectName)
		return
	}

	// Get algorithm container
	var algorithm database.Container
	if err := database.DB.Where("name = ? AND type = ? AND status = ?", req.Algorithm.Name, consts.ContainerTypeAlgorithm, true).First(&algorithm).Error; err != nil {
		dto.ErrorResponse(c, http.StatusNotFound, "Algorithm not found: "+req.Algorithm.Name)
		return
	}

	// Create execution payload based on datapack or dataset
	var payload map[string]any
	var datapackID *int
	var datasetID *int

	if req.Datapack != nil {
		// V1 compatible mode - use datapack (FaultInjectionSchedule)
		var datapack database.FaultInjectionSchedule
		if err := database.DB.Where("injection_name = ?", *req.Datapack).First(&datapack).Error; err != nil {
			dto.ErrorResponse(c, http.StatusNotFound, "Datapack not found: "+*req.Datapack)
			return
		}
		datapackID = &datapack.ID

		// Create v1 compatible payload
		payload = map[string]any{
			"algorithm": map[string]any{
				"name":  req.Algorithm.Name,
				"image": algorithm.Image,
				"tag":   algorithm.Tag,
			},
			"dataset":  *req.Datapack, // v1 uses "dataset" field for datapack name
			"env_vars": req.EnvVars,
		}
	} else {
		// V2 mode - use dataset
		var dataset database.Dataset
		if err := database.DB.Where("name = ? AND status = ?", *req.Dataset, 1).First(&dataset).Error; err != nil {
			dto.ErrorResponse(c, http.StatusNotFound, "Dataset not found: "+*req.Dataset)
			return
		}
		datasetID = &dataset.ID

		// Create v2 payload with dataset support
		payload = map[string]any{
			"algorithm": map[string]any{
				"name":  req.Algorithm.Name,
				"image": algorithm.Image,
				"tag":   algorithm.Tag,
			},
			"dataset_name": *req.Dataset, // v2 uses "dataset_name" field
			"dataset_id":   dataset.ID,
			"env_vars":     req.EnvVars,
		}
	}

	// Create unified task
	task := &dto.UnifiedTask{
		Type:      consts.TaskTypeRunAlgorithm,
		Payload:   payload,
		Immediate: true,
		GroupID:   c.GetString("groupID"),
		ProjectID: &project.ID,
	}

	// Get span context for tracing
	ctx, ok := c.Get("spanContext")
	var spanCtx context.Context
	if ok {
		spanCtx = ctx.(context.Context)
		task.SetGroupCtx(spanCtx)
	} else {
		spanCtx = context.Background()
	}

	// Submit task
	taskID, traceID, err := executor.SubmitTask(spanCtx, task)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to submit algorithm execution task: "+err.Error())
		return
	}

	// Build response
	response := dto.AlgorithmExecutionResponse{
		TraceID:     traceID,
		TaskID:      taskID,
		AlgorithmID: algorithm.ID,
		DatapackID:  datapackID,
		DatasetID:   datasetID,
		Status:      "submitted",
	}

	dto.JSONResponse(c, http.StatusAccepted, "Algorithm execution submitted successfully", response)
}

// SubmitBatchAlgorithmExecution submits batch algorithm execution for multiple datapacks or datasets
//
//	@Summary Submit batch algorithm execution
//	@Description Submit multiple algorithm execution tasks in batch. Supports mixing datapack (v1 compatible) and dataset (v2 feature) executions.
//	@Tags Algorithms
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param request body dto.BatchAlgorithmExecutionRequest true "Batch algorithm execution request"
//	@Success 202 {object} dto.GenericResponse[dto.BatchAlgorithmExecutionResponse] "Batch algorithm execution submitted successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 404 {object} dto.GenericResponse[any] "Project, algorithm, datapack or dataset not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/algorithms/execute/batch [post]
func SubmitBatchAlgorithmExecution(c *gin.Context) {
	// Check permissions
	userID, exists := c.Get("user_id")
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	checker := repository.NewPermissionChecker(userID.(int), nil)
	canWrite, err := checker.CanWriteResource(consts.ResourceContainer)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Permission check failed: "+err.Error())
		return
	}

	if !canWrite {
		dto.ErrorResponse(c, http.StatusForbidden, "No permission to execute algorithms")
		return
	}

	// Parse request body
	var req dto.BatchAlgorithmExecutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	// Validate request data
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Data validation failed: "+err.Error())
		return
	}

	// Get project
	project, err := repository.GetProject("name", req.ProjectName)
	if err != nil {
		dto.ErrorResponse(c, http.StatusNotFound, "Project not found: "+req.ProjectName)
		return
	}

	// Get span context for tracing
	ctx, ok := c.Get("spanContext")
	var spanCtx context.Context
	if ok {
		spanCtx = ctx.(context.Context)
	} else {
		spanCtx = context.Background()
	}

	groupID := c.GetString("groupID")
	var executions []dto.AlgorithmExecutionResponse

	// Process each execution request
	for _, execution := range req.Executions {
		// Get algorithm container
		var algorithm database.Container
		if err := database.DB.Where("name = ? AND type = ? AND status = ?", execution.Algorithm.Name, consts.ContainerTypeAlgorithm, true).First(&algorithm).Error; err != nil {
			dto.ErrorResponse(c, http.StatusNotFound, "Algorithm not found: "+execution.Algorithm.Name)
			return
		}

		// Create execution payload based on datapack or dataset
		var payload map[string]any
		var datapackID *int
		var datasetID *int

		if execution.Datapack != nil {
			// V1 compatible mode - use datapack (FaultInjectionSchedule)
			var datapack database.FaultInjectionSchedule
			if err := database.DB.Where("injection_name = ?", *execution.Datapack).First(&datapack).Error; err != nil {
				dto.ErrorResponse(c, http.StatusNotFound, "Datapack not found: "+*execution.Datapack)
				return
			}
			datapackID = &datapack.ID

			// Create v1 compatible payload
			payload = map[string]any{
				"algorithm": map[string]any{
					"name":  execution.Algorithm.Name,
					"image": algorithm.Image,
					"tag":   algorithm.Tag,
				},
				"dataset":  *execution.Datapack, // v1 uses "dataset" field for datapack name
				"env_vars": execution.EnvVars,
			}
		} else {
			// V2 mode - use dataset
			var dataset database.Dataset
			if err := database.DB.Where("name = ? AND status = ?", *execution.Dataset, 1).First(&dataset).Error; err != nil {
				dto.ErrorResponse(c, http.StatusNotFound, "Dataset not found: "+*execution.Dataset)
				return
			}
			datasetID = &dataset.ID

			// Create v2 payload with dataset support
			payload = map[string]any{
				"algorithm": map[string]any{
					"name":  execution.Algorithm.Name,
					"image": algorithm.Image,
					"tag":   algorithm.Tag,
				},
				"dataset_name": *execution.Dataset, // v2 uses "dataset_name" field
				"dataset_id":   dataset.ID,
				"env_vars":     execution.EnvVars,
			}
		}

		// Create unified task
		task := &dto.UnifiedTask{
			Type:      consts.TaskTypeRunAlgorithm,
			Payload:   payload,
			Immediate: true,
			GroupID:   groupID,
			ProjectID: &project.ID,
		}
		task.SetGroupCtx(spanCtx)

		// Submit task
		taskID, traceID, err := executor.SubmitTask(spanCtx, task)
		if err != nil {
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to submit algorithm execution task: "+err.Error())
			return
		}

		// Add to executions list
		executions = append(executions, dto.AlgorithmExecutionResponse{
			TraceID:     traceID,
			TaskID:      taskID,
			AlgorithmID: algorithm.ID,
			DatapackID:  datapackID,
			DatasetID:   datasetID,
			Status:      "submitted",
		})
	}

	// Build response
	response := dto.BatchAlgorithmExecutionResponse{
		GroupID:    groupID,
		Executions: executions,
		Message:    fmt.Sprintf("Successfully submitted %d algorithm executions", len(executions)),
	}

	dto.JSONResponse(c, http.StatusAccepted, "Batch algorithm execution submitted successfully", response)
}

// Helper function to parse integer parameters
func parseIntParam(s string) (int, error) {
	var result int
	for _, char := range s {
		if char < '0' || char > '9' {
			return 0, fmt.Errorf("invalid integer")
		}
		result = result*10 + int(char-'0')
	}
	return result, nil
}
