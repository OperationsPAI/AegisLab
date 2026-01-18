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
		configsRead := configs.Group("", middleware.RequireConfigurationRead)
		{
			configsRead.GET("", system.ListConfigs)                              // Search configurations with filters
			configsRead.GET("/:config_id", system.GetConfig)                     // Get configuration by ID
			configsRead.GET("/:config_id/histories", system.ListConfigHistories) // Get configuration change history
		}

		configsWrite := configs.Group("", middleware.RequireConfigurationWrite)
		{
			// Frequent operation: Update configuration value (for ops team)
			configsWrite.PATCH("/:config_id", system.UpdateConfigValue)                 // Update configuration value
			configsWrite.POST("/:config_id/value/rollback", system.RollbackConfigValue) // Rollback configuration value

			// Rare operation: Update configuration metadata (admin only)
			configsWrite.PUT("/:config_id/metadata", system.UpdateConfigMetadata)             // Update configuration metadata (schema)
			configsWrite.POST("/:config_id/metadata/rollback", system.RollbackConfigMetadata) // Rollback configuration metadata
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
		monitor.GET("/namespaces/locks", system.ListNamespaceLocks)
		monitor.GET("/tasks/queue", system.ListQueuedTasks)
	}
}
