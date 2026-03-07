package producer

import (
	"aegis/database"
	"aegis/dto"
	"fmt"

	"github.com/sirupsen/logrus"
)

// GetInjectionMetrics retrieves aggregated metrics for fault injections
func GetInjectionMetrics(req *dto.GetMetricsReq) (*dto.InjectionMetrics, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	logrus.WithFields(map[string]interface{}{
		"start_time": req.StartTime,
		"end_time":   req.EndTime,
		"fault_type": req.FaultType,
	}).Info("GetInjectionMetrics: starting")

	var injections []database.FaultInjection
	query := database.DB

	// Apply time range filter
	if req.StartTime != nil {
		query = query.Where("created_at >= ?", req.StartTime)
	}
	if req.EndTime != nil {
		query = query.Where("created_at <= ?", req.EndTime)
	}

	// Apply fault type filter
	if req.FaultType != nil {
		query = query.Where("fault_type = ?", *req.FaultType)
	}

	if err := query.Find(&injections).Error; err != nil {
		return nil, fmt.Errorf("failed to query injections: %w", err)
	}

	// Calculate metrics
	metrics := &dto.InjectionMetrics{
		TotalCount:       len(injections),
		StateDistrib:     make(map[string]int),
		FaultTypeDistrib: make(map[string]int),
	}

	var totalDuration float64
	successCount := 0
	failedCount := 0

	for _, inj := range injections {
		// Count by state
		stateName := fmt.Sprintf("%d", inj.State)
		metrics.StateDistrib[stateName]++

		// Count by fault type
		faultTypeName := fmt.Sprintf("%d", inj.FaultType)
		metrics.FaultTypeDistrib[faultTypeName]++

		// Calculate duration stats
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

		// Count success/failed
		switch inj.State {
		case 2: // success state
			successCount++
		case 3: // failed state
			failedCount++
		}
	}

	metrics.SuccessCount = successCount
	metrics.FailedCount = failedCount

	if metrics.TotalCount > 0 {
		metrics.SuccessRate = float64(successCount) / float64(metrics.TotalCount) * 100
		metrics.AvgDuration = totalDuration / float64(metrics.TotalCount)
	}

	logrus.WithField("metrics", metrics).Info("GetInjectionMetrics: completed")
	return metrics, nil
}

// GetExecutionMetrics retrieves aggregated metrics for algorithm executions
func GetExecutionMetrics(req *dto.GetMetricsReq) (*dto.ExecutionMetrics, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	logrus.WithFields(map[string]interface{}{
		"start_time":   req.StartTime,
		"end_time":     req.EndTime,
		"algorithm_id": req.AlgorithmID,
	}).Info("GetExecutionMetrics: starting")

	var executions []database.Execution
	query := database.DB

	// Apply time range filter
	if req.StartTime != nil {
		query = query.Where("created_at >= ?", req.StartTime)
	}
	if req.EndTime != nil {
		query = query.Where("created_at <= ?", req.EndTime)
	}

	// Apply algorithm filter
	if req.AlgorithmID != nil {
		query = query.Where("algorithm_id = ?", *req.AlgorithmID)
	}

	if err := query.Find(&executions).Error; err != nil {
		return nil, fmt.Errorf("failed to query executions: %w", err)
	}

	// Calculate metrics
	metrics := &dto.ExecutionMetrics{
		TotalCount:   len(executions),
		StateDistrib: make(map[string]int),
	}

	var totalDuration float64
	successCount := 0
	failedCount := 0

	for _, exec := range executions {
		// Count by state
		stateName := fmt.Sprintf("%d", exec.State)
		metrics.StateDistrib[stateName]++

		// Calculate duration stats
		if exec.Duration > 0 {
			totalDuration += exec.Duration

			if metrics.MinDuration == 0 || exec.Duration < metrics.MinDuration {
				metrics.MinDuration = exec.Duration
			}
			if exec.Duration > metrics.MaxDuration {
				metrics.MaxDuration = exec.Duration
			}
		}

		// Count success/failed
		switch exec.State {
		case 2: // success state
			successCount++
		case 3: // failed state
			failedCount++
		}
	}

	metrics.SuccessCount = successCount
	metrics.FailedCount = failedCount

	if metrics.TotalCount > 0 {
		metrics.SuccessRate = float64(successCount) / float64(metrics.TotalCount) * 100
		metrics.AvgDuration = totalDuration / float64(metrics.TotalCount)
	}

	logrus.WithField("metrics", metrics).Info("GetExecutionMetrics: completed")
	return metrics, nil
}

// GetAlgorithmMetrics retrieves comparative metrics across different algorithms
func GetAlgorithmMetrics(req *dto.GetMetricsReq) (*dto.AlgorithmMetrics, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	logrus.WithFields(map[string]interface{}{
		"start_time": req.StartTime,
		"end_time":   req.EndTime,
	}).Info("GetAlgorithmMetrics: starting")

	// Get all algorithms
	var algorithms []database.Container
	query := database.DB.Where("type = ?", 2) // Assuming 2 is algorithm type

	if err := query.Find(&algorithms).Error; err != nil {
		return nil, fmt.Errorf("failed to query algorithms: %w", err)
	}

	metrics := &dto.AlgorithmMetrics{
		Algorithms: make([]dto.AlgorithmMetricItem, 0, len(algorithms)),
	}

	// Calculate metrics for each algorithm
	for _, algo := range algorithms {
		var executions []database.Execution
		execQuery := database.DB.Where("algorithm_id = ?", algo.ID)

		// Apply time range filter
		if req.StartTime != nil {
			execQuery = execQuery.Where("created_at >= ?", req.StartTime)
		}
		if req.EndTime != nil {
			execQuery = execQuery.Where("created_at <= ?", req.EndTime)
		}

		if err := execQuery.Find(&executions).Error; err != nil {
			logrus.WithError(err).Warnf("failed to query executions for algorithm %d", algo.ID)
			continue
		}

		if len(executions) == 0 {
			continue
		}

		item := dto.AlgorithmMetricItem{
			AlgorithmID:    algo.ID,
			AlgorithmName:  algo.Name,
			ExecutionCount: len(executions),
		}

		var totalDuration float64
		successCount := 0
		failedCount := 0

		for _, exec := range executions {
			// Calculate duration stats
			if exec.Duration > 0 {
				totalDuration += exec.Duration
			}

			// Count success/failed
			switch exec.State {
			case 2: // success state
				successCount++
			case 3: // failed state
				failedCount++
			}
		}

		item.SuccessCount = successCount
		item.FailedCount = failedCount
		item.SuccessRate = float64(successCount) / float64(item.ExecutionCount) * 100
		if item.ExecutionCount > 0 {
			item.AvgDuration = totalDuration / float64(item.ExecutionCount)
		}

		metrics.Algorithms = append(metrics.Algorithms, item)
	}

	logrus.WithField("algorithm_count", len(metrics.Algorithms)).Info("GetAlgorithmMetrics: completed")
	return metrics, nil
}
