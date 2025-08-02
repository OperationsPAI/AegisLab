package v2

import (
	"context"
	"fmt"
	"net/http"

	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/database"
	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/executor"
	"github.com/LGU-SE-Internal/rcabench/middleware"
	"github.com/LGU-SE-Internal/rcabench/repository"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
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

		// Create new execution record using repository function
		executionID, err = repository.CreateExecutionResult("", algorithmID, req.DatapackID, nil)
		if err != nil {
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to create execution record: "+err.Error())
			return
		}

		isNewExecution = true
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
//	@Router /api/v2/algorithms/execute [post]
func SubmitAlgorithmExecution(c *gin.Context) {
	groupID := c.GetString("groupID")
	logrus.Infof("SubmitAlgorithmExecution called, groupID: %s", groupID)

	ctx, ok := c.Get(middleware.SpanContextKey)
	if !ok {
		logrus.Error("failed to get span context from gin.Context")
		dto.ErrorResponse(c, http.StatusInternalServerError, "failed to get span context")
		return
	}

	spanCtx := ctx.(context.Context)
	trace.SpanFromContext(spanCtx)



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

	var allExecutions []dto.AlgorithmExecutionResponse

	// Process each execution request
	for _, execution := range req.Executions {
		// Get algorithm container
		var algorithm database.Container
		if err := database.DB.Where("name = ? AND type = ? AND status = ?", execution.Algorithm.Name, consts.ContainerTypeAlgorithm, true).First(&algorithm).Error; err != nil {
			dto.ErrorResponse(c, http.StatusNotFound, "Algorithm not found: "+execution.Algorithm.Name)
			return
		}

		// Extract datapacks from request (either single datapack or dataset)
		datapacks, datasetID, err := extractDatapacks(execution.Datapack, execution.Dataset, execution.DatasetVersion)
		if err != nil {
			dto.ErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		// Submit tasks for all datapacks
		executions, err := submitAlgorithmTasks(spanCtx, groupID, &project.ID, &algorithm, execution.EnvVars, datapacks, datasetID, req.Labels)
		if err != nil {
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to submit algorithm execution tasks: "+err.Error())
			return
		}

		allExecutions = append(allExecutions, executions...)
	}

	// Build response
	response := dto.BatchAlgorithmExecutionResponse{
		GroupID:    groupID,
		Executions: allExecutions,
		Message:    fmt.Sprintf("Successfully submitted %d algorithm executions", len(allExecutions)),
	}

	dto.JSONResponse(c, http.StatusAccepted, "Batch algorithm execution submitted successfully", response)
}

// extractDatapacks extracts datapacks from either a single datapack name or dataset name
// Returns datapacks and optional dataset ID (if from dataset)
func extractDatapacks(datapackName *string, datasetName *string, datasetVersion *string) ([]database.FaultInjectionSchedule, *int, error) {
	if datapackName != nil {
		// Single datapack mode
		var datapack database.FaultInjectionSchedule
		if err := database.DB.Where("injection_name = ?", *datapackName).First(&datapack).Error; err != nil {
			return nil, nil, fmt.Errorf("datapack not found: %s", *datapackName)
		}
		return []database.FaultInjectionSchedule{datapack}, nil, nil
	} else if datasetName != nil {
		// Dataset mode - get all datapacks in the dataset
		var dataset database.Dataset

		// Use name and version to uniquely identify dataset
		if datasetVersion == nil || *datasetVersion == "" {
			return nil, nil, fmt.Errorf("dataset_version is required when querying by dataset name")
		}

		if err := database.DB.Where("name = ? AND version = ? AND status = ?", *datasetName, *datasetVersion, consts.DatasetEnabled).First(&dataset).Error; err != nil {
			return nil, nil, fmt.Errorf("dataset not found: %s:%s", *datasetName, *datasetVersion)
		}

		datasetFaultInjections, err := repository.GetDatasetFaultInjections(dataset.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get dataset datapacks: %s", err.Error())
		}

		if len(datasetFaultInjections) == 0 {
			return nil, nil, fmt.Errorf("dataset contains no datapacks")
		}

		// Extract datapacks from relations
		var datapacks []database.FaultInjectionSchedule
		for _, relation := range datasetFaultInjections {
			if relation.FaultInjectionSchedule != nil {
				datapacks = append(datapacks, *relation.FaultInjectionSchedule)
			}
		}
		return datapacks, &dataset.ID, nil
	}

	return nil, nil, fmt.Errorf("either datapack or dataset must be specified")
}

// submitAlgorithmTasks submits algorithm execution tasks for the given datapacks
func submitAlgorithmTasks(ctx context.Context, groupID string, projectID *int, algorithm *database.Container, envVars map[string]string, datapacks []database.FaultInjectionSchedule, datasetID *int, labels *dto.ExecutionLabels) ([]dto.AlgorithmExecutionResponse, error) {
	var executions []dto.AlgorithmExecutionResponse

	for _, datapack := range datapacks {
		// Create v1 compatible payload
		payload := map[string]any{
			"algorithm": map[string]any{
				"name":  algorithm.Name,
				"image": algorithm.Image,
				"tag":   algorithm.Tag,
			},
			"dataset":  datapack.InjectionName, // Use datapack name as dataset field
			"env_vars": envVars,
		}

		// Add labels to payload if provided
		if labels != nil {
			payload["labels"] = labels
		}

		// Create unified task
		task := &dto.UnifiedTask{
			Type:      consts.TaskTypeRunAlgorithm,
			Payload:   payload,
			Immediate: true,
			GroupID:   groupID,
			ProjectID: projectID,
		}
		task.SetGroupCtx(ctx)

		// Submit task
		taskID, traceID, err := executor.SubmitTask(ctx, task)
		if err != nil {
			return nil, fmt.Errorf("failed to submit task: %s", err.Error())
		}

		// Build execution response
		execution := dto.AlgorithmExecutionResponse{
			TraceID:     traceID,
			TaskID:      taskID,
			AlgorithmID: algorithm.ID,
			DatapackID:  &datapack.ID,
			Status:      "submitted",
		}

		// Set DatasetID if this is from a dataset
		if datasetID != nil {
			execution.DatasetID = datasetID
		}

		executions = append(executions, execution)
	}

	return executions, nil
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
