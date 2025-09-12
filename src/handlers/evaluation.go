package handlers

import (
	"net/http"

	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ListEvaluationRawData Get raw evaluation data for algorithms and datasets
//
//	@Summary		Get raw evaluation data
//	@Description	Supports three query modes: 1) Directly pass an array of algorithm-dataset pairs for precise query; 2) Pass lists of algorithms and datasets for Cartesian product query; 3) Query by execution ID list. The three modes are mutually exclusive, only one can be selected
//	@Tags			evaluation
//	@Produce		json
//	@Consumes		application/json
//	@Param			body	body		dto.RawDataReq	true	"Raw data query request, supports three modes: pairs array, (algorithms+datasets) Cartesian product, or execution_ids list"
//	@Success		200		{object}	dto.GenericResponse[dto.RawDataResp]	"Successfully returns the list of raw evaluation data"
//	@Failure		400		{object}	dto.GenericResponse[any]				"Request parameter error, such as incorrect JSON format, query mode conflict or empty parameter"
//	@Failure		500		{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v1/evaluations/raw-data [post]
func ListEvaluationRawData(c *gin.Context) {
	var req dto.RawDataReq
	if err := c.BindJSON(&req); err != nil {
		logrus.Errorf("failed to bind JSON request: %v", err)
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters")
		return
	}

	if err := req.Validate(); err != nil {
		logrus.Errorf("validation error: %v", err)
		dto.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	var items []dto.RawDataItem
	var err error
	if req.HasPairsMode() {
		items, err = repository.ListExecutionRawDatasByPairs(req)
	} else if req.HasExecutionMode() {
		items, err = repository.ListExecutionRawDataByIds(req)
	}

	if err != nil {
		logrus.Errorf("failed to get raw execution data: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get raw results")
		return
	}

	dto.SuccessResponse(c, items)
}

// GetGroundtruth Get ground truth for datasets
//
//	@Summary		Get ground truth for datasets
//	@Description	Get ground truth data for the given dataset array, used as benchmark data for algorithm evaluation. Supports batch query for ground truth information of multiple datasets
//	@Tags			evaluation
//	@Produce		json
//	@Consumes		application/json
//	@Param			body	body		dto.GroundTruthReq	true	"Ground truth query request, contains dataset list"
//	@Success		200		{object}	dto.GenericResponse[dto.GroundTruthResp]	"Successfully returns ground truth information for datasets"
//	@Failure		400		{object}	dto.GenericResponse[any]					"Request parameter error, such as incorrect JSON format or empty dataset array"
//	@Failure		500		{object}	dto.GenericResponse[any]					"Internal server error"
//	@Router			/api/v1/evaluations/groundtruth [post]
func GetGroundtruth(c *gin.Context) {
	var req dto.GroundTruthReq
	if err := c.BindJSON(&req); err != nil {
		logrus.Errorf("failed to bind JSON request: %v", err)
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters")
		return
	}

	if len(req.Datasets) == 0 {
		logrus.Error("datasets cannot be empty")
		dto.ErrorResponse(c, http.StatusBadRequest, "Datasets cannot be empty")
		return
	}

	res, err := repository.GetGroundtruthMap(req.Datasets)
	if err != nil {
		logrus.Errorf("failed to get ground truth map: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get ground truth")
		return
	}

	dto.SuccessResponse(c, dto.GroundTruthResp(res))
}

// GetSuccessfulExecutions Get all successful algorithm execution records
//
//	@Summary		Get successful algorithm execution records
//	@Description	Get all records in ExecutionResult with status ExecutionSuccess, supports time range filtering and quantity filtering
//	@Tags			evaluation
//	@Produce		json
//	@Param			start_time	query	string	false	"Start time, format: 2006-01-02T15:04:05Z07:00"
//	@Param			end_time	query	string	false	"End time, format: 2006-01-02T15:04:05Z07:00"
//	@Param			limit		query	int		false	"Limit"
//	@Param			offset		query	int		false	"Offset for pagination"
//	@Success		200			{object}	dto.GenericResponse[dto.SuccessfulExecutionsResp]	"Successfully returns the list of successful algorithm execution records"
//	@Failure		400			{object}	dto.GenericResponse[any]							"Request parameter error"
//	@Failure		500			{object}	dto.GenericResponse[any]							"Internal server error"
//	@Router			/api/v1/evaluations/executions [get]
func GetSuccessfulExecutions(c *gin.Context) {
	var req dto.SuccessfulExecutionsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		logrus.Errorf("failed to bind query parameters: %v", err)
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	executions, err := repository.ListSuccessfulExecutionsWithFilter(req)
	if err != nil {
		logrus.Errorf("failed to get successful executions: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get successful executions")
		return
	}

	dto.SuccessResponse(c, dto.SuccessfulExecutionsResp(executions))
}
