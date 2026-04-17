package router

import (
	"aegis/middleware"

	"github.com/gin-gonic/gin"
)

func SetupPortalV2Routes(v2 *gin.RouterGroup, handlers *Handlers) {
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
				injectionRead.POST("/search", handlers.Injection.SearchProjectInjections)
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
		teams.POST("", handlers.Team.CreateTeam)
		teams.GET("", handlers.Team.ListTeams)

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

	accessKeys := v2.Group("/access-keys", middleware.JWTAuth())
	{
		accessKeys.GET("", handlers.Auth.ListAccessKeys)
		accessKeys.POST("", handlers.Auth.CreateAccessKey)
		accessKeys.GET("/:access_key_id", handlers.Auth.GetAccessKey)
		accessKeys.DELETE("/:access_key_id", handlers.Auth.DeleteAccessKey)
		accessKeys.POST("/:access_key_id/rotate", handlers.Auth.RotateAccessKey)
		accessKeys.POST("/:access_key_id/disable", handlers.Auth.DisableAccessKey)
		accessKeys.POST("/:access_key_id/enable", handlers.Auth.EnableAccessKey)
	}
}
