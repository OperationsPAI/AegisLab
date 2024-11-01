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

	rca := handlers.Rcabench{}
	api := router.Group("/api/v1")
	api.GET("/", rca.Home)
	return router
}
