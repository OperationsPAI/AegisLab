package router

import (
	"dagger/rcabench/handlers"
	"dagger/rcabench/middleware"

	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"
)

func New() *gin.Engine {
	router := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000", "http://localhost:5173"} // 允许来自前端服务器的请求
	router.Use(middleware.Logging(), cors.New(config))
	router.LoadHTMLGlob("templates/*")
	router.POST("/submit", handlers.SubmitTask)
	router.GET("/status/:taskID", handlers.GetTaskStatus)
	router.GET("/tasks", handlers.ShowAllTasks)
	router.GET("/task/:taskID/details", handlers.GetTaskDetails)
	router.GET("/task/:taskID/logs", handlers.GetTaskLogs)
	router.GET("/algobench", handlers.GetAlgoBench)
	return router
}
