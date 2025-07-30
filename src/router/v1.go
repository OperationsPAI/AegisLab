package router

import (
	"github.com/LGU-SE-Internal/rcabench/handlers"
	"github.com/gin-gonic/gin"
)

// SetupV1Routes 设置 API v1 路由
func SetupV1Routes(router *gin.Engine) {
	r := router.Group("/api/v1")

	debug := r.Group("/debug")
	{
		debug.GET("/var", handlers.GetVar)
		debug.GET("/vars", handlers.GetAllVars)
		debug.POST("/var", handlers.SetVar)
		debug.GET("/ns/status", handlers.GetNSLock)
	}

	algorithms := r.Group("/algorithms")
	{
		algorithms.GET("", handlers.ListAlgorithms)
		algorithms.POST("", handlers.SubmitAlgorithmExecution)
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

	injections := r.Group("/injections")
	{
		injections.GET("", handlers.ListInjections)
		injections.GET("/conf", handlers.GetInjectionConf)
		injections.GET("/configs", handlers.ListDisplayConfigs)
		injections.GET("/mapping", handlers.GetInjectionFieldMapping)
		injections.GET("ns-resources", handlers.GetNsResourceMap)
		injections.GET("/query", handlers.QueryInjection)
		injections.POST("", handlers.SubmitFaultInjection)

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
		traces.GET("/completed", handlers.GetCompletedMap)

		tracesWithID := traces.Group("/:trace_id")
		{
			tracesWithID.GET("/stream", handlers.GetTraceStream)
		}
	}
}
