package router

import (
	"aegis/handlers/system"
	"aegis/middleware"

	"github.com/gin-gonic/gin"
)

// SetupSystemRoutes sets up system routes
func SetupSystemRoutes(router *gin.Engine) {
	audit := router.Group("/system/audit", middleware.JWTAuth(), middleware.RequireAuditRead)
	{
		audit.GET("", system.ListAuditLogs)
		audit.GET("/:id", system.GetAuditLog)
	}

	// Dynamic Configuration Management
	configs := router.Group("/system/configs", middleware.JWTAuth())
	{
		configsRead := configs.Group("", middleware.RequireConfigRead)
		{
			configsRead.GET("", system.ListConfigs)                              // Search configurations with filters
			configsRead.GET("/:config_id", system.GetConfig)                     // Get configuration by ID
			configsRead.GET("/:config_id/histories", system.ListConfigHistories) // Get configuration change history                // Get configuration statistics
		}

		configsWrite := configs.Group("", middleware.RequireConfigWrite)
		{
			configsWrite.PATCH("/:config_id", system.UpdateConfig)           // Update configuration
			configsWrite.POST("/:config_id/rollback", system.RollbackConfig) // Rollback configuration
		}
	}

	health := router.Group("/system/health")
	{
		health.GET("", system.GetHealth)
	}

	monitor := router.Group("/system/monitor", middleware.JWTAuth(), middleware.RequireSystemRead)
	{
		monitor.POST("/metrics", system.GetMetrics)
		monitor.GET("/info", system.GetSystemInfo)
		monitor.GET("/namespaces/locks", system.ListQueuedTasks)
		monitor.GET("/task/queue", system.ListQueuedTasks)
	}
}
