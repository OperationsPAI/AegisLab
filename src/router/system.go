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

		// Admin only
		auditAdmin := audit.Group("", middleware.RequireSystemAdmin)
		{
			auditAdmin.POST("", system.CreateAuditLog)
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
	}

	statistics := router.Group("/system/statistics", middleware.JWTAuth(), middleware.RequireSystemRead)
	{
		statistics.GET("", system.GetStatistics)
	}
}
