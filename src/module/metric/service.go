package metricmodule

import (
	"context"
	"fmt"

	"aegis/model"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetInjectionMetrics(_ context.Context, req *GetMetricsReq) (*InjectionMetrics, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	logrus.WithFields(map[string]interface{}{
		"start_time": req.StartTime,
		"end_time":   req.EndTime,
		"fault_type": req.FaultType,
	}).Info("GetInjectionMetrics: starting")

	injections, err := s.repo.ListFaultInjections(func(db *gorm.DB) *gorm.DB {
		query := db
		if req.StartTime != nil {
			query = query.Where("created_at >= ?", req.StartTime)
		}
		if req.EndTime != nil {
			query = query.Where("created_at <= ?", req.EndTime)
		}
		if req.FaultType != nil {
			query = query.Where("fault_type = ?", *req.FaultType)
		}
		return query
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query injections: %w", err)
	}

	metrics := buildInjectionMetrics(injections)
	logrus.WithField("metrics", metrics).Info("GetInjectionMetrics: completed")
	return metrics, nil
}

func (s *Service) GetExecutionMetrics(_ context.Context, req *GetMetricsReq) (*ExecutionMetrics, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	logrus.WithFields(map[string]interface{}{
		"start_time":   req.StartTime,
		"end_time":     req.EndTime,
		"algorithm_id": req.AlgorithmID,
	}).Info("GetExecutionMetrics: starting")

	executions, err := s.repo.ListExecutions(func(db *gorm.DB) *gorm.DB {
		query := db
		if req.StartTime != nil {
			query = query.Where("created_at >= ?", req.StartTime)
		}
		if req.EndTime != nil {
			query = query.Where("created_at <= ?", req.EndTime)
		}
		if req.AlgorithmID != nil {
			query = query.Where("algorithm_id = ?", *req.AlgorithmID)
		}
		return query
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query executions: %w", err)
	}

	metrics := buildExecutionMetrics(executions)
	logrus.WithField("metrics", metrics).Info("GetExecutionMetrics: completed")
	return metrics, nil
}

func (s *Service) GetAlgorithmMetrics(_ context.Context, req *GetMetricsReq) (*AlgorithmMetrics, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	logrus.WithFields(map[string]interface{}{
		"start_time": req.StartTime,
		"end_time":   req.EndTime,
	}).Info("GetAlgorithmMetrics: starting")

	algorithms, err := s.repo.ListAlgorithmContainers()
	if err != nil {
		return nil, fmt.Errorf("failed to query algorithms: %w", err)
	}

	metrics := &AlgorithmMetrics{
		Algorithms: make([]AlgorithmMetricItem, 0, len(algorithms)),
	}

	for _, algo := range algorithms {
		if req.AlgorithmID != nil && algo.ID != *req.AlgorithmID {
			continue
		}
		executions, err := s.repo.ListExecutions(func(db *gorm.DB) *gorm.DB {
			query := db.Where("algorithm_id = ?", algo.ID)
			if req.StartTime != nil {
				query = query.Where("created_at >= ?", req.StartTime)
			}
			if req.EndTime != nil {
				query = query.Where("created_at <= ?", req.EndTime)
			}
			return query
		})
		if err != nil {
			logrus.WithError(err).Warnf("failed to query executions for algorithm %d", algo.ID)
			continue
		}
		item, ok := buildAlgorithmMetricItem(algo, executions)
		if ok {
			metrics.Algorithms = append(metrics.Algorithms, item)
		}
	}

	logrus.WithField("algorithm_count", len(metrics.Algorithms)).Info("GetAlgorithmMetrics: completed")
	return metrics, nil
}

func buildInjectionMetrics(injections []model.FaultInjection) *InjectionMetrics {
	metrics := &InjectionMetrics{
		TotalCount:       len(injections),
		StateDistrib:     make(map[string]int),
		FaultTypeDistrib: make(map[string]int),
	}

	var totalDuration float64
	successCount := 0
	failedCount := 0

	for _, inj := range injections {
		stateName := fmt.Sprintf("%d", inj.State)
		metrics.StateDistrib[stateName]++

		faultTypeName := fmt.Sprintf("%d", inj.FaultType)
		metrics.FaultTypeDistrib[faultTypeName]++

		if inj.StartTime != nil && inj.EndTime != nil {
			duration := inj.EndTime.Sub(*inj.StartTime).Seconds()
			totalDuration += duration
			if metrics.MinDuration == 0 || duration < metrics.MinDuration {
				metrics.MinDuration = duration
			}
			if duration > metrics.MaxDuration {
				metrics.MaxDuration = duration
			}
		}

		switch inj.State {
		case 2:
			successCount++
		case 3:
			failedCount++
		}
	}

	metrics.SuccessCount = successCount
	metrics.FailedCount = failedCount
	if metrics.TotalCount > 0 {
		metrics.SuccessRate = float64(successCount) / float64(metrics.TotalCount) * 100
		metrics.AvgDuration = totalDuration / float64(metrics.TotalCount)
	}
	return metrics
}

func buildExecutionMetrics(executions []model.Execution) *ExecutionMetrics {
	metrics := &ExecutionMetrics{
		TotalCount:   len(executions),
		StateDistrib: make(map[string]int),
	}

	var totalDuration float64
	successCount := 0
	failedCount := 0

	for _, exec := range executions {
		stateName := fmt.Sprintf("%d", exec.State)
		metrics.StateDistrib[stateName]++
		if exec.Duration > 0 {
			totalDuration += exec.Duration
			if metrics.MinDuration == 0 || exec.Duration < metrics.MinDuration {
				metrics.MinDuration = exec.Duration
			}
			if exec.Duration > metrics.MaxDuration {
				metrics.MaxDuration = exec.Duration
			}
		}
		switch exec.State {
		case 2:
			successCount++
		case 3:
			failedCount++
		}
	}

	metrics.SuccessCount = successCount
	metrics.FailedCount = failedCount
	if metrics.TotalCount > 0 {
		metrics.SuccessRate = float64(successCount) / float64(metrics.TotalCount) * 100
		metrics.AvgDuration = totalDuration / float64(metrics.TotalCount)
	}
	return metrics
}

func buildAlgorithmMetricItem(algo model.Container, executions []model.Execution) (AlgorithmMetricItem, bool) {
	if len(executions) == 0 {
		return AlgorithmMetricItem{}, false
	}

	item := AlgorithmMetricItem{
		AlgorithmID:    algo.ID,
		AlgorithmName:  algo.Name,
		ExecutionCount: len(executions),
	}

	var totalDuration float64
	successCount := 0
	failedCount := 0
	for _, exec := range executions {
		if exec.Duration > 0 {
			totalDuration += exec.Duration
		}
		switch exec.State {
		case 2:
			successCount++
		case 3:
			failedCount++
		}
	}

	item.SuccessCount = successCount
	item.FailedCount = failedCount
	item.SuccessRate = float64(successCount) / float64(item.ExecutionCount) * 100
	item.AvgDuration = totalDuration / float64(item.ExecutionCount)
	return item, true
}
