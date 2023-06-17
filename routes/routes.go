package routes

import (
	"khoomi-api-io/khoomi_api/configs"
	"khoomi-api-io/khoomi_api/controllers"
	"khoomi-api-io/khoomi_api/middleware"

	"github.com/gin-gonic/gin"
)

func InitRoute() *gin.Engine {
	router := gin.Default()
	router.Use(middleware.CorsMiddleware())

	api := router.Group("/api", configs.KhoomiRateLimiter())
	{
		api.POST("/signup", controllers.CreateUser())
		api.POST("/auth", controllers.HandleUserAuthentication())
		api.DELETE("/logout", controllers.Logout())
		api.GET("/verify-email", controllers.VerifyEmail())
		api.POST("/send-password-reset", controllers.PasswordResetEmail())
		api.POST("/password-reset", controllers.PasswordReset())

		userRoutes(api)
		ShopRoutes(api)
		CategoryRoutes(api)
		ShippingRoutes(api)
	}

	return router
}

func userRoutes(api *gin.RouterGroup) {
	user := api.Group("/users")
	{
		user.GET("/", controllers.GetUserByIDOrEmail())
		user.GET("/:userId/shops", controllers.GetShopByOwnerUserId())
		secured := api.Group("/users").Use(middleware.Auth())
		{
			secured.GET("/ping", controllers.Ping)
			// change my password
			secured.PUT("/password-reset", controllers.ChangePassword())
			// user endpoint.
			secured.GET("/me", controllers.CurrentUser)
			secured.PUT("/:userid", controllers.UpdateFirstLastName())
			// notification setting endpoint
			secured.POST("/:userid/notication-setting", controllers.CreateUserNotificationSettings())
			secured.GET("/:userid/notication-setting", controllers.GetUserNotificationSettings())
			secured.PUT("/:userid/notication-setting", controllers.UpdateUserNotificationSettings())
			// user thumbnail endpoints.
			secured.PUT("/:userid/thumbnail", controllers.UploadThumbnail())
			secured.DELETE("/:userid/thumbnail", controllers.DeleteThumbnail())
			// user address endpoints.
			secured.POST("/:userid/addresses", controllers.CreateUserAddress())
			secured.PUT("/:userid/addresses", controllers.UpdateUserAddress())
			secured.GET("/:userid/addresses", controllers.GetUserAddresses())
			//secured.GET("/:userid/addresses", controllers.GetUserAddress())
			// email notification.
			secured.POST("/:userid/send-verify-email", controllers.SendVerifyEmail())
			// User birthdate
			secured.PUT("/:userid/birthdate", controllers.UpdateUserBirthdate())
			// Login histories.
			secured.GET("/:userid/login-history", controllers.GetLoginHistories())
			secured.DELETE("/:userid/login-history", controllers.DeleteLoginHistories())
			// Profile update
			secured.PUT("/:userid/update", controllers.UpdateUserSingleField())
			// favorites shops
			secured.POST("/:userid/shop", controllers.AddRemoveFavoriteShop())
			// wish list
			secured.GET("/:userid/wishlist", controllers.GetUserWishlist())
			secured.POST("/:userid/wishlist", controllers.AddWishListItem())
			secured.DELETE("/:userid/wishlist", controllers.RemoveWishListItem())
			/// payment informations
			secured.POST("/:userid/payment-information", controllers.CreatePaymentInformation())
			secured.GET("/:userid/payment-information", controllers.GetPaymentInformations())
			secured.DELETE("/:userid/payment-information/:paymentInfoId", controllers.DeletePaymentInformation())
		}

	}
}

func ShopRoutes(api *gin.RouterGroup) {
	shop := api.Group("/shops")
	shop.GET("/", controllers.GetShops())
	shop.GET("/:shopid", controllers.GetShop())
	shop.GET("/:shopid/about", controllers.GetShopAbout())
	shop.GET("/:shopid/reviews", controllers.GetShopReviews())
	shop.GET("/:shopid/members", controllers.GetShopMembers())
	shop.GET("/search", controllers.SearchShops())

	secured := api.Group("/shops").Use(middleware.Auth())
	{
		// create shop
		secured.POST("/", controllers.CreateShop())
		// check for shop username availability.
		secured.POST("/check/:username", controllers.CheckShopNameAvailability())
		// shop images
		secured.PUT("/:shopid/logo", controllers.UpdateShopLogo())
		secured.PUT("/:shopid/banner", controllers.UpdateShopBanner())
		// shop  about
		secured.POST("/:shopid/about", controllers.CreateShopAbout())
		secured.PUT("/:shopid/about", controllers.UpdateShopAbout())
		secured.PUT("/:shopid/about/status", controllers.UpdateShopAboutStatus())
		// shop vacation
		secured.PUT("/:shopid/vacation", controllers.UpdateShopVacation())
		// shop gallery
		secured.PUT("/:shopid/gallery", controllers.UpdateShopGallery())
		secured.DELETE("/:shopid/gallery", controllers.DeleteFromShopGallery())
		// shop announcement
		secured.PUT("/:shopid/announcement", controllers.UpdateShopAnnouncement())
		// shop favorers
		secured.PUT("/:shopid/favorers", controllers.AddShopFavorer())
		secured.DELETE("/:shopid/favorers", controllers.RemoveShopFavorer())
		// shop members
		shop.POST("/:shopid/members", controllers.JoinShopMembers())
		secured.DELETE("/:shopid/members", controllers.LeaveShopMembers())
		secured.DELETE("/:shopid/members/other", controllers.RemoveOtherMember())
		// shop review
		shop.POST("/:shopid/reviews", controllers.CreateShopReview())
		secured.DELETE("/:shopid/reviews", controllers.DeleteMyReview())
		secured.DELETE("/:shopid/reviews/other", controllers.DeleteOtherReview())
		// policies
		secured.POST("/:shopid/policies", controllers.CreateShopReturnPolicy())
		secured.PUT("/:shopid/policies", controllers.UpdateShopReturnPolicy())
		secured.GET("/:shopid/policies", controllers.GetShopReturnPolicy())
		secured.GET("/:shopid/policies/all", controllers.GetShopReturnPolicies())
		secured.DELETE("/:shopid/policies/", controllers.DeleteShopReturnPolicy())

	}
}

func CategoryRoutes(api *gin.RouterGroup) {
	category := api.Group("/categories")
	category.GET("/", controllers.GetAllCategories())
	category.GET("/search", controllers.SearchCategories())
	category.GET("/:id/children", controllers.GetCategoryChildren())
	category.GET("/:id/ancestor", controllers.GetCategoryAncestor())
	secured := api.Group("/categories").Use(middleware.Auth())
	{
		secured.POST("/", controllers.CreateCategorySingle())
		secured.POST("/multi", controllers.CreateCategoryMulti())
		secured.DELETE("/", controllers.DeleteAllCategories())
	}
}

func ShippingRoutes(api *gin.RouterGroup) {
	shipping := api.Group("/shipping/info")
	shipping.GET("/:infoId", controllers.GetShopShippingProfileInfo())
	secured := api.Group("/shipping/info").Use(middleware.Auth())
	{
		secured.POST("/", controllers.CreateShopShippingProfile())
		secured.PUT("/", controllers.UpdateShopShippingProfileInfo())
	}
}
