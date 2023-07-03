package routes

import (
	"khoomi-api-io/khoomi_api/configs"
	"khoomi-api-io/khoomi_api/controllers"
	"khoomi-api-io/khoomi_api/middleware"

	"github.com/gin-gonic/gin"
)

func InitRoute() *gin.Engine {
	// Create a new Gin router
	router := gin.Default()

	// Apply the CORS middleware
	router.Use(middleware.CorsMiddleware())

	// Create the "/api" group for API endpoints
	api := router.Group("/api", configs.KhoomiRateLimiter())
	{
		// Public endpoints
		api.POST("/signup", controllers.CreateUser())
		api.POST("/auth", controllers.HandleUserAuthentication())
		api.DELETE("/logout", controllers.Logout())
		api.GET("/verify-email", controllers.VerifyEmail())
		api.POST("/send-password-reset", controllers.PasswordResetEmail())
		api.POST("/password-reset", controllers.PasswordReset())

		// Protected endpoints
		userRoutes(api)
		ShopRoutes(api)
		CategoryRoutes(api)
		ShippingRoutes(api)
	}

	return router
}

func userRoutes(api *gin.RouterGroup) {
	// Define the "/users" group
	user := api.Group("/users")
	{
		// Endpoint to get user by ID or email
		user.GET("/", controllers.GetUserByIDOrEmail())
		// Endpoint to get shops by owner user ID
		user.GET("/:userId/shops", controllers.GetShopByOwnerUserId())

		// Secured endpoints that require authentication
		secured := user.Group("").Use(middleware.Auth())
		{
			// Ping endpoint
			secured.GET("/ping", controllers.Ping)

			// Change password endpoint
			secured.PUT("/me/password-change", controllers.ChangePassword())

			// Delete user request
			secured.POST("/me/delete", controllers.SendDeleteUserAccount())
			// Cancel delete user request
			secured.DELETE("/me/delete", controllers.CancelDeleteUserAccount())
			// Current user endpoint
			secured.GET("/me", controllers.CurrentUser)
			// Update first and last name endpoint
			secured.PUT("/me", controllers.UpdateMyProfile())

			// Notification settings endpoints
			secured.POST("/:userId/notification-settings", controllers.CreateUserNotificationSettings())
			secured.GET("/:userId/notification-settings", controllers.GetUserNotificationSettings())
			secured.PUT("/:userId/notification-settings", controllers.UpdateUserNotificationSettings())

			// User thumbnail endpoints
			secured.PUT("/:userId/thumbnail", controllers.UploadThumbnail())
			secured.DELETE("/:userId/thumbnail", controllers.DeleteThumbnail())

			// User address endpoints
			secured.POST("/:userId/addresses", controllers.CreateUserAddress())
			secured.PUT("/:userId/addresses/:addressId", controllers.UpdateUserAddress())
			secured.GET("/:userId/addresses", controllers.GetUserAddresses())

			// Send verify email endpoint
			secured.POST("/:userId/send-verify-email", controllers.SendVerifyEmail())

			// User birthdate endpoint
			secured.PUT("/:userId/birthdate", controllers.UpdateUserBirthdate())

			// Login histories endpoints
			secured.GET("/:userId/login-history", controllers.GetLoginHistories())
			secured.DELETE("/:userId/login-history", controllers.DeleteLoginHistories())

			// Profile update endpoint
			secured.PUT("/:userId/update", controllers.UpdateUserSingleField())

			// Favorite shop endpoints
			secured.POST("/:userId/favorite-shop", controllers.AddRemoveFavoriteShop())

			// Wishlist endpoints
			secured.GET("/:userId/wishlist", controllers.GetUserWishlist())
			secured.POST("/:userId/wishlist", controllers.AddWishListItem())
			secured.DELETE("/:userId/wishlist", controllers.RemoveWishListItem())

			// Payment information endpoints
			secured.POST("/:userId/payment-information/", controllers.CreatePaymentInformation())
			secured.GET("/:userId/payment-information", controllers.GetPaymentInformations())
			secured.PUT("/:userId/payment-information/:paymentInfoId", controllers.ChangeDefaultPaymentInformation())
			secured.DELETE("/:userId/payment-information/:paymentInfoId", controllers.DeletePaymentInformation())
		}
	}
}

func ShopRoutes(api *gin.RouterGroup) {
	// Define the "/shops" group
	shop := api.Group("/shops")
	{
		// Endpoint to get all shops
		shop.GET("/", controllers.GetShops())
		// Endpoint to get a specific shop by ID
		shop.GET("/:shopid", controllers.GetShop())
		// Endpoint to get shop about information
		shop.GET("/:shopid/about", controllers.GetShopAbout())
		// Endpoint to get shop reviews
		shop.GET("/:shopid/reviews", controllers.GetShopReviews())
		// Endpoint to get shop members
		shop.GET("/:shopid/members", controllers.GetShopMembers())
		// Endpoint to search for shops
		shop.GET("/search", controllers.SearchShops())
		// Endpoint to get shipping profile
		shop.GET("/shipping/:infoId", controllers.GetShopShippingProfileInfo())

		// Secured endpoints that require authentication
		secured := shop.Group("").Use(middleware.Auth())
		{
			// Endpoint to create a new shop
			secured.POST("/", controllers.CreateShop())
			// Endpoint to check shop username availability
			secured.POST("/check/:username", controllers.CheckShopNameAvailability())
			// Endpoint to update shop logo
			secured.PUT("/:shopid/logo", controllers.UpdateShopLogo())
			// Endpoint to update shop banner
			secured.PUT("/:shopid/banner", controllers.UpdateShopBanner())
			// Endpoint to create/update shop about information
			secured.POST("/:shopid/about", controllers.CreateShopAbout())
			secured.PUT("/:shopid/about", controllers.UpdateShopAbout())
			secured.PUT("/:shopid/about/status", controllers.UpdateShopAboutStatus())
			// Endpoint to update shop vacation status
			secured.PUT("/:shopid/vacation", controllers.UpdateShopVacation())
			// Endpoint to update shop gallery
			secured.PUT("/:shopid/gallery", controllers.UpdateShopGallery())
			secured.DELETE("/:shopid/gallery", controllers.DeleteFromShopGallery())
			// Endpoint to update shop announcement
			secured.PUT("/:shopid/announcement", controllers.UpdateShopAnnouncement())
			// Endpoint to add/remove shop favorers
			secured.PUT("/:shopid/favorers", controllers.AddShopFavorer())
			secured.DELETE("/:shopid/favorers", controllers.RemoveShopFavorer())
			// Endpoint to join/leave shop members
			shop.POST("/:shopid/members", controllers.JoinShopMembers())
			secured.DELETE("/:shopid/members", controllers.LeaveShopMembers())
			secured.DELETE("/:shopid/members/other", controllers.RemoveOtherMember())
			// Endpoint to create/delete shop reviews
			shop.POST("/:shopid/reviews", controllers.CreateShopReview())
			secured.DELETE("/:shopid/reviews", controllers.DeleteMyReview())
			secured.DELETE("/:shopid/reviews/other", controllers.DeleteOtherReview())
			// Endpoint to create/update/delete shop return policies
			secured.POST("/:shopid/policies", controllers.CreateShopReturnPolicy())
			secured.PUT("/:shopid/policies", controllers.UpdateShopReturnPolicy())
			secured.GET("/:shopid/policies", controllers.GetShopReturnPolicy())
			secured.GET("/:shopid/policies/all", controllers.GetShopReturnPolicies())
			secured.DELETE("/:shopid/policies", controllers.DeleteShopReturnPolicy())
			// Shipping routes
			secured.POST("/:shopid/shipping", controllers.CreateShopShippingProfile())
			secured.PUT("/:shopid/shipping", controllers.UpdateShopShippingProfileInfo())
		}

	}
}

func CategoryRoutes(api *gin.RouterGroup) {
	// Define the "/categories" group
	category := api.Group("/categories")
	{
		// Endpoint to get all categories
		category.GET("/", controllers.GetAllCategories())
		// Endpoint to search for categories
		category.GET("/search", controllers.SearchCategories())
		// Endpoint to get category children
		category.GET("/:id/children", controllers.GetCategoryChildren())
		// Endpoint to get category ancestor
		category.GET("/:id/ancestor", controllers.GetCategoryAncestor())

		// Secured endpoints that require authentication
		secured := category.Group("").Use(middleware.Auth())
		{
			// Endpoint to create a single category
			secured.POST("/", controllers.CreateCategorySingle())
			// Endpoint to create multiple categories
			secured.POST("/multi", controllers.CreateCategoryMulti())
			// Endpoint to delete all categories
			secured.DELETE("/", controllers.DeleteAllCategories())
		}
	}
}

func ShippingRoutes(api *gin.RouterGroup) {
	// Define the "/shipping/info" group
	shipping := api.Group("/shipping/info")
	{
		// Endpoint to get shop shipping profile info by ID
		shipping.GET("/:infoId", controllers.GetShopShippingProfileInfo())

		// Secured endpoints that require authentication
		secured := shipping.Group("").Use(middleware.Auth())
		{
			// Endpoint to create shop shipping profile
			secured.POST("/", controllers.CreateShopShippingProfile())
			// Endpoint to update shop shipping profile info
			secured.PUT("/", controllers.UpdateShopShippingProfileInfo())
		}
	}
}
