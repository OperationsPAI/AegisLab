package v2

import (
	"net/http"
	"strconv"

	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/database"
	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/repository"
	"github.com/gin-gonic/gin"
)

// SearchContainers handles complex container search with advanced filtering
//	@Summary Search containers
//	@Description Search containers with complex filtering, sorting and pagination. Supports all container types (algorithm, benchmark, etc.)
//	@Tags Containers
//	@Produce json
//	@Security BearerAuth
//	@Param request body dto.ContainerSearchRequest true "Container search request"
//	@Success 200 {object} dto.GenericResponse[dto.SearchResponse[dto.ContainerResponse]] "Containers retrieved successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/containers/search [post]
func SearchContainers(c *gin.Context) {
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
		dto.ErrorResponse(c, http.StatusForbidden, "No permission to read containers")
		return
	}

	var req dto.ContainerSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	// Convert to SearchRequest
	searchReq := req.ConvertToSearchRequest()

	// Validate search request
	if err := searchReq.ValidateSearchRequest(); err != nil {
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
		containerResponse := dto.ContainerResponse{
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

// ListContainers handles simple container listing
//	@Summary List containers
//	@Description Get a simple list of containers with basic filtering
//	@Tags Containers
//	@Produce json
//	@Security BearerAuth
//	@Param page query int false "Page number" default(1)
//	@Param size query int false "Page size" default(20)
//	@Param type query string false "Container type filter" Enums(algorithm,benchmark)
//	@Param status query bool false "Container status filter"
//	@Success 200 {object} dto.GenericResponse[dto.SearchResponse[dto.ContainerResponse]] "Containers retrieved successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid request"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/containers [get]
func ListContainers(c *gin.Context) {
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
		dto.ErrorResponse(c, http.StatusForbidden, "No permission to read containers")
		return
	}

	// Create a basic search request from query parameters
	req := dto.ContainerSearchRequest{
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

	// Parse filters from query parameters
	if containerType := c.Query("type"); containerType != "" {
		req.Type = &containerType
	}
	if statusStr := c.Query("status"); statusStr != "" {
		if status, err := strconv.ParseBool(statusStr); err == nil {
			req.IsActive = &status
		}
	}

	// Convert to SearchRequest
	searchReq := req.ConvertToSearchRequest()

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
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get container list: "+err.Error())
		return
	}

	// Convert database containers to response DTOs
	var containerResponses []dto.ContainerResponse
	for _, container := range searchResult.Items {
		containerResponse := dto.ContainerResponse{
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

// GetContainer handles getting a single container by ID
//	@Summary Get container by ID
//	@Description Get detailed information about a specific container
//	@Tags Containers
//	@Produce json
//	@Security BearerAuth
//	@Param id path int true "Container ID"
//	@Success 200 {object} dto.GenericResponse[dto.ContainerResponse] "Container retrieved successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid container ID"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 404 {object} dto.GenericResponse[any] "Container not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/containers/{id} [get]
func GetContainer(c *gin.Context) {
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
		dto.ErrorResponse(c, http.StatusForbidden, "No permission to read containers")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid container ID")
		return
	}

	var container database.Container
	if err := database.DB.First(&container, id).Error; err != nil {
		if err.Error() == "record not found" {
			dto.ErrorResponse(c, http.StatusNotFound, "Container not found")
		} else {
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get container: "+err.Error())
		}
		return
	}

	response := dto.ContainerResponse{
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

	dto.SuccessResponse(c, response)
}
