package router

import (
	"aegis/middleware"

	"github.com/gin-gonic/gin"
)

/*
===================================================================================
API v2 Design Specification - RESTful API Standard
===================================================================================

v2 API strictly adheres to RESTful design principles, contrasting with the disorganized design of v1.
v1 API design was rather arbitrary, with non-standard methods and paths. v2 will uniformly follow the standards below.

📋 HTTP Method Usage Specification:
- GET    : Query resources (idempotent, cacheable)
- POST   : Create resources / Complex queries (non-idempotent)
- PUT    : Full update of resources (idempotent)
- PATCH  : Partial update of resources (idempotent)
- DELETE : Delete resources (idempotent)

🎯 URL Design Specification:
1. Resource names use plural form
   ✅ GET /api/v2/users          ❌ GET /api/v2/user
   ✅ GET /api/v2/projects       ❌ GET /api/v2/project

2. Clear hierarchical relationships
   ✅ GET /api/v2/users/{id}/projects
   ✅ GET /api/v2/projects/{id}/members

3. Query parameter specification
   ✅ GET /api/v2/users?page=1&size=10&status=active
   ✅ GET /api/v2/tasks?project_id=123&type=injection

📊 Standard CRUD Operation Modes:
- GET    /api/v2/{resource}           # List query (supports pagination, filtering, sorting)
- POST   /api/v2/{resource}           # Create resource
- GET    /api/v2/{resource}/{id}      # Get single resource details
- PUT    /api/v2/{resource}/{id}      # Full update of resource
- PATCH  : Partial update of resource (idempotent)
- DELETE : Delete resource (idempotent)

🔍 Complex Query Handling:
For complex search conditions, use dedicated search endpoints:
- POST /api/v2/{resource}/search      # Complex condition search
- POST /api/v2/{resource}/query       # Advanced query
- POST /api/v2/{resource}/batch       # Batch operations

🎨 Business Operation Endpoints:
Semantic business operations use verb forms:
- POST /api/v2/users/{id}/activate    # Activate user
- POST /api/v2/tasks/{id}/cancel      # Cancel task
- POST /api/v2/injections/{id}/start  # Start fault injection
- POST /api/v2/containers/{id}/build  # Build container

📨 Response Format Specification:
1. Successful Response:
   {
     "code": 200,
     "message": "success",
     "data": {...},
     "timestamp": "2024-01-01T12:00:00Z"
   }

2. List Response:
   {
     "code": 200,
     "message": "success",
     "data": {
       "items": [...],
       "pagination": {
         "page": 1,
         "size": 10,
         "total": 100,
         "pages": 10
       }
     }
   }

3. Error Response:
   {
     "code": 400,
     "message": "validation failed",
     "errors": ["field xxx is required"],
     "timestamp": "2024-01-01T12:00:00Z"
   }

🔐 Authentication and Authorization Specification:
- Use JWT Bearer Token authentication
- Permission checks based on RBAC model
- Sensitive operations require secondary confirmation

⚡ Performance Optimization:
- GET requests support ETag caching
- List queries default to pagination (page=1, size=20)
- Supports field selection ?fields=id,name,status
- Supports associated queries ?include=project,labels

🔄 Version Compatibility:
- v2 API ensures backward compatibility
- Deprecated endpoints provide a 6-month transition period
- Major changes handled by new version numbers

Note: v1 API design is chaotic and does not follow a unified standard. It will gradually migrate to v2 specification later.
===================================================================================
*/

// SetupV2Routes sets up API v2 routes - stable version of the API
func SetupV2Routes(router *gin.Engine, handlers *Handlers) {
	middleware.StartCleanupRoutine()

	v2 := router.Group("/api/v2")
	SetupPublicV2Routes(v2, handlers)
	SetupSDKV2Routes(v2, handlers)
	SetupAdminV2Routes(v2, handlers)
	SetupPortalV2Routes(v2, handlers)
	SetupSystemV2Routes(v2, handlers)

	// =====================================================================
	// Admin Entity API Group
	// =====================================================================

	// Container Management - Container Entity
	containers := v2.Group("/containers", middleware.JWTAuth())
	{
		// Container Version sub-resource routes
		versions := containers.Group("/:container_id/versions")
		{

			// Container Version Read operations
			versionRead := versions.Group("", middleware.RequireContainerVersionRead)
			{
				versionRead.GET("/:version_id", handlers.Container.GetContainerVersion) // Get container version by ID
				versionRead.GET("", handlers.Container.ListContainerVersions)           // List container versions
			}

			// Container Version Create operations
			versions.POST("", middleware.RequireContainerVersionCreate, handlers.Container.CreateContainerVersion) // Create container version

			// Container Version Upload operations
			versions.POST("/:version_id/helm-chart", middleware.RequireContainerVersionUpload, handlers.Container.UploadHelmChart)      // Upload Helm chart tgz file
			versions.POST("/:version_id/helm-values", middleware.RequireContainerVersionUpload, handlers.Container.UploadHelmValueFile) // Upload Helm values file

			// Container Version Update operations
			versions.PATCH("/:version_id", middleware.RequireContainerVersionUpdate, handlers.Container.UpdateContainerVersion) // Update container version

			// Container Version Delete operations
			versions.DELETE("/:version_id", middleware.RequireContainerVersionDelete, handlers.Container.DeleteContainerVersion)
		}

		// Container Read operations
		containerRead := containers.Group("", middleware.RequireContainerRead)
		{
			containerRead.GET("/:container_id", handlers.Container.GetContainer) // Get container by ID
			containerRead.GET("", handlers.Container.ListContainers)             // List containers
		}

		// Container Create operations
		containers.POST("", middleware.RequireContainerCreate, handlers.Container.CreateContainer) // Create container

		// Container Execute operations (build requires execute permission)
		containers.POST("/build", middleware.RequireContainerExecute, handlers.Container.SubmitContainerBuilding) // Build container

		// Container Update operations
		containers.PATCH("/:container_id", middleware.RequireContainerUpdate, handlers.Container.UpdateContainer)                    // Update container
		containers.PATCH("/:container_id/labels", middleware.RequireContainerUpdate, handlers.Container.ManageContainerCustomLabels) // Manage container labels

		// Container Delete operations
		containers.DELETE("/:container_id", middleware.RequireContainerDelete, handlers.Container.DeleteContainer) // Delete container
	}

	// Dataset Management - Dataset Entity
	datasets := v2.Group("/datasets", middleware.JWTAuth())
	{
		// Dataset Version sub-resource routes
		versions := datasets.Group("/:dataset_id/versions")
		{
			versionRead := versions.Group("", middleware.RequireDatasetVersionRead)
			{
				versionRead.GET("", handlers.Dataset.ListDatasetVersions)                         // List dataset versions
				versionRead.GET("/:version_id", handlers.Dataset.GetDatasetVersion)               // Get dataset version by ID
				versionRead.GET("/:version_id/download", handlers.Dataset.DownloadDatasetVersion) // Download dataset version
			}

			// Dataset Version Create operations
			versions.POST("", middleware.RequireDatasetVersionCreate, handlers.Dataset.CreateDatasetVersion) // Create dataset version

			// Dataset Version Update operations
			versions.PATCH("/:version_id", middleware.RequireDatasetVersionUpdate, handlers.Dataset.UpdateDatasetVersion)                      // Update dataset version
			versions.PATCH("/:version_id/injections", middleware.RequireDatasetVersionUpdate, handlers.Dataset.ManageDatasetVersionInjections) // Manage dataset version injections

			versions.DELETE("/:version_id", middleware.RequireDatasetVersionDelete, handlers.Dataset.DeleteDatasetVersion) // Delete dataset version
		}

		// Dataset Read operations
		datasetRead := datasets.Group("", middleware.RequireDatasetRead)
		{
			datasetRead.GET("/:dataset_id", handlers.Dataset.GetDataset) // Get dataset by ID
			datasetRead.GET("", handlers.Dataset.ListDatasets)           // List datasets
		}

		// Dataset Create operations
		datasets.POST("", middleware.RequireDatasetCreate, handlers.Dataset.CreateDataset) // Create dataset

		// Dataset Update operations
		datasets.PATCH("/:dataset_id", middleware.RequireDatasetUpdate, handlers.Dataset.UpdateDataset)                    // Update dataset
		datasets.PATCH("/:dataset_id/labels", middleware.RequireDatasetUpdate, handlers.Dataset.ManageDatasetCustomLabels) // Manage dataset labels

		// Dataset Delete operations
		datasets.DELETE("/:dataset_id", middleware.RequireDatasetDelete, handlers.Dataset.DeleteDataset) // Delete dataset
	}

	// =====================================================================
	// Core Business Entity API Group
	// =====================================================================

	// Task Management - Task Entity
	tasks := v2.Group("/tasks")
	{
		taskWithAuth := tasks.Group("", middleware.JWTAuth())
		{

			// Task Read operations
			taskRead := taskWithAuth.Group("", middleware.RequireTaskRead)
			{
				taskRead.GET("", handlers.Task.List)         // List tasks
				taskRead.GET("/:task_id", handlers.Task.Get) // Get task by ID
			}

			// Task Delete operations
			taskWithAuth.POST("/batch-delete", middleware.RequireTaskDelete, handlers.Task.BatchDelete) // Batch delete tasks
		}

		// Task Log streaming (WebSocket) - auth via query param, not middleware
		tasks.GET("/:task_id/logs/ws", handlers.Task.LogsWS) // Stream task logs via WebSocket
	}

	// Fault Injection Management - FaultInjectionSchedule Entity
	// Note: These global routes are for system admins only. Regular users should access injections via /projects/:project_id
	injections := v2.Group("/injections", middleware.JWTAuth())
	{
		injectionSystemAdmin := injections.Group("", middleware.RequireSystemAdmin())
		{
			injectionSystemAdmin.GET("", handlers.Injection.ListInjections)           // List injections
			injectionSystemAdmin.POST("/search", handlers.Injection.SearchInjections) // Advanced search
		}

		// Manual upload (must be before /:id routes)
		injections.POST("/upload", handlers.Injection.UploadDatapack) // Upload manual datapack

		// Injection Read operations
		injections.GET("/:id", handlers.Injection.GetInjection)                        // Get injection by ID
		injections.GET("/:id/download", handlers.Injection.DownloadDatapack)           // Download injection datapack
		injections.GET("/:id/logs", handlers.Injection.GetInjectionLogs)               // Get injection execution logs
		injections.GET("/:id/files", handlers.Injection.ListDatapackFiles)             // Get injection file structure
		injections.GET("/:id/files/download", handlers.Injection.DownloadDatapackFile) // Download specific injection file
		injections.GET("/:id/files/query", handlers.Injection.QueryDatapackFile)       // Query parquet file content
		injections.GET("/metadata", handlers.Injection.GetInjectionMetadata)           // Get injection metadata

		// Injection Clone operations
		injections.POST("/:id/clone", handlers.Injection.CloneInjection) // Clone injection

		// Injection Update operations (label management, ground truth)
		injections.PUT("/:id/groundtruth", handlers.Injection.UpdateGroundtruth)         // Update ground truth
		injections.PATCH("/:id/labels", handlers.Injection.ManageInjectionCustomLabels)  // Manage injection custom labels
		injections.PATCH("/labels/batch", handlers.Injection.BatchManageInjectionLabels) // Batch manage injection labels

		// Injection Delete operations
		injections.POST("/batch-delete", handlers.Injection.BatchDeleteInjections) // Batch delete injections
	}

	// Execution Result Management - ExecutionResult Entity
	// Note: These global routes are for system admins only. Regular users should access executions via /projects/:project_id
	executions := v2.Group("/executions", middleware.JWTAuth())
	{
		executionSystemAdmin := executions.Group("", middleware.RequireSystemAdmin())
		{
			executionSystemAdmin.GET("", handlers.Execution.ListExecutions)                      // List executions
			executionSystemAdmin.GET("/labels", handlers.Execution.ListAvailableExecutionLabels) // List available execution labels
		}

		// Execution Read operations
		executions.GET("/:execution_id", handlers.Execution.GetExecution) // Get execution by ID

		// Execution Update operations (upload results and manage labels)
		executions.POST("/:execution_id/detector_results", handlers.Execution.UploadDetectorResults)       // Upload detector results
		executions.POST("/:execution_id/granularity_results", handlers.Execution.UploadGranularityResults) // Upload granularity results
		executions.PATCH("/:execution_id/labels", handlers.Execution.ManageExecutionCustomLabels)          // Manage execution custom labels

		// Execution Delete operations
		executions.POST("/batch-delete", handlers.Execution.BatchDeleteExecutions) // Batch delete executions
	}

	// Trace Management - Trace Entity
	traces := v2.Group("/traces", middleware.JWTAuth())
	{
		traces.GET("", handlers.Trace.ListTraces)                      // List traces
		traces.GET("/:trace_id", handlers.Trace.GetTrace)              // Get trace by ID
		traces.GET("/:trace_id/stream", handlers.Trace.GetTraceStream) // Get trace stream (SSE)
	}

	// Group Management - Group stream for real-time batch progress
	groups := v2.Group("/groups", middleware.JWTAuth())
	{
		groups.GET("/:group_id/stats", handlers.Group.GetGroupStats)   // Get group stats (can be used for progress tracking)
		groups.GET("/:group_id/stream", handlers.Group.GetGroupStream) // Stream group trace events (SSE)
	}

	// =====================================================================
	// Notification API Group
	// =====================================================================

	// Notification Management - Global workflow notifications
	notifications := v2.Group("/notifications", middleware.JWTAuth())
	{
		notifications.GET("/stream", handlers.Notification.GetStream) // Stream global notifications (SSE)
	}

	// =====================================================================
	// Analyzer Service API Group
	// =====================================================================

	// Analyzer related routes (placeholder for future expansion)
	analyzer := v2.Group("/analyzer", middleware.JWTAuth())
	_ = analyzer // Temporarily unused to avoid compilation errors

	// =====================================================================
	// Evaluation API Group
	// =====================================================================

	// Evaluation API Group
	evaluations := v2.Group("/evaluations", middleware.JWTAuth())
	{
		// GET /api/v2/evaluations - List persisted evaluations with pagination
		evaluations.GET("", handlers.Evaluation.ListEvaluations)

		// GET /api/v2/evaluations/:id - Get a single evaluation by ID
		evaluations.GET("/:id", handlers.Evaluation.GetEvaluation)

		// DELETE /api/v2/evaluations/:id - Delete an evaluation by ID
		evaluations.DELETE("/:id", handlers.Evaluation.DeleteEvaluation)

		// POST /api/v2/evaluations/datasets - Get algorithm evaluations on multiple datasets (requires dataset read permission)
		evaluations.POST("/datasets", middleware.RequireDatasetRead, handlers.Evaluation.ListDatasetEvaluationResults)

		// POST /api/v2/evaluations/datapacks - Get algorithm evaluations on multiple datapacks (requires dataset read permission)
		evaluations.POST("/datapacks", middleware.RequireDatasetRead, handlers.Evaluation.ListDatapackEvaluationResults)
	}

	// =====================================================================
	// Metrics API Group
	// =====================================================================

	// Metrics routes
	metrics := v2.Group("/metrics", middleware.JWTAuth())
	{
		metrics.GET("/injections", handlers.Metric.GetInjectionMetrics) // Get injection metrics
		metrics.GET("/executions", handlers.Metric.GetExecutionMetrics) // Get execution metrics
		metrics.GET("/algorithms", handlers.Metric.GetAlgorithmMetrics) // Get algorithm comparison metrics
	}

}
