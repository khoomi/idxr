package routers

import (
	auth "khoomi-api-io/api/internal/auth"
	"khoomi-api-io/api/internal/container"
	"khoomi-api-io/api/internal/middleware"
	"khoomi-api-io/api/pkg/controllers"
	"khoomi-api-io/api/pkg/services"

	"github.com/gin-gonic/gin"
)

// InitRefactoredRoute creates a new Gin router with service layer architecture
func InitRoute() *gin.Engine {
	serviceContainer := container.NewServiceContainer()
	router := gin.Default()
	router.Use(middleware.CorsMiddleware())

	api := router.Group("/v1", middleware.KhoomiRateLimiter())
	{
		setupAuthRoutes(api, serviceContainer)
		userRoutes(api, serviceContainer)
		shopRoutes(api, serviceContainer)
		listingRoutes(api, serviceContainer)
		cartRoutes(api, serviceContainer)
		// categoryRoutes(api) // Keep existing
	}

	return router
}

// setupAuthRoutes configures public authentication endpoints
func setupAuthRoutes(api *gin.RouterGroup, serviceContainer *container.ServiceContainer) {
	userController := serviceContainer.GetUserController()

	api.POST("/signup", userController.CreateUser)
	api.POST("/auth", userController.HandleUserAuthentication)
	api.POST("/auth/google", userController.HandleUserGoogleAuthentication)
	api.PUT("/auth/refresh-token", userController.RefreshToken)
	api.DELETE("/logout", userController.Logout)
	api.GET("/verify-email", userController.VerifyEmail)
	api.POST("/send-password-reset", userController.PasswordResetEmail)
	api.POST("/password-reset", userController.PasswordReset)
}

// userRoutesRefactored configures user-related endpoints
func userRoutes(api *gin.RouterGroup, serviceContainer *container.ServiceContainer) {
	userController := serviceContainer.GetUserController()
	addressService := serviceContainer.GetUserAddressController()
	favoriteService := serviceContainer.GetUserFavoriteController()
	api.GET("/ping", controllers.Ping)

	user := api.Group("/users")

	user.GET("/:userid", userController.GetUser)

	{
		secured := user.Group("").Use(auth.Auth())
		// Current user session
		secured.GET("/me", userController.ActiveSessionUser)
		// Session management
		secured.GET("/:userid/session", userController.GetMyActiveSession)
		secured.PUT("/:userid/change-password", userController.ChangePassword)

		// Account deletion
		secured.GET("/:userid/deletion", userController.IsAccountPendingDeletion)
		secured.POST("/:userid/deletion", userController.SendDeleteUserAccount)
		secured.DELETE("/:userid/deletion", userController.CancelDeleteUserAccount)

		// Profile management
		secured.PUT("/:userid/", userController.UpdateMyProfile)
		secured.PUT("/:userid/single", userController.UpdateUserSingleField)
		secured.PUT("/:userid/thumbnail", userController.UploadThumbnail)
		secured.DELETE("/:userid/thumbnail/:url", userController.DeleteThumbnail)
		secured.PUT("/:userid/birthdate", userController.UpdateUserBirthdate)
		secured.POST("/:userid/send-verify-email", userController.SendVerifyEmail)

		// Notification settings
		secured.POST("/:userid/notification-settings", controllers.CreateUserNotificationSettings())
		secured.GET("/:userid/notification-settings", controllers.GetUserNotificationSettings())
		secured.PUT("/:userid/notification-settings", controllers.UpdateUserNotificationSettings())

		// Address management
		secured.POST("/:userid/addresses", addressService.CreateUserAddress())
		secured.PUT("/:userid/addresses/:id", addressService.UpdateUserAddress())
		secured.GET("/:userid/addresses", addressService.GetUserAddresses())
		secured.DELETE("/:userid/addresses/:id", addressService.DeleteUserAddress())
		secured.PUT("/:userid/addresses/:id/default", addressService.ChangeDefaultAddress())

		// Security & login history
		secured.GET("/:userid/login-history", userController.GetLoginHistories)
		secured.DELETE("/:userid/login-history", userController.DeleteLoginHistories)
		secured.PUT("/:userid/login-notification", userController.UpdateSecurityNotificationSetting)
		secured.GET("/:userid/login-notification", userController.GetSecurityNotificationSetting)

		// Wishlist management
		secured.GET("/:userid/wishlist", userController.GetUserWishlist)
		secured.POST("/:userid/wishlist", userController.AddWishListItem)
		secured.DELETE("/:userid/wishlist", userController.RemoveWishListItem)

		// Favorites
		secured.GET("/favorite/shops", favoriteService.IsShopFavorited())
		secured.POST("/favorite/shops", favoriteService.ToggleFavoriteShop())
		secured.GET("/favorite/listings", favoriteService.IsListingFavorited())
		secured.POST("/favorite/listings", favoriteService.ToggleFavoriteListing())
	}

	// Payment routes (separate group for clarity)
	setupUserPaymentRoutes(user, serviceContainer)
}

// setupUserPaymentRoutes configures user payment-related endpoints
func setupUserPaymentRoutes(user *gin.RouterGroup, serviceContainer *container.ServiceContainer) {
	paymentController := serviceContainer.GetPaymentController()
	secured := user.Group("").Use(auth.Auth())

	// Payment cards
	payment := user.Group("/:userid/payment/cards").Use(auth.Auth())
	payment.POST("/", paymentController.CreatePaymentCard())
	payment.GET("/", paymentController.GetPaymentCards())
	payment.PUT("/:id/default", paymentController.ChangeDefaultPaymentCard())
	payment.DELETE("/:id", paymentController.DeletePaymentCard())

	// Seller payment information
	secured.POST("/:userid/payment-information/", paymentController.CreateSellerPaymentInformation())
	secured.GET("/:userid/payment-information/onboarded", paymentController.CompletedPaymentOnboarding())
	secured.GET("/:userid/payment-information", paymentController.GetSellerPaymentInformations())
	secured.PUT("/:userid/payment-information/:paymentInfoId/default", paymentController.ChangeDefaultSellerPaymentInformation())
	secured.DELETE("/:userid/payment-information/:paymentInfoId", paymentController.DeleteSellerPaymentInformation())
}

// shopRoutesRefactored configures shop-related endpoints
func shopRoutes(api *gin.RouterGroup, serviceContainer *container.ServiceContainer) {
	reviewController := serviceContainer.GetReviewController()
	emailService := serviceContainer.GetEmailService()
	shopController := serviceContainer.GetShopController()
	verificationController := serviceContainer.GetVerificationController()
	shippingController := serviceContainer.GetShippingController()

	api.GET("/:userid/shops", shopController.GetShopByOwnerUserId())

	shop := api.Group("/shops")
	// Public shop endpoints - use original controllers (they handle complex aggregations)
	shop.GET("/", shopController.GetShops())
	shop.GET("/:shopid", shopController.GetShop())
	shop.GET("/search", shopController.SearchShops())
	shop.GET("/:shopid/reviews", reviewController.GetShopReviews()) // Keep refactored review controller
	shop.GET("/:shopid/followers", shopController.GetShopFollowers())
	shop.GET("/:shopid/shippings", shippingController.GetShopShippingProfileInfos())
	shop.GET("/:shopid/shippings/:shippingid", shippingController.GetShopShippingProfileInfo())
	shop.GET("/:shopid/listings", controllers.GetShopListings())

	// Protected shop endpoints - use original controllers (they handle file uploads, transactions, etc.)
	secured := shop.Group("").Use(auth.Auth())
	{
		// Shop creation and basic management
		secured.POST("", shopController.CreateShop(emailService))
		secured.GET("/check/:username", shopController.CheckShopNameAvailability())
		secured.PUT("/:shopid/information", shopController.UpdateShopInformation())
		secured.PUT("/:shopid/status", shopController.UpdateMyShopStatus())
		secured.PUT("/:shopid/field", shopController.UpdateShopField())
		secured.POST("/:shopid/address", shopController.UpdateShopAddress())

		// Shop content management
		secured.PUT("/:shopid/about", shopController.UpdateShopAbout())
		secured.PUT("/:shopid/announcement", shopController.UpdateShopAnnouncement())
		secured.PUT("/:shopid/vacation", shopController.UpdateShopVacation())

		// Shop media management
		secured.PUT("/:shopid/logo", shopController.UpdateShopLogo())
		secured.PUT("/:shopid/banner", shopController.UpdateShopBanner())
		secured.PUT("/:shopid/gallery", shopController.UpdateShopGallery())
		secured.DELETE("/:shopid/gallery", shopController.DeleteFromShopGallery())

		// Shop following system
		secured.POST("/:shopid/followers", shopController.FollowShop())
		secured.DELETE("/:shopid/followers", shopController.UnfollowShop())
		secured.DELETE("/:shopid/followers/other", shopController.RemoveOtherFollower())
		secured.GET("/:shopid/followers/following", shopController.IsFollowingShop())

		// Shop business setup
		secured.POST("/:shopid/shippings", shippingController.CreateShopShippingProfile())
		secured.DELETE("/:shopid/shippings/:shippingid", shippingController.DeleteShippingProfile())
		secured.PUT("/:shopid/shippings/:shippingid", shippingController.UpdateShippingProfile())
		secured.PUT("/:shopid/shippings/:shippingid/default", shippingController.ChangeDefaultShippingProfile())
		secured.POST("/:shopid/verification", verificationController.CreateSellerVerificationProfile())
		secured.GET("/:shopid/verification", verificationController.GetSellerVerificationProfile())
		secured.POST("/:shopid/compliance", shopController.CreateShopComplianceInformation())
		secured.GET("/:shopid/compliance", shopController.GetShopComplianceInformation())

		// Shop notification management
		secured.GET("/:shopid/notifications/settings", shopController.GetShopNotificationSettings())
		secured.PUT("/:shopid/notifications/settings", shopController.UpdateShopNotificationSettings())
		secured.GET("/:shopid/notifications", shopController.GetShopNotifications())
		secured.POST("/:shopid/notifications", shopController.CreateShopNotification())
		secured.PUT("/:shopid/notifications/:notificationid/read", shopController.MarkShopNotificationAsRead())
		secured.PUT("/:shopid/notifications/read-all", shopController.MarkAllShopNotificationsAsRead())
		secured.DELETE("/:shopid/notifications/:notificationid", shopController.DeleteShopNotification())
	}

	// Shop policies and listings (separate groups for clarity)
	setupShopPoliciesRoutes(shop, serviceContainer)
	setupShopListingsRoutes(shop, emailService)
}

// setupShopPoliciesRoutes configures shop policy endpoints
func setupShopPoliciesRoutes(shop *gin.RouterGroup, serviceContainer *container.ServiceContainer) {
	shopController := serviceContainer.GetShopController()
	secured := shop.Group("").Use(auth.Auth())

	secured.POST("/:shopid/policies", shopController.CreateShopReturnPolicy())
	secured.PUT("/:shopid/policies", shopController.UpdateShopReturnPolicy())
	secured.GET("/:shopid/policies", shopController.GetShopReturnPolicies())
	secured.GET("/:shopid/policies/:policyid", shopController.GetShopReturnPolicy())
	secured.DELETE("/:shopid/policies/:policyid", shopController.DeleteShopReturnPolicy())
}

// setupShopListingsRoutes configures shop listing endpoints
func setupShopListingsRoutes(shop *gin.RouterGroup, emailService services.EmailService) {
	secured := shop.Group("").Use(auth.Auth())

	secured.POST("/:shopid/listings", controllers.CreateListingWithEmailService(emailService))
	secured.GET("/:shopid/listings/summary", controllers.GetMyListingsSummary())
	secured.GET("/:shopid/check-listing-onboarding", controllers.HasUserCreatedListingOnboarding())
}

// listingRoutesRefactored configures listing-related endpoints
func listingRoutes(api *gin.RouterGroup, serviceContainer *container.ServiceContainer) {
	listing := api.Group("/listings")
	reviewController := serviceContainer.GetReviewController()
	// Note: ListingController methods need to be refactored from functions to struct methods

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
func cartRoutes(api *gin.RouterGroup, serviceContainer *container.ServiceContainer) {
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
