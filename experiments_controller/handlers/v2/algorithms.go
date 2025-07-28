package v2

import (
	"fmt"
	"net/http"

	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/database"
	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/repository"
	"github.com/gin-gonic/gin"
)

// SearchAlgorithms handles complex algorithm search with advanced filtering
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
