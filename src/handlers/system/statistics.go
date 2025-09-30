package system

import (
	"aegis/dto"
	"aegis/repository"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

// GetStatistics handles system statistics
//
//	@Summary Get system statistics
//	@Description Get comprehensive system statistics and metrics
//	@Tags System
//	@Produce json
//	@Success 200 {object} dto.GenericResponse[dto.SystemStatisticsResponse] "Statistics retrieved successfully"
//	@Failure 500 {object} dto.GenericResponse[any] "Internal server error"
//	@Router /system/statistics [get]
func GetStatistics(c *gin.Context) {
	var response dto.SystemStatisticsResponse

	// Get real system memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Get real user statistics
	userStats := dto.UserStatistics{}
	if allUsers, _, err := repository.ListUsers(1, 10000, nil); err == nil {
		userStats.Total = len(allUsers)
		for _, user := range allUsers {
			if user.IsActive {
				userStats.Active++
			} else {
				userStats.Inactive++
			}
			// Check if user was created today
			if user.CreatedAt.After(time.Now().Add(-24 * time.Hour)) {
				userStats.NewToday++
			}
			// Check if user was created this week
			if user.CreatedAt.After(time.Now().Add(-7 * 24 * time.Hour)) {
				userStats.NewThisWeek++
			}
		}
	}

	// Get real role statistics - using available repository functions
	roleStats := dto.ProjectStatistics{} // Reusing structure for roles
	if allRoles, _, err := repository.ListRoles(1, 10000, "", nil); err == nil {
		roleStats.Total = len(allRoles)
		for _, role := range allRoles {
			if role.Status == 1 {
				roleStats.Active++
			} else {
				roleStats.Inactive++
			}
			if role.CreatedAt.After(time.Now().Add(-24 * time.Hour)) {
				roleStats.NewToday++
			}
		}
	}

	// Get real permission statistics
	permStats := dto.TaskStatistics{} // Reusing structure for permissions
	if allPerms, _, err := repository.ListPermissions(1, 10000, "", nil, nil); err == nil {
		permStats.Total = len(allPerms)
		for _, perm := range allPerms {
			if perm.Status == 1 {
				permStats.Completed++ // Using as "active"
			} else {
				permStats.Failed++ // Using as "inactive"
			}
		}
	}

	// Get container statistics
	containerStats, err := repository.GetContainerStatistics()
	if err != nil {
		// Log error but continue with default values
		containerStats = map[string]int64{"total": 0, "active": 0, "deleted": 0}
	}

	// Get dataset statistics
	datasetStats, err := repository.GetDatasetStatistics()
	if err != nil {
		// Log error but continue with default values
		datasetStats = map[string]int64{"total": 0, "active": 0, "deleted": 0}
	}

	// Injection statistics are now handled in the response section using GetInjectionDetailedStats

	// Get execution statistics
	executionStats, err := repository.GetExecutionStatistics()
	if err != nil {
		executionStats = map[string]int64{"total": 0, "pending": 0, "running": 0, "completed": 0, "failed": 0}
	}

	// Get task statistics
	taskStats, err := repository.GetTaskStatistics()
	if err != nil {
		taskStats = map[string]int64{"total": 0}
	}

	// Get project statistics
	projectStats, err := repository.GetProjectStatistics()
	if err != nil {
		projectStats = map[string]int64{"total": 0, "active": 0, "inactive": 0, "new_today": 0}
	}

	response = dto.SystemStatisticsResponse{
		Users: userStats,
		Projects: dto.ProjectStatistics{
			Total:    int(projectStats["total"]),
			Active:   int(projectStats["active"]),
			Inactive: int(projectStats["inactive"]),
			NewToday: int(projectStats["new_today"]),
		},
		Tasks: dto.TaskStatistics{
			Total:     int(taskStats["total"]),
			Pending:   int(taskStats["pending"]),
			Running:   int(taskStats["running"]),
			Completed: int(taskStats["completed"]),
			Failed:    int(taskStats["failed"]),
		},
		Containers: dto.ContainerStatistics{
			Total:   int(containerStats["total"]),
			Active:  int(containerStats["active"]),
			Deleted: int(containerStats["deleted"]),
		},
		Datasets: func() dto.DatasetStatistics {
			totalSize, err := repository.GetDatasetTotalSize()
			if err != nil {
				totalSize = 0
			}
			return dto.DatasetStatistics{
				Total:     int(datasetStats["total"]),
				Public:    int(datasetStats["active"]),
				Private:   int(datasetStats["deleted"]),
				TotalSize: totalSize,
			}
		}(),
		Injections: func() dto.InjectionStatistics {
			detailedStats, err := repository.GetInjectionDetailedStats()
			if err != nil {
				detailedStats = map[string]int64{"total": 0, "scheduled": 0, "running": 0, "completed": 0, "failed": 0}
			}
			return dto.InjectionStatistics{
				Total:     int(detailedStats["total"]),
				Scheduled: int(detailedStats["scheduled"]),
				Running:   int(detailedStats["running"]),
				Completed: int(detailedStats["completed"]),
				Failed:    int(detailedStats["failed"]),
			}
		}(),
		Executions: dto.ExecutionStatistics{
			Total:      int(executionStats["total"]),
			Successful: int(executionStats["completed"]),
			Failed:     int(executionStats["failed"]),
		},
		GeneratedAt: time.Now(),
	}

	dto.SuccessResponse(c, response)
}
