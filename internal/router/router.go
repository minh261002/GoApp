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
	categoryHandler := handler.NewCategoryHandler()
	productHandler := handler.NewProductHandler()
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

		// Category routes (public for reading, protected for writing)
		categories := v1.Group("/categories")
		{
			// Public routes (no authentication required)
			categories.GET("", categoryHandler.GetAllCategories)
			categories.GET("/tree", categoryHandler.GetCategoryTree)
			categories.GET("/level/:level", categoryHandler.GetCategoriesByLevel)
			categories.GET("/parent", categoryHandler.GetCategoriesByParent)
			categories.GET("/root", categoryHandler.GetRootCategories)
			categories.GET("/leaf", categoryHandler.GetLeafCategories)
			categories.GET("/search", categoryHandler.SearchCategories)
			categories.GET("/slug/:slug", categoryHandler.GetCategoryBySlug)
			categories.GET("/:id", categoryHandler.GetCategoryByID)
			categories.GET("/:id/children", categoryHandler.GetCategoryWithChildren)
			categories.GET("/:id/breadcrumbs", categoryHandler.GetCategoryBreadcrumbs)
			categories.GET("/:id/descendants", categoryHandler.GetCategoryDescendants)
			categories.GET("/:id/ancestors", categoryHandler.GetCategoryAncestors)
		}

		// Product routes (public for reading, protected for writing)
		products := v1.Group("/products")
		{
			// Public routes (no authentication required)
			products.GET("", productHandler.GetAllProducts)
			products.GET("/featured", productHandler.GetFeaturedProducts)
			products.GET("/brand/:brand_id", productHandler.GetProductsByBrand)
			products.GET("/category/:category_id", productHandler.GetProductsByCategory)
			products.GET("/search", productHandler.SearchProducts)
			products.GET("/low-stock", productHandler.GetLowStockProducts)
			products.GET("/stats", productHandler.GetProductStats)
			products.GET("/slug/:slug", productHandler.GetProductBySlug)
			products.GET("/sku/:sku", productHandler.GetProductBySKU)
			products.GET("/:id", productHandler.GetProductByID)
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

			// Category management routes (require authentication)
			categoryManagement := protected.Group("/categories")
			{
				categoryManagement.POST("", categoryHandler.CreateCategory)
				categoryManagement.PUT("/:id", categoryHandler.UpdateCategory)
				categoryManagement.DELETE("/:id", categoryHandler.DeleteCategory)
				categoryManagement.PATCH("/:id/status", categoryHandler.UpdateCategoryStatus)
				categoryManagement.PATCH("/bulk-status", categoryHandler.BulkUpdateCategoryStatus)
			}

			// Product management routes (require authentication)
			productManagement := protected.Group("/products")
			{
				productManagement.POST("", productHandler.CreateProduct)
				productManagement.PUT("/:id", productHandler.UpdateProduct)
				productManagement.DELETE("/:id", productHandler.DeleteProduct)
				productManagement.PATCH("/:id/stock", productHandler.UpdateProductStock)
				productManagement.PATCH("/:id/status", productHandler.UpdateProductStatus)
				productManagement.PATCH("/bulk-status", productHandler.BulkUpdateProductStatus)
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
