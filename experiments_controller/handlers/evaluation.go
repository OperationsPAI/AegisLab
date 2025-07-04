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
//	@Description	根据算法和数据集的笛卡尔积获取对应的原始评估数据，包括粒度记录和真实值信息。支持批量查询多个算法在多个数据集上的执行结果
//	@Tags			evaluation
//	@Produce		json
//	@Consumes		application/json
//	@Param			body	body		dto.RawDataReq	true	"原始数据查询请求，包含算法列表和数据集列表"
//	@Success		200		{object}	dto.GenericResponse[[]dto.RawDataItem]	"成功返回原始评估数据列表"
//	@Failure		400		{object}	dto.GenericResponse[any]				"请求参数错误，如JSON格式不正确、算法或数据集数组为空"
//	@Failure		500		{object}	dto.GenericResponse[any]				"服务器内部错误"
//	@Router			/api/v1/evaluations/raw-data [post]
func GetEvaluationRawData(c *gin.Context) {
	// TODO 可以同时输入算法和数据集的笛卡尔积或者是算法执行ID
	var req dto.RawDataReq
	if err := c.BindJSON(&req); err != nil {
		logrus.Errorf("failed to bind JSON request: %v", err)
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters")
		return
	}

	if len(req.Algorithms) == 0 || len(req.Datasets) == 0 {
		logrus.Error("algorithms or datasets cannot be empty")
		dto.ErrorResponse(c, http.StatusBadRequest, "Algorithms or datasets cannot be empty")
		return
	}

	items, err := repository.ListExecutionRawData(req.CartesianProduct())
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
