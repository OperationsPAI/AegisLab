package router

import (
	"github.com/CUHK-SE-Group/rcabench/middleware"

	"github.com/CUHK-SE-Group/rcabench/handlers"

	"github.com/gin-contrib/cors"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
)

func New() *gin.Engine {
	router := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000", "http://localhost:5173"} // 允许来自前端服务器的请求
	router.Use(middleware.Logging(), middleware.GroupID(), middleware.SSEPath(), cors.New(config))
	r := router.Group("/api/v1")

	algorithms := r.Group("/algorithms")
	{
		algorithms.GET("", handlers.GetAlgorithmList)
		algorithms.POST("", handlers.SubmitAlgorithmExecution)
		algorithms.POST("build", handlers.BuildAlgorithm)
	}

	datasets := r.Group("/datasets")
	{
		datasets.DELETE("", handlers.DeleteDataset)
		datasets.GET("", handlers.GetDatasetList)
		datasets.GET("query", handlers.QueryDataset)
		datasets.GET("/download", handlers.DownloadDataset)
		datasets.POST("", handlers.SubmitDatasetBuilding)
	}

	evaluations := r.Group("/evaluations")
	{
		evaluations.GET("", handlers.GetEvaluationList)
	}

	injections := r.Group("/injections")
	{
		injections.GET("", handlers.GetInjectionList)
		injections.GET("/conf", handlers.GetInjectionConf)
		injections.POST("", handlers.SubmitFaultInjection)

		tasks := injections.Group("/:task_id")
		{
			tasks.PUT("/cancel", handlers.CancelInjection)
		}
	}

	tasks := r.Group("/tasks")
	{
		tasksWithID := tasks.Group("/:task_id")
		{
			tasksWithID.GET("", handlers.GetTaskDetail)
			tasksWithID.GET("/stream", handlers.GetTaskStream)
		}
	}

	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return router
}
