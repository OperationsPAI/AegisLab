package router

import (
	"aegis/middleware"

	"github.com/gin-gonic/gin"
)

func SetupPortalV2Routes(v2 *gin.RouterGroup, handlers *Handlers) {
	containers := v2.Group("/containers", middleware.JWTAuth())
	{
		containerRead := containers.Group("", middleware.RequireContainerRead)
		{
			containerRead.GET("", handlers.Container.ListContainers)
			containerRead.GET("/:container_id", handlers.Container.GetContainer)
		}

		containers.POST("", middleware.RequireContainerCreate, handlers.Container.CreateContainer)
		containers.PATCH("/:container_id", middleware.RequireContainerUpdate, handlers.Container.UpdateContainer)
		containers.PATCH("/:container_id/labels", middleware.RequireContainerUpdate, handlers.Container.ManageContainerCustomLabels)
		containers.DELETE("/:container_id", middleware.RequireContainerDelete, handlers.Container.DeleteContainer)
		containers.POST("/build", middleware.RequireContainerExecute, handlers.Container.SubmitContainerBuilding)

		containerVersions := containers.Group("/:container_id/versions")
		{
			containerVersionRead := containerVersions.Group("", middleware.RequireContainerVersionRead)
			{
				containerVersionRead.GET("", handlers.Container.ListContainerVersions)
				containerVersionRead.GET("/:version_id", handlers.Container.GetContainerVersion)
			}

			containerVersions.POST("", middleware.RequireContainerVersionCreate, handlers.Container.CreateContainerVersion)
			containerVersions.PATCH("/:version_id", middleware.RequireContainerVersionUpdate, handlers.Container.UpdateContainerVersion)
			containerVersions.DELETE("/:version_id", middleware.RequireContainerVersionDelete, handlers.Container.DeleteContainerVersion)
			containerVersions.POST("/:version_id/helm-chart", middleware.RequireContainerVersionUpload, handlers.Container.UploadHelmChart)
			containerVersions.POST("/:version_id/helm-values", middleware.RequireContainerVersionUpload, handlers.Container.UploadHelmValueFile)
		}
	}

	datasets := v2.Group("/datasets", middleware.JWTAuth())
	{
		datasetRead := datasets.Group("", middleware.RequireDatasetRead)
		{
			datasetRead.GET("", handlers.Dataset.ListDatasets)
			datasetRead.GET("/:dataset_id", handlers.Dataset.GetDataset)
			datasetRead.POST("/search", handlers.Dataset.SearchDataset)
		}

		datasets.POST("", middleware.RequireDatasetCreate, handlers.Dataset.CreateDataset)
		datasets.PATCH("/:dataset_id", middleware.RequireDatasetUpdate, handlers.Dataset.UpdateDataset)
		datasets.PATCH("/:dataset_id/labels", middleware.RequireDatasetUpdate, handlers.Dataset.ManageDatasetCustomLabels)
		datasets.DELETE("/:dataset_id", middleware.RequireDatasetDelete, handlers.Dataset.DeleteDataset)

		datasetVersions := datasets.Group("/:dataset_id/versions")
		{
			datasetVersionRead := datasetVersions.Group("", middleware.RequireDatasetVersionRead)
			{
				datasetVersionRead.GET("", handlers.Dataset.ListDatasetVersions)
				datasetVersionRead.GET("/:version_id", handlers.Dataset.GetDatasetVersion)
			}

			datasetVersions.POST("", middleware.RequireDatasetVersionCreate, handlers.Dataset.CreateDatasetVersion)
			datasetVersions.PATCH("/:version_id", middleware.RequireDatasetVersionUpdate, handlers.Dataset.UpdateDatasetVersion)
			datasetVersions.DELETE("/:version_id", middleware.RequireDatasetVersionDelete, handlers.Dataset.DeleteDatasetVersion)
		}
	}

	projects := v2.Group("/projects", middleware.JWTAuth())
	{
		projects.POST("/:project_id/injections/search", middleware.RequireProjectRead, handlers.Injection.SearchProjectInjections)

		projectRead := projects.Group("", middleware.RequireProjectRead)
		{
			projectRead.GET("/:project_id", handlers.Project.GetProjectDetail)
			projectRead.GET("", handlers.Project.ListProjects)
		}

		projects.POST("", middleware.RequireProjectCreate, handlers.Project.CreateProject)
		projects.PATCH("/:project_id", middleware.RequireProjectUpdate, handlers.Project.UpdateProject)
		projects.PATCH("/:project_id/labels", middleware.RequireProjectUpdate, handlers.Project.ManageProjectCustomLabels)
		projects.DELETE("/:project_id", middleware.RequireProjectDelete, handlers.Project.DeleteProject)
	}

	teams := v2.Group("/teams", middleware.JWTAuth())
	{
		teams.POST("", middleware.RequireTeamCreate, handlers.Team.CreateTeam)
		teams.GET("", middleware.RequireTeamRead, handlers.Team.ListTeams)

		teamAdmin := teams.Group("/:team_id", middleware.RequireTeamAdminAccess)
		{
			teamAdmin.PATCH("", handlers.Team.UpdateTeam)
			teamAdmin.DELETE("", handlers.Team.DeleteTeam)

			teamManagement := teamAdmin.Group("/members")
			teamManagement.POST("", handlers.Team.AddTeamMember)
			teamManagement.DELETE("/:user_id", handlers.Team.RemoveTeamMember)
			teamManagement.PATCH("/:user_id/role", handlers.Team.UpdateTeamMemberRole)
		}

		teamMember := teams.Group("", middleware.RequireTeamMemberAccess)
		{
			teamMember.GET("/:team_id", handlers.Team.GetTeamDetail)
			teamMember.GET("/:team_id/members", handlers.Team.ListTeamMembers)
			teamMember.GET("/:team_id/projects", handlers.Team.ListTeamProjects)
		}
	}

	labels := v2.Group("/labels", middleware.JWTAuth())
	{
		labelRead := labels.Group("", middleware.RequireLabelRead)
		{
			labelRead.GET("/:label_id", handlers.Label.GetLabelDetail)
			labelRead.GET("", handlers.Label.ListLabels)
		}

		labels.POST("", middleware.RequireLabelCreate, handlers.Label.CreateLabel)
		labels.PATCH("/:label_id", middleware.RequireLabelUpdate, handlers.Label.UpdateLabel)
		labels.DELETE("/:label_id", middleware.RequireLabelDelete, handlers.Label.DeleteLabel)
		labels.POST("/batch-delete", middleware.RequireLabelDelete, handlers.Label.BatchDeleteLabels)
	}

	evaluations := v2.Group("/evaluations", middleware.JWTAuth())
	{
		evaluations.DELETE("/:id", handlers.Evaluation.DeleteEvaluation)
	}

	executions := v2.Group("/executions", middleware.JWTAuth())
	{
		executions.GET("/labels", middleware.RequireAPIKeyScopesAny("sdk:*", "sdk:executions:*", "sdk:executions:read"), handlers.Execution.ListAvailableExecutionLabels)
		executions.POST("/batch-delete", handlers.Execution.BatchDeleteExecutions)
	}

	injections := v2.Group("/injections", middleware.JWTAuth())
	{
		injections.PATCH("/labels/batch", handlers.Injection.BatchManageInjectionLabels)
		injections.POST("/batch-delete", handlers.Injection.BatchDeleteInjections)
		injections.POST("/upload", handlers.Injection.UploadDatapack)
		injections.PUT("/:id/groundtruth", handlers.Injection.UpdateGroundtruth)
	}

	notifications := v2.Group("/notifications", middleware.JWTAuth())
	{
		notifications.GET("/stream", handlers.Notification.GetStream)
	}

	tasks := v2.Group("/tasks", middleware.JWTAuth())
	{
		taskRead := tasks.Group("", middleware.RequireTaskRead)
		{
			taskRead.GET("", handlers.Task.ListTasks)
			taskRead.GET("/:task_id", handlers.Task.GetTask)
			taskRead.GET("/:task_id/logs/ws", handlers.Task.GetTaskLogsWS)
		}

		tasks.POST("/batch-delete", middleware.RequireTaskDelete, handlers.Task.BatchDelete)
	}

	groups := v2.Group("/groups", middleware.JWTAuth(), middleware.RequireTraceRead)
	{
		groups.GET("/:group_id/stats", handlers.Group.GetGroupStats)
		groups.GET("/:group_id/stream", handlers.Group.GetGroupStream)
	}

	traces := v2.Group("/traces", middleware.JWTAuth(), middleware.RequireTraceRead)
	{
		traces.GET("", handlers.Trace.ListTraces)
		traces.GET("/:trace_id", handlers.Trace.GetTrace)
		traces.GET("/:trace_id/stream", handlers.Trace.GetTraceStream)
	}

	accessKeys := v2.Group("/api-keys", middleware.JWTAuth(), middleware.RequireHumanUserAuth())
	{
		accessKeys.GET("", handlers.Auth.ListAPIKeys)
		accessKeys.POST("", handlers.Auth.CreateAPIKey)
		accessKeys.GET("/:id", handlers.Auth.GetAPIKey)
		accessKeys.DELETE("/:id", handlers.Auth.DeleteAPIKey)
		accessKeys.POST("/:id/rotate", handlers.Auth.RotateAPIKey)
		accessKeys.POST("/:id/disable", handlers.Auth.DisableAPIKey)
		accessKeys.POST("/:id/enable", handlers.Auth.EnableAPIKey)
		accessKeys.POST("/:id/revoke", handlers.Auth.RevokeAPIKey)
	}
}
