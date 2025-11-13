package router

import (
	"aegis/handlers/system"
	"aegis/middleware"

	"github.com/gin-gonic/gin"
)

// SetupSystemRoutes sets up system routes
func SetupSystemRoutes(router *gin.Engine) {
	health := router.Group("/system/health")
	{
		health.GET("", system.GetHealth)
	}

	audit := router.Group("/system/audit", middleware.JWTAuth(), middleware.RequireAuditRead)
	{
		audit.GET("", system.ListAuditLogs)
		audit.GET("/:id", system.GetAuditLog)
	}

	monitor := router.Group("/system/monitor", middleware.JWTAuth(), middleware.RequireSystemRead)
	{
		monitor.POST("/metrics", system.GetMetrics)
		monitor.GET("/info", system.GetSystemInfo)
		monitor.GET("/namespaces/locks", system.ListQueuedTasks)
		monitor.GET("/task/queue", system.ListQueuedTasks)
	}
}
