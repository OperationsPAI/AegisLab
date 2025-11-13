package v2

import (
	"net/http"

	"aegis/dto"
	"aegis/handlers"
	"aegis/middleware"
	"aegis/service/analyzer"

	"github.com/gin-gonic/gin"
)

// ListDatapackEvaluationResults retrieves evaluation data for multiple algorithm-datapack pairs
//
//	@Summary		List Datapack Evaluation Results
//	@Description	Retrieve evaluation data for multiple algorithm-datapack pairs.
//	@Tags			Evaluations
//	@ID				evaluate_algorithm_on_datapacks
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.BatchEvaluateDatapackReq						true	"Batch evaluation request containing multiple algorithm-datapack pairs"
//	@Success		200		{object}	dto.GenericResponse[dto.BatchEvaluateDatapackResp]	"Batch algorithm datapack evaluation data retrieved successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]							"Invalid request format/parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]							"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]							"Permission denied"
//	@Failure		500		{object}	dto.GenericResponse[any]							"Internal server error"
//	@Router			/api/v2/evaluations/datapacks [post]
func ListDatapackEvaluationResults(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists || userID <= 0 {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req dto.BatchEvaluateDatapackReq
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := analyzer.ListDatapackEvaluationResults(&req, userID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse(c, http.StatusOK, "Batch algorithm datapack evaluation data retrieved successfully", resp)
}

// ListDatasetEvaluationResults retrieves evaluation data for multiple algorithm-dataset pairs
//
//	@Summary		List Dataset Evaluation Results
//	@Description	Retrieve evaluation data for multiple algorithm-dataset pairs.
//	@Tags			Evaluations
//	@ID				evaluate_algorithm_on_datasets
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.BatchEvaluateDatapackReq						true	"Batch evaluation request containing multiple algorithm-dataset pairs"
//	@Success		200		{object}	dto.GenericResponse[dto.BatchEvaluateDatasetResp]	"Batch algorithm dataset evaluation data retrieved successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]							"Invalid request format/parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]							"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]							"Permission denied"
//	@Failure		500		{object}	dto.GenericResponse[any]							"Internal server error"
//	@Router			/api/v2/evaluations/datasets [post]
func ListDatasetEvaluationResults(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists || userID <= 0 {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req dto.BatchEvaluateDatasetReq
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := analyzer.ListDatasetEvaluationResults(&req, userID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse(c, http.StatusOK, "Batch algorithm dataset evaluation data retrieved successfully", resp)
}
