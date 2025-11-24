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
				versionRead.GET("/:version_id", v2handlers.GetContainerVersion) // Get container version by ID
				versionRead.GET("", v2handlers.ListContainerVersions)           // List container versions
			}

			// Container Version Write operations
			versionWrite := versions.Group("", middleware.RequireContainerVersionWrite)
			{
				versionWrite.POST("", v2handlers.CreateContainerVersion)              // Create container version
				versionWrite.PATCH("/:version_id", v2handlers.UpdateContainerVersion) // Update container version
			}

			// Container Version Delete operations
			versions.DELETE("/:version_id", middleware.RequireContainerVersionDelete, v2handlers.DeleteContainerVersion)
		}

		// Container Read operations
		containerRead := containers.Group("", middleware.RequireContainerRead)
		{
			containerRead.GET("/:container_id", v2handlers.GetContainer) // Get container by ID
			containerRead.GET("", v2handlers.ListContainers)             // List containers
			containerRead.POST("/search", v2handlers.SearchContainers)   // Advanced search
		}

		// Container Write operations
		containerWrite := containers.Group("", middleware.RequireContainerWrite)
		{
			containerWrite.POST("", v2handlers.CreateContainer)                                   // Create container
			containers.POST("/build", v2handlers.SubmitContainerBuilding)                         // Build container
			containerWrite.PATCH("/:container_id", v2handlers.UpdateContainer)                    // Update container
			containerWrite.PATCH("/:container_id/labels", v2handlers.ManageContainerCustomLabels) // Manage container labels
		}

		// Container Delete operations
		containers.DELETE("/:container_id", middleware.RequireContainerDelete, v2handlers.DeleteContainer) // Delete container
	}

	// Dataset Management - Dataset Entity
	datasets := v2.Group("/datasets", middleware.JWTAuth())
	{
		// Dataset Version sub-resource routes
		versions := datasets.Group("/:dataset_id/versions")
		{
			versionRead := versions.Group("", middleware.RequireDatasetVersionRead)
			{
				versionRead.GET("/:version_id", v2handlers.GetDatasetVersion) // Get dataset version by ID
				versionRead.GET("", v2handlers.ListDatasetVersions)           // List dataset versions
			}

			versionWrite := versions.Group("", middleware.RequireDatasetVersionWrite)
			{
				versionWrite.POST("", v2handlers.CreateDatasetVersion)                                   // Create dataset version
				versionWrite.PATCH("/:version_id", v2handlers.UpdateDatasetVersion)                      // Update dataset version
				versionWrite.PATCH("/:version_id/injections", v2handlers.ManageDatasetVersionInjections) // Manage dataset version injections
			}

			versions.DELETE("/:version_id", middleware.RequireDatasetVersionDelete, v2handlers.DeleteDatasetVersion) // Delete dataset version
		}

		// Dataset Read operations
		datasetRead := datasets.Group("", middleware.RequireDatasetRead)
		{
			datasetRead.GET("/:dataset_id", v2handlers.GetDataset)                      // Get dataset by ID
			datasetRead.GET("", v2handlers.ListDatasets)                                // List datasets
			datasetRead.GET("/:dataset_id/download", v2handlers.DownloadDatasetVersion) // Download dataset version
		}

		// Dataset Write operations
		datasetWrite := datasets.Group("", middleware.RequireDatasetWrite)
		{
			datasetWrite.POST("", v2handlers.CreateDataset)                                 // Create dataset
			datasetWrite.PATCH("/:dataset_id", v2handlers.UpdateDataset)                    // Update dataset
			datasetWrite.PATCH("/:dataset_id/labels", v2handlers.ManageDatasetCustomLabels) // Manage dataset labels
		}

		// Dataset Delete operations
		datasets.DELETE("/:dataset_id", middleware.RequireDatasetDelete, v2handlers.DeleteDataset) // Delete dataset
	}

	// Project Management - Project Entity
	projects := v2.Group("/projects", middleware.JWTAuth())
	{
		// Project Read operations
		projectRead := projects.Group("", middleware.RequireProjectRead)
		{
			projectRead.GET("/:project_id", v2handlers.GetProjectDetail) // Get project by ID
			projectRead.GET("", v2handlers.ListProjects)                 // List projects
		}

		// Project Write operations
		projectWrite := projects.Group("", middleware.RequireProjectWrite)
		{
			projectWrite.POST("", v2handlers.CreateProject)                                 // Create project
			projectWrite.PATCH("/:project_id", v2handlers.UpdateProject)                    // Update project
			projectWrite.PATCH("/:project_id/labels", v2handlers.ManageProjectCustomLabels) // Manage project labels
		}

		// Project Delete operations
		projects.DELETE("/:project_id", middleware.RequireProjectDelete, v2handlers.DeleteProject) // Delete project
	}

	// Label Management - Label Entity
	labels := v2.Group("/labels", middleware.JWTAuth())
	{
		// Label Read operations
		labelRead := labels.Group("", middleware.RequireLabelRead)
		{
			labelRead.GET("/:label_id", v2handlers.GetLabelDetail) // Get label by ID
			labelRead.GET("", v2handlers.ListLabels)               // List labels
		}

		// Label Write operations
		labelWrite := labels.Group("", middleware.RequireLabelWrite)
		{
			labelWrite.POST("", v2handlers.CreateLabel)            // Create label
			labelWrite.PATCH("/:label_id", v2handlers.UpdateLabel) // Update label
		}

		// Label Delete operations
		labels.DELETE("/:label_id", middleware.RequireLabelDelete, v2handlers.DeleteLabel)        // Delete label
		labels.POST("/batch-delete", middleware.RequireLabelDelete, v2handlers.BatchDeleteLabels) // Batch delete labels
	}

	// User Management - User Entity
	users := v2.Group("/users", middleware.JWTAuth())
	{
		// User-Role relationship routes
		roles := users.Group("/:user_id/roles")
		{
			roles.POST("/:role_id", middleware.RequireUserWrite, v2handlers.AssignUserRole)     // Assign role to user
			roles.DELETE("/:role_id", middleware.RequireUserWrite, v2handlers.RemoveGlobalRole) // Remove role from user
		}

		// User-Project relationship routes
		projects := users.Group("/:user_id/projects")
		{
			projects.POST("/:project_id/roles/:role_id", middleware.RequireUserWrite, v2handlers.AssignUserProject) // Assign user to project
			projects.DELETE("/:project_id", middleware.RequireUserWrite, v2handlers.RemoveUserProject)              // Remove user from project
		}

		// User-Permission relationship routes
		permissions := users.Group("/:user_id/permissions")
		{
			permissions.POST("/assign", middleware.RequireUserWrite, v2handlers.AssignUserPermission) // Assign permission to user
			permissions.POST("/remove", middleware.RequireUserWrite, v2handlers.RemoveUserPermission) // Remove permission from user
		}

		// User-Container relationship routes
		containers := users.Group("/:user_id/containers")
		{
			containers.POST("/:container_id/roles/:role_id", middleware.RequireUserWrite, v2handlers.AssignUserContainer) // Assign container to user
			containers.DELETE("/:container_id", middleware.RequireUserWrite, v2handlers.RemoveUserContainer)              // Remove container from user
		}

		// User-Dataset relationship routes
		datasets := users.Group("/:user_id/datasets")
		{
			datasets.POST("/:dataset_id/roles/:role_id", middleware.RequireUserWrite, v2handlers.AssignUserDataset) // Assign dataset to user
			datasets.DELETE("/:dataset_id", middleware.RequireUserWrite, v2handlers.RemoveUserDataset)              // Remove dataset from user
		}

		// User Read operations
		userRead := users.Group("", middleware.RequireUserRead)
		{
			userRead.GET("", v2handlers.ListUsersV2)                                                             // List users
			userRead.GET("/:user_id/detail", middleware.RequireAdminOrUserOwnership, v2handlers.GetUserDetailV2) // Get user by ID
			userRead.POST("/search", v2handlers.SearchUsers)                                                     // Search users
		}

		// User Write operations
		userWrite := users.Group("", middleware.RequireUserWrite)
		{
			userWrite.POST("", v2handlers.CreateUser)           // Create user
			userWrite.PATCH("/:user_id", v2handlers.UpdateUser) // Update user
		}

		// User Delete operations
		users.DELETE("/:user_id", middleware.RequireUserDelete, v2handlers.DeleteUser) // Delete user
	}

	// =====================================================================
	// Authentication and Authorization API Group
	// =====================================================================

	// Role Management - Role Entity
	roles := v2.Group("/roles", middleware.JWTAuth())
	{
		// Role-Permission relationship routes
		permissions := roles.Group("/:role_id/permissions")
		{
			permissions.POST("/assign", middleware.RequireRoleWrite, v2handlers.AssignRolePermission)      // Assign permissions to role
			permissions.POST("/remove", middleware.RequireRoleWrite, v2handlers.RemovePermissionsFromRole) // Remove permissions from role
		}

		// Role-User relationship routes
		users := roles.Group("/:role_id/users")
		{
			users.GET("", middleware.RequireRoleRead, v2handlers.ListUsersFromRole) // List users with this role
		}

		// Role Read operations
		roleRead := roles.Group("", middleware.RequireRoleRead)
		{
			roleRead.GET("/:role_id", v2handlers.GetRole)    // Get role by ID
			roleRead.GET("", v2handlers.ListRoles)           // List roles
			roleRead.POST("/search", v2handlers.SearchRoles) // Search roles
		}

		// Role Write operations
		roleWrite := roles.Group("", middleware.RequireRoleWrite)
		{
			roleWrite.POST("", v2handlers.CreateRole)           // Create role
			roleWrite.PATCH("/:role_id", v2handlers.UpdateRole) // Update role
		}

		// Role Delete operations
		roles.DELETE("/:role_id", middleware.RequireRoleDelete, v2handlers.DeleteRole) // Delete role
	}

	// Permission Management - Permission Entity
	permissions := v2.Group("/permissions", middleware.JWTAuth())
	{
		// Permission-Role relationship routes
		roles := permissions.Group("/:permission_id/roles")
		{
			roles.GET("", middleware.RequirePermissionRead, v2handlers.ListRolesFromPermission) // List roles assigned to permission
		}

		// Permission Read operations
		permRead := permissions.Group("", middleware.RequirePermissionRead)
		{
			permRead.GET("", v2handlers.ListPermissions)              // List permissions
			permRead.GET("/:permission_id", v2handlers.GetPermission) // Get permission by ID
			permRead.POST("/search", v2handlers.SearchPermissions)    // Search permissions
		}

		// Permission Write operations
		permWrite := permissions.Group("", middleware.RequirePermissionWrite)
		{
			permWrite.POST("", v2handlers.CreatePermission)               // Create permission
			permWrite.PUT("/:permission_id", v2handlers.UpdatePermission) // Update permission
		}

		// Permission Delete operations
		permissions.DELETE("/:permission_id", middleware.RequirePermissionDelete, v2handlers.DeletePermission) // Delete permission
	}

	// Resource Management - Resource Entity
	resources := v2.Group("/resources", middleware.JWTAuth())
	{
		// Resource-Permission relationship routes
		permissions := resources.Group("/:resource_id/permissions")
		{
			permissions.GET("", v2handlers.ListResourcePermissions) // List permissions assigned to resource
		}

		// Resource Read operations
		resources.GET("/:resource_id", v2handlers.GetResourceDetail) // Get resource by ID
		resources.GET("", v2handlers.ListResources)                  // List resources
	}

	// =====================================================================
	// Core Business Entity API Group
	// =====================================================================

	// Task Management - Task Entity
	tasks := v2.Group("/tasks", middleware.JWTAuth())
	{
		// Task Read operations
		taskRead := tasks.Group("", middleware.RequireTaskRead)
		{
			taskRead.GET("", v2handlers.ListTasks)        // List tasks
			taskRead.GET("/:task_id", v2handlers.GetTask) // Get task by ID
		}

		// Task Delete operations
		tasks.POST("/batch-delete", middleware.RequireTaskDelete, v2handlers.BatchDeleteTasks) // Batch delete tasks
	}

	// Fault Injection Management - FaultInjectionSchedule Entity
	injections := v2.Group("/injections", middleware.JWTAuth())
	{
		// Injection Analysis sub-group
		analysis := injections.Group("/analysis", middleware.RequireInjectionRead)
		{
			analysis.GET("/no-issues", v2handlers.ListFaultInjectionNoIssues)     // Get fault injections with no issues
			analysis.GET("/with-issues", v2handlers.ListFaultInjectionWithIssues) // Get fault injections with issues
		}

		// Injection Read operations
		injectionRead := injections.Group("", middleware.RequireInjectionRead)
		{
			injectionRead.GET("", v2handlers.ListInjections)                // List injections
			injectionRead.GET("/:id", v2handlers.GetInjection)              // Get injection by ID
			injectionRead.GET("/metadata", v2handlers.GetInjectionMetadata) // Get injection metadata
			injectionRead.POST("/search", v2handlers.SearchInjections)      // Advanced search
		}

		// Injection Write operations
		injectionWrite := injections.Group("", middleware.RequireInjectionWrite)
		{
			injectionWrite.POST("/inject", v2handlers.SubmitFaultInjection)             // Submit new injection
			injectionWrite.POST("/build", v2handlers.SubmitDatapackBuilding)            // Submit new datapack building
			injectionWrite.PATCH("/:id/labels", v2handlers.ManageInjectionCustomLabels) // Manage injection custom labels
		}

		// Injection Delete operations
		injections.POST("/batch-delete", middleware.RequireInjectionDelete, v2handlers.BatchDeleteInjections) // Batch delete injections
	}

	// Execution Result Management - ExecutionResult Entity
	executions := v2.Group("/executions", middleware.JWTAuth())
	{
		// Execution Read operations
		executions.GET("", v2handlers.ListExecutions)                      // List executions
		executions.GET("/:execution_id", v2handlers.GetExecution)          // Get execution by ID
		executions.GET("/labels", v2handlers.ListAvaliableExecutionLabels) // List available execution labels

		// Execution Write operations
		executions.POST("/execute", v2handlers.SubmitAlgorithmExecution)                           // Submit algorithm execution
		executions.POST("/:execution_id/detector_results", v2handlers.UploadDetectorResults)       // Upload detector results
		executions.POST("/:execution_id/granularity_results", v2handlers.UploadGranularityResults) // Upload granularity results
		executions.PATCH("/:execution_id/labels", v2handlers.ManageExecutionCustomLabels)          // Manage execution custom labels

		// Execution Delete operations
		executions.POST("/batch-delete", v2handlers.BatchDeleteExecutions) // Batch delete executions
	}

	// Trace Management - Trace Entity
	traces := v2.Group("/traces", middleware.JWTAuth())
	{
		traces.GET("/:trace_id/stream", v2handlers.GetTraceStream) // Get trace stream (SSE)
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
		// POST /api/v2/evaluations/datasets - Get algorithm evaluations on multiple datasets (requires dataset read permission)
		evaluations.POST("/datasets", middleware.RequireDatasetRead, v2handlers.ListDatasetEvaluationResults)

		// POST /api/v2/evaluations/datapacks - Get algorithm evaluations on multiple datapacks (requires dataset read permission)
		evaluations.POST("/datapacks", middleware.RequireDatasetRead, v2handlers.ListDatapackEvaluationResults)
	}
}
