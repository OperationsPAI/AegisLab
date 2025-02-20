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
	router.Use(middleware.Logging(), middleware.GroupID(), cors.New(config))
	r := router.Group("/api/v1")

	algorithms := r.Group("/algorithms")
	{
		algorithms.GET("", handlers.GetAlgorithmList)
		algorithms.POST("", handlers.SubmitAlgorithmExecution)
	}

	datasets := r.Group("/datasets")
	{
		datasets.DELETE("/:dataset_id", handlers.DeleteDataset)
		datasets.GET("", handlers.GetDatasetList)
		datasets.GET("/download", handlers.DownloadDataset)
		datasets.POST("", handlers.SubmitDatasetBuilding)
		datasets.POST("/upload", handlers.UploadDataset)
	}

	evaluations := r.Group("/evaluations")
	{
		evaluations.GET("", handlers.GetEvaluationList)
		evaluations.POST("", handlers.SubmitEvaluation)

		tasks := evaluations.Group("/:evaluation_id")
		{
			tasks.GET("/logs", handlers.GetEvaluationLogs)
			tasks.GET("/results", handlers.GetEvaluationResults)
			tasks.GET("/status", handlers.GetEvaluationStatus)
			tasks.PUT("/cancel", handlers.CancelEvaluation)
		}
	}

	injections := r.Group("/injections")
	{
		injections.GET("", handlers.GetInjectionList)
		injections.GET("/parameters", handlers.GetInjectionParameters)
		injections.POST("", handlers.SubmitFaultInjection)
		injections.GET("/namespace_pods", handlers.GetNamespacePods)

		tasks := injections.Group("/:injection_id")
		{
			tasks.GET("/status", handlers.GetInjectionStatus)
			tasks.PUT("/cancel", handlers.CancelInjection)
		}
	}

	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return router
}
