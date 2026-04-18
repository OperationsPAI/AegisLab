package router

import (
	"aegis/middleware"

	"github.com/gin-gonic/gin"
)

// SetupSystemRoutes sets up system routes
func SetupSystemRoutes(router *gin.Engine, handlers *Handlers) {
	audit := router.Group("/system/audit", middleware.JWTAuth(), middleware.RequireAuditRead)
	{
		audit.GET("", handlers.System.ListAuditLogs)
		audit.GET("/:id", handlers.System.GetAuditLog)
	}

	// Dynamic Configuration Management
	configs := router.Group("/system/configs", middleware.JWTAuth())
	{
		configsRead := configs.Group("", middleware.RequireConfigurationRead)
		{
			configsRead.GET("", handlers.System.ListConfigs)                              // Search configurations with filters
			configsRead.GET("/:config_id", handlers.System.GetConfig)                     // Get configuration by ID
			configsRead.GET("/:config_id/histories", handlers.System.ListConfigHistories) // Get configuration change history
		}

		// Configuration Update operations
		configs.PATCH("/:config_id", middleware.RequireConfigurationUpdate, handlers.System.UpdateConfigValue)                 // Update configuration value
		configs.POST("/:config_id/value/rollback", middleware.RequireConfigurationUpdate, handlers.System.RollbackConfigValue) // Rollback configuration value

		// Configuration Configure operations (metadata management, higher privilege)
		configs.PUT("/:config_id/metadata", middleware.RequireConfigurationConfigure, handlers.System.UpdateConfigMetadata)             // Update configuration metadata (schema)
		configs.POST("/:config_id/metadata/rollback", middleware.RequireConfigurationConfigure, handlers.System.RollbackConfigMetadata) // Rollback configuration metadata
	}

	health := router.Group("/system/health")
	{
		health.GET("", handlers.System.GetHealth)
	}

	monitor := router.Group("/system/monitor", middleware.JWTAuth(), middleware.RequireSystemRead)
	{
		monitor.POST("/metrics", handlers.System.GetMetrics)
		monitor.GET("/info", handlers.System.GetSystemInfo)
		monitor.GET("/namespaces/locks", handlers.System.ListNamespaceLocks)
		monitor.GET("/tasks/queue", handlers.System.ListQueuedTasks)
	}
}

func SetupSystemV2Routes(v2 *gin.RouterGroup, handlers *Handlers) {
	system := v2.Group("/system", middleware.JWTAuth(), middleware.RequireSystemRead)
	{
		system.GET("/metrics", handlers.SystemMetric.GetSystemMetrics)                // Get current system metrics
		system.GET("/metrics/history", handlers.SystemMetric.GetSystemMetricsHistory) // Get historical system metrics
	}
}
