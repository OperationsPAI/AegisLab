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
	router.Use(middleware.Logging(), cors.New(config))
	r := router.Group("/api/v1")
	// router.POST("/submit", handlers.SubmitTask)
	// router.GET("/status/:taskID", handlers.GetTaskStatus)
	// router.GET("/tasks", handlers.ShowAllTasks)
	// router.GET("/task/:taskID/details", handlers.GetTaskDetails)
	// router.GET("/task/:taskID/logs", handlers.GetTaskLogs)
	// router.GET("/algobench", handlers.GetAlgoBench)
	// router.GET("/datasets", handlers.GetDatasets)
	// router.GET("/injection", handlers.GetInjectionPara)
	// router.GET("/namespacepod", handlers.GetNamespacePod)
	// router.DELETE("/task/:taskID", handlers.WithdrawTask)

	algor := r.Group("/algo")
	{
		algor.GET("", handlers.GetAlgorithmList)
		algor.POST("", handlers.SubmitAlgorithmExecution)
	}

	datasetr := r.Group("/dataset")
	{
		datasetr.GET("", handlers.GetDatasetList)
		datasetr.POST("/download", handlers.DownloadDataset)
		datasetr.POST("/upload", handlers.UploadDataset)
		datasetr.DELETE("/:datasetID", handlers.DeleteDataset)
	}

	evaluationr := r.Group("/evaluation")
	{
		evaluationr.GET("", handlers.GetEvaluationList)
		evaluationr.POST("", handlers.SubmitEvaluation)
		evaluationr.POST("/:taskID/cancel", handlers.CancelEvaluation)
		evaluationr.POST("/:taskID/logs", handlers.GetEvaluationLogs)
		evaluationr.POST("/:taskID/result", handlers.GetEvaluationResults)
		evaluationr.POST("/:taskID/status", handlers.GetEvaluationStatus)
	}

	injectr := r.Group("/injection")
	{
		injectr.GET("", handlers.GetInjectionList)
		injectr.POST("", handlers.SubmitFaultInjection)
		injectr.GET("/para", handlers.GetInjectionPara)
		injectr.POST("/:taskID/cancel", handlers.CancelInjection)
		injectr.GET("/:taskID/status", handlers.GetInjectionStatus)
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return router
}
