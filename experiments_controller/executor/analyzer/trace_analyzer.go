package analyzer

import (
	"context"

	"github.com/CUHK-SE-Group/rcabench/repository"
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
	groupTrace, err := repository.GetGroupToTraceIDsMap()
	if err != nil {
		logrus.WithError(err).Error("Failed to get group to trace IDs map")
		return nil, err
	}

	for groupID, traceIDs := range groupTrace {
		logrus.WithField("group_id", groupID).WithField("trace_ids", traceIDs).Info("Group to Trace IDs")
		for _, traceID := range traceIDs {
			stat, err := repository.GetTraceStatistic(ctx, traceID)
			if err != nil {
				logrus.WithError(err).WithField("trace_id", traceID).Error("Failed to get trace statistic")
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
