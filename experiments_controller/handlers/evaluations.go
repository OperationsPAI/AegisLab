package handlers

import (
	"dagger/rcabench/database"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"net/http"
)

type TaskWithResults struct {
	Algo       string      `json:"algo"`
	Executions []Execution `json:"executions"`
}

type Execution struct {
	Dataset           database.FaultInjectionSchedule `json:"dataset"`
	ExecutionRecord   database.ExecutionResult        `json:"execution_record"`
	GranularityResult []database.GranularityResult    `json:"granularity_results"`
	DetectorResult    database.Detector               `json:"detector_result"`
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
		return nil, err
	}

	var detectorResult database.Detector
	if err := db.Where("execution_id = ?", detectorExecution.ID).First(&detectorResult).Error; err != nil {
		return nil, err
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

	// 转化为TaskWithResults结构
	var result []TaskWithResults
	for algo, execs := range groupedResults {
		result = append(result, TaskWithResults{
			Algo:       algo,
			Executions: execs,
		})
	}

	// 返回结果
	c.JSON(http.StatusOK, result)
}
