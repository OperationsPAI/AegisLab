package router

import (
	"aegis/middleware"

	"github.com/gin-gonic/gin"
)

func SetupPublicV2Routes(v2 *gin.RouterGroup, handlers *Handlers) {
	auth := v2.Group("/auth")
	{
		auth.POST("/login", handlers.Auth.Login)          // User login
		auth.POST("/register", handlers.Auth.Register)    // User registration
		auth.POST("/refresh", handlers.Auth.RefreshToken) // Token refresh

		// These require authentication
		authProtected := auth.Group("", middleware.JWTAuth(), middleware.RequireHumanUserAuth())
		{
			authProtected.POST("/logout", handlers.Auth.Logout)                  // User logout
			authProtected.POST("/change-password", handlers.Auth.ChangePassword) // Change password
			authProtected.GET("/profile", handlers.Auth.GetProfile)              // Get current user profile
		}
	}
}
