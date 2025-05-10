package routers

import (
	auth "khoomi-api-io/api/internal/auth"
	"khoomi-api-io/api/internal/middleware"
	"khoomi-api-io/api/pkg/controllers"

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
		api.POST("/auth/google", controllers.HandleUserGoogleAuthentication())
		api.PUT("/auth/refresh-token", controllers.RefreshToken())
		api.DELETE("/logout", controllers.Logout())
		api.GET("/verify-email", controllers.VerifyEmail())
		api.POST("/send-password-reset", controllers.PasswordResetEmail())
		api.POST("/password-reset", controllers.PasswordReset())

		// Protected endpoints
		userRoutes(api)
		shopRoutes(api)
		listingRoutes(api)
		categoryRoutes(api)
	}

	return router
}

func userRoutes(api *gin.RouterGroup) {
	// Define the "/users" group
	// Ping endpoint
	api.GET("/ping", controllers.Ping)
	user := api.Group("/users")
	// Endpoint to get user by ID or email
	user.GET("/:userid", controllers.GetUser())
	{
		// Endpoint to get shops by owner user ID
		user.GET("/:userid/shops", controllers.GetShopByOwnerUserId())

		// Secured endpoints that require authentication
		secured := user.Group("").Use(auth.Auth())
		{
			// Get my session
			secured.GET("/:userid/session", controllers.GetMyActiveSession())

			// Change password endpoint
			secured.PUT("/:userid/change-password", controllers.ChangePassword())

			// Get delete user request
			secured.GET("/:userid/deletion", controllers.IsAccountPendingDeletion())
			// Send delet user request
			secured.POST("/:userid/deletion", controllers.SendDeleteUserAccount())
			// Cancel delet user request
			secured.DELETE("/:userid/deletion", controllers.CancelDeleteUserAccount())

			// Update first and last name endpoint
			secured.PUT("/:userid/", controllers.UpdateMyProfile())
			secured.PUT("/:userid/single", controllers.UpdateUserSingleField())

			// Notification settings endpoints
			secured.POST("/:userid/notification-settings", controllers.CreateUserNotificationSettings())
			secured.GET("/:userid/notification-settings", controllers.GetUserNotificationSettings())
			secured.PUT("/:userid/notification-settings", controllers.UpdateUserNotificationSettings())

			// User thumbnail endpoints
			secured.PUT("/:userid/thumbnail", controllers.UploadThumbnail())
			secured.DELETE("/:userid/thumbnail/:url", controllers.DeleteThumbnail())

			// User address endpoints
			secured.POST("/:userid/addresses", controllers.CreateUserAddress())
			secured.PUT("/:userid/addresses/:id", controllers.UpdateUserAddress())
			secured.GET("/:userid/addresses", controllers.GetUserAddresses())
			secured.DELETE("/:userid/addresses/:id", controllers.DeleteUserAddress())
			secured.PUT("/:userid/addresses/:id/default", controllers.ChangeDefaultAddress())

			// Send verify email endpoint
			secured.POST("/:userid/send-verify-email", controllers.SendVerifyEmail())

			// User birthdate endpoint
			secured.PUT("/:userid/birthdate", controllers.UpdateUserBirthdate())

			// Login histories endpoints
			secured.GET("/:userid/login-history", controllers.GetLoginHistories())
			secured.DELETE("/:userid/login-history", controllers.DeleteLoginHistories())
			secured.PUT("/:userid/login-notification", controllers.UpdateSecurityNotificationSetting())
			secured.GET("/:userid/login-notification", controllers.GetSecurityNotificationSetting())

			// Favorite shop endpoints
			secured.POST("/:userid/favorite-shop", controllers.AddRemoveFavoriteShop())

			// Wishlist endpoints
			secured.GET("/:userid/wishlist", controllers.GetUserWishlist())
			secured.POST("/:userid/wishlist", controllers.AddWishListItem())
			secured.DELETE("/:userid/wishlist", controllers.RemoveWishListItem())

			// Payment information endpoints
			payment := user.Group("/:userid/payment/cards").Use(auth.Auth())
			payment.POST("/", controllers.CreatePaymentCard())
			payment.GET("/", controllers.GetPaymentCards())
			payment.PUT("/:id/default", controllers.ChangeDefaultPaymentCard())
			payment.DELETE("/:id", controllers.DeletePaymentCard())

			// Seller Payment information endpoints
			secured.POST("/:userid/payment-information/", controllers.CreateSellerPaymentInformation())
			secured.GET("/:userid/payment-information/onboarded", controllers.CompletedPaymentOnboarding())
			secured.GET("/:userid/payment-information", controllers.GetSellerPaymentInformations())
			secured.PUT("/:userid/payment-information/:paymentInfoId/default", controllers.ChangeDefaultSellerPaymentInformation())
			secured.DELETE("/:userid/payment-information/:paymentInfoId", controllers.DeleteSellerPaymentInformation())

		}
	}
}

func shopRoutes(api *gin.RouterGroup) {
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
		shop.GET("/:shopid/shipping/all", controllers.GetShopShippingProfileInfos())
		shop.GET("/:shopid/shipping/:id", controllers.GetShopShippingProfileInfo())

		// Secured endpoints that require authentication
		secured := shop.Group("").Use(auth.Auth())
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
			secured.GET("/:shopid/followers/following", controllers.IsFollowingShop())
			// Endpoint to create/delete shop reviews
			shop.POST("/:shopid/reviews", controllers.CreateShopReview())
			secured.DELETE("/:shopid/reviews", controllers.DeleteMyReview())
			secured.DELETE("/:shopid/reviews/other", controllers.DeleteOtherReview())
			// Endpoint to create/update/delete shop return policies
			secured.POST("/:shopid/policies", controllers.CreateShopReturnPolicy())
			secured.PUT("/:shopid/policies", controllers.UpdateShopReturnPolicy())
			secured.GET("/:shopid/policies/:policyid", controllers.GetShopReturnPolicy())
			secured.GET("/:shopid/policies", controllers.GetShopReturnPolicies())
			secured.DELETE("/:shopid/policies/:policyid", controllers.DeleteShopReturnPolicy())
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
			secured := listing.Group("").Use(auth.Auth())
			{
				// Endpoint to create a single listing
				secured.POST("/:shopid/listings", controllers.CreateListing())
				secured.GET("/:shopid/listings/summary", controllers.GetMyListingsSummary())
				secured.GET("/:shopid/check-listing-onboarding", controllers.HasUserCreatedListingOnboarding())
			}
		}

	}
}

func listingRoutes(api *gin.RouterGroup) {
	listing := api.Group("/listings")
	// Get all listings -> /api/listings/?limit=50&skip=0&sort=date.created_at
	listing.GET("/", controllers.GetListings())
	// Get single listing by listingid -> /api/listings/{listingId}
	listing.GET("/:listingid", controllers.GetListing())
	// Secured endpoints that require authentication
	secured := listing.Group("").Use(auth.Auth())
	{
		secured.DELETE("/", controllers.DeleteListings())
		secured.PUT("/deactivate", controllers.DeactivateListings())
	}
}

func cartRoutes(api *gin.RouterGroup) {
	cart := api.Group("/carts")
	// Secured endpoints that require authentication
	secured := cart.Group("").Use(auth.Auth())
	{
		secured.GET("/", controllers.GetCartItems())
		secured.POST("/", controllers.SaveCartItem())
		secured.DELETE("/:cartId", controllers.DeleteCartItem())
		secured.DELETE("/many", controllers.DeleteCartItems())
		secured.DELETE("/clear", controllers.DeleteCartItems())
	}
}

func categoryRoutes(api *gin.RouterGroup) {
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
		secured := category.Group("").Use(auth.Auth())
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
