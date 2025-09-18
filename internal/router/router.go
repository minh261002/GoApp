package router

import (
	"go_app/internal/handler"
	"go_app/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	// Initialize handlers
	authHandler := handler.NewAuthHandler()
	authMiddleware := middleware.NewAuthMiddleware()

	// API v1 group
	v1 := r.Group("/api/v1")
	{
		// Auth routes (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/forgot-password", authHandler.ForgotPassword)
			auth.POST("/reset-password", authHandler.ResetPassword)
			auth.POST("/verify-email", authHandler.VerifyEmail)
		}

		// Protected routes
		protected := v1.Group("/")
		protected.Use(authMiddleware.AuthMiddleware())
		{
			// Auth protected routes
			authProtected := protected.Group("/auth")
			{
				authProtected.GET("/profile", authHandler.GetProfile)
				authProtected.POST("/logout", authHandler.Logout)
			}

			// Admin routes
			admin := protected.Group("/admin")
			admin.Use(authMiddleware.AdminMiddleware())
			{
				// Add admin routes here
				admin.GET("/users", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Admin users endpoint"})
				})
			}

			// Moderator routes
			moderator := protected.Group("/moderator")
			moderator.Use(authMiddleware.ModeratorMiddleware())
			{
				// Add moderator routes here
				moderator.GET("/dashboard", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Moderator dashboard"})
				})
			}
		}
	}
}
