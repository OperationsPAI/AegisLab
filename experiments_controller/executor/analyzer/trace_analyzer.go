package analyzer

import (
	"context"

	"github.com/LGU-SE-Internal/rcabench/repository"
	"github.com/sirupsen/logrus"
)

type Statistics struct {
	TotalTraces int
	HasAnomaly  int
	NoAnomaly   int
	AvgDuration float64
	MinDuration float64
	MaxDuration float64
}

func AnalyzeTrace(ctx context.Context) (*Statistics, error) {
	stats := &Statistics{}
	groupToTraceIDs, err := repository.GetGroupToTraceIDsMap()
	if err != nil {
		logrus.WithError(err).Error("Failed to get group to trace IDs map")
		return nil, err
	}

	for _, traceIDs := range groupToTraceIDs {
		for _, traceID := range traceIDs {
			stat, err := repository.GetTraceStatistic(ctx, traceID)
			if err != nil {
				logrus.WithError(err).WithField("trace_id", traceID).Debug("Failed to get trace statistic")
				continue
			}

			stats.TotalTraces++
			if stat.DetectAnomaly {
				stats.HasAnomaly++
			} else {
				stats.NoAnomaly++
			}

			stats.AvgDuration += stat.TotalDuration
			if stats.MinDuration == 0 || stat.TotalDuration < stats.MinDuration {
				stats.MinDuration = stat.TotalDuration
			}
			if stat.TotalDuration > stats.MaxDuration {
				stats.MaxDuration = stat.TotalDuration
			}
		}
	}

	stats.AvgDuration /= float64(stats.TotalTraces)
	return stats, nil
}
