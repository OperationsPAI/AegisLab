package router

import (
	"aegis/consts"
	"aegis/middleware"

	"github.com/gin-gonic/gin"
)

func SetupAdminV2Routes(v2 *gin.RouterGroup, handlers *Handlers) {
	users := v2.Group("/users", middleware.JWTAuth())
	{
		roles := users.Group("/:user_id/roles")
		{
			roles.POST("/:role_id", middleware.RequireUserAssign, handlers.User.AssignRole)
			roles.DELETE("/:role_id", middleware.RequireUserAssign, handlers.User.RemoveRole)
		}

		projects := users.Group("/:user_id/projects")
		{
			projects.POST("/:project_id/roles/:role_id", middleware.RequireUserAssign, handlers.User.AssignProject)
			projects.DELETE("/:project_id", middleware.RequireUserAssign, handlers.User.RemoveProject)
		}

		permissions := users.Group("/:user_id/permissions")
		{
			permissions.POST("/assign", middleware.RequireUserAssign, handlers.User.AssignPermissions)
			permissions.POST("/remove", middleware.RequireUserAssign, handlers.User.RemovePermissions)
		}

		containers := users.Group("/:user_id/containers")
		{
			containers.POST("/:container_id/roles/:role_id", middleware.RequireUserAssign, handlers.User.AssignContainer)
			containers.DELETE("/:container_id", middleware.RequireUserAssign, handlers.User.RemoveContainer)
		}

		datasets := users.Group("/:user_id/datasets")
		{
			datasets.POST("/:dataset_id/roles/:role_id", middleware.RequireUserAssign, handlers.User.AssignDataset)
			datasets.DELETE("/:dataset_id", middleware.RequireUserAssign, handlers.User.RemoveDataset)
		}

		userRead := users.Group("", middleware.RequireUserRead)
		{
			userRead.GET("", handlers.User.ListUsers)
			userRead.GET("/:user_id/detail", middleware.RequireAdminOrUserOwnership, handlers.User.GetUserDetail)
		}

		users.POST("", middleware.RequireUserCreate, handlers.User.CreateUser)
		users.PATCH("/:user_id", middleware.RequireUserUpdate, handlers.User.UpdateUser)
		users.DELETE("/:user_id", middleware.RequireUserDelete, handlers.User.DeleteUser)
	}

	roles := v2.Group("/roles", middleware.JWTAuth())
	{
		permissions := roles.Group("/:role_id/permissions")
		{
			permissions.POST("/assign", middleware.RequireRoleGrant, handlers.RBAC.AssignRolePermissions)
			permissions.POST("/remove", middleware.RequireRoleRevoke, handlers.RBAC.RemoveRolePermissions)
		}

		users := roles.Group("/:role_id/users")
		{
			users.GET("", middleware.RequireRoleRead, handlers.RBAC.ListUsersFromRole)
		}

		roleRead := roles.Group("", middleware.RequireRoleRead)
		{
			roleRead.GET("/:role_id", handlers.RBAC.GetRole)
			roleRead.GET("", handlers.RBAC.ListRoles)
		}

		roles.POST("", middleware.RequireRoleCreate, handlers.RBAC.CreateRole)
		roles.PATCH("/:role_id", middleware.RequireRoleUpdate, handlers.RBAC.UpdateRole)
		roles.DELETE("/:role_id", middleware.RequireRoleDelete, handlers.RBAC.DeleteRole)
	}

	permissions := v2.Group("/permissions", middleware.JWTAuth())
	{
		roles := permissions.Group("/:permission_id/roles")
		{
			roles.GET("", middleware.RequirePermissionRead, handlers.RBAC.ListRolesFromPermission)
		}

		permRead := permissions.Group("", middleware.RequirePermissionRead)
		{
			permRead.GET("", handlers.RBAC.ListPermissions)
			permRead.GET("/:permission_id", handlers.RBAC.GetPermission)
		}
	}

	resources := v2.Group("/resources", middleware.JWTAuth())
	{
		resourceRead := resources.Group("", middleware.RequirePermissionRead)
		{
			permissions := resourceRead.Group("/:resource_id/permissions")
			{
				permissions.GET("", handlers.RBAC.ListResourcePermissions)
			}

			resourceRead.GET("/:resource_id", handlers.RBAC.GetResource)
			resourceRead.GET("", handlers.RBAC.ListResources)
		}
	}

	systems := v2.Group("/systems", middleware.JWTAuth())
	{
		systemRead := systems.Group("", middleware.RequireSystemRead)
		{
			systemRead.GET("", handlers.ChaosSystem.ListSystems)
			systemRead.GET("/:id", handlers.ChaosSystem.GetSystem)
			systemRead.GET("/:id/metadata", handlers.ChaosSystem.ListMetadata)
		}

		systemConfigure := systems.Group("", middleware.RequireSystemConfigure)
		{
			systemConfigure.POST("", handlers.ChaosSystem.CreateSystem)
			systemConfigure.PUT("/:id", handlers.ChaosSystem.UpdateSystem)
			systemConfigure.POST("/:id/metadata", handlers.ChaosSystem.UpsertMetadata)
		}

		systems.DELETE("/:id", middleware.RequirePermission(consts.PermSystemManage), handlers.ChaosSystem.DeleteSystem)
	}

	system := v2.Group("/system", middleware.JWTAuth(), middleware.RequireSystemRead)
	{
		system.GET("/metrics", handlers.SystemMetric.GetSystemMetrics)
		system.GET("/metrics/history", handlers.SystemMetric.GetSystemMetricsHistory)
		audit := system.Group("/audit", middleware.RequireAuditRead)
		{
			audit.GET("", handlers.System.ListAuditLogs)
			audit.GET("/:id", handlers.System.GetAuditLog)
		}

		configs := system.Group("/configs")
		{
			configsRead := configs.Group("", middleware.RequireConfigurationRead)
			{
				configsRead.GET("", handlers.System.ListConfigs)
				configsRead.GET("/:config_id", handlers.System.GetConfig)
				configsRead.GET("/:config_id/histories", handlers.System.ListConfigHistories)
			}

			configs.PATCH("/:config_id", middleware.RequireConfigurationUpdate, handlers.System.UpdateConfigValue)
			configs.POST("/:config_id/value/rollback", middleware.RequireConfigurationUpdate, handlers.System.RollbackConfigValue)
			configs.PUT("/:config_id/metadata", middleware.RequireConfigurationConfigure, handlers.System.UpdateConfigMetadata)
			configs.POST("/:config_id/metadata/rollback", middleware.RequireConfigurationConfigure, handlers.System.RollbackConfigMetadata)
		}

		system.GET("/health", handlers.System.GetHealth)

		monitor := system.Group("/monitor")
		monitor.POST("/metrics", handlers.System.GetMetrics)
		monitor.GET("/info", handlers.System.GetSystemInfo)
		monitor.GET("/namespaces/locks", handlers.System.ListNamespaceLocks)
		monitor.GET("/tasks/queue", handlers.System.ListQueuedTasks)
	}
}
