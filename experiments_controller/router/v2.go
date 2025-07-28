package router

import (
	v2handlers "github.com/LGU-SE-Internal/rcabench/handlers/v2"
	"github.com/LGU-SE-Internal/rcabench/middleware"
	"github.com/gin-gonic/gin"
)

/*
===================================================================================
API v2 设计规范 - RESTful API 标准
===================================================================================

v2 API 采用严格的 RESTful 设计规范，与 v1 的杂乱设计形成对比。
v1 API 设计较为随意，方法和路径不规范，v2 将统一按照以下标准执行。

📋 HTTP 方法使用规范：
- GET    : 查询资源（幂等，可缓存）
- POST   : 创建资源 / 复杂查询（非幂等）
- PUT    : 完整更新资源（幂等）
- PATCH  : 部分更新资源（幂等）
- DELETE : 删除资源（幂等）

🎯 URL 设计规范：
1. 资源名称使用复数形式
   ✅ GET /api/v2/users          ❌ GET /api/v2/user
   ✅ GET /api/v2/projects       ❌ GET /api/v2/project

2. 层级关系明确
   ✅ GET /api/v2/users/{id}/projects
   ✅ GET /api/v2/projects/{id}/members

3. 查询参数规范
   ✅ GET /api/v2/users?page=1&size=10&status=active
   ✅ GET /api/v2/tasks?project_id=123&type=injection

📊 标准 CRUD 操作模式：
- GET    /api/v2/{resource}           # 列表查询（支持分页、过滤、排序）
- POST   /api/v2/{resource}           # 创建资源
- GET    /api/v2/{resource}/{id}      # 获取单个资源详情
- PUT    /api/v2/{resource}/{id}      # 完整更新资源
- PATCH  /api/v2/{resource}/{id}      # 部分更新资源
- DELETE /api/v2/{resource}/{id}      # 删除资源

🔍 复杂查询处理：
对于复杂搜索条件，使用专门的搜索端点：
- POST /api/v2/{resource}/search      # 复杂条件搜索
- POST /api/v2/{resource}/query       # 高级查询
- POST /api/v2/{resource}/batch       # 批量操作

🎨 业务操作端点：
语义化的业务操作使用动词形式：
- POST /api/v2/users/{id}/activate    # 激活用户
- POST /api/v2/tasks/{id}/cancel      # 取消任务
- POST /api/v2/injections/{id}/start  # 开始故障注入
- POST /api/v2/containers/{id}/build  # 构建容器

📨 响应格式规范：
1. 成功响应：
   {
     "code": 200,
     "message": "success",
     "data": {...},
     "timestamp": "2024-01-01T12:00:00Z"
   }

2. 列表响应：
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

3. 错误响应：
   {
     "code": 400,
     "message": "validation failed",
     "errors": ["field xxx is required"],
     "timestamp": "2024-01-01T12:00:00Z"
   }

🔐 认证授权规范：
- 使用 JWT Bearer Token 认证
- 权限检查基于 RBAC 模型
- 敏感操作需要二次确认

⚡ 性能优化：
- GET 请求支持 ETag 缓存
- 列表查询默认分页（page=1, size=20）
- 支持字段选择 ?fields=id,name,status
- 支持关联查询 ?include=project,labels

🔄 版本兼容：
- v2 API 保证向后兼容
- 废弃的端点提供 6 个月过渡期
- 重大变更通过新版本号处理

注意：v1 API 设计较为混乱，不遵循统一标准，后续逐步迁移到 v2 规范。
===================================================================================
*/

// SetupV2Routes 设置 API v2 路由 - 稳定版本的 API
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

	// 权限认证相关 API 组 (部分实现，其他供将来扩展)
	roles := v2.Group("/roles", middleware.JWTAuth()) // 角色管理 - Role 实体
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

	users := v2.Group("/users", middleware.JWTAuth()) // 用户管理 - User 实体
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

	permissions := v2.Group("/permissions", middleware.JWTAuth(), middleware.RequirePermissionRead) // 权限管理 - Permission 实体
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

	resources := v2.Group("/resources") // 资源管理 - Resource 实体

	// 核心业务实体 API 组

	// 任务管理 - Task 实体
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

	// 容器管理 - Container 实体
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

	// 算法管理 - Algorithms (算法是容器的一个特殊类型)
	algorithms := v2.Group("/algorithms", middleware.JWTAuth())
	{
		// Read operations - permission checked in handler
		// GET /api/v2/algorithms?page=1&size=10 - Only active algorithms with type=algorithm
		algorithms.GET("", v2handlers.ListAlgorithms)

		// POST /api/v2/algorithms/search - Advanced search for algorithms (containers with type=algorithm) - permission checked in handler
		algorithms.POST("/search", v2handlers.SearchAlgorithms)
	}

	// 其他业务实体 API 组
	injections := v2.Group("/injections") // 故障注入管理 - FaultInjectionSchedule 实体

	// 数据集管理 - Dataset 实体
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

	executions := v2.Group("/executions") // 执行结果管理 - ExecutionResult 实体
	labels := v2.Group("/labels")         // 标签管理 - Label 实体
	projects := v2.Group("/projects")     // 项目管理 - Project 实体

	// 分析检测相关 API 组 (供将来扩展)
	detectors := v2.Group("/detectors")     // 检测器管理 - Detector 实体
	granularity := v2.Group("/granularity") // 粒度结果管理 - GranularityResult 实体
	traces := v2.Group("/traces")           // 追踪管理 - 与 TraceID 相关
	analyzer := v2.Group("/analyzer")       // 分析器相关

	// 暂时使用空赋值避免编译错误，后续逐步实现具体路由
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
