package evaluationmodule

import (
	"aegis/httpx"
	"net/http"

	"aegis/consts"
	"aegis/dto"
	"aegis/middleware"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service HandlerService
}

func NewHandler(service HandlerService) *Handler {
	return &Handler{service: service}
}

// ListDatapackEvaluationResults retrieves evaluation data for multiple algorithm-datapack pairs
//
//	@Summary		List Datapack Evaluation Results
//	@Description	Retrieve evaluation data for multiple algorithm-datapack pairs.
//	@Tags			Evaluations
//	@ID				evaluate_algorithm_on_datapacks
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		BatchEvaluateDatapackReq							true	"Batch evaluation request containing multiple algorithm-datapack pairs"
//	@Success		200		{object}	dto.GenericResponse[BatchEvaluateDatapackResp]		"Batch algorithm datapack evaluation data retrieved successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]							"Invalid request format/parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]							"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]							"Permission denied"
//	@Failure		500		{object}	dto.GenericResponse[any]							"Internal server error"
//	@Router			/api/v2/evaluations/datapacks [post]
//	@x-api-type		{}
func (h *Handler) ListDatapackEvaluationResults(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists || userID <= 0 {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req BatchEvaluateDatapackReq
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := h.service.ListDatapackEvaluationResults(c.Request.Context(), &req, userID)
	if httpx.HandleServiceError(c, err) {
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
//	@Param			request	body		BatchEvaluateDatasetReq								true	"Batch evaluation request containing multiple algorithm-dataset pairs"
//	@Success		200		{object}	dto.GenericResponse[BatchEvaluateDatasetResp]		"Batch algorithm dataset evaluation data retrieved successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]							"Invalid request format/parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]							"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]							"Permission denied"
//	@Failure		500		{object}	dto.GenericResponse[any]							"Internal server error"
//	@Router			/api/v2/evaluations/datasets [post]
//	@x-api-type		{}
func (h *Handler) ListDatasetEvaluationResults(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists || userID <= 0 {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req BatchEvaluateDatasetReq
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := h.service.ListDatasetEvaluationResults(c.Request.Context(), &req, userID)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse(c, http.StatusOK, "Batch algorithm dataset evaluation data retrieved successfully", resp)
}

// ListEvaluations handles listing persisted evaluations with pagination
//
//	@Summary		List evaluations
//	@Description	Get a paginated list of persisted evaluation results
//	@Tags			Evaluations
//	@ID				list_evaluations
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page	query		int															false	"Page number"	default(1)
//	@Param			size	query		int															false	"Page size"		default(20)
//	@Success		200		{object}	dto.GenericResponse[dto.ListResp[EvaluationResp]]		"Evaluations retrieved successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]									"Invalid request format or parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]									"Authentication required"
//	@Failure		500		{object}	dto.GenericResponse[any]									"Internal server error"
//	@Router			/api/v2/evaluations [get]
//	@x-api-type		{}
func (h *Handler) ListEvaluations(c *gin.Context) {
	var req ListEvaluationReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := h.service.ListEvaluations(c.Request.Context(), &req)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// GetEvaluation handles getting a single evaluation by ID
//
//	@Summary		Get evaluation by ID
//	@Description	Get detailed information about a specific evaluation
//	@Tags			Evaluations
//	@ID				get_evaluation_by_id
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int										true	"Evaluation ID"
//	@Success		200	{object}	dto.GenericResponse[EvaluationResp]	"Evaluation retrieved successfully"
//	@Failure		400	{object}	dto.GenericResponse[any]				"Invalid evaluation ID"
//	@Failure		404	{object}	dto.GenericResponse[any]				"Evaluation not found"
//	@Failure		500	{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/evaluations/{id} [get]
//	@x-api-type		{}
func (h *Handler) GetEvaluation(c *gin.Context) {
	id, ok := httpx.ParsePositiveID(c, c.Param(consts.URLPathID), "evaluation ID")
	if !ok {
		return
	}

	resp, err := h.service.GetEvaluation(c.Request.Context(), id)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// DeleteEvaluation handles deleting a single evaluation by ID
//
//	@Summary		Delete evaluation by ID
//	@Description	Soft-delete an evaluation by its ID
//	@Tags			Evaluations
//	@ID				delete_evaluation_by_id
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int							true	"Evaluation ID"
//	@Success		200	{object}	dto.GenericResponse[any]	"Evaluation deleted successfully"
//	@Failure		400	{object}	dto.GenericResponse[any]	"Invalid evaluation ID"
//	@Failure		404	{object}	dto.GenericResponse[any]	"Evaluation not found"
//	@Failure		500	{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/evaluations/{id} [delete]
//	@x-api-type		{}
func (h *Handler) DeleteEvaluation(c *gin.Context) {
	id, ok := httpx.ParsePositiveID(c, c.Param(consts.URLPathID), "evaluation ID")
	if !ok {
		return
	}

	if httpx.HandleServiceError(c, h.service.DeleteEvaluation(c.Request.Context(), id)) {
		return
	}

	dto.JSONResponse[any](c, http.StatusOK, "Evaluation deleted successfully", nil)
}
