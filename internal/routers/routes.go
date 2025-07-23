package routers

import (
	auth "khoomi-api-io/api/internal/auth"
	"khoomi-api-io/api/internal/container"
	"khoomi-api-io/api/internal/middleware"
	"khoomi-api-io/api/pkg/controllers"

	"github.com/gin-gonic/gin"
)

// InitRefactoredRoute creates a new Gin router with service layer architecture
func InitRoute() *gin.Engine {
	serviceContainer := container.NewServiceContainer()
	router := gin.Default()
	router.Use(middleware.CorsMiddleware())

	api := router.Group("/v1", middleware.KhoomiRateLimiter())
	{
		// Public authentication routes
		setupAuthRoutes(api)

		// Protected feature routes
		userRoutesRefactored(api)
		shopRoutesRefactored(api, serviceContainer)
		listingRoutesRefactored(api, serviceContainer)
		cartRoutesRefactored(api, serviceContainer)
		// categoryRoutes(api) // Keep existing
	}

	return router
}

// setupAuthRoutes configures public authentication endpoints
func setupAuthRoutes(api *gin.RouterGroup) {
	api.POST("/signup", controllers.CreateUser())
	api.POST("/auth", controllers.HandleUserAuthentication())
	api.POST("/auth/google", controllers.HandleUserGoogleAuthentication())
	api.PUT("/auth/refresh-token", controllers.RefreshToken())
	api.DELETE("/logout", controllers.Logout())
	api.GET("/verify-email", controllers.VerifyEmail())
	api.POST("/send-password-reset", controllers.PasswordResetEmail())
	api.POST("/password-reset", controllers.PasswordReset())
}

// userRoutesRefactored configures user-related endpoints
func userRoutesRefactored(api *gin.RouterGroup) {
	api.GET("/ping", controllers.Ping)

	user := api.Group("/users")

	user.GET("/:userid", controllers.GetUser())
	user.GET("/:userid/shops", controllers.GetShopByOwnerUserId())

	{
		secured := user.Group("").Use(auth.Auth())
		// Session management
		secured.GET("/:userid/session", controllers.GetMyActiveSession())
		secured.PUT("/:userid/change-password", controllers.ChangePassword())

		// Account deletion
		secured.GET("/:userid/deletion", controllers.IsAccountPendingDeletion())
		secured.POST("/:userid/deletion", controllers.SendDeleteUserAccount())
		secured.DELETE("/:userid/deletion", controllers.CancelDeleteUserAccount())

		// Profile management
		secured.PUT("/:userid/", controllers.UpdateMyProfile())
		secured.PUT("/:userid/single", controllers.UpdateUserSingleField())
		secured.PUT("/:userid/thumbnail", controllers.UploadThumbnail())
		secured.DELETE("/:userid/thumbnail/:url", controllers.DeleteThumbnail())
		secured.PUT("/:userid/birthdate", controllers.UpdateUserBirthdate())
		secured.POST("/:userid/send-verify-email", controllers.SendVerifyEmail())

		// Notification settings
		secured.POST("/:userid/notification-settings", controllers.CreateUserNotificationSettings())
		secured.GET("/:userid/notification-settings", controllers.GetUserNotificationSettings())
		secured.PUT("/:userid/notification-settings", controllers.UpdateUserNotificationSettings())

		// Address management
		secured.POST("/:userid/addresses", controllers.CreateUserAddress())
		secured.PUT("/:userid/addresses/:id", controllers.UpdateUserAddress())
		secured.GET("/:userid/addresses", controllers.GetUserAddresses())
		secured.DELETE("/:userid/addresses/:id", controllers.DeleteUserAddress())
		secured.PUT("/:userid/addresses/:id/default", controllers.ChangeDefaultAddress())

		// Security & login history
		secured.GET("/:userid/login-history", controllers.GetLoginHistories())
		secured.DELETE("/:userid/login-history", controllers.DeleteLoginHistories())
		secured.PUT("/:userid/login-notification", controllers.UpdateSecurityNotificationSetting())
		secured.GET("/:userid/login-notification", controllers.GetSecurityNotificationSetting())

		// Wishlist management
		secured.GET("/:userid/wishlist", controllers.GetUserWishlist())
		secured.POST("/:userid/wishlist", controllers.AddWishListItem())
		secured.DELETE("/:userid/wishlist", controllers.RemoveWishListItem())

		// Favorites
		secured.GET("/favorite/shops", controllers.IsShopFavorited())
		secured.POST("/favorite/shops", controllers.ToggleFavoriteShop())
		secured.GET("/favorite/listings", controllers.IsListingFavorited())
		secured.POST("/favorite/listings", controllers.ToggleFavoriteListing())
	}

	// Payment routes (separate group for clarity)
	setupUserPaymentRoutes(user)
}

// setupUserPaymentRoutes configures user payment-related endpoints
func setupUserPaymentRoutes(user *gin.RouterGroup) {
	secured := user.Group("").Use(auth.Auth())

	// Payment cards
	payment := user.Group("/:userid/payment/cards").Use(auth.Auth())
	payment.POST("/", controllers.CreatePaymentCard())
	payment.GET("/", controllers.GetPaymentCards())
	payment.PUT("/:id/default", controllers.ChangeDefaultPaymentCard())
	payment.DELETE("/:id", controllers.DeletePaymentCard())

	// Seller payment information
	secured.POST("/:userid/payment-information/", controllers.CreateSellerPaymentInformation())
	secured.GET("/:userid/payment-information/onboarded", controllers.CompletedPaymentOnboarding())
	secured.GET("/:userid/payment-information", controllers.GetSellerPaymentInformations())
	secured.PUT("/:userid/payment-information/:paymentInfoId/default", controllers.ChangeDefaultSellerPaymentInformation())
	secured.DELETE("/:userid/payment-information/:paymentInfoId", controllers.DeleteSellerPaymentInformation())
}

// shopRoutesRefactored configures shop-related endpoints
func shopRoutesRefactored(api *gin.RouterGroup, serviceContainer *container.ServiceContainer) {
	shop := api.Group("/shops")
	reviewController := serviceContainer.GetReviewController()

	// Public shop endpoints - use original controllers (they handle complex aggregations)
	shop.GET("/", controllers.GetShops())
	shop.GET("/:shopid", controllers.GetShop())
	shop.GET("/search", controllers.SearchShops())
	shop.GET("/:shopid/reviews", reviewController.GetShopReviews()) // Keep refactored review controller
	shop.GET("/:shopid/followers", controllers.GetShopFollowers())
	shop.GET("/:shopid/shippings", controllers.GetShopShippingProfileInfos())
	shop.GET("/:shopid/shippings/:id", controllers.GetShopShippingProfileInfo())
	shop.GET("/:shopid/listings", controllers.GetShopListings())

	// Protected shop endpoints - use original controllers (they handle file uploads, transactions, etc.)
	secured := shop.Group("").Use(auth.Auth())
	{
		// Shop creation and basic management
		secured.POST("", controllers.CreateShop())
		secured.GET("/check/:username", controllers.CheckShopNameAvailability())
		secured.PUT("/:shopid/information", controllers.UpdateShopInformation())
		secured.PUT("/:shopid/status", controllers.UpdateMyShopStatus())
		secured.PUT("/:shopid/field", controllers.UpdateShopField())
		secured.POST("/:shopid/address", controllers.UpdateShopAddress())

		// Shop content management
		secured.PUT("/:shopid/about", controllers.UpdateShopAbout())
		secured.PUT("/:shopid/announcement", controllers.UpdateShopAnnouncement())
		secured.PUT("/:shopid/vacation", controllers.UpdateShopVacation())

		// Shop media management
		secured.PUT("/:shopid/logo", controllers.UpdateShopLogo())
		secured.PUT("/:shopid/banner", controllers.UpdateShopBanner())
		secured.PUT("/:shopid/gallery", controllers.UpdateShopGallery())
		secured.DELETE("/:shopid/gallery", controllers.DeleteFromShopGallery())

		// Shop following system
		secured.POST("/:shopid/followers", controllers.FollowShop())
		secured.DELETE("/:shopid/followers", controllers.UnfollowShop())
		secured.DELETE("/:shopid/followers/other", controllers.RemoveOtherFollower())
		secured.GET("/:shopid/followers/following", controllers.IsFollowingShop())

		// Shop business setup
		secured.POST("/:shopid/shipping", controllers.CreateShopShippingProfile())
		secured.POST("/:shopid/verification", controllers.CreateSellerVerificationProfile())
		secured.GET("/:shopid/verification", controllers.GetSellerVerificationProfile())
		secured.POST("/:shopid/compliance", controllers.CreateShopComplianceInformation())
		secured.GET("/:shopid/compliance", controllers.GetShopComplianceInformation())
	}

	// Shop policies and listings (separate groups for clarity)
	setupShopPoliciesRoutes(shop)
	setupShopListingsRoutes(shop)
}

// setupShopPoliciesRoutes configures shop policy endpoints
func setupShopPoliciesRoutes(shop *gin.RouterGroup) {
	secured := shop.Group("").Use(auth.Auth())

	secured.POST("/:shopid/policies", controllers.CreateShopReturnPolicy())
	secured.PUT("/:shopid/policies", controllers.UpdateShopReturnPolicy())
	secured.GET("/:shopid/policies", controllers.GetShopReturnPolicies())
	secured.GET("/:shopid/policies/:policyid", controllers.GetShopReturnPolicy())
	secured.DELETE("/:shopid/policies/:policyid", controllers.DeleteShopReturnPolicy())
}

// setupShopListingsRoutes configures shop listing endpoints
func setupShopListingsRoutes(shop *gin.RouterGroup) {
	secured := shop.Group("").Use(auth.Auth())

	secured.POST("/:shopid/listings", controllers.CreateListing())
	secured.GET("/:shopid/listings/summary", controllers.GetMyListingsSummary())
	secured.GET("/:shopid/check-listing-onboarding", controllers.HasUserCreatedListingOnboarding())
}

// listingRoutesRefactored configures listing-related endpoints
func listingRoutesRefactored(api *gin.RouterGroup, serviceContainer *container.ServiceContainer) {
	listing := api.Group("/listings")
	reviewController := serviceContainer.GetReviewController()

	listing.GET("/", controllers.GetListings())
	listing.GET("/:listingid", controllers.GetListing())
	listing.GET("/:listingid/reviews", reviewController.GetListingReviews())
	{
		reviews := listing.Group("/:listingid/reviews").Use(auth.Auth())
		reviews.POST("/", reviewController.CreateListingReview())
		reviews.DELETE("/", reviewController.DeleteMyListingReview())
		reviews.DELETE("/:reviewid", reviewController.DeleteOtherListingReview())
	}
	{
		secured := listing.Group("/:listingid").Use(auth.Auth())
		secured.DELETE("/", controllers.DeleteListings())
		secured.PUT("/deactivate", controllers.DeactivateListings())
	}
}

// cartRoutesRefactored configures cart-related endpoints
func cartRoutesRefactored(api *gin.RouterGroup, serviceContainer *container.ServiceContainer) {
	cart := api.Group("/:userid/carts")
	cartController := serviceContainer.GetCartController()
	{
		secured := cart.Group("").Use(auth.Auth())
		secured.GET("/", cartController.GetCartItems())
		secured.POST("/", cartController.SaveCartItem())
		secured.DELETE("/:cartId", cartController.DeleteCartItem())
		secured.DELETE("/many", cartController.DeleteCartItems())
		secured.DELETE("/clear", cartController.ClearCartItems())

		// Cart item quantity management
		secured.PUT("/:cartId/quantity/inc", cartController.IncreaseCartItemQuantity())
		secured.PUT("/:cartId/quantity/dec", cartController.DecreaseCartItemQuantity())

		// Cart validation
		secured.GET("/validate", cartController.ValidateCartItems())
	}
}
