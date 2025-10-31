package v2

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"aegis/database"
	"aegis/dto"
	"aegis/repository"

	"github.com/gin-gonic/gin"
)

// GetProject gets a single project
//
//	@Summary Get project by ID
//	@Description Get detailed information about a specific project
//	@Tags Projects
//	@Produce json
//	@Security BearerAuth
//	@Param id path int true "Project ID"
//	@Param include_containers query bool false "Include related containers"
//	@Param include_datasets query bool false "Include related datasets"
//	@Param include_injections query bool false "Include related fault injections"
//	@Param include_labels query bool false "Include related labels"
//	@Success 200 {object} dto.GenericResponse[dto.ProjectV2Response] "Project retrieved successfully"
//	@Failure 400 {object} dto.GenericResponse[any] "Invalid project ID"
//	@Failure 403 {object} dto.GenericResponse[any] "Permission denied"
//	@Failure 404 {object} dto.GenericResponse[any] "Project not found"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /api/v2/projects/{id} [get]
func GetProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	var req dto.ProjectV2GetReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters: "+err.Error())
		return
	}

	// Get project using repository function which excludes deleted projects
	project, err := repository.GetProjectByID(database.DB, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			dto.ErrorResponse(c, http.StatusNotFound, "Project not found")
		} else {
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get project: "+err.Error())
		}
		return
	}

	response := dto.ToProjectV2Response(project, false)

	// Load containers if requested
	if req.IncludeContainers {
		relationMap, err := repository.GetProjectContainersMap([]int{project.ID})
		if err != nil {
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to load containers: "+err.Error())
			return
		}

		containers, ok := relationMap[project.ID]
		if !ok {
			dto.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("No container found for project %d", project.ID))
			return
		}

		response.Containers = containers
	}

	// Load datasets if requested
	if req.IncludeDatasets {
		relationMap, err := repository.GetProjectDatasetsMap([]int{project.ID})
		if err != nil {
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to load datasets: "+err.Error())
			return
		}

		datasets, ok := relationMap[project.ID]
		if !ok {
			dto.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("No dataset found for project %d", project.ID))
			return
		}

		response.Datasets = datasets
	}

	// Load injections if requested
	if req.IncludeInjections {
		relationMap, err := repository.GetProjectInjetionsMap([]int{project.ID})
		if err != nil {
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to load injections: "+err.Error())
			return
		}

		injections, ok := relationMap[project.ID]
		if !ok {
			dto.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("No injection found for project %d", project.ID))
			return
		}

		items, err := toInjectionV2ResponsesWithLabels(injections, false)
		if err != nil {
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to convert injections: "+err.Error())
			return
		}

		response.Injections = items
	}

	// Load labels if requested
	if req.IncludeLabels {
		labelMap, err := repository.GetProjectLabelsMap([]int{project.ID})
		if err != nil {
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to load labels: "+err.Error())
			return
		}

		labels, ok := labelMap[project.ID]
		if !ok {
			dto.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("No labels found for project %d", project.ID))
			return
		}

		response.Labels = labels
	}

	dto.SuccessResponse(c, response)
}
