package router

import (
	"go_app/internal/handler"
	"go_app/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Service is healthy",
		})
	})

	// Static file serving for uploads
	r.Static("/uploads", "./uploads")

	// Initialize handlers
	authHandler := handler.NewAuthHandler()
	brandHandler := handler.NewBrandHandler()
	uploadHandler := handler.NewUploadHandler()
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

		// Brand routes (public for reading, protected for writing)
		brands := v1.Group("/brands")
		{
			// Public routes (no authentication required)
			brands.GET("", brandHandler.GetAllBrands)
			brands.GET("/active", brandHandler.GetActiveBrands)
			brands.GET("/search", brandHandler.SearchBrands)
			brands.GET("/slug/:slug", brandHandler.GetBrandBySlug)
			brands.GET("/:id", brandHandler.GetBrandByID)
		}

		// Upload routes (public for reading, protected for writing)
		upload := v1.Group("/upload")
		{
			// Public routes (no authentication required)
			upload.GET("/info", uploadHandler.GetFileInfo)
			upload.GET("/stats", uploadHandler.GetUploadStats)
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

			// Brand management routes (require authentication)
			brandManagement := protected.Group("/brands")
			{
				brandManagement.POST("", brandHandler.CreateBrand)
				brandManagement.PUT("/:id", brandHandler.UpdateBrand)
				brandManagement.DELETE("/:id", brandHandler.DeleteBrand)
				brandManagement.PATCH("/:id/status", brandHandler.UpdateBrandStatus)
				brandManagement.PATCH("/bulk-status", brandHandler.BulkUpdateBrandStatus)
			}

			// Upload management routes (require authentication)
			uploadManagement := protected.Group("/upload")
			{
				uploadManagement.POST("", uploadHandler.UploadFile)
				uploadManagement.POST("/image", uploadHandler.UploadImage)
				uploadManagement.POST("/brand-logo", uploadHandler.UploadBrandLogo)
				uploadManagement.POST("/document", uploadHandler.UploadDocument)
				uploadManagement.POST("/multiple", uploadHandler.UploadMultipleFiles)
				uploadManagement.DELETE("", uploadHandler.DeleteFile)
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
