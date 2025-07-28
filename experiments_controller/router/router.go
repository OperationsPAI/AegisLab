package router

import (
	"github.com/LGU-SE-Internal/rcabench/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func New() *gin.Engine {
	router := gin.Default()

	// CORS 配置
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true

	// 中间件设置
	router.Use(
		middleware.Logging(),
		middleware.GroupID(),
		middleware.SSEPath(),
		cors.New(config),
		middleware.TracerMiddleware(),
	)

	// 设置 API 路由
	SetupV1Routes(router)
	SetupV2Routes(router)

	// Swagger 文档
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}
