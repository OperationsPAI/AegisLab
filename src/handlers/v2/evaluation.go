package v2

import (
	"fmt"
	"net/http"

	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/repository"
	"github.com/gin-gonic/gin"
)

// GetAlgorithmDatasetEvaluation retrieves evaluation data for a specific algorithm on a specific dataset
// @Summary Get Algorithm Dataset Evaluation
// @Description Get all execution results with predictions and ground truth for a specific algorithm on a specific dataset
// @Tags evaluation
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param algorithm path string true "Algorithm name"
// @Param dataset path string true "Dataset name"
// @Param dataset_version query string false "Dataset version (optional, defaults to v1.0)"
// @Param tag query string false "Tag label filter"
// @Success 200 {object} dto.GenericResponse[dto.AlgorithmDatasetEvaluationResp] "Algorithm dataset evaluation data"
// @Failure 400 {object} dto.GenericResponse[any] "Bad request"
// @Failure 401 {object} dto.GenericResponse[any] "Unauthorized"
// @Failure 403 {object} dto.GenericResponse[any] "Forbidden"
// @Failure 404 {object} dto.GenericResponse[any] "Algorithm or dataset not found"
// @Failure 500 {object} dto.GenericResponse[any] "Internal server error"
// @Router /api/v2/evaluations/algorithms/{algorithm}/datasets/{dataset} [get]
func GetAlgorithmDatasetEvaluation(c *gin.Context) {
	algorithm := c.Param("algorithm")
	dataset := c.Param("dataset")

	// Build request with path parameters and query parameters
	req := dto.AlgorithmDatasetEvaluationReq{
		Algorithm: algorithm,
		Dataset:   dataset,
	}

	// Parse label filters and dataset version from query parameters
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	// Validate request
	if req.Algorithm == "" {
		dto.ErrorResponse(c, http.StatusBadRequest, "Algorithm parameter is required")
		return
	}

	if req.Dataset == "" {
		dto.ErrorResponse(c, http.StatusBadRequest, "Dataset parameter is required")
		return
	}

	// Get evaluation data from repository
	result, err := repository.GetAlgorithmDatasetEvaluation(req)
	if err != nil {
		// Check if it's a not found error
		datasetVersion := req.DatasetVersion
		if datasetVersion == "" {
			datasetVersion = "v1.0"
		}
		if err.Error() == "dataset '"+req.Dataset+":"+datasetVersion+"' not found" {
			dto.ErrorResponse(c, http.StatusNotFound, "Dataset not found")
			return
		}

		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve evaluation data")
		return
	}

	dto.SuccessResponse(c, result)
}

// GetAvailableLabelKeys returns the list of available label keys for execution filtering
// @Summary Get Available Label Keys
// @Description Get the list of available label keys that can be used for filtering execution results
// @Tags evaluation
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} dto.GenericResponse[[]string] "Available label keys"
// @Failure 401 {object} dto.GenericResponse[any] "Unauthorized"
// @Failure 403 {object} dto.GenericResponse[any] "Forbidden"
// @Failure 500 {object} dto.GenericResponse[any] "Internal server error"
// @Router /api/v2/evaluations/label-keys [get]
func GetAvailableLabelKeys(c *gin.Context) {
	labelKeys := dto.GetAvailableLabelKeys()
	dto.SuccessResponse(c, labelKeys)
}

// GetAlgorithmDatapackEvaluation retrieves evaluation data for a specific algorithm on a specific datapack
// @Summary Get Algorithm Datapack Evaluation
// @Description Get execution result with predictions and ground truth for a specific algorithm on a specific datapack
// @Tags evaluation
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param algorithm path string true "Algorithm name"
// @Param datapack path string true "Datapack name"
// @Param tag query string false "Tag label filter"
// @Success 200 {object} dto.GenericResponse[dto.AlgorithmDatapackEvaluationResp] "Algorithm datapack evaluation data"
// @Failure 400 {object} dto.GenericResponse[any] "Bad request"
// @Failure 401 {object} dto.GenericResponse[any] "Unauthorized"
// @Failure 403 {object} dto.GenericResponse[any] "Forbidden"
// @Failure 404 {object} dto.GenericResponse[any] "Algorithm or datapack not found"
// @Failure 500 {object} dto.GenericResponse[any] "Internal server error"
// @Router /api/v2/evaluations/algorithms/{algorithm}/datapacks/{datapack} [get]
func GetAlgorithmDatapackEvaluation(c *gin.Context) {
	algorithm := c.Param("algorithm")
	datapack := c.Param("datapack")

	// Build request with path parameters and query parameters
	req := dto.AlgorithmDatapackEvaluationReq{
		Algorithm: algorithm,
		Datapack:  datapack,
	}

	// Parse label filters from query parameters
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	// Validate request
	if req.Algorithm == "" {
		dto.ErrorResponse(c, http.StatusBadRequest, "Algorithm parameter is required")
		return
	}

	if req.Datapack == "" {
		dto.ErrorResponse(c, http.StatusBadRequest, "Datapack parameter is required")
		return
	}

	// Get evaluation data from repository
	result, err := repository.GetAlgorithmDatapackEvaluation(req)
	if err != nil {
		// Check if it's a not found error
		if err.Error() == "datapack '"+req.Datapack+"' not found" {
			dto.ErrorResponse(c, http.StatusNotFound, "Datapack not found")
			return
		}

		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve evaluation data")
		return
	}

	dto.SuccessResponse(c, result)
}

// GetDatapackDetectorResults retrieves detector results for multiple datapacks
// @Summary Get Datapack Detector Results
// @Description Get detector analysis results for multiple datapacks. If a datapack has multiple executions, returns the latest one.
// @Tags evaluation
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.DatapackDetectorReq true "Datapack detector request"
// @Success 200 {object} dto.GenericResponse[dto.DatapackDetectorResp] "Datapack detector results"
// @Failure 400 {object} dto.GenericResponse[any] "Bad request"
// @Failure 401 {object} dto.GenericResponse[any] "Unauthorized"
// @Failure 403 {object} dto.GenericResponse[any] "Forbidden"
// @Failure 500 {object} dto.GenericResponse[any] "Internal server error"
// @Router /api/v2/evaluations/datapacks/detector [post]
func GetDatapackDetectorResults(c *gin.Context) {
	var req dto.DatapackDetectorReq

	// Parse request body
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if len(req.Datapacks) == 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "At least one datapack is required")
		return
	}

	// Validate datapack names
	for i, datapack := range req.Datapacks {
		if datapack == "" {
			dto.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Datapack name cannot be empty at index %d", i))
			return
		}
	}

	// Get detector results from repository
	result, err := repository.GetDatapackDetectorResults(req)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve detector results")
		return
	}

	dto.SuccessResponse(c, result)
}
