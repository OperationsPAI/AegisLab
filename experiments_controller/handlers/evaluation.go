package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/CUHK-SE-Group/rcabench/dto"
	"github.com/CUHK-SE-Group/rcabench/executor"
	"github.com/CUHK-SE-Group/rcabench/repository"
	"golang.org/x/sync/errgroup"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

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
		dto.ErrorResponse(c, http.StatusBadRequest, formatErrorMessage(err, map[string]string{}))
		return
	}

	metricSet := convertQueryArrayToSet(req.Metrics)
	rank := 5
	if req.Rank != nil {
		rank = *req.Rank
	}

	groupedResults, err := getGroupedResults(req.ExecutionIDs, req.Algoritms, req.Levels, rank)
	if err != nil {
		message := fmt.Sprintf("failed to get executions")
		logrus.Errorf("%s: %v", message, err)
		dto.ErrorResponse(c, http.StatusInternalServerError, message)
		return
	}

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
					item.Conclusions = append(item.Conclusions, conclusion)
				}
			}
		}

		items = append(items, item)
	}

	dto.SuccessResponse(c, dto.EvaluationListResp{Results: items})
}

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

func getGroupedResults(executionIDs []int, algorithms, levels []string, rank int) (map[string][]dto.Execution, error) {
	var items []dto.DatasetItemWithID
	var records []dto.ExecutionRecordWithDatasetID

	g, _ := errgroup.WithContext(context.Background())
	g.Go(func() error {
		var err error
		items, err = repository.ListDatasetByExecutionIDs(executionIDs)
		return fmt.Errorf("failed to retrieve fault injection records: %v", err)
	})

	g.Go(func() error {
		var err error
		records, err = repository.ListExecutionRecordByExecID(executionIDs, algorithms, levels, rank)
		return fmt.Errorf("failed to retrieve granularity results: %v", err)
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	recordMap := make(map[int]dto.ExecutionRecord)
	for _, record := range records {
		recordMap[record.DatasetID] = record.ExecutionRecord
	}

	grouped := make(map[string][]dto.Execution)
	for _, item := range items {
		record := recordMap[item.ID]
		execution := dto.Execution{
			Dataset:            item.DatasetItem,
			GranularityRecords: record.GranularityRecords,
		}
		grouped[record.Algorithm] = append(grouped[record.Algorithm], execution)
	}

	return grouped, nil
}
