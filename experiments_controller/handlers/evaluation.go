package handlers

import (
	"net/http"

	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GetEvaluationRawData 获取算法和数据集的原始评估数据
//
//	@Summary		获取原始评估数据
//	@Description	支持三种查询模式：1) 直接传入算法-数据集对数组进行精确查询；2) 传入算法和数据集列表进行笛卡尔积查询；3) 通过执行ID列表查询。三种模式互斥，只能选择其中一种
//	@Tags			evaluation
//	@Produce		json
//	@Consumes		application/json
//	@Param			body	body		dto.RawDataReq	true	"原始数据查询请求，支持三种模式：pairs数组、(algorithms+datasets)笛卡尔积、或execution_ids列表"
//	@Success		200		{object}	dto.GenericResponse[dto.RawDataResp]	"成功返回原始评估数据列表"
//	@Failure		400		{object}	dto.GenericResponse[any]				"请求参数错误，如JSON格式不正确、查询模式冲突或参数为空"
//	@Failure		500		{object}	dto.GenericResponse[any]				"服务器内部错误"
//	@Router			/api/v1/evaluations/raw-data [post]
func GetEvaluationRawData(c *gin.Context) {
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
		items, err = repository.ListExecutionRawDatasByPairs(req.Pairs)
	} else if req.HasCartesianMode() {
		items, err = repository.ListExecutionRawDatasByPairs(req.CartesianProduct())
	} else if req.HasExecutionMode() {
		items, err = repository.ListExecutionRawDataByIds(req.ExecutionIDs)
	}

	if err != nil {
		logrus.Errorf("failed to get raw execution data: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get raw results")
		return
	}

	dto.SuccessResponse(c, items)
}

// GetGroundtruth 获取数据集的 ground truth
//
//	@Summary		获取数据集的 ground truth
//	@Description	根据数据集数组获取对应的 ground truth 数据，用于算法评估的基准数据。支持批量查询多个数据集的真实标签信息
//	@Tags			evaluation
//	@Produce		json
//	@Consumes		application/json
//	@Param			body	body		dto.GroundTruthReq	true	"Ground truth查询请求，包含数据集列表"
//	@Success		200		{object}	dto.GenericResponse[dto.GroundTruthResp]	"成功返回数据集的ground truth信息"
//	@Failure		400		{object}	dto.GenericResponse[any]					"请求参数错误，如JSON格式不正确、数据集数组为空"
//	@Failure		500		{object}	dto.GenericResponse[any]					"服务器内部错误"
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

// GetSuccessfulExecutions 获取所有成功执行的算法记录
//
//	@Summary		获取成功执行的算法记录
//	@Description	获取所有ExecutionResult中status为ExecutionSuccess的记录，返回ID、Algorithm、Dataset三个字段
//	@Tags			evaluation
//	@Produce		json
//	@Success		200	{object}	dto.GenericResponse[dto.SuccessfulExecutionsResp]	"成功返回成功执行的算法记录列表"
//	@Failure		500	{object}	dto.GenericResponse[any]							"服务器内部错误"
//	@Router			/api/v1/evaluations/successful-executions [get]
func GetSuccessfulExecutions(c *gin.Context) {
	executions, err := repository.ListSuccessfulExecutions()
	if err != nil {
		logrus.Errorf("failed to get successful executions: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get successful executions")
		return
	}

	dto.SuccessResponse(c, dto.SuccessfulExecutionsResp(executions))
}
