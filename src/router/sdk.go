package router

import (
	"aegis/middleware"

	"github.com/gin-gonic/gin"
)

func SetupSDKV2Routes(v2 *gin.RouterGroup, handlers *Handlers) {
	sdkEval := v2.Group("/sdk/evaluations", middleware.JWTAuth())
	{
		sdkEval.GET("", handlers.SDK.ListEvaluations)
		sdkEval.GET("/experiments", handlers.SDK.ListExperiments)
		sdkEval.GET("/:id", handlers.SDK.GetEvaluation)
	}

	sdkData := v2.Group("/sdk/datasets", middleware.JWTAuth())
	{
		sdkData.GET("", handlers.SDK.ListDatasetSamples)
	}
}
