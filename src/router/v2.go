package router

import (
	v2handlers "aegis/handlers/v2"
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
func SetupV2Routes(router *gin.Engine) {
	middleware.StartCleanupRoutine()

	v2 := router.Group("/api/v2")
	// Authentication routes (with auth rate limiting)
	auth := v2.Group("/auth")
	{
		auth.POST("/login", v2handlers.Login)          // User login
		auth.POST("/register", v2handlers.Register)    // User registration
		auth.POST("/refresh", v2handlers.RefreshToken) // Token refresh

		// These require authentication
		authProtected := auth.Group("", middleware.JWTAuth())
		{
			authProtected.POST("/logout", v2handlers.Logout)                  // User logout
			authProtected.POST("/change-password", v2handlers.ChangePassword) // Change password
			authProtected.GET("/profile", v2handlers.GetProfile)              // Get current user profile
		}
	}

	// Authentication and Authorization API Group (partially implemented, others for future expansion)
	roles := v2.Group("/roles", middleware.JWTAuth()) // Role Management - Role Entity
	{
		permissions := roles.Group("/:id/permissions") // Role-Permission relationships
		{
			permissions.POST("", v2handlers.AssignPermissionsToRole)     // Assign permissions to role
			permissions.DELETE("", v2handlers.RemovePermissionsFromRole) // Remove permissions from role
		}

		users := roles.Group("/:id/users") // Role-User relationships
		{
			users.GET("", v2handlers.ListUsersFromRole) // List users with this role
		}

		roleRead := roles.Group("", middleware.RequireRoleRead)
		{
			roleRead.GET("/:id", v2handlers.GetRole)         // Get role by ID
			roleRead.GET("", v2handlers.ListRoles)           // List roles
			roleRead.POST("/search", v2handlers.SearchRoles) // Search roles
		}

		// Write operations (admin only)
		roleWrite := roles.Group("", middleware.RequireRoleWrite)
		{
			roleWrite.POST("", v2handlers.CreateRole)      // Create role
			roleWrite.PATCH("/:id", v2handlers.UpdateRole) // Update role
		}

		// Delete operations (admin only)
		roles.DELETE("/:id", middleware.RequireRoleDelete, v2handlers.DeleteRole) // Delete role
	}

	users := v2.Group("/users", middleware.JWTAuth()) // User Management - User Entity
	{
		roles := users.Group("/:id/roles") // User-Role relationships
		{
			roles.POST("", v2handlers.AssignGlobalRole)            // Assign role to user
			roles.DELETE("/:role_id", v2handlers.RemoveGlobalRole) // Remove role from user
		}

		projects := users.Group("/:id/projects") // User-Project relationships
		{
			projects.POST("", v2handlers.AssignUserToProject)                 // Assign user to project
			projects.DELETE("/:project_id", v2handlers.RemoveUserFromProject) // Remove user from project
		}

		userRead := users.Group("", middleware.RequireUserRead)
		{
			userRead.GET("", middleware.RequireUserRead, v2handlers.ListUsers)               // List users
			userRead.GET("/:id", middleware.RequireAdminOrUserOwnership, v2handlers.GetUser) // Get user by ID
			userRead.POST("/search", middleware.RequireUserRead, v2handlers.SearchUsers)     // Search users
		}

		// Write operations
		userWrite := users.Group("", middleware.RequireUserWrite)
		{
			userWrite.POST("", v2handlers.CreateUser)    // Create user
			userWrite.PUT("/:id", v2handlers.UpdateUser) // Update user
		}

		// Delete operations (admin only)
		users.DELETE("/:id", middleware.RequireUserDelete, v2handlers.DeleteUser) // Delete user
	}

	permissions := v2.Group("/permissions", middleware.JWTAuth()) // Permission Management - Permission Entity
	{
		roles := permissions.Group("/:id/roles") // Permission-Role relationships
		{
			roles.GET("", v2handlers.ListPermissionRoles) // List roles assigned to permission
		}

		permRead := permissions.Group("", middleware.RequirePermissionRead)
		{
			permRead.GET("", v2handlers.ListPermissions)           // List permissions
			permRead.GET("/:id", v2handlers.GetPermission)         // Get permission by ID
			permRead.POST("/search", v2handlers.SearchPermissions) // Search permissions
		}

		// Write operations (admin only)
		permWrite := permissions.Group("", middleware.RequirePermissionWrite)
		{
			permWrite.POST("", v2handlers.CreatePermission)      // Create permission
			permWrite.PATCH("/:id", v2handlers.UpdatePermission) // Update permission
		}

		// Delete operations (admin only)
		permissions.DELETE("/:id", middleware.RequirePermissionDelete, v2handlers.DeletePermission) // Delete permission
	}

	resources := v2.Group("/resources") // Resource Management - Resource Entity
	{
		permissions := resources.Group("/:id/permissions") // Resource-Permission relationships
		{
			permissions.GET("", v2handlers.ListResourcePermissions) // List permissions assigned to resource
		}
	}

	// Core Business Entity API Group

	// Task Management - Task Entity
	tasks := v2.Group("/tasks", middleware.JWTAuth())
	{
		// Read operations
		// GET /api/v2/tasks?page=1&size=20&task_type=RestartService&status=Completed
		tasks.GET("", middleware.RequireTaskRead, v2handlers.ListTasks)

		// GET /api/v2/tasks/{id}?include=logs
		tasks.GET("/:id", middleware.RequireTaskRead, v2handlers.GetTask)

		// POST /api/v2/tasks/search - Advanced search with complex filters
		tasks.POST("/search", middleware.RequireTaskRead, v2handlers.SearchTasks)

		// POST /api/v2/tasks/queue - Get tasks in ready/delayed queues (admin only for system-wide view)
		tasks.POST("/queue", middleware.RequireSystemRead, v2handlers.GetQueuedTasks)
	}

	// Container Management - Container Entity
	containers := v2.Group("/containers", middleware.JWTAuth())
	{
		versions := containers.Group("/:container_id/versions")
		{
			versions.GET("/:version_id", middleware.RequireContainerVersionRead, v2handlers.GetContainerVersion)
			versions.POST("", middleware.RequireContainerVersionWrite, v2handlers.CreateContainerVersion)
			versions.PATCH("/:version_id", middleware.RequireContainerVersionWrite, v2handlers.UpdateContainerVersion)
			versions.DELETE("/:version_id", middleware.RequireContainerVersionDelete, v2handlers.DeleteContainerVersion)
		}

		containerRead := containers.Group("", middleware.RequireContainerRead)
		{
			containerRead.GET("/:container_id", v2handlers.GetContainer) // Get container by ID
			containerRead.GET("", v2handlers.ListContainers)             // List containers
			containerRead.POST("/search", v2handlers.SearchContainers)   // Advanced search
		}

		containerWrite := containers.Group("", middleware.RequireContainerWrite)
		{
			containerWrite.POST("", v2handlers.CreateContainer)
			containerWrite.PATCH("/:container_id", middleware.RequireContainerWrite, v2handlers.UpdateContainer)
		}

		containers.POST("/build", v2handlers.BuildContainer)
		containers.DELETE("/:container_id", middleware.RequireContainerDelete, v2handlers.DeleteContainer)
	}

	// Algorithm Management - Algorithms (Algorithm is a special type of container)
	algorithms := v2.Group("/algorithms", middleware.JWTAuth())
	{
		// Read operations
		// GET /api/v2/algorithms?page=1&size=10 - Only active algorithms with type=algorithm
		algorithms.GET("", middleware.RequireContainerRead, v2handlers.ListAlgorithms)

		// POST /api/v2/algorithms/search - Advanced search for algorithms (containers with type=algorithm)
		algorithms.POST("/search", middleware.RequireContainerRead, v2handlers.SearchAlgorithms)

		// Algorithm execution operations
		// POST /api/v2/algorithms/execute - Submit single algorithm execution (supports both datapack and dataset)
		algorithms.POST("/execute", middleware.RequireContainerWrite, v2handlers.SubmitAlgorithmExecution)

		// Algorithm result upload operations
		// POST /api/v2/algorithms/{algorithm_id}/executions/{execution_id}/detectors - Upload detector results
		algorithms.POST("/:algorithm_id/executions/:execution_id/detectors", middleware.RequireContainerWrite, v2handlers.UploadDetectorResults)

		// POST /api/v2/algorithms/{algorithm_id}/results - Upload granularity results (supports auto-execution creation via query param)
		algorithms.POST("/:algorithm_id/results", middleware.RequireContainerWrite, v2handlers.UploadGranularityResults)
	}

	// Other Business Entity API Group
	injections := v2.Group("/injections", middleware.JWTAuth()) // Fault Injection Management - FaultInjectionSchedule Entity
	{
		// Create operations
		injections.POST("", middleware.RequireFaultInjectionWrite, v2handlers.CreateInjection) // Create injections (batch supported)

		// Read operations
		injections.GET("", middleware.RequireFaultInjectionRead, v2handlers.ListInjections)           // List injections
		injections.GET("/:id", middleware.RequireFaultInjectionRead, v2handlers.GetInjection)         // Get injection by ID
		injections.POST("/search", middleware.RequireFaultInjectionRead, v2handlers.SearchInjections) // Advanced search

		// Write operations
		injections.PUT("/:id", middleware.RequireFaultInjectionWrite, v2handlers.UpdateInjection)                        // Update injection
		injections.PATCH("/:name/tags", middleware.RequireFaultInjectionWrite, v2handlers.ManageInjectionTags)           // Manage injection labels
		injections.PATCH("/:name/labels", middleware.RequireFaultInjectionWrite, v2handlers.ManageInjectionCustomLabels) // Manage injection custom labels
		injections.DELETE("/:id", middleware.RequireFaultInjectionDelete, v2handlers.DeleteInjection)                    // Delete injection (soft delete)
		injections.POST("/batch-delete", middleware.RequireFaultInjectionDelete, v2handlers.BatchDeleteInjections)       // Batch delete injections

	}

	// Dataset Management - Dataset Entity
	datasets := v2.Group("/datasets", middleware.JWTAuth())
	{
		datasets.GET("", middleware.RequireDatasetRead, v2handlers.ListDatasets)
		datasets.GET("/:id", middleware.RequireDatasetRead, v2handlers.GetDataset)
		datasets.GET("/:id/download", middleware.RequireDatasetRead, v2handlers.DownloadDataset)
		datasets.POST("/search", middleware.RequireDatasetRead, v2handlers.SearchDatasets)
		datasets.POST("", middleware.RequireDatasetWrite, v2handlers.CreateDataset)
		datasets.PUT("/:id", middleware.RequireDatasetWrite, v2handlers.UpdateDataset)
		datasets.PATCH("/:id/injections", middleware.RequireDatasetWrite, v2handlers.ManageDatasetInjections)
		datasets.PATCH("/:id/labels", middleware.RequireDatasetWrite, v2handlers.ManageDatasetLabels)
		datasets.DELETE("/:id", middleware.RequireDatasetDelete, v2handlers.DeleteDataset)
	}

	// Execution Result Management - ExecutionResult Entity
	executions := v2.Group("/executions")
	labels := v2.Group("/labels") // Label Management - Label Entity
	{
		labels.POST("", v2handlers.CreateLabels)
	}

	// Project Management - Project Entity
	projects := v2.Group("/projects", middleware.JWTAuth())
	{
		projects.GET("/:id", v2handlers.GetProject)
	}

	// Evaluation API Group
	evaluations := v2.Group("/evaluations", middleware.JWTAuth())
	{
		// GET /api/v2/evaluations/label-keys - Get available label keys for filtering (requires system read permission)
		evaluations.GET("/label-keys", middleware.RequireDatasetRead, v2handlers.GetAvailableLabelKeys)

		// POST /api/v2/evaluations/datasets - Get algorithm evaluations on multiple datasets (requires system read permission)
		evaluations.POST("/datasets", middleware.RequireDatasetRead, v2handlers.GetDatasetEvaluationResults)

		// POST /api/v2/evaluations/datapacks - Get algorithm evaluations on multiple datapacks (requires system read permission)
		evaluations.POST("/datapacks", middleware.RequireDatasetRead, v2handlers.GetDatapackEvaluationResults)

		// POST /api/v2/evaluations/datapacks/detector - Get detector results for multiple datapacks (requires system read permission)
		evaluations.POST("/datapacks/detector", middleware.RequireDatasetRead, v2handlers.GetDatapackDetectorResults)
	}

	// Trace Management
	traces := v2.Group("/traces", middleware.JWTAuth())
	{
		traces.GET("/:id/stream", v2handlers.GetTraceStream)
	}

	// Analysis and Detection related API Group (for future expansion)
	detectors := v2.Group("/detectors")     // Detector Management - Detector Entity
	granularity := v2.Group("/granularity") // Granularity Result Management - GranularityResult Entity
	analyzer := v2.Group("/analyzer")       // Analyzer related

	// Temporarily use empty assignment to avoid compilation errors, specific routes will be implemented gradually later
	_ = injections
	_ = executions
	_ = labels
	_ = projects
	_ = resources
	_ = detectors
	_ = granularity
	_ = traces
	_ = analyzer
}
