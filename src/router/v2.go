package router

import (
	v2handlers "github.com/LGU-SE-Internal/rcabench/handlers/v2"
	"github.com/LGU-SE-Internal/rcabench/middleware"
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

	// Start rate limiting cleanup routine
	middleware.StartCleanupRoutine()

	v2 := router.Group("/api/v2")
	{
		// Apply general rate limiting to all v2 routes
		v2.Use(middleware.GeneralRateLimit)
	}

	// Authentication routes (with auth rate limiting)
	auth := v2.Group("/auth")
	{
		auth.Use(middleware.AuthRateLimit)             // Special rate limiting for auth
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

	// System management routes
	health := v2.Group("/health")
	{
		health.GET("", v2handlers.GetHealth) // System health check (no auth required)
	}

	statistics := v2.Group("/statistics", middleware.JWTAuth(), middleware.RequireSystemRead)
	{
		statistics.GET("", v2handlers.GetStatistics) // System statistics
	}

	audit := v2.Group("/audit", middleware.JWTAuth(), middleware.RequireAuditRead)
	{
		audit.GET("", v2handlers.ListAuditLogs)   // List audit logs
		audit.GET("/:id", v2handlers.GetAuditLog) // Get audit log by ID

		// Admin only
		auditAdmin := audit.Group("", middleware.RequireSystemAdmin)
		{
			auditAdmin.POST("", v2handlers.CreateAuditLog) // Create audit log (internal)
		}
	}

	monitor := v2.Group("/monitor", middleware.JWTAuth(), middleware.RequireSystemRead)
	{
		monitor.POST("/metrics", v2handlers.GetMetrics) // Query monitoring metrics
		monitor.GET("/info", v2handlers.GetSystemInfo)  // Get system information
	}

	// Relation management routes
	relations := v2.Group("/relations", middleware.JWTAuth(), middleware.StrictRateLimit)
	{
		relations.GET("", middleware.RequireSystemRead, v2handlers.ListRelations)                    // List relationships
		relations.GET("/statistics", middleware.RequireSystemRead, v2handlers.GetRelationStatistics) // Relationship statistics

		// Admin operations
		adminRelations := relations.Group("", middleware.RequireSystemAdmin)
		{
			adminRelations.POST("/batch", v2handlers.BatchRelationOperations) // Batch operations

			// User-Role relationships
			adminRelations.POST("/user-roles", v2handlers.AssignUserRole)   // Assign role to user
			adminRelations.DELETE("/user-roles", v2handlers.RemoveUserRole) // Remove role from user

			// Role-Permission relationships
			adminRelations.POST("/role-permissions", v2handlers.AssignRolePermissions)   // Assign permissions to role
			adminRelations.DELETE("/role-permissions", v2handlers.RemoveRolePermissions) // Remove permissions from role

			// User-Permission relationships (direct)
			adminRelations.POST("/user-permissions", v2handlers.AssignUserPermission)   // Assign permission to user
			adminRelations.DELETE("/user-permissions", v2handlers.RemoveUserPermission) // Remove permission from user
		}
	}

	// Authentication and Authorization API Group (partially implemented, others for future expansion)
	roles := v2.Group("/roles", middleware.JWTAuth()) // Role Management - Role Entity
	{
		roles.GET("", middleware.RequireRoleRead, v2handlers.ListRoles)              // List roles
		roles.GET("/:id", middleware.RequireRoleRead, v2handlers.GetRole)            // Get role by ID
		roles.GET("/:id/users", middleware.RequireRoleRead, v2handlers.GetRoleUsers) // Get users with this role
		roles.POST("/search", middleware.RequireRoleRead, v2handlers.SearchRoles)    // Search roles

		// Write operations (admin only)
		roleWrite := roles.Group("", middleware.RequireRoleWrite)
		{
			roleWrite.POST("", v2handlers.CreateRole)                                  // Create role
			roleWrite.PUT("/:id", v2handlers.UpdateRole)                               // Update role
			roleWrite.POST("/:id/permissions", v2handlers.AssignPermissionsToRole)     // Assign permissions to role
			roleWrite.DELETE("/:id/permissions", v2handlers.RemovePermissionsFromRole) // Remove permissions from role
		}

		// Delete operations (admin only)
		roles.DELETE("/:id", middleware.RequireRoleDelete, v2handlers.DeleteRole) // Delete role
	}

	users := v2.Group("/users", middleware.JWTAuth()) // User Management - User Entity
	{
		users.GET("", middleware.RequireUserRead, v2handlers.ListUsers)               // List users
		users.GET("/:id", middleware.RequireAdminOrUserOwnership, v2handlers.GetUser) // Get user by ID
		users.POST("/search", middleware.RequireUserRead, v2handlers.SearchUsers)     // Search users

		// Write operations
		userWrite := users.Group("", middleware.RequireUserWrite)
		{
			userWrite.POST("", v2handlers.CreateUser)                                       // Create user
			userWrite.PUT("/:id", v2handlers.UpdateUser)                                    // Update user
			userWrite.POST("/:id/projects", v2handlers.AssignUserToProject)                 // Assign user to project
			userWrite.DELETE("/:id/projects/:project_id", v2handlers.RemoveUserFromProject) // Remove user from project
		}

		// Delete operations (admin only)
		users.DELETE("/:id", middleware.RequireUserDelete, v2handlers.DeleteUser) // Delete user
	}

	permissions := v2.Group("/permissions", middleware.JWTAuth(), middleware.RequirePermissionRead) // Permission Management - Permission Entity
	{
		permissions.GET("", v2handlers.ListPermissions)                                // List permissions
		permissions.GET("/:id", v2handlers.GetPermission)                              // Get permission by ID
		permissions.POST("/search", v2handlers.SearchPermissions)                      // Search permissions
		permissions.GET("/:id/roles", v2handlers.GetPermissionRoles)                   // Get roles with this permission
		permissions.GET("/resource/:resource_id", v2handlers.GetPermissionsByResource) // Get permissions by resource

		// Write operations (admin only)
		permWrite := permissions.Group("", middleware.RequirePermissionWrite)
		{
			permWrite.POST("", v2handlers.CreatePermission)    // Create permission
			permWrite.PUT("/:id", v2handlers.UpdatePermission) // Update permission
		}

		// Delete operations (admin only)
		permissions.DELETE("/:id", middleware.RequirePermissionDelete, v2handlers.DeletePermission) // Delete permission
	}

	resources := v2.Group("/resources") // Resource Management - Resource Entity

	// Core Business Entity API Group

	// Task Management - Task Entity
	tasks := v2.Group("/tasks", middleware.JWTAuth())
	{
		// Read operations - permission checked in handler
		// GET /api/v2/tasks?page=1&size=20&task_type=RestartService&status=Completed
		tasks.GET("", v2handlers.ListTasks)

		// GET /api/v2/tasks/{id}?include=logs - permission checked in handler
		tasks.GET("/:id", v2handlers.GetTask)

		// POST /api/v2/tasks/search - Advanced search with complex filters - permission checked in handler
		tasks.POST("/search", v2handlers.SearchTasks)

		// POST /api/v2/tasks/queue - Get tasks in ready/delayed queues (admin only for system-wide view)
		tasks.POST("/queue", middleware.RequireSystemRead, v2handlers.GetQueuedTasks)
	}

	// Container Management - Container Entity
	containers := v2.Group("/containers", middleware.JWTAuth())
	{
		// Create operation - permission checked in handler
		// POST /api/v2/containers - Create a new container
		containers.POST("", v2handlers.CreateContainer)

		// Read operations - permission checked in handler
		// GET /api/v2/containers?page=1&size=20&type=algorithm&status=true
		containers.GET("", v2handlers.ListContainers)

		// GET /api/v2/containers/{id} - permission checked in handler
		containers.GET("/:id", v2handlers.GetContainer)

		// POST /api/v2/containers/search - Advanced search with complex filters - permission checked in handler
		containers.POST("/search", v2handlers.SearchContainers)
	}

	// Algorithm Management - Algorithms (Algorithm is a special type of container)
	algorithms := v2.Group("/algorithms", middleware.JWTAuth())
	{
		// Read operations - permission checked in handler
		// GET /api/v2/algorithms?page=1&size=10 - Only active algorithms with type=algorithm
		algorithms.GET("", v2handlers.ListAlgorithms)

		// POST /api/v2/algorithms/search - Advanced search for algorithms (containers with type=algorithm) - permission checked in handler
		algorithms.POST("/search", v2handlers.SearchAlgorithms)

		// Algorithm result upload operations - permission checked in handler
		// POST /api/v2/algorithms/{algorithm_id}/executions/{execution_id}/detectors - Upload detector results
		algorithms.POST("/:algorithm_id/executions/:execution_id/detectors", v2handlers.UploadDetectorResults)

		// POST /api/v2/algorithms/{algorithm_id}/results - Upload granularity results (supports auto-execution creation)
		algorithms.POST("/:algorithm_id/results", v2handlers.UploadGranularityResults)

		// POST /api/v2/algorithms/{algorithm_id}/executions/{execution_id}/results - Upload granularity results to existing execution
		algorithms.POST("/:algorithm_id/executions/:execution_id/results", v2handlers.UploadGranularityResults)
	}

	// Other Business Entity API Group
	injections := v2.Group("/injections", middleware.JWTAuth()) // Fault Injection Management - FaultInjectionSchedule Entity
	{
		// Create operations - permission checked in handler
		injections.POST("", v2handlers.CreateInjection) // Create injections (batch supported)

		// Read operations - permission checked in handler
		injections.GET("", v2handlers.ListInjections)           // List injections
		injections.GET("/:id", v2handlers.GetInjection)         // Get injection by ID
		injections.POST("/search", v2handlers.SearchInjections) // Advanced search

		// Write operations - permission checked in handler
		injections.PUT("/:id", v2handlers.UpdateInjection)    // Update injection
		injections.DELETE("/:id", v2handlers.DeleteInjection) // Delete injection (soft delete)
	}

	// Dataset Management - Dataset Entity
	datasets := v2.Group("/datasets", middleware.JWTAuth())
	{
		datasets.GET("", v2handlers.ListDatasets)
		datasets.GET("/:id", v2handlers.GetDataset)
		datasets.POST("/search", v2handlers.SearchDatasets)
		datasets.POST("", v2handlers.CreateDataset)
		datasets.PUT("/:id", v2handlers.UpdateDataset)
		datasets.PATCH("/:id/injections", v2handlers.ManageDatasetInjections)
		datasets.PATCH("/:id/labels", v2handlers.ManageDatasetLabels)
		datasets.DELETE("/:id", v2handlers.DeleteDataset)
	}

	executions := v2.Group("/executions") // Execution Result Management - ExecutionResult Entity
	labels := v2.Group("/labels")         // Label Management - Label Entity
	projects := v2.Group("/projects")     // Project Management - Project Entity

	// Analysis and Detection related API Group (for future expansion)
	detectors := v2.Group("/detectors")     // Detector Management - Detector Entity
	granularity := v2.Group("/granularity") // Granularity Result Management - GranularityResult Entity
	traces := v2.Group("/traces")           // Trace Management - Related to TraceID
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
