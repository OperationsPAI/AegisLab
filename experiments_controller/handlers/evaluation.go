package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/executor"
	"github.com/LGU-SE-Internal/rcabench/repository"
	"github.com/k0kubun/pp/v3"
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
//	@Success		200				{object}	dto.GenericResponse[dto.EvaluationListResp]	"成功响应"
//	@Failure		400				{object}	dto.GenericResponse[any]		"参数校验失败"
//	@Failure		500				{object}	dto.GenericResponse[any]		"服务器内部错误"
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
		logrus.Error(err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "failed to get executions")
		return
	}

	var items []dto.EvaluationItem
	for algorithm, executions := range groupedResults {
		groundTruthMaps, err := executor.ParseConfigAndGetGroundTruthMap(executions)
		if err != nil {
			message := "failed to read grountruths"
			logrus.Errorf("%s: %v", message, err)
			dto.ErrorResponse(c, http.StatusInternalServerError, message)
			return
		}

		pp.Println(groundTruthMaps)

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
				item.Conclusions = append(item.Conclusions, conclusions...)
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
		if err != nil {
			return fmt.Errorf("failed to retrieve fault injection records: %v", err)
		}

		return nil
	})

	g.Go(func() error {
		var err error
		records, err = repository.ListExecutionRecordByExecID(executionIDs, algorithms, levels, rank)
		if err != nil {
			return fmt.Errorf("failed to retrieve granularity results: %v", err)
		}

		return nil
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

// func GetCalculatedEvaluationResults(c *gin.Context) {
// 	var req dto.CalculatedResultsReq
// 	if err := c.BindQuery(&req); err != nil {
// 		dto.ErrorResponse(c, http.StatusBadRequest, formatErrorMessage(err, map[string]string{}))
// 		return
// 	}

// 	results, err := repository.ListCalculatedResults(req.ExecutionIDs, req.Algorithms, req.Levels)
// 	if err != nil {
// 		logrus.Error(err)
// 		dto.ErrorResponse(c, http.StatusInternalServerError, "failed to get calculated results")
// 		return
// 	}

// 	dto.SuccessResponse(c, dto.CalculatedResultsResp{Results: results})
// }

func GetEvaluationRawData(c *gin.Context) {
	var req dto.RawDataReq
	if err := c.BindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters")
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
