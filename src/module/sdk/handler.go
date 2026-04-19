package sdk

import (
	"aegis/httpx"
	"net/http"

	"aegis/consts"
	"aegis/dto"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// ListSDKEvaluations handles listing SDK evaluation samples with pagination
//
//	@Summary		List SDK evaluation samples
//	@Description	Get a paginated list of SDK evaluation samples, optionally filtered by exp_id and stage
//	@Tags			Evaluations
//	@ID				list_sdk_evaluations
//	@Produce		json
//	@Security		BearerAuth
//	@Param			exp_id	query		string													false	"Experiment ID filter"
//	@Param			stage	query		string													false	"Stage filter (init, rollout, judged)"
//	@Param			page	query		int														false	"Page number"	default(1)
//	@Param			size	query		int														false	"Page size"		default(20)
//	@Success		200		{object}	dto.GenericResponse[dto.ListResp[SDKEvaluationSample]]	"SDK evaluations retrieved successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]								"Invalid request format or parameters"
//	@Failure		500		{object}	dto.GenericResponse[any]								"Internal server error"
//	@Router			/api/v2/sdk/evaluations [get]
//	@x-api-type		{"sdk":"true"}
func (h *Handler) ListEvaluations(c *gin.Context) {
	var req ListSDKEvaluationReq
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

// GetSDKEvaluation handles getting a single SDK evaluation sample by ID
//
//	@Summary		Get SDK evaluation sample by ID
//	@Description	Get detailed information about a specific SDK evaluation sample
//	@Tags			Evaluations
//	@ID				get_sdk_evaluation
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int											true	"SDK Evaluation Sample ID"
//	@Success		200	{object}	dto.GenericResponse[SDKEvaluationSample]	"SDK evaluation sample retrieved successfully"
//	@Failure		400	{object}	dto.GenericResponse[any]					"Invalid evaluation ID"
//	@Failure		404	{object}	dto.GenericResponse[any]					"SDK evaluation sample not found"
//	@Failure		500	{object}	dto.GenericResponse[any]					"Internal server error"
//	@Router			/api/v2/sdk/evaluations/{id} [get]
//	@x-api-type		{"sdk":"true"}
func (h *Handler) GetEvaluation(c *gin.Context) {
	id, ok := httpx.ParsePositiveID(c, c.Param(consts.URLPathID), "SDK evaluation ID")
	if !ok {
		return
	}
	resp, err := h.service.GetEvaluation(c.Request.Context(), id)
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.SuccessResponse(c, resp)
}

// ListSDKExperiments handles listing all distinct experiment IDs
//
//	@Summary		List SDK experiment IDs
//	@Description	Get all distinct experiment IDs from SDK evaluation data
//	@Tags			Evaluations
//	@ID				list_sdk_experiments
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	dto.GenericResponse[SDKExperimentListResp]	"SDK experiments retrieved successfully"
//	@Failure		500	{object}	dto.GenericResponse[any]					"Internal server error"
//	@Router			/api/v2/sdk/evaluations/experiments [get]
//	@x-api-type		{"sdk":"true"}
func (h *Handler) ListExperiments(c *gin.Context) {
	resp, err := h.service.ListExperiments(c.Request.Context())
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.SuccessResponse(c, resp)
}

// ListSDKDatasetSamples handles listing SDK dataset samples with pagination
//
//	@Summary		List SDK dataset samples
//	@Description	Get a paginated list of SDK dataset samples, optionally filtered by dataset name
//	@Tags			Datasets
//	@ID				list_sdk_dataset_samples
//	@Produce		json
//	@Security		BearerAuth
//	@Param			dataset	query		string												false	"Dataset name filter"
//	@Param			page	query		int													false	"Page number"	default(1)
//	@Param			size	query		int													false	"Page size"		default(20)
//	@Success		200		{object}	dto.GenericResponse[dto.ListResp[SDKDatasetSample]]	"SDK dataset samples retrieved successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]							"Invalid request format or parameters"
//	@Failure		500		{object}	dto.GenericResponse[any]							"Internal server error"
//	@Router			/api/v2/sdk/datasets [get]
//	@x-api-type		{"sdk":"true"}
func (h *Handler) ListDatasetSamples(c *gin.Context) {
	var req ListSDKDatasetSampleReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}
	resp, err := h.service.ListDatasetSamples(c.Request.Context(), &req)
	if httpx.HandleServiceError(c, err) {
		return
	}
	dto.SuccessResponse(c, resp)
}
