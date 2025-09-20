package router

import (
	"go_app/internal/handler"
	"go_app/internal/model"
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
	inventoryHandler := handler.NewInventoryHandler()
	permissionHandler := handler.NewPermissionHandler()
	orderHandler := handler.NewOrderHandler()
	addressHandler := handler.NewAddressHandler()
	reviewHandler := handler.NewReviewHandler()
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

		// Inventory routes (public for reading, protected for writing)
		inventory := v1.Group("/inventory")
		{
			// Public routes (no authentication required)
			inventory.GET("/stock-levels", inventoryHandler.GetAllStockLevels)
			inventory.GET("/stock-levels/product/:product_id", inventoryHandler.GetStockLevelByProduct)
			inventory.GET("/low-stock", inventoryHandler.GetLowStockProducts)
			inventory.GET("/out-of-stock", inventoryHandler.GetOutOfStockProducts)
			inventory.GET("/stats", inventoryHandler.GetInventoryStats)
			inventory.GET("/alerts", inventoryHandler.GetLowStockAlerts)
			inventory.GET("/value", inventoryHandler.GetStockValue)
		}

		// Permission routes (public for checking, protected for management)
		permissions := v1.Group("/permissions")
		{
			// Public routes (no authentication required)
			permissions.POST("/check", permissionHandler.CheckPermission)
		}

		// Order routes (public for reading, protected for writing)
		orders := v1.Group("/orders")
		{
			// Public routes (no authentication required)
			orders.GET("/order-number/:order_number", orderHandler.GetOrderByOrderNumber)
		}

		// Address routes (public for reading, protected for writing)
		addresses := v1.Group("/addresses")
		{
			// Public routes (no authentication required)
			addresses.GET("/city/:city", addressHandler.GetAddressesByCity)
			addresses.GET("/district/:district", addressHandler.GetAddressesByDistrict)
			addresses.GET("/nearby", addressHandler.GetAddressesNearby)
			addresses.GET("/search", addressHandler.SearchAddresses)
			addresses.GET("/stats", addressHandler.GetAddressStats)
			addresses.GET("/stats/city", addressHandler.GetAddressStatsByCity)
		}

		// Review routes (public for reading, protected for writing)
		reviews := v1.Group("/reviews")
		{
			// Public routes (no authentication required)
			reviews.GET("/product/:product_id", reviewHandler.GetReviewsByProduct)
			reviews.GET("/product/:product_id/verified", reviewHandler.GetVerifiedReviews)
			reviews.GET("/product/:product_id/rating", reviewHandler.GetAverageRating)
			reviews.GET("/product/:product_id/distribution", reviewHandler.GetRatingDistribution)
			reviews.GET("/product/:product_id/stats", reviewHandler.GetProductReviewStats)
			reviews.GET("/recent", reviewHandler.GetRecentReviews)
			reviews.GET("/search", reviewHandler.SearchReviews)
			reviews.GET("/stats", reviewHandler.GetReviewStats)
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

			// Brand management routes (require authentication and permissions)
			brandManagement := protected.Group("/brands")
			{
				// Create brand - requires write permission
				brandManagement.POST("", middleware.WritePermissionMiddleware(model.ResourceTypeBrand), brandHandler.CreateBrand)
				// Update brand - requires write permission
				brandManagement.PUT("/:id", middleware.WritePermissionMiddleware(model.ResourceTypeBrand), brandHandler.UpdateBrand)
				// Delete brand - requires delete permission
				brandManagement.DELETE("/:id", middleware.DeletePermissionMiddleware(model.ResourceTypeBrand), brandHandler.DeleteBrand)
				// Update status - requires write permission
				brandManagement.PATCH("/:id/status", middleware.WritePermissionMiddleware(model.ResourceTypeBrand), brandHandler.UpdateBrandStatus)
				// Bulk update status - requires write permission
				brandManagement.PATCH("/bulk-status", middleware.WritePermissionMiddleware(model.ResourceTypeBrand), brandHandler.BulkUpdateBrandStatus)
			}

			// Category management routes (require authentication and permissions)
			categoryManagement := protected.Group("/categories")
			{
				// Create category - requires write permission
				categoryManagement.POST("", middleware.WritePermissionMiddleware(model.ResourceTypeCategory), categoryHandler.CreateCategory)
				// Update category - requires write permission
				categoryManagement.PUT("/:id", middleware.WritePermissionMiddleware(model.ResourceTypeCategory), categoryHandler.UpdateCategory)
				// Delete category - requires delete permission
				categoryManagement.DELETE("/:id", middleware.DeletePermissionMiddleware(model.ResourceTypeCategory), categoryHandler.DeleteCategory)
				// Update status - requires write permission
				categoryManagement.PATCH("/:id/status", middleware.WritePermissionMiddleware(model.ResourceTypeCategory), categoryHandler.UpdateCategoryStatus)
				// Bulk update status - requires write permission
				categoryManagement.PATCH("/bulk-status", middleware.WritePermissionMiddleware(model.ResourceTypeCategory), categoryHandler.BulkUpdateCategoryStatus)
			}

			// Product management routes (require authentication and permissions)
			productManagement := protected.Group("/products")
			{
				// Create product - requires write permission
				productManagement.POST("", middleware.WritePermissionMiddleware(model.ResourceTypeProduct), productHandler.CreateProduct)
				// Update product - requires write permission
				productManagement.PUT("/:id", middleware.WritePermissionMiddleware(model.ResourceTypeProduct), productHandler.UpdateProduct)
				// Delete product - requires delete permission
				productManagement.DELETE("/:id", middleware.DeletePermissionMiddleware(model.ResourceTypeProduct), productHandler.DeleteProduct)
				// Update stock - requires write permission
				productManagement.PATCH("/:id/stock", middleware.WritePermissionMiddleware(model.ResourceTypeProduct), productHandler.UpdateProductStock)
				// Update status - requires write permission
				productManagement.PATCH("/:id/status", middleware.WritePermissionMiddleware(model.ResourceTypeProduct), productHandler.UpdateProductStatus)
				// Bulk update status - requires write permission
				productManagement.PATCH("/bulk-status", middleware.WritePermissionMiddleware(model.ResourceTypeProduct), productHandler.BulkUpdateProductStatus)
			}

			// Upload management routes (require authentication and permissions)
			uploadManagement := protected.Group("/upload")
			{
				// Upload file - requires write permission
				uploadManagement.POST("", middleware.WritePermissionMiddleware(model.ResourceTypeUpload), uploadHandler.UploadFile)
				// Upload image - requires write permission
				uploadManagement.POST("/image", middleware.WritePermissionMiddleware(model.ResourceTypeUpload), uploadHandler.UploadImage)
				// Upload brand logo - requires write permission
				uploadManagement.POST("/brand-logo", middleware.WritePermissionMiddleware(model.ResourceTypeUpload), uploadHandler.UploadBrandLogo)
				// Upload document - requires write permission
				uploadManagement.POST("/document", middleware.WritePermissionMiddleware(model.ResourceTypeUpload), uploadHandler.UploadDocument)
				// Upload multiple files - requires write permission
				uploadManagement.POST("/multiple", middleware.WritePermissionMiddleware(model.ResourceTypeUpload), uploadHandler.UploadMultipleFiles)
				// Delete file - requires delete permission
				uploadManagement.DELETE("", middleware.DeletePermissionMiddleware(model.ResourceTypeUpload), uploadHandler.DeleteFile)
			}

			// Inventory management routes (require authentication and permissions)
			inventoryManagement := protected.Group("/inventory")
			{
				// Inventory Movements
				// Create movement - requires write permission
				inventoryManagement.POST("/movements", middleware.WritePermissionMiddleware(model.ResourceTypeInventory), inventoryHandler.CreateMovement)
				// Get movements - requires read permission
				inventoryManagement.GET("/movements", middleware.ReadPermissionMiddleware(model.ResourceTypeInventory), inventoryHandler.GetMovements)
				// Get movement by ID - requires read permission
				inventoryManagement.GET("/movements/:id", middleware.ReadPermissionMiddleware(model.ResourceTypeInventory), inventoryHandler.GetMovementByID)
				// Update movement - requires write permission
				inventoryManagement.PUT("/movements/:id", middleware.WritePermissionMiddleware(model.ResourceTypeInventory), inventoryHandler.UpdateMovement)
				// Delete movement - requires delete permission
				inventoryManagement.DELETE("/movements/:id", middleware.DeletePermissionMiddleware(model.ResourceTypeInventory), inventoryHandler.DeleteMovement)
				// Approve movement - requires manage permission
				inventoryManagement.PATCH("/movements/:id/approve", middleware.ManagePermissionMiddleware(model.ResourceTypeInventory), inventoryHandler.ApproveMovement)
				// Complete movement - requires manage permission
				inventoryManagement.PATCH("/movements/:id/complete", middleware.ManagePermissionMiddleware(model.ResourceTypeInventory), inventoryHandler.CompleteMovement)
				// Get movements by product - requires read permission
				inventoryManagement.GET("/movements/product/:product_id", middleware.ReadPermissionMiddleware(model.ResourceTypeInventory), inventoryHandler.GetMovementsByProduct)
				// Get movements by reference - requires read permission
				inventoryManagement.GET("/movements/reference/:reference", middleware.ReadPermissionMiddleware(model.ResourceTypeInventory), inventoryHandler.GetMovementsByReference)

				// Stock Levels
				// Update stock level settings - requires manage permission
				inventoryManagement.PATCH("/stock-levels/product/:product_id/settings", middleware.ManagePermissionMiddleware(model.ResourceTypeInventory), inventoryHandler.UpdateStockLevelSettings)
				// Reserve stock - requires write permission
				inventoryManagement.POST("/stock-levels/product/:product_id/reserve", middleware.WritePermissionMiddleware(model.ResourceTypeInventory), inventoryHandler.ReserveStock)
				// Release stock - requires write permission
				inventoryManagement.POST("/stock-levels/product/:product_id/release", middleware.WritePermissionMiddleware(model.ResourceTypeInventory), inventoryHandler.ReleaseStock)

				// Inventory Adjustments
				// Create adjustment - requires manage permission
				inventoryManagement.POST("/adjustments", middleware.ManagePermissionMiddleware(model.ResourceTypeInventory), inventoryHandler.CreateAdjustment)
				// Get adjustments - requires read permission
				inventoryManagement.GET("/adjustments", middleware.ReadPermissionMiddleware(model.ResourceTypeInventory), inventoryHandler.GetAdjustments)
				// Get adjustment by ID - requires read permission
				inventoryManagement.GET("/adjustments/:id", middleware.ReadPermissionMiddleware(model.ResourceTypeInventory), inventoryHandler.GetAdjustmentByID)
				// Get adjustments by product - requires read permission
				inventoryManagement.GET("/adjustments/product/:product_id", middleware.ReadPermissionMiddleware(model.ResourceTypeInventory), inventoryHandler.GetAdjustmentsByProduct)

				// Reports
				// Get movement stats - requires read permission
				inventoryManagement.GET("/movements/stats", middleware.ReadPermissionMiddleware(model.ResourceTypeInventory), inventoryHandler.GetMovementStats)
			}

			// Admin routes (require admin role and system permissions)
			admin := protected.Group("/admin")
			admin.Use(authMiddleware.AdminMiddleware())
			{
				// User management - requires user admin permission
				admin.GET("/users", middleware.AdminPermissionMiddleware(model.ResourceTypeUser), func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Admin users endpoint"})
				})
				// System management - requires system admin permission
				admin.GET("/system", middleware.AdminPermissionMiddleware(model.ResourceTypeSystem), func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "System management endpoint"})
				})
			}

			// Permission management routes (require authentication and system permissions)
			permissionManagement := protected.Group("/permissions")
			{
				// Permissions - require system manage permission
				permissionManagement.POST("", middleware.ManagePermissionMiddleware(model.ResourceTypeSystem), permissionHandler.CreatePermission)
				permissionManagement.GET("", middleware.ReadPermissionMiddleware(model.ResourceTypeSystem), permissionHandler.GetAllPermissions)
				permissionManagement.GET("/name/:name", middleware.ReadPermissionMiddleware(model.ResourceTypeSystem), permissionHandler.GetPermissionByName)
				permissionManagement.GET("/resource/:resource", middleware.ReadPermissionMiddleware(model.ResourceTypeSystem), permissionHandler.GetPermissionsByResource)
				permissionManagement.GET("/:id", middleware.ReadPermissionMiddleware(model.ResourceTypeSystem), permissionHandler.GetPermissionByID)
				permissionManagement.PUT("/:id", middleware.ManagePermissionMiddleware(model.ResourceTypeSystem), permissionHandler.UpdatePermission)
				permissionManagement.DELETE("/:id", middleware.ManagePermissionMiddleware(model.ResourceTypeSystem), permissionHandler.DeletePermission)

				// Roles - require system manage permission
				permissionManagement.POST("/roles", middleware.ManagePermissionMiddleware(model.ResourceTypeSystem), permissionHandler.CreateRole)
				permissionManagement.GET("/roles", middleware.ReadPermissionMiddleware(model.ResourceTypeSystem), permissionHandler.GetAllRoles)
				permissionManagement.GET("/roles/name/:name", middleware.ReadPermissionMiddleware(model.ResourceTypeSystem), permissionHandler.GetRoleByName)
				permissionManagement.GET("/roles/:id", middleware.ReadPermissionMiddleware(model.ResourceTypeSystem), permissionHandler.GetRoleByID)
				permissionManagement.PUT("/roles/:id", middleware.ManagePermissionMiddleware(model.ResourceTypeSystem), permissionHandler.UpdateRole)
				permissionManagement.DELETE("/roles/:id", middleware.ManagePermissionMiddleware(model.ResourceTypeSystem), permissionHandler.DeleteRole)

				// Role Permissions - require system manage permission
				permissionManagement.POST("/roles/:role_id/permissions/:permission_id", middleware.ManagePermissionMiddleware(model.ResourceTypeSystem), permissionHandler.AssignPermissionToRole)
				permissionManagement.DELETE("/roles/:role_id/permissions/:permission_id", middleware.ManagePermissionMiddleware(model.ResourceTypeSystem), permissionHandler.RevokePermissionFromRole)
				permissionManagement.GET("/roles/:role_id/permissions", middleware.ReadPermissionMiddleware(model.ResourceTypeSystem), permissionHandler.GetRolePermissions)
				permissionManagement.PUT("/roles/:role_id/permissions", middleware.ManagePermissionMiddleware(model.ResourceTypeSystem), permissionHandler.UpdateRolePermissions)

				// User Permissions - require user manage permission
				permissionManagement.POST("/users/:user_id/permissions", middleware.ManagePermissionMiddleware(model.ResourceTypeUser), permissionHandler.AssignPermissionToUser)
				permissionManagement.DELETE("/users/:user_id/permissions/:permission_id", middleware.ManagePermissionMiddleware(model.ResourceTypeUser), permissionHandler.RevokePermissionFromUser)
				permissionManagement.GET("/users/:user_id/permissions", middleware.ReadPermissionMiddleware(model.ResourceTypeUser), permissionHandler.GetUserPermissions)
				permissionManagement.GET("/users/:user_id/effective-permissions", middleware.ReadPermissionMiddleware(model.ResourceTypeUser), permissionHandler.GetUserEffectivePermissions)
				permissionManagement.GET("/users/:user_id/permissions/resource/:resource", middleware.ReadPermissionMiddleware(model.ResourceTypeUser), permissionHandler.GetUserPermissionsForResource)

				// Audit & Logging - require system read permission
				permissionManagement.GET("/logs", middleware.ReadPermissionMiddleware(model.ResourceTypeSystem), permissionHandler.GetPermissionLogs)

				// Statistics - require system read permission
				permissionManagement.GET("/stats", middleware.ReadPermissionMiddleware(model.ResourceTypeSystem), permissionHandler.GetPermissionStats)
				permissionManagement.GET("/users/:user_id/stats", middleware.ReadPermissionMiddleware(model.ResourceTypeUser), permissionHandler.GetUserPermissionStats)

				// Utility - require system admin permission
				permissionManagement.POST("/initialize", middleware.AdminPermissionMiddleware(model.ResourceTypeSystem), permissionHandler.InitializeDefaultPermissions)
				permissionManagement.POST("/users/:user_id/sync-role/:role_name", middleware.ManagePermissionMiddleware(model.ResourceTypeUser), permissionHandler.SyncUserRole)
			}

			// Order management routes (require authentication and permissions)
			orderManagement := protected.Group("/orders")
			{
				// Order CRUD - requires order permissions
				orderManagement.POST("", middleware.WritePermissionMiddleware(model.ResourceTypeOrder), orderHandler.CreateOrder)
				orderManagement.GET("", middleware.ReadPermissionMiddleware(model.ResourceTypeOrder), orderHandler.GetAllOrders)
				orderManagement.GET("/:id", middleware.ReadPermissionMiddleware(model.ResourceTypeOrder), orderHandler.GetOrderByID)
				orderManagement.PUT("/:id", middleware.WritePermissionMiddleware(model.ResourceTypeOrder), orderHandler.UpdateOrder)
				orderManagement.DELETE("/:id", middleware.DeletePermissionMiddleware(model.ResourceTypeOrder), orderHandler.DeleteOrder)

				// Order status management - requires manage permission
				orderManagement.POST("/:id/cancel", middleware.ManagePermissionMiddleware(model.ResourceTypeOrder), orderHandler.CancelOrder)
				orderManagement.POST("/:id/confirm", middleware.ManagePermissionMiddleware(model.ResourceTypeOrder), orderHandler.ConfirmOrder)
				orderManagement.POST("/:id/ship", middleware.ManagePermissionMiddleware(model.ResourceTypeOrder), orderHandler.ShipOrder)
				orderManagement.POST("/:id/deliver", middleware.ManagePermissionMiddleware(model.ResourceTypeOrder), orderHandler.DeliverOrder)

				// Order items - requires read permission
				orderManagement.GET("/:id/items", middleware.ReadPermissionMiddleware(model.ResourceTypeOrder), orderHandler.GetOrderItems)

				// User orders - requires read permission
				orderManagement.GET("/user/:user_id", middleware.ReadPermissionMiddleware(model.ResourceTypeOrder), orderHandler.GetOrdersByUser)
			}

			// Admin order management routes (require admin role and order permissions)
			adminOrderManagement := protected.Group("/admin/orders")
			adminOrderManagement.Use(authMiddleware.AdminMiddleware())
			{
				// Admin can create orders for any user - requires order write permission
				adminOrderManagement.POST("/user/:user_id", middleware.WritePermissionMiddleware(model.ResourceTypeOrder), orderHandler.CreateOrderForUser)
			}

			// Cart management routes (require authentication)
			cartManagement := protected.Group("/carts")
			{
				// Cart CRUD - requires order write permission
				cartManagement.POST("", middleware.WritePermissionMiddleware(model.ResourceTypeOrder), orderHandler.CreateCart)
				cartManagement.GET("", middleware.ReadPermissionMiddleware(model.ResourceTypeOrder), orderHandler.GetCart)
				cartManagement.PUT("/:cart_id", middleware.WritePermissionMiddleware(model.ResourceTypeOrder), orderHandler.UpdateCart)
				cartManagement.DELETE("/:cart_id", middleware.DeletePermissionMiddleware(model.ResourceTypeOrder), orderHandler.DeleteCart)
				cartManagement.POST("/:cart_id/clear", middleware.WritePermissionMiddleware(model.ResourceTypeOrder), orderHandler.ClearCart)

				// Cart items - requires order write permission
				cartManagement.POST("/:cart_id/items", middleware.WritePermissionMiddleware(model.ResourceTypeOrder), orderHandler.AddToCart)
				cartManagement.PUT("/:cart_id/items/:item_id", middleware.WritePermissionMiddleware(model.ResourceTypeOrder), orderHandler.UpdateCartItem)
				cartManagement.DELETE("/:cart_id/items/:item_id", middleware.WritePermissionMiddleware(model.ResourceTypeOrder), orderHandler.RemoveFromCart)
				cartManagement.GET("/:cart_id/items", middleware.ReadPermissionMiddleware(model.ResourceTypeOrder), orderHandler.GetCartItems)

				// Convert cart to order - requires order write permission
				cartManagement.POST("/:cart_id/convert-to-order", middleware.WritePermissionMiddleware(model.ResourceTypeOrder), orderHandler.ConvertCartToOrder)
			}

			// Order statistics routes (require read permission)
			orderStats := protected.Group("/order-stats")
			{
				orderStats.GET("", middleware.ReadPermissionMiddleware(model.ResourceTypeOrder), orderHandler.GetOrderStats)
				orderStats.GET("/user/:user_id", middleware.ReadPermissionMiddleware(model.ResourceTypeOrder), orderHandler.GetOrderStatsByUser)
				orderStats.GET("/revenue", middleware.ReadPermissionMiddleware(model.ResourceTypeOrder), orderHandler.GetRevenueStats)
			}

			// Address management routes (require authentication and permissions)
			addressManagement := protected.Group("/addresses")
			{
				// Address CRUD - requires address permissions
				addressManagement.POST("", middleware.WritePermissionMiddleware(model.ResourceTypeAddress), addressHandler.CreateAddress)
				addressManagement.GET("", middleware.ReadPermissionMiddleware(model.ResourceTypeAddress), addressHandler.GetAllAddresses)
				addressManagement.GET("/:id", middleware.ReadPermissionMiddleware(model.ResourceTypeAddress), addressHandler.GetAddressByID)
				addressManagement.PUT("/:id", middleware.WritePermissionMiddleware(model.ResourceTypeAddress), addressHandler.UpdateAddress)
				addressManagement.DELETE("/:id", middleware.DeletePermissionMiddleware(model.ResourceTypeAddress), addressHandler.DeleteAddress)

				// User addresses - requires read permission
				addressManagement.GET("/user/:user_id", middleware.ReadPermissionMiddleware(model.ResourceTypeAddress), addressHandler.GetAddressesByUser)
				addressManagement.GET("/user/:user_id/default", middleware.ReadPermissionMiddleware(model.ResourceTypeAddress), addressHandler.GetDefaultAddressByUser)
				addressManagement.GET("/user/:user_id/type/:type", middleware.ReadPermissionMiddleware(model.ResourceTypeAddress), addressHandler.GetAddressesByType)
				addressManagement.GET("/user/:user_id/active", middleware.ReadPermissionMiddleware(model.ResourceTypeAddress), addressHandler.GetActiveAddressesByUser)
				addressManagement.POST("/user/:user_id/:address_id/set-default", middleware.WritePermissionMiddleware(model.ResourceTypeAddress), addressHandler.SetDefaultAddress)

				// Address statistics - requires read permission
				addressManagement.GET("/user/:user_id/stats", middleware.ReadPermissionMiddleware(model.ResourceTypeAddress), addressHandler.GetAddressStatsByUser)
			}

			// Review management routes (require authentication and permissions)
			reviewManagement := protected.Group("/reviews")
			{
				// Review CRUD - requires review permissions
				reviewManagement.POST("", middleware.WritePermissionMiddleware(model.ResourceTypeReview), reviewHandler.CreateReview)
				reviewManagement.GET("", middleware.ReadPermissionMiddleware(model.ResourceTypeReview), reviewHandler.GetAllReviews)
				reviewManagement.GET("/:id", middleware.ReadPermissionMiddleware(model.ResourceTypeReview), reviewHandler.GetReviewByID)
				reviewManagement.PUT("/:id", middleware.WritePermissionMiddleware(model.ResourceTypeReview), reviewHandler.UpdateReview)
				reviewManagement.DELETE("/:id", middleware.DeletePermissionMiddleware(model.ResourceTypeReview), reviewHandler.DeleteReview)

				// User reviews - requires read permission
				reviewManagement.GET("/user/:user_id", middleware.ReadPermissionMiddleware(model.ResourceTypeReview), reviewHandler.GetReviewsByUser)
				reviewManagement.GET("/user/:user_id/stats", middleware.ReadPermissionMiddleware(model.ResourceTypeReview), reviewHandler.GetUserReviewStats)

				// Review management - requires read permission
				reviewManagement.GET("/status/:status", middleware.ReadPermissionMiddleware(model.ResourceTypeReview), reviewHandler.GetReviewsByStatus)
				reviewManagement.GET("/type/:type", middleware.ReadPermissionMiddleware(model.ResourceTypeReview), reviewHandler.GetReviewsByType)

				// Helpful votes - requires write permission
				reviewManagement.POST("/:review_id/vote", middleware.WritePermissionMiddleware(model.ResourceTypeReview), reviewHandler.CreateHelpfulVote)
				reviewManagement.PUT("/:review_id/vote", middleware.WritePermissionMiddleware(model.ResourceTypeReview), reviewHandler.UpdateHelpfulVote)
				reviewManagement.DELETE("/:review_id/vote", middleware.WritePermissionMiddleware(model.ResourceTypeReview), reviewHandler.DeleteHelpfulVote)
				reviewManagement.GET("/:review_id/votes", middleware.ReadPermissionMiddleware(model.ResourceTypeReview), reviewHandler.GetHelpfulVotesByReview)

				// Review images - requires read permission
				reviewManagement.GET("/:review_id/images", middleware.ReadPermissionMiddleware(model.ResourceTypeReview), reviewHandler.GetReviewImagesByReview)
				reviewManagement.DELETE("/images/:image_id", middleware.DeletePermissionMiddleware(model.ResourceTypeReview), reviewHandler.DeleteReviewImage)
			}

			// Review moderation routes (require moderator role and review permissions)
			reviewModeration := protected.Group("/moderator/reviews")
			reviewModeration.Use(authMiddleware.ModeratorMiddleware())
			{
				// Review moderation - requires review manage permission
				reviewModeration.POST("/:review_id/moderate", middleware.ManagePermissionMiddleware(model.ResourceTypeReview), reviewHandler.ModerateReview)
				reviewModeration.GET("/pending", middleware.ReadPermissionMiddleware(model.ResourceTypeReview), reviewHandler.GetPendingReviews)
				reviewModeration.GET("/moderated/:moderator_id", middleware.ReadPermissionMiddleware(model.ResourceTypeReview), reviewHandler.GetModeratedReviews)
			}

			// Moderator routes (require moderator role and appropriate permissions)
			moderator := protected.Group("/moderator")
			moderator.Use(authMiddleware.ModeratorMiddleware())
			{
				// Moderator dashboard - requires read permission for multiple resources
				moderator.GET("/dashboard", middleware.MultiplePermissionMiddleware([]middleware.PermissionRequirement{
					middleware.BrandReadPermission,
					middleware.CategoryReadPermission,
					middleware.ProductReadPermission,
					middleware.InventoryReadPermission,
				}), func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Moderator dashboard"})
				})
				// Content moderation - requires write permission for content management
				moderator.GET("/content", middleware.WritePermissionMiddleware(model.ResourceTypeProduct), func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Content moderation"})
				})
			}
		}
	}
}
