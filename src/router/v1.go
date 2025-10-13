package router

import (
	"aegis/handlers"
	"aegis/middleware"

	"github.com/gin-gonic/gin"
)

// SetupV1Routes Setup API v1 routes
func SetupV1Routes(router *gin.Engine) {
	middleware.StartCleanupRoutine()

	r := router.Group("/api/v1")

	debug := r.Group("/debug")
	{
		debug.GET("/var", handlers.GetVar)
		debug.GET("/vars", handlers.GetAllVars)
		debug.POST("/var", handlers.SetVar)
		debug.GET("/ns/status", handlers.GetNSLock)
	}

	analyzer := r.Group("/analyzer")
	{
		analyzer.GET("/injections", handlers.AnalyzeInjections)
		analyzer.GET("/traces", handlers.AnalyzeTraces)
	}

	containers := r.Group("/containers")
	{
		containers.POST("", handlers.SubmitContainerBuilding)
	}

	datasets := r.Group("/datasets")
	{
		datasets.DELETE("", handlers.DeleteDataset)
		datasets.GET("/download", handlers.DownloadDataset)
		datasets.POST("", handlers.SubmitDatasetBuilding)
	}

	evaluations := r.Group("/evaluations")
	{
		evaluations.POST("groundtruth", handlers.GetGroundtruth)
		evaluations.POST("raw-data", handlers.ListEvaluationRawData)
		evaluations.GET("executions", handlers.GetSuccessfulExecutions)
	}

	r.GET("/injections/query", handlers.QueryInjection)

	injections := r.Group("/injections", middleware.JWTAuth())
	{
		injections.GET("", handlers.ListInjections)
		injections.GET("/conf", handlers.GetInjectionConf)
		injections.GET("/configs", handlers.ListDisplayConfigs)
		injections.GET("/mapping", handlers.GetInjectionFieldMapping)
		injections.GET("ns-resources", handlers.GetNsResourceMap)
		injections.POST("", middleware.RequireFaultInjectionWrite, handlers.SubmitFaultInjection)

		analysis := injections.Group("/analysis")
		{
			analysis.GET("/no-issues", handlers.GetFaultInjectionNoIssues)
			analysis.GET("/with-issues", handlers.GetFaultInjectionWithIssues)
			analysis.GET("/stats", handlers.GetInjectionStats)
		}

		tasks := injections.Group("/:task_id")
		{
			tasks.PUT("/cancel", handlers.CancelInjection)
		}
	}

	tasks := r.Group("/tasks")
	{
		tasks.GET("", handlers.ListTasks)
		tasks.GET("/queue", handlers.GetQueuedTasks)

		tasksWithID := tasks.Group("/:task_id")
		{
			tasksWithID.GET("", handlers.GetTaskDetail)
		}
	}

	traces := r.Group("/traces")
	{
		tracesWithID := traces.Group("/:trace_id")
		{
			tracesWithID.GET("/stream", handlers.GetTraceStream)
		}
	}
}
