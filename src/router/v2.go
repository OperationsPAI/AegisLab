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

			// Container Version Create operations
			versions.POST("", middleware.RequireContainerVersionCreate, v2handlers.CreateContainerVersion) // Create container version

			// Container Version Upload operations
			versions.POST("/:version_id/helm-chart", middleware.RequireContainerVersionUpload, v2handlers.UploadHelmChart)      // Upload Helm chart tgz file
			versions.POST("/:version_id/helm-values", middleware.RequireContainerVersionUpload, v2handlers.UploadHelmValueFile) // Upload Helm values file

			// Container Version Update operations
			versions.PATCH("/:version_id", middleware.RequireContainerVersionUpdate, v2handlers.UpdateContainerVersion) // Update container version

			// Container Version Delete operations
			versions.DELETE("/:version_id", middleware.RequireContainerVersionDelete, v2handlers.DeleteContainerVersion)
		}

		// Container Read operations
		containerRead := containers.Group("", middleware.RequireContainerRead)
		{
			containerRead.GET("/:container_id", v2handlers.GetContainer) // Get container by ID
			containerRead.GET("", v2handlers.ListContainers)             // List containers
		}

		// Container Create operations
		containers.POST("", middleware.RequireContainerCreate, v2handlers.CreateContainer) // Create container

		// Container Execute operations (build requires execute permission)
		containers.POST("/build", middleware.RequireContainerExecute, v2handlers.SubmitContainerBuilding) // Build container

		// Container Update operations
		containers.PATCH("/:container_id", middleware.RequireContainerUpdate, v2handlers.UpdateContainer)                    // Update container
		containers.PATCH("/:container_id/labels", middleware.RequireContainerUpdate, v2handlers.ManageContainerCustomLabels) // Manage container labels

		// Container Delete operations
		containers.DELETE("/:container_id", middleware.RequireContainerDelete, v2handlers.DeleteContainer) // Delete container
	}

	// Container Version flat resource — direct-by-version-id operations without
	// the parent container id in the URL. Used by aegisctl `container version
	// set-image` to rewrite the four image-reference columns of a single row.
	containerVersions := v2.Group("/container-versions", middleware.JWTAuth())
	{
		containerVersions.PATCH("/:id/image", middleware.RequireContainerVersionUpdate, v2handlers.SetContainerVersionImage)
	}

	// Dataset Management - Dataset Entity
	datasets := v2.Group("/datasets", middleware.JWTAuth())
	{
		// Dataset Version sub-resource routes
		versions := datasets.Group("/:dataset_id/versions")
		{
			versionRead := versions.Group("", middleware.RequireDatasetVersionRead)
			{
				versionRead.GET("", v2handlers.ListDatasetVersions)                         // List dataset versions
				versionRead.GET("/:version_id", v2handlers.GetDatasetVersion)               // Get dataset version by ID
				versionRead.GET("/:version_id/download", v2handlers.DownloadDatasetVersion) // Download dataset version
			}

			// Dataset Version Create operations
			versions.POST("", middleware.RequireDatasetVersionCreate, v2handlers.CreateDatasetVersion) // Create dataset version

			// Dataset Version Update operations
			versions.PATCH("/:version_id", middleware.RequireDatasetVersionUpdate, v2handlers.UpdateDatasetVersion)                      // Update dataset version
			versions.PATCH("/:version_id/injections", middleware.RequireDatasetVersionUpdate, v2handlers.ManageDatasetVersionInjections) // Manage dataset version injections

			versions.DELETE("/:version_id", middleware.RequireDatasetVersionDelete, v2handlers.DeleteDatasetVersion) // Delete dataset version
		}

		// Dataset Read operations
		datasetRead := datasets.Group("", middleware.RequireDatasetRead)
		{
			datasetRead.GET("/:dataset_id", v2handlers.GetDataset) // Get dataset by ID
			datasetRead.GET("", v2handlers.ListDatasets)           // List datasets
		}

		// Dataset Create operations
		datasets.POST("", middleware.RequireDatasetCreate, v2handlers.CreateDataset) // Create dataset

		// Dataset Update operations
		datasets.PATCH("/:dataset_id", middleware.RequireDatasetUpdate, v2handlers.UpdateDataset)                    // Update dataset
		datasets.PATCH("/:dataset_id/labels", middleware.RequireDatasetUpdate, v2handlers.ManageDatasetCustomLabels) // Manage dataset labels

		// Dataset Delete operations
		datasets.DELETE("/:dataset_id", middleware.RequireDatasetDelete, v2handlers.DeleteDataset) // Delete dataset
	}

	// Project Management - Project Entity
	projects := v2.Group("/projects", middleware.JWTAuth())
	{
		injections := projects.Group("/:project_id/injections")
		{
			injectionRead := injections.Group("", middleware.RequireProjectRead)
			{
				analysis := injectionRead.Group("/analysis")
				{
					analysis.GET("/no-issues", v2handlers.ListFaultInjectionNoIssues)
					analysis.GET("/with-issues", v2handlers.ListFaultInjectionWithIssues)
				}

				injectionRead.GET("", v2handlers.ListProjectInjections)
				injectionRead.POST("/search", v2handlers.SearchInjections)
			}

			injectionExecute := injections.Group("", middleware.RequireProjectInjectionExecute)
			{
				injectionExecute.POST("/inject", v2handlers.SubmitProjectFaultInjection)
				injectionExecute.POST("/build", v2handlers.SubmitProjectDatapackBuilding)
			}
		}

		executions := projects.Group("/:project_id/executions")
		{
			executionRead := executions.Group("", middleware.RequireProjectRead)
			{
				executionRead.GET("", v2handlers.ListProjectExecutions)
			}

			executionExecute := executions.Group("", middleware.RequireProjectExecutionExecute)
			{
				executionExecute.POST("/execute", v2handlers.SubmitAlgorithmExecution)
			}
		}

		// Project Read operations
		projectRead := projects.Group("", middleware.RequireProjectRead)
		{
			projectRead.GET("/:project_id", v2handlers.GetProjectDetail) // Get project by ID
			projectRead.GET("", v2handlers.ListProjects)                 // List projects
		}

		// Project Create operations
		projects.POST("", middleware.RequireProjectCreate, v2handlers.CreateProject) // Create project

		// Project Update operations
		projects.PATCH("/:project_id", middleware.RequireProjectUpdate, v2handlers.UpdateProject)                    // Update project
		projects.PATCH("/:project_id/labels", middleware.RequireProjectUpdate, v2handlers.ManageProjectCustomLabels) // Manage project labels

		// Project Delete operations
		projects.DELETE("/:project_id", middleware.RequireProjectDelete, v2handlers.DeleteProject) // Delete project
	}

	// Team Management - Team Entity
	teams := v2.Group("/teams", middleware.JWTAuth())
	{
		// Anyone can create a team (will become team admin automatically)
		teams.POST("", v2handlers.CreateTeam)

		// List teams - returns public teams + user's teams (no special permission needed)
		teams.GET("", v2handlers.ListTeams)

		// Team Write/Delete/Manage operations - requires team admin OR system admin
		TeamAdmin := teams.Group("/:team_id", middleware.RequireTeamAdminAccess)
		{
			TeamAdmin.PATCH("", v2handlers.UpdateTeam)  // Update team
			TeamAdmin.DELETE("", v2handlers.DeleteTeam) // Delete team

			// Team Member Management - only team admins can manage members
			TeamManagement := TeamAdmin.Group("/members")
			TeamManagement.POST("", v2handlers.AddTeamMember)                       // Add team member
			TeamManagement.DELETE("/:user_id", v2handlers.RemoveTeamMember)         // Remove team member
			TeamManagement.PATCH("/:user_id/role", v2handlers.UpdateTeamMemberRole) // Update team member role
		}

		// Team Read operations - requires being a member OR team is public OR system admin
		TeamMember := teams.Group("", middleware.RequireTeamMemberAccess)
		{
			TeamMember.GET("/:team_id", v2handlers.GetTeamDetail)             // Get team by ID
			TeamMember.GET("/:team_id/members", v2handlers.ListTeamMembers)   // List team members
			TeamMember.GET("/:team_id/projects", v2handlers.ListTeamProjects) // List team projects
		}
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

		// Label Create operations
		labels.POST("", middleware.RequireLabelCreate, v2handlers.CreateLabel) // Create label

		// Label Update operations
		labels.PATCH("/:label_id", middleware.RequireLabelUpdate, v2handlers.UpdateLabel) // Update label

		// Label Delete operations
		labels.DELETE("/:label_id", middleware.RequireLabelDelete, v2handlers.DeleteLabel)        // Delete label
		labels.POST("/batch-delete", middleware.RequireLabelDelete, v2handlers.BatchDeleteLabels) // Batch delete labels
	}

	// User Management - User Entity
	users := v2.Group("/users", middleware.JWTAuth())
	{
		// User-Role relationship routes (assign roles requires assign permission)
		roles := users.Group("/:user_id/roles")
		{
			roles.POST("/:role_id", middleware.RequireUserAssign, v2handlers.AssignUserRole)     // Assign role to user
			roles.DELETE("/:role_id", middleware.RequireUserAssign, v2handlers.RemoveGlobalRole) // Remove role from user
		}

		// User-Project relationship routes (assign requires assign permission)
		projects := users.Group("/:user_id/projects")
		{
			projects.POST("/:project_id/roles/:role_id", middleware.RequireUserAssign, v2handlers.AssignUserProject) // Assign user to project
			projects.DELETE("/:project_id", middleware.RequireUserAssign, v2handlers.RemoveUserProject)              // Remove user from project
		}

		// User-Permission relationship routes (assign requires assign permission)
		permissions := users.Group("/:user_id/permissions")
		{
			permissions.POST("/assign", middleware.RequireUserAssign, v2handlers.AssignUserPermission) // Assign permission to user
			permissions.POST("/remove", middleware.RequireUserAssign, v2handlers.RemoveUserPermission) // Remove permission from user
		}

		// User-Container relationship routes (assign requires assign permission)
		containers := users.Group("/:user_id/containers")
		{
			containers.POST("/:container_id/roles/:role_id", middleware.RequireUserAssign, v2handlers.AssignUserContainer) // Assign container to user
			containers.DELETE("/:container_id", middleware.RequireUserAssign, v2handlers.RemoveUserContainer)              // Remove container from user
		}

		// User-Dataset relationship routes (assign requires assign permission)
		datasets := users.Group("/:user_id/datasets")
		{
			datasets.POST("/:dataset_id/roles/:role_id", middleware.RequireUserAssign, v2handlers.AssignUserDataset) // Assign dataset to user
			datasets.DELETE("/:dataset_id", middleware.RequireUserAssign, v2handlers.RemoveUserDataset)              // Remove dataset from user
		}

		// User Read operations
		userRead := users.Group("", middleware.RequireUserRead)
		{
			userRead.GET("", v2handlers.ListUsersV2)                                                             // List users
			userRead.GET("/:user_id/detail", middleware.RequireAdminOrUserOwnership, v2handlers.GetUserDetailV2) // Get user by ID
		}

		// User Create operations
		users.POST("", middleware.RequireUserCreate, v2handlers.CreateUser) // Create user

		// User Update operations
		users.PATCH("/:user_id", middleware.RequireUserUpdate, v2handlers.UpdateUser) // Update user

		// User Delete operations
		users.DELETE("/:user_id", middleware.RequireUserDelete, v2handlers.DeleteUser) // Delete user
	}

	// =====================================================================
	// Authentication and Authorization API Group
	// =====================================================================

	// Role Management - Role Entity
	roles := v2.Group("/roles", middleware.JWTAuth())
	{
		// Role-Permission relationship routes (grant/revoke)
		permissions := roles.Group("/:role_id/permissions")
		{
			permissions.POST("/assign", middleware.RequireRoleGrant, v2handlers.AssignRolePermission)       // Assign permissions to role
			permissions.POST("/remove", middleware.RequireRoleRevoke, v2handlers.RemovePermissionsFromRole) // Remove permissions from role
		}

		// Role-User relationship routes
		users := roles.Group("/:role_id/users")
		{
			users.GET("", middleware.RequireRoleRead, v2handlers.ListUsersFromRole) // List users with this role
		}

		// Role Read operations
		roleRead := roles.Group("", middleware.RequireRoleRead)
		{
			roleRead.GET("/:role_id", v2handlers.GetRole) // Get role by ID
			roleRead.GET("", v2handlers.ListRoles)        // List roles
		}

		// Role Create operations
		roles.POST("", middleware.RequireRoleCreate, v2handlers.CreateRole) // Create role

		// Role Update operations
		roles.PATCH("/:role_id", middleware.RequireRoleUpdate, v2handlers.UpdateRole) // Update role

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
		}
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
	tasks := v2.Group("/tasks")
	{
		taskWithAuth := tasks.Group("", middleware.JWTAuth())
		{

			// Task Read operations
			taskRead := taskWithAuth.Group("", middleware.RequireTaskRead)
			{
				taskRead.GET("", v2handlers.ListTasks)        // List tasks
				taskRead.GET("/:task_id", v2handlers.GetTask) // Get task by ID
			}

			// Task Delete operations
			taskWithAuth.POST("/batch-delete", middleware.RequireTaskDelete, v2handlers.BatchDeleteTasks) // Batch delete tasks

			// Task Update/Execute operations
			taskWithAuth.POST("/:task_id/expedite", middleware.RequireTaskExecute, v2handlers.ExpediteTask) // Expedite pending task
		}

		// Task Log streaming (WebSocket) - auth via query param, not middleware
		tasks.GET("/:task_id/logs/ws", v2handlers.GetTaskLogsWS) // Stream task logs via WebSocket
	}

	// Fault Injection Management - FaultInjectionSchedule Entity
	// Note: These global routes are for system admins only. Regular users should access injections via /projects/:project_id
	injections := v2.Group("/injections", middleware.JWTAuth())
	{
		injectionSystemAdmin := injections.Group("", middleware.RequireSystemAdmin())
		{
			injectionSystemAdmin.GET("", v2handlers.ListInjections)           // List injections
			injectionSystemAdmin.POST("/search", v2handlers.SearchInjections) // Advanced search
		}

		// Manual upload (must be before /:id routes)
		injections.POST("/upload", v2handlers.UploadDatapack) // Upload manual datapack

		// DSL translation endpoints (must be before /:id routes)
		injections.GET("/systems", v2handlers.GetSystemMapping)     // Get system type mapping
		injections.POST("/translate", v2handlers.TranslateFaultSpecs) // Translate fault specs to Nodes

		// Injection Read operations
		injections.GET("/:id", v2handlers.GetInjection)                        // Get injection by ID
		injections.GET("/:id/download", v2handlers.DownloadDatapack)           // Download injection datapack
		injections.GET("/:id/logs", v2handlers.GetInjectionLogs)                // Get injection execution logs
		injections.GET("/:id/files", v2handlers.ListDatapackFiles)             // Get injection file structure
		injections.GET("/:id/files/download", v2handlers.DownloadDatapackFile) // Download specific injection file
		injections.GET("/:id/files/query", v2handlers.QueryDatapackFile)       // Query parquet file content
		injections.GET("/metadata", v2handlers.GetInjectionMetadata)           // Get injection metadata

		// Injection Clone operations
		injections.POST("/:id/clone", v2handlers.CloneInjection) // Clone injection

		// Injection Update operations (label management, ground truth)
		injections.PUT("/:id/groundtruth", v2handlers.UpdateGroundtruth)         // Update ground truth
		injections.PATCH("/:id/labels", v2handlers.ManageInjectionCustomLabels)  // Manage injection custom labels
		injections.PATCH("/labels/batch", v2handlers.BatchManageInjectionLabels) // Batch manage injection labels

		// Injection Delete operations
		injections.POST("/batch-delete", v2handlers.BatchDeleteInjections) // Batch delete injections
	}

	// Execution Result Management - ExecutionResult Entity
	// Note: These global routes are for system admins only. Regular users should access executions via /projects/:project_id
	executions := v2.Group("/executions", middleware.JWTAuth())
	{
		executionSystemAdmin := executions.Group("", middleware.RequireSystemAdmin())
		{
			executionSystemAdmin.GET("", v2handlers.ListExecutions)                      // List executions
			executionSystemAdmin.GET("/labels", v2handlers.ListAvaliableExecutionLabels) // List available execution labels
		}

		// Execution Read operations
		executions.GET("/:execution_id", v2handlers.GetExecution) // Get execution by ID

		// Execution Update operations (upload results and manage labels)
		executions.POST("/:execution_id/detector_results", v2handlers.UploadDetectorResults)       // Upload detector results
		executions.POST("/:execution_id/granularity_results", v2handlers.UploadGranularityResults) // Upload granularity results
		executions.PATCH("/:execution_id/labels", v2handlers.ManageExecutionCustomLabels)          // Manage execution custom labels

		// Execution Delete operations
		executions.POST("/batch-delete", v2handlers.BatchDeleteExecutions) // Batch delete executions
	}

	// Trace Management - Trace Entity
	traces := v2.Group("/traces", middleware.JWTAuth())
	{
		traces.GET("", v2handlers.ListTraces)                      // List traces
		traces.GET("/:trace_id", v2handlers.GetTrace)              // Get trace by ID
		traces.GET("/:trace_id/stream", v2handlers.GetTraceStream) // Get trace stream (SSE)
	}

	// Group Management - Group stream for real-time batch progress
	groups := v2.Group("/groups", middleware.JWTAuth())
	{
		groups.GET("/:group_id/stats", v2handlers.GetAlgorithmMetrics) // Get group stats (can be used for progress tracking)
		groups.GET("/:group_id/stream", v2handlers.GetGroupStream)     // Stream group trace events (SSE)
	}

	// =====================================================================
	// Notification API Group
	// =====================================================================

	// Notification Management - Global workflow notifications
	notifications := v2.Group("/notifications", middleware.JWTAuth())
	{
		notifications.GET("/stream", v2handlers.GetNotificationStream) // Stream global notifications (SSE)
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
		evaluations.GET("", v2handlers.ListEvaluations)

		// GET /api/v2/evaluations/:id - Get a single evaluation by ID
		evaluations.GET("/:id", v2handlers.GetEvaluation)

		// DELETE /api/v2/evaluations/:id - Delete an evaluation by ID
		evaluations.DELETE("/:id", v2handlers.DeleteEvaluation)

		// POST /api/v2/evaluations/datasets - Get algorithm evaluations on multiple datasets (requires dataset read permission)
		evaluations.POST("/datasets", middleware.RequireDatasetRead, v2handlers.ListDatasetEvaluationResults)

		// POST /api/v2/evaluations/datapacks - Get algorithm evaluations on multiple datapacks (requires dataset read permission)
		evaluations.POST("/datapacks", middleware.RequireDatasetRead, v2handlers.ListDatapackEvaluationResults)
	}

	// =====================================================================
	// SDK Evaluation API Group (read-only access to Python SDK tables)
	// =====================================================================

	sdkEval := v2.Group("/sdk/evaluations", middleware.JWTAuth())
	{
		sdkEval.GET("", v2handlers.ListSDKEvaluations)
		sdkEval.GET("/experiments", v2handlers.ListSDKExperiments)
		sdkEval.GET("/:id", v2handlers.GetSDKEvaluation)
	}

	sdkData := v2.Group("/sdk/datasets", middleware.JWTAuth())
	{
		sdkData.GET("", v2handlers.ListSDKDatasetSamples)
	}

	// =====================================================================
	// Metrics API Group
	// =====================================================================

	// Metrics routes
	metrics := v2.Group("/metrics", middleware.JWTAuth())
	{
		metrics.GET("/injections", v2handlers.GetInjectionMetrics) // Get injection metrics
		metrics.GET("/executions", v2handlers.GetExecutionMetrics) // Get execution metrics
		metrics.GET("/algorithms", v2handlers.GetAlgorithmMetrics) // Get algorithm comparison metrics
	}

	// =====================================================================
	// Pedestal Helm Config API Group
	// =====================================================================
	//
	// CRUD + dry-run verification over the helm_configs table, keyed by
	// container_version_id. Used by `aegisctl pedestal helm` to fix bad
	// repo URLs without running `mysql -e UPDATE` and without triggering
	// a real restart_pedestal task.
	pedestal := v2.Group("/pedestal", middleware.JWTAuth())
	{
		helm := pedestal.Group("/helm")
		{
			helm.GET("/:container_version_id", v2handlers.GetPedestalHelmConfig)
			helm.POST("/:container_version_id/verify", v2handlers.VerifyPedestalHelmConfig)
			// Mutating route — admin/upload permission (same tier as helm-chart upload).
			helm.PUT("/:container_version_id", middleware.RequireContainerVersionUpload, v2handlers.UpsertPedestalHelmConfig)
		}
	}

	// =====================================================================
	// Chaos Systems API Group
	// =====================================================================

	// Chaos System Management - System Entity
	systems := v2.Group("/systems", middleware.JWTAuth())
	{
		systems.GET("", v2handlers.ListChaosSystemsHandler)
		systems.POST("", v2handlers.CreateChaosSystemHandler)
		systems.GET("/:id", v2handlers.GetChaosSystemHandler)
		systems.PUT("/:id", v2handlers.UpdateChaosSystemHandler)
		systems.DELETE("/:id", v2handlers.DeleteChaosSystemHandler)
		systems.POST("/:id/metadata", v2handlers.UpsertChaosSystemMetadataHandler)
		systems.GET("/:id/metadata", v2handlers.ListChaosSystemMetadataHandler)
	}

	// =====================================================================
	// System Metrics API Group
	// =====================================================================

	// System metrics routes
	system := v2.Group("/system", middleware.JWTAuth())
	{
		system.GET("/metrics", v2handlers.GetSystemMetrics)                // Get current system metrics
		system.GET("/metrics/history", v2handlers.GetSystemMetricsHistory) // Get historical system metrics
	}

	// =====================================================================
	// Rate Limiter Admin API Group (OperationsPAI/aegis#21)
	// =====================================================================

	rateLimiters := v2.Group("/rate-limiters", middleware.JWTAuth())
	{
		// status: any authenticated user
		rateLimiters.GET("", v2handlers.ListRateLimiters)

		// reset + gc: system admin only
		rateLimiterAdmin := rateLimiters.Group("", middleware.RequireSystemAdmin())
		{
			rateLimiterAdmin.DELETE("/:bucket", v2handlers.ResetRateLimiter)
			rateLimiterAdmin.POST("/gc", v2handlers.GCRateLimiters)
		}
	}
}
