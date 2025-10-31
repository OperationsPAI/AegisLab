package v2

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"aegis/config"
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/executor"
	"aegis/middleware"
	"aegis/repository"
	"aegis/service"

	"github.com/gin-gonic/gin"
)

// BuildContainer handles container building with source code
//
//	@Summary Build container
//	@Description Build a container from provided source code (e.g., GitHub repository).
//	@Tags Containers
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param request body dto.BuildContainerRequest true "Container build request"
//	@Success 200 {object} dto.GenericResponse[dto.SubmitResp] "Container build task submitted successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 404 {object} dto.GenericResponse[any] "Required files not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/containers/build [post]
func BuildContainer(c *gin.Context) {
	var req dto.BuildContainerRequest
	if err := c.ShouldBind(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	sourcePath, err := service.ProcessGitHubSource(&req)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to process build source: "+err.Error())
		return
	}

	if err := req.ValidateInfoContent(sourcePath); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := req.Options.ValidateRequiredFiles(sourcePath); err != nil {
		dto.ErrorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	ctx, ok := c.Get(middleware.SpanContextKey)
	if !ok {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get span context")
		return
	}
	spanCtx := ctx.(context.Context)

	imageRef := fmt.Sprintf("%s/%s/%s:%s", config.GetString("harbor.registry"), config.GetString("harbor.namespace"), req.ImageName, req.Tag)
	task := &dto.UnifiedTask{
		Type: consts.TaskTypeBuildContainer,
		Payload: map[string]any{
			consts.BuildImageRef:     imageRef,
			consts.BuildSourcePath:   sourcePath,
			consts.BuildBuildOptions: req.Options,
		},
		Immediate: true,
	}
	task.SetGroupCtx(spanCtx)

	taskID, traceID, err := executor.SubmitTask(spanCtx, task)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to submit container building task: %s", err.Error()))
		return
	}

	dto.JSONResponse(c, http.StatusOK,
		"Container building task submitted successfully",
		dto.SubmitResp{Traces: []dto.Trace{{TraceID: traceID, HeadTaskID: taskID, Index: 0}}},
	)
}

// CreateContainer handles container creation for v2 API
//
//	@Summary Create container
//	@Description Create a new container without build configuration.
//	@Tags Containers
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param request body dto.CreateContainerRequest true "Container creation request"
//	@Success 201 {object} dto.GenericResponse[dto.ContainerResponse] "Container created successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 401 {object} dto.GenericResponse[any] "Authentication required"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 409 {object} dto.GenericResponse[any] "Conflict error"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/containers [post]
func CreateContainer(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req dto.CreateContainerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := service.CreateContainer(&req, userID)
	if !handleServiceError(c, err) {
		dto.JSONResponse[any](c, http.StatusCreated, "Container created successfully", resp)
	}
}

// CreateContainerVersion handles container version creation for v2 API
//
//	@Summary Create container version
//	@Description Create a new container version for an existing container.
//	@Tags Containers
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param container_id path int true "Container ID"
//	@Param request body dto.CreateContainerVersionRequest true "Container version creation request"
//	@Success 201 {object} dto.GenericResponse[dto.ContainerVersionResponse] "Container version created successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 401 {object} dto.GenericResponse[any] "Authentication required"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 409 {object} dto.GenericResponse[any] "Conflict error"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/containers/{container_id}/versions [post]
func CreateContainerVersion(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	containerIDStr := c.Param("container_id")
	containerID, err := strconv.Atoi(containerIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid container ID")
		return
	}

	var req dto.CreateContainerVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := service.CreateContainerVersion(&req, containerID, userID)
	if !handleServiceError(c, err) {
		dto.JSONResponse[any](c, http.StatusCreated, "Container version created successfully", resp)
	}
}

// DeleteContainer handles container deletion
//
//	@Summary Delete container
//	@Description Delete a container (soft delete by setting status to false)
//	@Tags Containers
//	@Produce json
//	@Security BearerAuth
//	@Param container_id path int true "Container ID"
//	@Success 204 {object} dto.GenericResponse[any] "Container deleted successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid container ID"
//	@Failure 401 {object} dto.GenericResponse[any] "Authentication required"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 404 {object} dto.GenericResponse[any] "Container not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/containers/{container_id} [delete]
func DeleteContainer(c *gin.Context) {
	containerIDStr := c.Param("container_id")
	containerID, err := strconv.Atoi(containerIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid container ID")
		return
	}

	err = service.DeleteContainer(containerID)
	if !handleServiceError(c, err) {
		dto.JSONResponse[any](c, http.StatusCreated, "Container deleted successfully", nil)
	}
}

// DeleteContainerVersion handles container version deletion
//
//	@Summary Delete container version
//	@Description Delete a container version (soft delete by setting status to false)
//	@Tags Containers
//	@Produce json
//	@Security BearerAuth
//	@Param container_id path int true "Container ID"
//	@Param version_id path int true "Container Version ID"
//	@Success 204 {object} dto.GenericResponse[any] "Container version deleted successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid container ID/container version ID"
//	@Failure 401 {object} dto.GenericResponse[any] "Authentication required"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 404 {object} dto.GenericResponse[any] "Container or version not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/containers/{container_id}/versions/{version_id} [delete]
func DeleteContainerVersion(c *gin.Context) {
	containerIDStr := c.Param("container_id")
	containerID, err := strconv.Atoi(containerIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid container ID")
		return
	}

	versionIDStr := c.Param("version_id")
	versionID, err := strconv.Atoi(versionIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid container version ID")
		return
	}

	err = service.DeleteContainerVersion(containerID, versionID)
	if !handleServiceError(c, err) {
		dto.JSONResponse[any](c, http.StatusNoContent, "Container version deleted successfully", nil)
	}
}

// GetContainer handles getting a single container by ID
//
//	@Summary Get container by ID
//	@Description Get detailed information about a specific container
//	@Tags Containers
//	@Produce json
//	@Security BearerAuth
//	@Param container_id path int true "Container ID"
//	@Success 200 {object} dto.GenericResponse[dto.ContainerDetailResponse] "Container retrieved successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid container ID"
//	@Failure 401 {object} dto.GenericResponse[any] "Authentication required"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 404 {object} dto.GenericResponse[any] "Container not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/containers/{container_id} [get]
func GetContainer(c *gin.Context) {
	containerIDStr := c.Param("container_id")
	containerID, err := strconv.Atoi(containerIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid container ID")
		return
	}

	resp, err := service.GetContainerDetail(containerID)
	if err != nil {
		switch err {
		case consts.ErrNotFound:
			dto.ErrorResponse(c, http.StatusNotFound, "Container not found")
		default:
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get container: "+err.Error())
		}
		return
	}

	dto.SuccessResponse(c, resp)
}

// GetContainerVersion handles getting a single container version by ID
//
//	@Summary Get container version by ID
//	@Description Get detailed information about a specific container version
//	@Tags Containers
//	@Produce json
//	@Security BearerAuth
//	@Param container_id path int true "Container ID"
//	@Param version_id path int true "Container Version ID"
//	@Success 200 {object} dto.GenericResponse[dto.ContainerVersionDetailResponse] "Container version retrieved successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid container ID/container version ID"
//	@Failure 401 {object} dto.GenericResponse[any] "Authentication required"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 404 {object} dto.GenericResponse[any] "Container or version not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/containers/{container_id}/versions/{version_id} [get]
func GetContainerVersion(c *gin.Context) {
	containerIDStr := c.Param("container_id")
	containerID, err := strconv.Atoi(containerIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid container ID")
		return
	}

	versionIDStr := c.Param("version_id")
	versionID, err := strconv.Atoi(versionIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid container version ID")
		return
	}

	resp, err := service.GetContainerVersionDetail(containerID, versionID)
	if err != nil {
		switch err {
		case consts.ErrNotFound:
			dto.ErrorResponse(c, http.StatusNotFound, "Container or version not found")
		default:
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get container version: "+err.Error())
		}
		return
	}

	dto.SuccessResponse(c, resp)
}

// ListContainers handles listing containers with pagination and filtering
//
//	@Summary List containers
//	@Description Get paginated list of containers with optional filtering
//	@Tags Containers
//	@Produce json
//	@Security BearerAuth
//	@Param page query int false "Page number" default(1)
//	@Param size query int false "Page size" default(20)
//	@Param type query string false "Container type filter" Enums(algorithm,benchmark)
//	@Param status query bool false "Container status filter"
//	@Success 200 {object} dto.GenericResponse[dto.ListResponse[dto.ContainerResponse]] "Containers retrieved successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request format or parameters"
//	@Failure 401 {object} dto.GenericResponse[any] "Authentication required"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/containers [get]
func ListContainers(c *gin.Context) {
	var req dto.ListContainerRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := service.ListContainers(&req)
	if handleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// SearchContainers handles complex container search with advanced filtering
//
//	@Summary Search containers
//	@Description Search containers with complex filtering, sorting and pagination. Supports all container types (algorithm, benchmark, etc.)
//	@Tags Containers
//	@Produce json
//	@Security BearerAuth
//	@Param request body dto.SearchContainerRequest true "Container search request"
//	@Success 200 {object} dto.GenericResponse[dto.SearchResponse[dto.ContainerResponse]] "Containers retrieved successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/containers/search [post]
func SearchContainers(c *gin.Context) {
	var req dto.SearchContainerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	// Convert to SearchRequest
	searchReq := req.ConvertToSearchRequest()

	// Validate search request
	if err := searchReq.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid search parameters: "+err.Error())
		return
	}

	// Execute search using query builder
	searchResult, err := repository.ExecuteSearch(database.DB, searchReq, database.Container{})
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to search containers: "+err.Error())
		return
	}

	// Convert database containers to response DTOs
	var containerResponses []dto.ContainerResponse
	for _, container := range searchResult.Items {
		var containerResponse dto.ContainerResponse
		containerResponse.ConvertFromContainer(&container)
		containerResponses = append(containerResponses, containerResponse)
	}

	// Build final response
	response := dto.SearchResponse[dto.ContainerResponse]{
		Items:      containerResponses,
		Pagination: searchResult.Pagination,
		Filters:    searchResult.Filters,
		Sort:       searchResult.Sort,
	}

	dto.SuccessResponse(c, response)
}

// UpdateContainer handles container updates
//
//	@Summary Update container
//	@Description Update an existing container's information
//	@Tags Containers
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param container_id path int true "Container ID"
//	@Param request body dto.UpdateContainerRequest true "Container update request"
//	@Success 202 {object} dto.GenericResponse[dto.ContainerResponse] "Container updated successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid container ID/request"
//	@Failure 401 {object} dto.GenericResponse[any] "Authentication required"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 404 {object} dto.GenericResponse[any] "Container not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/containers/{container_id} [patch]
func UpdateContainer(c *gin.Context) {
	containerIDStr := c.Param("container_id")
	containerID, err := strconv.Atoi(containerIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid container ID")
		return
	}

	var req dto.UpdateContainerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	resp, err := service.UpdateContainer(&req, containerID)
	if err != nil {
		switch err {
		case consts.ErrNotFound:
			dto.ErrorResponse(c, http.StatusNotFound, "Container not found")
		default:
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to update container: "+err.Error())
		}
		return
	}

	dto.JSONResponse[any](c, http.StatusAccepted, "Container updated successfully", resp)
}

// UpdateContainerVersion handles container version updates
//
//	@Summary Update container version
//	@Description Update an existing container version's information
//	@Tags Containers
//	@Accept json
//	@Produce json
//	@Security BearerAuth
//	@Param container_id path int true "Container ID"
//	@Param version_id path int true "Container Version ID"
//	@Param request body dto.UpdateContainerVersionRequest true "Container version update request"
//	@Success 202 {object} dto.GenericResponse[dto.ContainerVersionResponse] "Container version updated successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid container ID/container version ID/request"
//	@Failure 401 {object} dto.GenericResponse[any] "Authentication required"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 404 {object} dto.GenericResponse[any] "Container not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/containers/{container_id}/versions/{version_id} [patch]
func UpdateContainerVersion(c *gin.Context) {
	containerIDStr := c.Param("container_id")
	containerID, err := strconv.Atoi(containerIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid container ID")
		return
	}

	versionIDStr := c.Param("version_id")
	versionID, err := strconv.Atoi(versionIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid container version ID")
		return
	}

	var req dto.UpdateContainerVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	resp, err := service.UpdateContainerVersion(&req, containerID, versionID)
	if err != nil {
		switch err {
		case consts.ErrNotFound:
			dto.ErrorResponse(c, http.StatusNotFound, "Model not found")
		default:
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to update container: "+err.Error())
		}
		return
	}

	dto.JSONResponse[any](c, http.StatusAccepted, "Container version updated successfully", resp)
}
