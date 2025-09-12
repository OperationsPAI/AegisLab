package v2

import (
	"fmt"
	"net/http"

	"rcabench/dto"
	"rcabench/repository"
	"github.com/gin-gonic/gin"
)

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

// GetDatasetEvaluationResults retrieves evaluation data for multiple algorithm-dataset pairs
// @Summary Get Batch Algorithm Dataset Evaluation
// @Description Get execution results with predictions and ground truth for multiple algorithm-dataset pairs in a single request. Returns the latest execution for each datapack if multiple executions exist.
// @Tags evaluation
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.DatasetEvaluationBatchReq true "Batch evaluation request containing multiple algorithm-dataset pairs"
// @Success 200 {object} dto.GenericResponse[dto.DatasetEvaluationBatchResp] "Batch algorithm dataset evaluation data"
// @Failure 400 {object} dto.GenericResponse[any] "Bad request"
// @Failure 401 {object} dto.GenericResponse[any] "Unauthorized"
// @Failure 403 {object} dto.GenericResponse[any] "Forbidden"
// @Failure 500 {object} dto.GenericResponse[any] "Internal server error"
// @Router /api/v2/evaluations/datasets [post]
func GetDatasetEvaluationResults(c *gin.Context) {
	var req dto.DatasetEvaluationBatchReq
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	results, err := repository.GetDatasetEvaluationBatch(req)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve evaluation data")
		return
	}

	dto.SuccessResponse(c, results)
}

// GetDatapackEvaluationResults retrieves evaluation data for multiple algorithm-datapack pairs
// @Summary Get Batch Algorithm Datapack Evaluation
// @Description Get execution results with predictions and ground truth for multiple algorithm-datapack pairs in a single request. Returns the latest execution for each pair if multiple executions exist.
// @Tags evaluation
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.DatapackEvaluationBatchReq true "Batch evaluation request containing multiple algorithm-datapack pairs"
// @Success 200 {object} dto.GenericResponse[dto.DatapackEvaluationBatchResp] "Batch algorithm datapack evaluation data"
// @Failure 400 {object} dto.GenericResponse[any] "Bad request"
// @Failure 401 {object} dto.GenericResponse[any] "Unauthorized"
// @Failure 403 {object} dto.GenericResponse[any] "Forbidden"
// @Failure 500 {object} dto.GenericResponse[any] "Internal server error"
// @Router /api/v2/evaluations/datapacks [post]
func GetDatapackEvaluationResults(c *gin.Context) {
	var req dto.DatapackEvaluationBatchReq
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Get evaluation data from repository
	results, err := repository.GetDatapackEvaluationBatch(req)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve evaluation data")
		return
	}

	dto.SuccessResponse(c, results)
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
