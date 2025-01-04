package handlers

import (
	"dagger/rcabench/database"
	"dagger/rcabench/executor"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/k0kubun/pp/v3"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
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
	Conclusion        []Conclusion                    `json:"conclusion"`
}

type Conclusion struct {
	Level  string `json:"level"`  // 例如 service level
	Metric string `json:"metric"` // 例如 topk
	Hit    bool   `json:"hit"`
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
		Conclusion:        make([]Conclusion, 0),
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

	// 转化为TaskWithResults结构
	var result []TaskWithResults
	for algo, execs := range groupedResults {
		result = append(result, TaskWithResults{
			Algo:       algo,
			Executions: execs,
		})
	}

	for _, res := range result {
		for idx := range res.Executions {
			if res.Executions[idx].DetectorResult.Issues != "" {
				record := map[string]int{
					"top1": 0,
					"top3": 0,
					"top5": 0,
				}
				for _, g := range res.Executions[idx].GranularityResult {
					var payload map[string]interface{}
					err := json.Unmarshal([]byte(res.Executions[idx].Dataset.Config), &payload)
					if err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unmarshal config"})
						return
					}
					conf, err := executor.ParseFaultInjectionPayload(payload)
					if err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse payload"})
						return
					}
					fmt.Println(g.Result, conf.Pod)
					if g.Result == conf.Pod {
						if g.Rank == 1 {
							record["top1"] += 1
						}
						if g.Rank <= 3 {
							record["top3"] += 1
						}
						if g.Rank <= 5 {
							record["top5"] += 1
						}
					}

				}
				for metric, hit := range record {
					res.Executions[idx].Conclusion = append(res.Executions[idx].Conclusion, Conclusion{
						Level:  res.Executions[idx].GranularityResult[0].Level,
						Metric: metric,
						Hit:    hit > 0,
					})
				}

				fmt.Println(res.Executions[idx].Conclusion)
			}
		}
	}

	// 返回结果
	c.JSON(http.StatusOK, result)
}
