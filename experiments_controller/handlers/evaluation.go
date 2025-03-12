package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/CUHK-SE-Group/rcabench/database"
	"github.com/CUHK-SE-Group/rcabench/dto"
	"github.com/CUHK-SE-Group/rcabench/executor"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// 将查询参数数组转换为集合
func convertQueryArrayToSet(params []string) map[string]bool {
	set := make(map[string]bool)

	for _, param := range params {
		if param != "" {
			set[param] = true
		}
	}

	return set
}

// 查询Execution相关数据并返回Execution对象
func fetchExecution(executionID, rank int) (*executor.Execution, error) {
	db := database.DB

	var execution database.ExecutionResult
	if err := db.Where("id = ?", executionID).First(&execution).Error; err != nil {
		return nil, fmt.Errorf("Algorithm does not execute")
	}

	var dataset database.FaultInjectionSchedule
	if err := db.Where("id = ?", execution.Dataset).First(&dataset).Error; err != nil {
		return nil, fmt.Errorf("Dataset id %d is not found", execution.Dataset)
	}

	// 查找detector相关的ExecutionResult
	var detectorExecution database.ExecutionResult
	if err := db.Where("dataset = ? AND algo = ?", execution.Dataset, "detector").First(&detectorExecution).Error; err != nil {
		return nil, fmt.Errorf("Detector is not runned for dataset id %d, error: %v", execution.Dataset, err)
	}

	var detectorResult database.Detector
	if err := db.Where("execution_id = ? AND issues != ?", detectorExecution.ID, "").First(&detectorResult).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	var granularityResults []database.GranularityResult
	if err := db.Where("execution_id = ? AND rank <= ?", executionID, rank).Find(&granularityResults).Error; err != nil {
		return nil, err
	}

	return &executor.Execution{
		Dataset:            dataset,
		DetectorResult:     detectorResult,
		ExecutionRecord:    execution,
		GranularityResults: granularityResults,
	}, nil
}

// GetTaskResults 获取每种算法的执行历史记录
//
//	@Summary		获取每种算法的执行历史记录
//	@Description	返回每种算法的执行历史记录
//	@Tags			evaluation
//	@Produce		application/json
//	@Param			execution_ids	query		[]int						false	"执行结果 ID 数组"
//	@Param			algorithms		query		[]string					false	"算法名称数组"
//	@Param			levels			query		[]string					false	"级别名称数组"
//	@Param			metrics			query		[]string					false	"指标名称数组"
//	@Success		200				{object}	GenericResponse[EvaluationListResp]	"成功响应"
//	@Failure		400				{object}	GenericResponse[any]		"参数校验失败"
//	@Failure		500				{object}	GenericResponse[any]		"服务器内部错误"
//	@Router			/api/v1/evaluations [get]
func GetEvaluationList(c *gin.Context) {
	var req dto.EvaluationListReq
	if err := c.BindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, dto.FormatErrorMessage(err, map[string]string{}))
		return
	}

	algoSet := convertQueryArrayToSet(req.Algoritms)
	levelSet := convertQueryArrayToSet(req.Levels)
	metricSet := convertQueryArrayToSet(req.Metrics)
	rank := 5
	if req.Rank != nil {
		rank = *req.Rank
	}

	if len(req.ExecutionIDs) == 0 {
		err := database.DB.
			Model(&database.GranularityResult{}).
			Select("DISTINCT execution_id").
			Pluck("execution_id", &req.ExecutionIDs).Error
		if err != nil {
			message := "Failed to query distinct execution_ids"
			logrus.WithError(err).Error(message)
			dto.ErrorResponse(c, http.StatusInternalServerError, message)
			return
		}
	}

	// 使用map按算法分组Execution结果
	groupedResults := make(map[string][]executor.Execution)
	for _, executionID := range req.ExecutionIDs {
		execution, err := fetchExecution(executionID, rank)
		if err != nil {
			logrus.WithError(err).WithField("execution_id", executionID).Error("Failed to fetch execution details")
			continue
		}

		algo := execution.ExecutionRecord.Algo
		if len(algoSet) == 0 || algoSet[algo] {
			groupedResults[algo] = append(groupedResults[algo], *execution)
		}
	}

	// 转化为TaskWithResults结构, 表示每个算法，在不同的执行里的信息
	var items []dto.EvaluationItem
	for algorithm, executions := range groupedResults {
		item := dto.EvaluationItem{
			Algorithm:  algorithm,
			Executions: executions,
		}

		for metric, evalFunc := range executor.GetMetrics() {
			if len(metricSet) == 0 || metricSet[metric] {
				conclusions, err := evalFunc(executions)
				if err != nil {
					message := fmt.Sprintf("Failed to calculate metric %s", metric)
					logrus.WithError(err).Errorf(message)
					dto.ErrorResponse(c, http.StatusInternalServerError, message)
					return
				}

				for _, conclusion := range conclusions {
					if len(levelSet) == 0 || levelSet[conclusion.Level] {
						item.Conclusions = append(item.Conclusions, *conclusion)
					}
				}
			}
		}

		items = append(items, item)
	}

	dto.SuccessResponse(c, dto.EvaluationListResp{Results: items})
}
