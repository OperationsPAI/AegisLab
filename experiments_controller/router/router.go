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
	router.LoadHTMLGlob("templates/*")
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
		algor.GET("/getlist", handlers.GetAlgorithmList)
	}

	datasetr := r.Group("/dataset")
	{
		datasetr.GET("/getlist", handlers.GetDatasetList)
		datasetr.DELETE("/delete", handlers.DeleteDataset)
		datasetr.POST("/download", handlers.DownloadDataset)
		datasetr.POST("/upload", handlers.UploadDataset)
	}

	evaluationr := r.Group("/evaluation")
	{
		evaluationr.POST("/getlist", handlers.GetEvaluationList)
		evaluationr.POST("/cancel", handlers.CancelEvaluation)
		evaluationr.POST("/getstatus", handlers.GetEvaluationStatus)
		evaluationr.POST("/getlogs", handlers.GetEvaluationLogs)
		evaluationr.POST("/submit", handlers.SubmitEvaluation)
		evaluationr.POST("/getresult", handlers.GetEvaluationResults)
	}

	injectr := r.Group("/injection")
	{
		injectr.POST("/submit", handlers.InjectFault)
		injectr.POST("/getlist", handlers.GetInjectionList)
		injectr.POST("/getstatus", handlers.GetInjectionStatus)
		injectr.POST("/cancel", handlers.CancelInjection)
		injectr.POST("/getpara", handlers.GetInjectionPara)
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return router
}
