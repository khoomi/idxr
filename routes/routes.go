package routes

import (
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
	api := router.Group("/api", middleware.KhoomiRateLimiter())
	{
		// Public endpoints
		api.POST("/signup", controllers.CreateUser())
		api.POST("/auth", controllers.HandleUserAuthentication())
		api.PUT("/auth/refresh-token", controllers.RefreshToken())
		api.DELETE("/logout", controllers.Logout())
		api.GET("/verify-email", controllers.VerifyEmail())
		api.POST("/send-password-reset", controllers.PasswordResetEmail())
		api.POST("/password-reset", controllers.PasswordReset())

		// Protected endpoints
		userRoutes(api)
		ShopRoutes(api)
		ListingRoutes(api)
		CategoryRoutes(api)
	}

	return router
}

func userRoutes(api *gin.RouterGroup) {
	// Define the "/users" group
	user := api.Group("/users")
	{
		// Endpoint to get user by ID or email
		user.GET("/:userid/", controllers.GetUser())
		// Endpoint to get shops by owner user ID
		user.GET("/:userId/shops", controllers.GetShopByOwnerUserId())

		// Secured endpoints that require authentication
		secured := user.Group("").Use(middleware.Auth())
		{
			// Ping endpoint
			secured.GET("/ping", controllers.Ping)

			// Change password endpoint
			secured.PUT("/me/password-change", controllers.ChangePassword())

			// Get delete user request
			secured.GET("/me/delete", controllers.IsAccountPendingDeletion())
			// Send delete user request
			secured.POST("/me/delete", controllers.SendDeleteUserAccount())
			// Cancel delete user request
			secured.DELETE("/me/delete", controllers.CancelDeleteUserAccount())

			// Current user endpoint
			secured.GET("/me", controllers.CurrentUser)
			// Update first and last name endpoint
			secured.PUT("/me", controllers.UpdateMyProfile())

			// Notification settings endpoints
			secured.POST("/me/notification-settings", controllers.CreateUserNotificationSettings())
			secured.GET("/me/notification-settings", controllers.GetUserNotificationSettings())
			secured.PUT("/me/notification-settings", controllers.UpdateUserNotificationSettings())

			// User thumbnail endpoints
			secured.PUT("/:userId/thumbnail", controllers.UploadThumbnail())
			secured.DELETE("/:userId/thumbnail", controllers.DeleteThumbnail())

			// User address endpoints
			secured.POST("/:userId/addresses", controllers.CreateUserAddress())
			secured.PUT("/:userId/addresses/:addressId", controllers.UpdateUserAddress())
			secured.GET("/:userId/addresses", controllers.GetUserAddresses())
			secured.DELETE("/:userId/addresses/:addressId", controllers.DeleteUserAddress())

			// Send verify email endpoint
			secured.POST("/:userId/send-verify-email", controllers.SendVerifyEmail())

			// User birthdate endpoint
			secured.PUT("/:userId/birthdate", controllers.UpdateUserBirthdate())

			// Login histories endpoints
			secured.GET("/:userId/login-history", controllers.GetLoginHistories())
			secured.DELETE("/:userId/login-history", controllers.DeleteLoginHistories())
			secured.PUT("/:userId/login-notification", controllers.UpdateSecurityNotificationSetting())
			secured.GET("/:userId/login-notification", controllers.GetSecurityNotificationSetting())

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
			secured.GET("/:userId/payment-information/onboarded", controllers.CompletedPaymentOnboarding())
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
		// Endpoint to get shop followers
		shop.GET("/:shopid/followers", controllers.GetShopFollowers())
		// Endpoint to search for shops
		shop.GET("/search", controllers.SearchShops())
		// Endpoint to get shipping profile
		shop.GET("/:shopid/shipping/", controllers.GetShopShippingProfileInfos())
		shop.GET("/:shopid/shipping/:shippingProfileId", controllers.GetShopShippingProfileInfo())

		// Secured endpoints that require authentication
		secured := shop.Group("").Use(middleware.Auth())
		{
			// Endpoint to create a new shop
			secured.POST("", controllers.CreateShop())
			// Shop status
			secured.PUT("/:shopid/status", controllers.UpdateMyShopStatus())
			// update shop information
			secured.PUT("/:shopid/information", controllers.UpdateShopInformation())
			// Endpoint to check shop username availability
			secured.GET("/check/:username", controllers.CheckShopNameAvailability())
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
			// Endpoint to follow shop
			shop.POST("/:shopid/followers", controllers.FollowShop())
			secured.DELETE("/:shopid/followers", controllers.UnfollowShop())
			secured.DELETE("/:shopid/followers/other", controllers.RemoveOtherFollower())
			secured.GET("/:shopid/followers/is-following", controllers.IsFollowingShop())
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
			// Verification routes
			secured.POST("/:shopid/verification", controllers.CreateSellerVerificationProfile())
			secured.GET("/:shopid/verification", controllers.GetSellerVerificationProfile())
			// Compliance information endpoints
			secured.POST("/:shopid/compliance", controllers.CreateShopComplianceInformation())
			secured.GET("/:shopid/compliance", controllers.GetShopComplianceInformation())
		}

		listing := shop.Group("")
		{
			// Get shop listings -> /api/shops/{shopid}/listings/?limit=50&skip=0&sort=date.created_at
			shop.GET("/:shopid/listings", controllers.GetShopListings())
			secured := listing.Group("").Use(middleware.Auth())
			{
				// Endpoint to create a single listing
				secured.POST("/:shopid", controllers.CreateListing())
				secured.GET("/:shopid/check-listing-onboarding", controllers.HasUserCreatedListingOnboarding())
			}
		}

	}
}

func ListingRoutes(api *gin.RouterGroup) {
	// Define the "/listing" group
	listing := api.Group("/listings")
	// Get all listings -> /api/listings/?limit=50&skip=0&sort=date.created_at
	listing.GET("/", controllers.GetListings())
	// Get single listing by listingid -> /api/listings/{listingId}
	listing.GET("/:listingid", controllers.GetListing())
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
