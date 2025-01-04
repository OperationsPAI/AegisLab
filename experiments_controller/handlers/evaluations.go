package handlers

import (
	"dagger/rcabench/database"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/k0kubun/pp/v3"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Execution struct {
	Dataset           database.FaultInjectionSchedule `json:"dataset"`
	ExecutionRecord   database.ExecutionResult        `json:"execution_record"`
	GranularityResult []database.GranularityResult    `json:"granularity_results"`
	DetectorResult    database.Detector               `json:"detector_result"`
}

type Conclusion struct {
	Level  string  `json:"level"`  // 例如 service level
	Metric string  `json:"metric"` // 例如 topk
	Rate   float64 `json:"rate"`
}

type ConclusionACatK struct {
	Level  string `json:"level"`  // 例如 service level
	Metric string `json:"metric"` // 例如 topk
	Hit    []int  `json:"hit"`
}

type TaskWithResults struct {
	Algo        string      `json:"algo"`
	Executions  []Execution `json:"executions"`
	Conclusions []*Conclusion
}

// 查询Execution相关数据并返回Execution对象
func fetchExecutionDetails(db *gorm.DB, granularityID int) (*Execution, error) {
	var execution database.ExecutionResult
	if err := db.Where("id = ?", granularityID).First(&execution).Error; err != nil {
		return nil, err
	}

	var dataset database.FaultInjectionSchedule
	if err := db.Where("id = ?", execution.Dataset).First(&dataset).Error; err != nil {
		return nil, err
	}

	// 查找detector相关的ExecutionResult
	var detectorExecution database.ExecutionResult
	if err := db.Where("dataset = ? AND algo = ?", execution.Dataset, "detector").First(&detectorExecution).Error; err != nil {
		return nil, fmt.Errorf("detector is not runned for dataset %v, error: %v", execution.Dataset, err)
	}

	var detectorResult database.Detector
	if err := db.Where("execution_id = ? AND issues != ?", detectorExecution.ID, "").First(&detectorResult).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	var granularityResult []database.GranularityResult
	if err := db.Where("execution_id = ? AND rank <= ?", granularityID, 5).Find(&granularityResult).Error; err != nil {
		return nil, err
	}

	return &Execution{
		Dataset:           dataset,
		ExecutionRecord:   execution,
		GranularityResult: granularityResult,
		DetectorResult:    detectorResult,
	}, nil
}

// GetTaskResults 获取每种算法的执行历史记录
func GetTaskResults(c *gin.Context) {
	db := database.DB

	// 获取 distinct execution_ids
	var executionIDs []int
	if err := db.Model(&database.GranularityResult{}).
		Select("DISTINCT execution_id").
		Pluck("execution_id", &executionIDs).Error; err != nil {
		logrus.WithError(err).Error("Failed to query distinct execution_ids")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query distinct execution_ids"})
		return
	}

	// 使用map按算法分组Execution结果
	groupedResults := make(map[string][]Execution)
	for _, granularityID := range executionIDs {
		executionDetail, err := fetchExecutionDetails(db, granularityID)
		if err != nil {
			logrus.WithError(err).WithField("execution_id", granularityID).Error("Failed to fetch execution details")
			continue
		}

		// 按算法分组
		groupedResults[executionDetail.ExecutionRecord.Algo] = append(groupedResults[executionDetail.ExecutionRecord.Algo], *executionDetail)
	}

	pp.Println(groupedResults)

	// 转化为TaskWithResults结构, 表示每个算法，在不同的执行里的信息
	var result []TaskWithResults
	for algo, execs := range groupedResults {
		taskResult := TaskWithResults{
			Algo:       algo,
			Executions: execs,
		}
		for metric, eval := range GetMetrics() {
			conclusions, err := eval(execs)
			if err != nil {
				logrus.WithError(err).Errorf("Failed to calculate metric %s", metric)
				c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to calculate metric %s", metric)})
				return
			}
			taskResult.Conclusions = append(taskResult.Conclusions, conclusions...)
		}
		result = append(result, taskResult)
	}

	// 返回结果
	c.JSON(http.StatusOK, result)
}
