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

	// CORS configuration
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true

	// Middleware setup
	router.Use(
		middleware.Logging(),
		middleware.GroupID(),
		middleware.SSEPath(),
		cors.New(config),
		middleware.TracerMiddleware(),
	)

	// Set up API routes
	SetupV1Routes(router)
	SetupV2Routes(router)

	// Swagger documentation
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}
