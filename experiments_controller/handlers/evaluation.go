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
//	@Description	根据算法和数据集的笛卡尔积获取对应的原始评估数据，包括粒度记录和真实值信息
//	@Tags			evaluation
//	@Accept			json
//	@Produce		application/json
//	@Param			algorithms	query		[]string	true	"算法数组"
//	@Param			datasets	query		[]string	true	"数据集数组"
//	@Success		200			{object}	dto.GenericResponse[[]dto.RawDataItem]	"成功响应"
//	@Failure		400			{object}	dto.GenericResponse[any]	"参数校验失败"
//	@Failure		500			{object}	dto.GenericResponse[any]	"服务器内部错误"
//	@Router			/api/v1/evaluations/raw-data [get]
func GetEvaluationRawData(c *gin.Context) {
	// TODO 可以同时输入算法和数据集的笛卡尔积或者是算法执行ID
	var req dto.RawDataReq
	if err := c.BindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters")
		return
	}

	if len(req.Algorithms) == 0 || len(req.Datasets) == 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Algorithms or datasets cannot be empty")
		return
	}

	items, err := repository.ListExecutionRawData(req.CartesianProduct())
	if err != nil {
		logrus.Error(err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get raw results")
		return
	}

	dto.SuccessResponse(c, items)
}

// GetGroundtruth 获取 dataset 的 groudtruth
//
//	@Summary		获取数据集的 ground truth
//	@Description	根据数据集数组获取对应的 ground truth 数据
//	@Tags			evaluation
//	@Accept			json
//	@Produce		application/json
//	@Param			datasets	query		[]string	true	"数据集数组"
//	@Success		200			{object}	dto.GenericResponse[dto.GroundTruthResp]	"成功响应"
//	@Failure		400			{object}	dto.GenericResponse[any]	"参数校验失败"
//	@Failure		500			{object}	dto.GenericResponse[any]	"服务器内部错误"
//	@Router			/api/v1/evaluations/groundtruth [get]
func GetGroundtruth(c *gin.Context) {
	var req dto.GroundTruthReq
	if err := c.BindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters")
		return
	}
	res, err := repository.GetGroundtruthMap(req.Datasets)
	if err != nil {
		logrus.Error(err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get ground truth")
		return
	}

	dto.SuccessResponse(c, dto.GroundTruthResp(res))
}
