package dto

import "time"

// ProjectStatistics holds statistics for a project
type ProjectStatistics struct {
	InjectionCount  int
	ExecutionCount  int
	LastInjectionAt *time.Time
	LastExecutionAt *time.Time
}
