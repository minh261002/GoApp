package router

import (
	"go_app/internal/handler"
	"go_app/internal/model"
	"go_app/internal/repository"
	"go_app/internal/service"
	"go_app/pkg/database"
	"go_app/pkg/middleware"
	"go_app/pkg/payment"
	"go_app/pkg/shipping"

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
	couponHandler := handler.NewCouponHandler()
	bannerHandler := handler.NewBannerHandler()
	sliderHandler := handler.NewSliderHandler()
	wishlistHandler := handler.NewWishlistHandler()
	// Initialize search service
	searchRepo := repository.NewSearchRepository()
	productRepo := repository.NewProductRepository()
	categoryRepo := repository.NewCategoryRepository()
	brandRepo := repository.NewBrandRepository()
	userRepo := repository.NewUserRepository()
	searchService := service.NewSearchService(searchRepo, productRepo, categoryRepo, brandRepo, userRepo)
	searchHandler := handler.NewSearchHandler(searchService)

	// Initialize notification service
	notificationRepo := repository.NewNotificationRepository()
	notificationService := service.NewNotificationService(notificationRepo, userRepo)
	notificationHandler := handler.NewNotificationHandler(notificationService)

	// Initialize email service
	emailRepo := repository.NewEmailRepository(database.GetDB())
	emailService := service.NewEmailService(emailRepo)
	emailHandler := handler.NewEmailHandler(emailService)

	// Initialize shipping service
	shippingRepo := repository.NewShippingRepository(database.GetDB())
	ghtkConfig := shipping.GHTKConfig{
		BaseURL:    "https://services.ghtk.vn",
		Token:      "YOUR_GHTK_TOKEN", // TODO: Get from config
		ShopID:     "YOUR_SHOP_ID",    // TODO: Get from config
		Timeout:    30,
		IsTestMode: false,
	}
	shippingService := service.NewShippingService(shippingRepo, repository.NewOrderRepository(), ghtkConfig)
	shippingHandler := handler.NewShippingHandler(shippingService)

	// Initialize order service
	orderService := service.NewOrderService()

	// Initialize cart handler (using order service)
	cartHandler := handler.NewCartHandler(orderService)

	// Initialize payment gateway service
	payOSConfig := payment.PayOSConfig{
		ClientID:    "YOUR_PAYOS_CLIENT_ID",    // TODO: Get from config
		APIKey:      "YOUR_PAYOS_API_KEY",      // TODO: Get from config
		ChecksumKey: "YOUR_PAYOS_CHECKSUM_KEY", // TODO: Get from config
		BaseURL:     "https://api-merchant.payos.vn",
	}
	paymentGatewayService := service.NewPaymentGatewayService(payOSConfig)
	paymentHandler := handler.NewPaymentHandler(orderService, paymentGatewayService)

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

		// Coupon routes (public for reading, protected for writing)
		coupons := v1.Group("/coupons")
		{
			// Public routes (no authentication required)
			coupons.GET("/active", couponHandler.GetActiveCoupons)
			coupons.GET("/type/:type", couponHandler.GetCouponsByType)
			coupons.GET("/search", couponHandler.SearchCoupons)
			coupons.GET("/stats", couponHandler.GetCouponStats)
		}

		// Point routes (public for reading, protected for writing)
		points := v1.Group("/points")
		{
			// Public routes (no authentication required)
			points.GET("/user/:user_id", couponHandler.GetPointByUserID)
			points.GET("/user/:user_id/balance", couponHandler.GetUserPointBalance)
			points.GET("/user/:user_id/transactions", couponHandler.GetPointTransactionsByUser)
			points.GET("/user/:user_id/history", couponHandler.GetPointHistory)
			points.GET("/user/:user_id/expired", couponHandler.GetExpiredPoints)
			points.GET("/user/:user_id/expiring", couponHandler.GetExpiringPoints)
			points.GET("/user/:user_id/stats", couponHandler.GetUserPointStats)
			points.GET("/stats", couponHandler.GetPointStats)
			points.GET("/top-earners", couponHandler.GetTopEarners)
		}

		// Banner public routes (no authentication required)
		banners := v1.Group("/banners")
		{
			// Public banner routes
			banners.GET("/active", bannerHandler.GetActiveBanners)
			banners.GET("/type/:type", bannerHandler.GetBannersByType)
			banners.GET("/position/:position", bannerHandler.GetBannersByPosition)
			banners.GET("/audience/:audience", bannerHandler.GetBannersByTargetAudience)
			banners.GET("/device/:device_type", bannerHandler.GetBannersByDeviceType)
			banners.GET("/search", bannerHandler.SearchBanners)
			banners.POST("/click", bannerHandler.TrackBannerClick)
			banners.POST("/view", bannerHandler.TrackBannerView)
		}

		// Slider public routes (no authentication required)
		sliders := v1.Group("/sliders")
		{
			// Public slider routes
			sliders.GET("/active", sliderHandler.GetActiveSliders)
			sliders.GET("/type/:type", sliderHandler.GetSlidersByType)
			sliders.GET("/audience/:audience", sliderHandler.GetSlidersByTargetAudience)
			sliders.GET("/device/:device_type", sliderHandler.GetSlidersByDeviceType)
			sliders.GET("/search", sliderHandler.SearchSliders)
			sliders.POST("/view", sliderHandler.TrackSliderView)
			sliders.POST("/item/click", sliderHandler.TrackSliderItemClick)
		}

		// Wishlist public routes (no authentication required)
		wishlists := v1.Group("/wishlists")
		{
			// Public wishlist routes
			wishlists.GET("/public", wishlistHandler.GetPublicWishlists)
			wishlists.GET("/search", wishlistHandler.SearchWishlists)
			wishlists.GET("/slug/:slug", wishlistHandler.GetWishlistBySlug)
			wishlists.GET("/:id", wishlistHandler.GetWishlistByID)
			wishlists.GET("/:id/items", wishlistHandler.GetWishlistItems)
			wishlists.GET("/share/:token", wishlistHandler.GetWishlistShareByToken)
			wishlists.GET("/stats", wishlistHandler.GetWishlistStats)
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

				// Payment routes for orders
				orderManagement.POST("/:id/payment/link", middleware.WritePermissionMiddleware(model.ResourceTypeOrder), paymentHandler.CreatePaymentLink)
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
				cartManagement.POST("", middleware.WritePermissionMiddleware(model.ResourceTypeOrder), cartHandler.CreateCart)
				cartManagement.GET("", middleware.ReadPermissionMiddleware(model.ResourceTypeOrder), cartHandler.GetCart)
				cartManagement.PUT("/:id", middleware.WritePermissionMiddleware(model.ResourceTypeOrder), cartHandler.UpdateCart)
				cartManagement.DELETE("/:id", middleware.DeletePermissionMiddleware(model.ResourceTypeOrder), cartHandler.DeleteCart)
				cartManagement.POST("/:id/clear", middleware.WritePermissionMiddleware(model.ResourceTypeOrder), cartHandler.ClearCart)

				// Cart items - requires order write permission
				cartManagement.POST("/:id/items", middleware.WritePermissionMiddleware(model.ResourceTypeOrder), cartHandler.AddToCart)
				cartManagement.PUT("/:id/items/:item_id", middleware.WritePermissionMiddleware(model.ResourceTypeOrder), cartHandler.UpdateCartItem)
				cartManagement.DELETE("/:id/items/:item_id", middleware.WritePermissionMiddleware(model.ResourceTypeOrder), cartHandler.RemoveFromCart)
				cartManagement.GET("/:id/items", middleware.ReadPermissionMiddleware(model.ResourceTypeOrder), cartHandler.GetCartItems)

				// Cart sync - requires order write permission
				cartManagement.POST("/:id/sync", middleware.WritePermissionMiddleware(model.ResourceTypeOrder), cartHandler.SyncCartWithUser)

				// Convert cart to order - requires order write permission
				cartManagement.POST("/:id/convert-to-order", middleware.WritePermissionMiddleware(model.ResourceTypeOrder), orderHandler.ConvertCartToOrder)
			}

			// Cart statistics routes (require read permission)
			cartStats := protected.Group("/carts")
			cartStats.Use(authMiddleware.AdminMiddleware())
			{
				cartStats.GET("/stats", middleware.ReadPermissionMiddleware(model.ResourceTypeOrder), cartHandler.GetCartStats)
			}

			// Order statistics routes (require read permission)
			orderStats := protected.Group("/order-stats")
			{
				orderStats.GET("", middleware.ReadPermissionMiddleware(model.ResourceTypeOrder), orderHandler.GetOrderStats)
				orderStats.GET("/user/:user_id", middleware.ReadPermissionMiddleware(model.ResourceTypeOrder), orderHandler.GetOrderStatsByUser)
				orderStats.GET("/revenue", middleware.ReadPermissionMiddleware(model.ResourceTypeOrder), orderHandler.GetRevenueStats)
			}

			// Payment management routes (require authentication and permissions)
			paymentManagement := protected.Group("/payments")
			{
				// Payment processing - requires order write permission
				paymentManagement.GET("/process/:order_code", middleware.WritePermissionMiddleware(model.ResourceTypeOrder), paymentHandler.ProcessPayment)
				paymentManagement.POST("/cancel/:order_code", middleware.WritePermissionMiddleware(model.ResourceTypeOrder), paymentHandler.CancelPayment)
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

			// Coupon management routes (require authentication and permissions)
			couponManagement := protected.Group("/coupons")
			{
				// Coupon CRUD - requires coupon permissions
				couponManagement.POST("", middleware.WritePermissionMiddleware(model.ResourceTypeCoupon), couponHandler.CreateCoupon)
				couponManagement.GET("", middleware.ReadPermissionMiddleware(model.ResourceTypeCoupon), couponHandler.GetAllCoupons)
				couponManagement.GET("/:id", middleware.ReadPermissionMiddleware(model.ResourceTypeCoupon), couponHandler.GetCouponByID)
				couponManagement.GET("/code/:code", middleware.ReadPermissionMiddleware(model.ResourceTypeCoupon), couponHandler.GetCouponByCode)
				couponManagement.PUT("/:id", middleware.WritePermissionMiddleware(model.ResourceTypeCoupon), couponHandler.UpdateCoupon)
				couponManagement.DELETE("/:id", middleware.DeletePermissionMiddleware(model.ResourceTypeCoupon), couponHandler.DeleteCoupon)

				// Coupon management - requires read permission
				couponManagement.GET("/expired", middleware.ReadPermissionMiddleware(model.ResourceTypeCoupon), couponHandler.GetExpiredCoupons)
				couponManagement.GET("/status/:status", middleware.ReadPermissionMiddleware(model.ResourceTypeCoupon), couponHandler.GetCouponsByStatus)

				// Coupon usage - requires write permission
				couponManagement.POST("/validate", middleware.ReadPermissionMiddleware(model.ResourceTypeCoupon), couponHandler.ValidateCoupon)
				couponManagement.POST("/use", middleware.WritePermissionMiddleware(model.ResourceTypeCoupon), couponHandler.UseCoupon)
				couponManagement.GET("/:coupon_id/usages", middleware.ReadPermissionMiddleware(model.ResourceTypeCoupon), couponHandler.GetCouponUsagesByCoupon)
				couponManagement.GET("/:coupon_id/usage-stats", middleware.ReadPermissionMiddleware(model.ResourceTypeCoupon), couponHandler.GetCouponUsageStats)

				// User coupon usages - requires read permission
				couponManagement.GET("/user/:user_id/usages", middleware.ReadPermissionMiddleware(model.ResourceTypeCoupon), couponHandler.GetCouponUsagesByUser)
				couponManagement.GET("/order/:order_id/usages", middleware.ReadPermissionMiddleware(model.ResourceTypeCoupon), couponHandler.GetCouponUsagesByOrder)
			}

			// Point management routes (require authentication and permissions)
			pointManagement := protected.Group("/points")
			{
				// Point operations - requires point permissions
				pointManagement.POST("/earn", middleware.WritePermissionMiddleware(model.ResourceTypePoint), couponHandler.EarnPoints)
				pointManagement.POST("/redeem", middleware.WritePermissionMiddleware(model.ResourceTypePoint), couponHandler.RedeemPoints)
				pointManagement.POST("/refund", middleware.WritePermissionMiddleware(model.ResourceTypePoint), couponHandler.RefundPoints)
				pointManagement.POST("/adjust", middleware.WritePermissionMiddleware(model.ResourceTypePoint), couponHandler.AdjustPoints)
				pointManagement.POST("/expire", middleware.WritePermissionMiddleware(model.ResourceTypePoint), couponHandler.ExpirePoints)

				// Point queries - requires read permission
				pointManagement.GET("/transaction/:id", middleware.ReadPermissionMiddleware(model.ResourceTypePoint), couponHandler.GetPointTransactionByID)
			}

			// Banner management routes (require authentication and permissions)
			bannerManagement := protected.Group("/banners")
			{
				// Banner CRUD - requires banner permissions
				bannerManagement.POST("", middleware.WritePermissionMiddleware(model.ResourceTypeBanner), bannerHandler.CreateBanner)
				bannerManagement.GET("", middleware.ReadPermissionMiddleware(model.ResourceTypeBanner), bannerHandler.GetAllBanners)
				bannerManagement.GET("/:id", middleware.ReadPermissionMiddleware(model.ResourceTypeBanner), bannerHandler.GetBannerByID)
				bannerManagement.PUT("/:id", middleware.WritePermissionMiddleware(model.ResourceTypeBanner), bannerHandler.UpdateBanner)
				bannerManagement.DELETE("/:id", middleware.WritePermissionMiddleware(model.ResourceTypeBanner), bannerHandler.DeleteBanner)

				// Banner management - requires banner permissions
				bannerManagement.GET("/type/:type", middleware.ReadPermissionMiddleware(model.ResourceTypeBanner), bannerHandler.GetBannersByType)
				bannerManagement.GET("/position/:position", middleware.ReadPermissionMiddleware(model.ResourceTypeBanner), bannerHandler.GetBannersByPosition)
				bannerManagement.GET("/status/:status", middleware.ReadPermissionMiddleware(model.ResourceTypeBanner), bannerHandler.GetBannersByStatus)
				bannerManagement.GET("/audience/:audience", middleware.ReadPermissionMiddleware(model.ResourceTypeBanner), bannerHandler.GetBannersByTargetAudience)
				bannerManagement.GET("/device/:device_type", middleware.ReadPermissionMiddleware(model.ResourceTypeBanner), bannerHandler.GetBannersByDeviceType)
				bannerManagement.GET("/search", middleware.ReadPermissionMiddleware(model.ResourceTypeBanner), bannerHandler.SearchBanners)

				// Banner analytics - requires read permission
				bannerManagement.GET("/stats", middleware.ReadPermissionMiddleware(model.ResourceTypeBanner), bannerHandler.GetBannerStats)
				bannerManagement.GET("/:id/analytics", middleware.ReadPermissionMiddleware(model.ResourceTypeBanner), bannerHandler.GetBannerAnalytics)
				bannerManagement.GET("/expired", middleware.ReadPermissionMiddleware(model.ResourceTypeBanner), bannerHandler.GetExpiredBanners)
				bannerManagement.GET("/to-activate", middleware.ReadPermissionMiddleware(model.ResourceTypeBanner), bannerHandler.GetBannersToActivate)
				bannerManagement.PUT("/:id/status", middleware.WritePermissionMiddleware(model.ResourceTypeBanner), bannerHandler.UpdateBannerStatus)
			}

			// Slider management routes (require authentication and permissions)
			sliderManagement := protected.Group("/sliders")
			{
				// Slider CRUD - requires slider permissions
				sliderManagement.POST("", middleware.WritePermissionMiddleware(model.ResourceTypeSlider), sliderHandler.CreateSlider)
				sliderManagement.GET("", middleware.ReadPermissionMiddleware(model.ResourceTypeSlider), sliderHandler.GetAllSliders)
				sliderManagement.GET("/:id", middleware.ReadPermissionMiddleware(model.ResourceTypeSlider), sliderHandler.GetSliderByID)
				sliderManagement.PUT("/:id", middleware.WritePermissionMiddleware(model.ResourceTypeSlider), sliderHandler.UpdateSlider)
				sliderManagement.DELETE("/:id", middleware.WritePermissionMiddleware(model.ResourceTypeSlider), sliderHandler.DeleteSlider)

				// Slider management - requires slider permissions
				sliderManagement.GET("/type/:type", middleware.ReadPermissionMiddleware(model.ResourceTypeSlider), sliderHandler.GetSlidersByType)
				sliderManagement.GET("/status/:status", middleware.ReadPermissionMiddleware(model.ResourceTypeSlider), sliderHandler.GetSlidersByStatus)
				sliderManagement.GET("/audience/:audience", middleware.ReadPermissionMiddleware(model.ResourceTypeSlider), sliderHandler.GetSlidersByTargetAudience)
				sliderManagement.GET("/device/:device_type", middleware.ReadPermissionMiddleware(model.ResourceTypeSlider), sliderHandler.GetSlidersByDeviceType)
				sliderManagement.GET("/search", middleware.ReadPermissionMiddleware(model.ResourceTypeSlider), sliderHandler.SearchSliders)

				// Slider analytics - requires read permission
				sliderManagement.GET("/stats", middleware.ReadPermissionMiddleware(model.ResourceTypeSlider), sliderHandler.GetSliderStats)
				sliderManagement.GET("/:id/analytics", middleware.ReadPermissionMiddleware(model.ResourceTypeSlider), sliderHandler.GetSliderAnalytics)
				sliderManagement.GET("/expired", middleware.ReadPermissionMiddleware(model.ResourceTypeSlider), sliderHandler.GetExpiredSliders)
				sliderManagement.GET("/to-activate", middleware.ReadPermissionMiddleware(model.ResourceTypeSlider), sliderHandler.GetSlidersToActivate)
				sliderManagement.PUT("/:id/status", middleware.WritePermissionMiddleware(model.ResourceTypeSlider), sliderHandler.UpdateSliderStatus)

				// Slider items - requires slider permissions
				sliderManagement.POST("/:slider_id/items", middleware.WritePermissionMiddleware(model.ResourceTypeSlider), sliderHandler.CreateSliderItem)
				sliderManagement.GET("/:slider_id/items", middleware.ReadPermissionMiddleware(model.ResourceTypeSlider), sliderHandler.GetSliderItemsBySlider)
				sliderManagement.GET("/items/:id", middleware.ReadPermissionMiddleware(model.ResourceTypeSlider), sliderHandler.GetSliderItemByID)
				sliderManagement.PUT("/items/:id", middleware.WritePermissionMiddleware(model.ResourceTypeSlider), sliderHandler.UpdateSliderItem)
				sliderManagement.DELETE("/items/:id", middleware.WritePermissionMiddleware(model.ResourceTypeSlider), sliderHandler.DeleteSliderItem)
				sliderManagement.PUT("/:slider_id/reorder", middleware.WritePermissionMiddleware(model.ResourceTypeSlider), sliderHandler.ReorderSliderItems)
			}

			// Wishlist management routes (require authentication and permissions)
			wishlistManagement := protected.Group("/wishlists")
			{
				// Wishlist CRUD - requires wishlist permissions
				wishlistManagement.POST("", middleware.WritePermissionMiddleware(model.ResourceTypeWishlist), wishlistHandler.CreateWishlist)
				wishlistManagement.GET("", middleware.ReadPermissionMiddleware(model.ResourceTypeWishlist), wishlistHandler.GetWishlistsByUser)
				wishlistManagement.GET("/:id", middleware.ReadPermissionMiddleware(model.ResourceTypeWishlist), wishlistHandler.GetWishlistByID)
				wishlistManagement.PUT("/:id", middleware.WritePermissionMiddleware(model.ResourceTypeWishlist), wishlistHandler.UpdateWishlist)
				wishlistManagement.DELETE("/:id", middleware.WritePermissionMiddleware(model.ResourceTypeWishlist), wishlistHandler.DeleteWishlist)

				// Wishlist management - requires wishlist permissions
				wishlistManagement.PUT("/:id/set-default", middleware.WritePermissionMiddleware(model.ResourceTypeWishlist), wishlistHandler.SetDefaultWishlist)
				wishlistManagement.GET("/search", middleware.ReadPermissionMiddleware(model.ResourceTypeWishlist), wishlistHandler.SearchWishlists)

				// Wishlist items - requires wishlist permissions
				wishlistManagement.POST("/items", middleware.WritePermissionMiddleware(model.ResourceTypeWishlist), wishlistHandler.AddItemToWishlist)
				wishlistManagement.GET("/items/:id", middleware.ReadPermissionMiddleware(model.ResourceTypeWishlist), wishlistHandler.GetWishlistItemByID)
				wishlistManagement.PUT("/items/:id", middleware.WritePermissionMiddleware(model.ResourceTypeWishlist), wishlistHandler.UpdateWishlistItem)
				wishlistManagement.DELETE("/items/:id", middleware.WritePermissionMiddleware(model.ResourceTypeWishlist), wishlistHandler.DeleteWishlistItem)
				wishlistManagement.PUT("/:wishlist_id/reorder", middleware.WritePermissionMiddleware(model.ResourceTypeWishlist), wishlistHandler.ReorderWishlistItems)
				wishlistManagement.PUT("/items/:item_id/move", middleware.WritePermissionMiddleware(model.ResourceTypeWishlist), wishlistHandler.MoveItemToWishlist)

				// Favorites - requires wishlist permissions
				wishlistManagement.POST("/favorites", middleware.WritePermissionMiddleware(model.ResourceTypeWishlist), wishlistHandler.AddToFavorites)
				wishlistManagement.GET("/favorites", middleware.ReadPermissionMiddleware(model.ResourceTypeWishlist), wishlistHandler.GetFavoritesByUser)
				wishlistManagement.GET("/favorites/:id", middleware.ReadPermissionMiddleware(model.ResourceTypeWishlist), wishlistHandler.GetFavoriteByID)
				wishlistManagement.PUT("/favorites/:id", middleware.WritePermissionMiddleware(model.ResourceTypeWishlist), wishlistHandler.UpdateFavorite)
				wishlistManagement.DELETE("/favorites/:id", middleware.WritePermissionMiddleware(model.ResourceTypeWishlist), wishlistHandler.RemoveFromFavorites)
				wishlistManagement.DELETE("/favorites/product/:product_id", middleware.WritePermissionMiddleware(model.ResourceTypeWishlist), wishlistHandler.RemoveFromFavoritesByProduct)

				// Wishlist sharing - requires wishlist permissions
				wishlistManagement.POST("/:id/share", middleware.WritePermissionMiddleware(model.ResourceTypeWishlist), wishlistHandler.ShareWishlist)
				wishlistManagement.GET("/:id/shares", middleware.ReadPermissionMiddleware(model.ResourceTypeWishlist), wishlistHandler.GetWishlistSharesByWishlist)
				wishlistManagement.GET("/shares", middleware.ReadPermissionMiddleware(model.ResourceTypeWishlist), wishlistHandler.GetWishlistSharesByUser)

				// Analytics - requires read permission
				wishlistManagement.GET("/:id/analytics", middleware.ReadPermissionMiddleware(model.ResourceTypeWishlist), wishlistHandler.GetWishlistAnalytics)
				wishlistManagement.GET("/stats", middleware.ReadPermissionMiddleware(model.ResourceTypeWishlist), wishlistHandler.GetUserWishlistStats)
				wishlistManagement.POST("/:id/track-view", wishlistHandler.TrackWishlistView)
				wishlistManagement.POST("/items/:id/track-view", wishlistHandler.TrackWishlistItemView)
				wishlistManagement.POST("/items/:id/track-click", wishlistHandler.TrackWishlistItemClick)

				// Price tracking - requires read permission
				wishlistManagement.POST("/update-prices", middleware.ReadPermissionMiddleware(model.ResourceTypeWishlist), wishlistHandler.UpdateWishlistItemPrices)
				wishlistManagement.GET("/price-changes", middleware.ReadPermissionMiddleware(model.ResourceTypeWishlist), wishlistHandler.GetItemsWithPriceChanges)
				wishlistManagement.GET("/price-notifications", middleware.ReadPermissionMiddleware(model.ResourceTypeWishlist), wishlistHandler.GetItemsForPriceNotification)
			}

			// Search routes (public and protected)
			search := v1.Group("/search")
			{
				// Public search routes
				search.GET("/products", searchHandler.SearchProducts)
				search.GET("/suggestions", searchHandler.GetSearchSuggestions)
				search.GET("/filter-options", searchHandler.GetFilterOptions)
				search.GET("/popular", searchHandler.GetPopularSearches)
				search.GET("/trends", searchHandler.GetSearchTrends)
			}

			// Protected search routes
			searchProtected := protected.Group("/search")
			{
				// Search analytics - requires read permission
				searchProtected.GET("/analytics", middleware.ReadPermissionMiddleware(model.ResourceTypeProduct), searchHandler.GetSearchAnalytics)
				searchProtected.GET("/logs", middleware.ReadPermissionMiddleware(model.ResourceTypeProduct), searchHandler.GetSearchLogs)
				searchProtected.GET("/index-stats", middleware.ReadPermissionMiddleware(model.ResourceTypeProduct), searchHandler.GetSearchIndexStats)

				// Search index management - requires write permission
				searchProtected.POST("/index/create", middleware.WritePermissionMiddleware(model.ResourceTypeProduct), searchHandler.CreateSearchIndex)
				searchProtected.PUT("/index/update/:product_id", middleware.WritePermissionMiddleware(model.ResourceTypeProduct), searchHandler.UpdateSearchIndex)
				searchProtected.DELETE("/index/delete/:product_id", middleware.WritePermissionMiddleware(model.ResourceTypeProduct), searchHandler.DeleteSearchIndex)
				searchProtected.POST("/index/rebuild", middleware.WritePermissionMiddleware(model.ResourceTypeProduct), searchHandler.RebuildSearchIndex)
				searchProtected.DELETE("/logs/delete", middleware.WritePermissionMiddleware(model.ResourceTypeProduct), searchHandler.DeleteSearchLogs)
			}

			// Notification routes (require authentication)
			notifications := protected.Group("/notifications")
			{
				// Notification CRUD
				notifications.POST("", notificationHandler.CreateNotification)
				notifications.GET("", notificationHandler.GetNotificationsByUser)
				notifications.GET("/:id", notificationHandler.GetNotificationByID)
				notifications.PUT("/:id", notificationHandler.UpdateNotification)
				notifications.DELETE("/:id", notificationHandler.DeleteNotification)

				// Notification actions
				notifications.POST("/:id/read", notificationHandler.MarkAsRead)
				notifications.POST("/:id/unread", notificationHandler.MarkAsUnread)
				notifications.POST("/:id/archive", notificationHandler.MarkAsArchived)
				notifications.POST("/:id/unarchive", notificationHandler.MarkAsUnarchived)

				// Bulk actions
				notifications.POST("/bulk/read", notificationHandler.BulkMarkAsRead)
				notifications.POST("/bulk/archive", notificationHandler.BulkMarkAsArchived)

				// Statistics and search
				notifications.GET("/unread-count", notificationHandler.GetUnreadNotificationCount)
				notifications.GET("/stats", notificationHandler.GetNotificationStats)
				notifications.GET("/search", notificationHandler.SearchNotifications)
			}

			// Notification Templates routes (require authentication)
			notificationTemplates := protected.Group("/notification-templates")
			{
				notificationTemplates.POST("", notificationHandler.CreateNotificationTemplate)
				notificationTemplates.GET("", notificationHandler.GetNotificationTemplates)
				notificationTemplates.GET("/:id", notificationHandler.GetNotificationTemplateByID)
				notificationTemplates.PUT("/:id", notificationHandler.UpdateNotificationTemplate)
				notificationTemplates.DELETE("/:id", notificationHandler.DeleteNotificationTemplate)
			}

			// Notification Preferences routes (require authentication)
			notificationPreferences := protected.Group("/notification-preferences")
			{
				notificationPreferences.POST("", notificationHandler.CreateNotificationPreference)
				notificationPreferences.GET("", notificationHandler.GetNotificationPreferencesByUser)
				notificationPreferences.PUT("/:id", notificationHandler.UpdateNotificationPreference)
				notificationPreferences.DELETE("/:id", notificationHandler.DeleteNotificationPreference)
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

		// Payment public routes (no authentication required)
		payments := v1.Group("/payments")
		{
			// Public payment routes
			payments.GET("/methods", paymentHandler.GetPaymentMethods)
			payments.POST("/webhook/:payment_method", paymentHandler.HandleWebhook)
		}

		// Shipping public routes (no authentication required)
		shipping := v1.Group("/shipping")
		{
			// Public shipping routes
			shipping.POST("/calculate", shippingHandler.CalculateShipping)
			shipping.POST("/calculate/ghtk", shippingHandler.CalculateShippingWithGHTK)
			shipping.GET("/providers/active", shippingHandler.GetActiveShippingProviders)
			shipping.GET("/orders/tracking/:tracking_code", shippingHandler.GetShippingOrderByTrackingCode)
			shipping.POST("/webhook/:provider", shippingHandler.HandleShippingWebhook)
		}

		// Email management routes (require authentication)
		emailManagement := protected.Group("/email")
		{
			// Email templates - requires system write permission
			emailTemplates := emailManagement.Group("/templates")
			emailTemplates.Use(middleware.WritePermissionMiddleware(model.ResourceTypeSystem))
			{
				emailTemplates.POST("", emailHandler.CreateEmailTemplate)
				emailTemplates.GET("", emailHandler.GetAllEmailTemplates)
				emailTemplates.GET("/type/:type", emailHandler.GetEmailTemplatesByType)
				emailTemplates.GET("/:id", emailHandler.GetEmailTemplateByID)
				emailTemplates.PUT("/:id", emailHandler.UpdateEmailTemplate)
				emailTemplates.DELETE("/:id", emailHandler.DeleteEmailTemplate)
			}

			// Email sending - requires system write permission
			emailSending := emailManagement.Group("/send")
			emailSending.Use(middleware.WritePermissionMiddleware(model.ResourceTypeSystem))
			{
				emailSending.POST("", emailHandler.SendEmail)
				emailSending.POST("/template/:template_name", emailHandler.SendEmailWithTemplate)
				emailSending.POST("/process-queue", emailHandler.ProcessEmailQueue)
				emailSending.POST("/retry-failed", emailHandler.RetryFailedEmails)
				emailSending.POST("/test-retry", emailHandler.TestEmailRetry)
			}

			// Email monitoring - requires system read permission
			emailMonitoring := emailManagement.Group("/monitor")
			emailMonitoring.Use(middleware.ReadPermissionMiddleware(model.ResourceTypeSystem))
			{
				emailMonitoring.GET("/queue-stats", emailHandler.GetEmailQueueStats)
				emailMonitoring.GET("/logs", emailHandler.GetEmailLogs)
				emailMonitoring.GET("/stats", emailHandler.GetEmailStats)
			}

			// Email config - requires system write permission
			emailConfigs := emailManagement.Group("/configs")
			emailConfigs.Use(middleware.WritePermissionMiddleware(model.ResourceTypeSystem))
			{
				emailConfigs.POST("", emailHandler.CreateEmailConfig)
				emailConfigs.GET("", emailHandler.GetAllEmailConfigs)
				emailConfigs.GET("/active", emailHandler.GetActiveEmailConfig)
				emailConfigs.GET("/:id", emailHandler.GetEmailConfigByID)
				emailConfigs.PUT("/:id", emailHandler.UpdateEmailConfig)
				emailConfigs.DELETE("/:id", emailHandler.DeleteEmailConfig)
			}
		}

		// Shipping management routes (require authentication)
		shippingManagement := protected.Group("/shipping")
		{
			// Shipping providers - requires system write permission
			shippingProviders := shippingManagement.Group("/providers")
			shippingProviders.Use(middleware.WritePermissionMiddleware(model.ResourceTypeSystem))
			{
				shippingProviders.POST("", shippingHandler.CreateShippingProvider)
				shippingProviders.GET("", shippingHandler.GetAllShippingProviders)
				shippingProviders.GET("/:id", shippingHandler.GetShippingProviderByID)
				shippingProviders.PUT("/:id", shippingHandler.UpdateShippingProvider)
				shippingProviders.DELETE("/:id", shippingHandler.DeleteShippingProvider)
			}

			// Shipping orders - requires order write permission
			shippingOrders := shippingManagement.Group("/orders")
			shippingOrders.Use(middleware.WritePermissionMiddleware(model.ResourceTypeOrder))
			{
				shippingOrders.POST("", shippingHandler.CreateShippingOrder)
				shippingOrders.GET("", shippingHandler.GetShippingOrders)
				shippingOrders.GET("/:id", shippingHandler.GetShippingOrderByID)
				shippingOrders.GET("/order/:order_id", shippingHandler.GetShippingOrderByOrderID)
				shippingOrders.POST("/:id/cancel", shippingHandler.CancelShippingOrder)
			}

			// Shipping tracking - requires order read permission
			shippingTracking := shippingManagement.Group("/tracking")
			shippingTracking.Use(middleware.ReadPermissionMiddleware(model.ResourceTypeOrder))
			{
				shippingTracking.GET("/:order_id", shippingHandler.GetShippingTracking)
			}

			// Shipping statistics - requires system read permission
			shippingStats := shippingManagement.Group("/stats")
			shippingStats.Use(middleware.ReadPermissionMiddleware(model.ResourceTypeSystem))
			{
				shippingStats.GET("", shippingHandler.GetShippingStats)
				shippingStats.GET("/provider/:provider_id", shippingHandler.GetShippingStatsByProvider)
			}
		}
	}
}
