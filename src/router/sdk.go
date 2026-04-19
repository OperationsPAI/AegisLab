package router

import (
	"aegis/middleware"

	"github.com/gin-gonic/gin"
)

func SetupSDKV2Routes(v2 *gin.RouterGroup, handlers *Handlers) {
	auth := v2.Group("/auth")
	{
		auth.POST("/api-key/token", handlers.Auth.ExchangeAPIKeyToken)
	}

	sdkEval := v2.Group("/sdk/evaluations", middleware.JWTAuth(), middleware.RequireAPIKeyScopesAny("sdk:*", "sdk:evaluations:*", "sdk:evaluations:read"))
	{
		sdkEval.GET("", handlers.SDK.ListEvaluations)
		sdkEval.GET("/experiments", handlers.SDK.ListExperiments)
		sdkEval.GET("/:id", handlers.SDK.GetEvaluation)
	}

	sdkData := v2.Group("/sdk/datasets", middleware.JWTAuth(), middleware.RequireAPIKeyScopesAny("sdk:*", "sdk:datasets:*", "sdk:datasets:read"))
	{
		sdkData.GET("", handlers.SDK.ListDatasetSamples)
	}

	datasets := v2.Group("/datasets", middleware.JWTAuth())
	{
		datasetVersions := datasets.Group("/:dataset_id/versions")
		{
			datasetVersions.GET("/:version_id/download", middleware.RequireDatasetVersionDownload, middleware.RequireAPIKeyScopesAny("sdk:*", "sdk:datasets:*", "sdk:datasets:read"), handlers.Dataset.DownloadDatasetVersion)
		}

		datasets.PATCH("/:dataset_id/version/:version_id/injections", middleware.RequireDatasetVersionUpdate, middleware.RequireAPIKeyScopesAny("sdk:*", "sdk:datasets:*", "sdk:datasets:write"), handlers.Dataset.ManageDatasetVersionInjections)
	}

	projects := v2.Group("/projects", middleware.JWTAuth())
	{
		injections := projects.Group("/:project_id/injections")
		{
			injectionRead := injections.Group("", middleware.RequireProjectRead)
			{
				analysis := injectionRead.Group("/analysis")
				{
					analysis.GET("/no-issues", handlers.Injection.ListProjectFaultInjectionNoIssues)
					analysis.GET("/with-issues", handlers.Injection.ListProjectFaultInjectionWithIssues)
				}

				injectionRead.GET("", handlers.Injection.ListProjectInjections)
			}

			injectionExecute := injections.Group("", middleware.RequireProjectInjectionExecute)
			{
				injectionExecute.POST("/inject", handlers.Injection.SubmitProjectFaultInjection)
				injectionExecute.POST("/build", handlers.Injection.SubmitProjectDatapackBuilding)
			}
		}

		executions := projects.Group("/:project_id/executions")
		{
			executionRead := executions.Group("", middleware.RequireProjectRead)
			{
				executionRead.GET("", handlers.Execution.ListProjectExecutions)
			}

			executionExecute := executions.Group("", middleware.RequireProjectExecutionExecute)
			{
				executionExecute.POST("/execute", handlers.Execution.SubmitAlgorithmExecution)
			}
		}
	}

	evaluations := v2.Group("/evaluations", middleware.JWTAuth())
	{
		evaluations.POST("/datapacks", middleware.RequireAPIKeyScopesAny("sdk:*", "sdk:evaluations:*", "sdk:evaluations:read"), handlers.Evaluation.ListDatapackEvaluationResults)
		evaluations.POST("/datasets", middleware.RequireAPIKeyScopesAny("sdk:*", "sdk:evaluations:*", "sdk:evaluations:read"), handlers.Evaluation.ListDatasetEvaluationResults)
		evaluations.GET("", middleware.RequireAPIKeyScopesAny("sdk:*", "sdk:evaluations:*", "sdk:evaluations:read"), handlers.Evaluation.ListEvaluations)
		evaluations.GET("/:id", middleware.RequireAPIKeyScopesAny("sdk:*", "sdk:evaluations:*", "sdk:evaluations:read"), handlers.Evaluation.GetEvaluation)
	}

	executions := v2.Group("/executions", middleware.JWTAuth())
	{
		executions.GET("/:id", middleware.RequireAPIKeyScopesAny("sdk:*", "sdk:executions:*", "sdk:executions:read"), handlers.Execution.GetExecution)
		executions.PATCH("/:id/labels", middleware.RequireAPIKeyScopesAny("sdk:*", "sdk:executions:*", "sdk:executions:write"), handlers.Execution.ManageExecutionCustomLabels)
	}

	runtime := v2.Group("/executions", middleware.RequireServiceTokenAuth())
	{
		runtime.POST("/:execution_id/detector_results", handlers.Execution.UploadDetectorResults)
		runtime.POST("/:execution_id/granularity_results", handlers.Execution.UploadGranularityResults)
	}

	injections := v2.Group("/injections", middleware.JWTAuth())
	{
		injections.GET("/metadata", handlers.Injection.GetInjectionMetadata)
		injections.GET("/:id", handlers.Injection.GetInjection)
		injections.POST("/:id/clone", handlers.Injection.CloneInjection)
		injections.GET("/:id/download", handlers.Injection.DownloadDatapack)
		injections.GET("/:id/files", handlers.Injection.ListDatapackFiles)
		injections.GET("/:id/files/download", handlers.Injection.DownloadDatapackFile)
		injections.GET("/:id/files/query", handlers.Injection.QueryDatapackFile)
		injections.PATCH("/:id/labels", handlers.Injection.ManageInjectionCustomLabels)
	}

	metrics := v2.Group("/metrics", middleware.JWTAuth())
	{
		metrics.GET("/algorithms", middleware.RequireAPIKeyScopesAny("sdk:*", "sdk:metrics:*", "sdk:metrics:read"), handlers.Metric.GetAlgorithmMetrics)
		metrics.GET("/executions", middleware.RequireAPIKeyScopesAny("sdk:*", "sdk:metrics:*", "sdk:metrics:read"), handlers.Metric.GetExecutionMetrics)
		metrics.GET("/injections", middleware.RequireAPIKeyScopesAny("sdk:*", "sdk:metrics:*", "sdk:metrics:read"), handlers.Metric.GetInjectionMetrics)
	}
}
