package router

import (
	_ "aegis/docs/openapi2"
	"aegis/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func New(handlers *Handlers, middlewareService middleware.Service) *gin.Engine {
	router := gin.Default()

	// CORS configuration
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "Cache-Control", "X-Request-Id"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH", "HEAD"}
	config.AllowCredentials = true
	config.ExposeHeaders = []string{"Content-Length", "Content-Type", "X-Request-Id"}

	// Middleware setup
	router.Use(
		middleware.InjectService(middlewareService),
		middleware.RequestID(),
		middleware.GroupID(),
		middleware.SSEPath(),
		cors.New(config),
		middleware.TracerMiddleware(),
	)

	middleware.StartCleanupRoutine()

	v2 := router.Group("/api/v2")
	SetupPublicV2Routes(v2, handlers)
	SetupSDKV2Routes(v2, handlers)
	SetupAdminV2Routes(v2, handlers)
	SetupPortalV2Routes(v2, handlers)

	// Swagger documentation
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}
