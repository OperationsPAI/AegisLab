package router

import (
	"github.com/LGU-SE-Internal/rcabench/middleware"

	"github.com/LGU-SE-Internal/rcabench/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func New() *gin.Engine {
	router := gin.Default()
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	router.Use(middleware.Logging(), middleware.GroupID(), middleware.SSEPath(), cors.New(config), middleware.TracerMiddleware())
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
		algorithms.GET("", handlers.GetAlgorithmList)
		algorithms.POST("", handlers.SubmitAlgorithmExecution)
		algorithms.POST("/build", handlers.SubmitAlgorithmBuilding)
	}

	datasets := r.Group("/datasets")
	{
		datasets.DELETE("", handlers.DeleteDataset)
		datasets.GET("", handlers.GetDatasetList)
		datasets.GET("/download", handlers.DownloadDataset)
		datasets.GET("/query", handlers.QueryDataset)
		datasets.POST("", handlers.SubmitDatasetBuilding)
	}

	evaluations := r.Group("/evaluations")
	{
		evaluations.POST("groundtruth", handlers.GetGroundtruth)
		evaluations.POST("raw-data", handlers.GetEvaluationRawData)
	}

	injections := r.Group("/injections")
	{
		injections.GET("", handlers.ListInjections)
		injections.GET("/conf", handlers.GetInjectionConf)
		injections.GET("/configs", handlers.ListDisplayConfigs)
		injections.GET("/mapping", handlers.GetInjectionFieldMapping)
		injections.GET("key-resource", handlers.GetKeyResourceMap)
		injections.GET("ns-resources", handlers.GetNsResourceMap)
		injections.GET("/query", handlers.QueryInjection)
		injections.POST("", handlers.SubmitFaultInjection)

		analysis := injections.Group("/analysis")
		{
			analysis.GET("/no-issues", handlers.GetFaultInjectionNoIssues)
			analysis.GET("/with-issues", handlers.GetFaultInjectionWithIssues)
			analysis.GET("/statistics", handlers.GetFaultInjectionStatistics)
		}

		tasks := injections.Group("/:task_id")
		{
			tasks.PUT("/cancel", handlers.CancelInjection)
		}
	}

	tasks := r.Group("/tasks")
	{
		tasks.GET("/queue", handlers.GetQueuedTasks)
		tasks.GET("/list", handlers.ListTasks)

		tasksWithID := tasks.Group("/:task_id")
		{
			tasksWithID.GET("", handlers.GetTaskDetail)
		}
	}

	traces := r.Group("/traces")
	{
		traces.GET("/analyze", handlers.AnalyzeTrace)
		traces.GET("/completed", handlers.GetCompletedMap)

		tracesWithID := traces.Group("/:trace_id")
		{
			tracesWithID.GET("/stream", handlers.GetTraceStream)
		}
	}

	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return router
}
