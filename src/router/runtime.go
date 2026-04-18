package router

import (
	"aegis/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRuntimeV2Routes(v2 *gin.RouterGroup, handlers *Handlers) {
	runtime := v2.Group("/executions", middleware.JWTAuth(), middleware.RequireServiceTokenAuth())
	{
		runtime.POST("/:execution_id/detector_results", handlers.Execution.UploadDetectorResults)
		runtime.POST("/:execution_id/granularity_results", handlers.Execution.UploadGranularityResults)
	}
}
