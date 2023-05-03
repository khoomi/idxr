package user

import (
	"github.com/gin-gonic/gin"
	"khoomi-api-io/khoomi_api/controllers"
	"khoomi-api-io/khoomi_api/middleware"
)

func Routes(api *gin.RouterGroup) {
	user := api.Group("/users")
	{
		user.GET("/:userId", controllers.GetUser())
		secured := api.Group("/user").Use(middleware.Auth())
		{
			secured.GET("/ping", controllers.Ping)
			// user endpoint.
			secured.GET("/", controllers.CurrentUser)
			secured.PUT("/", controllers.UpdateFirstLastName())
			// user thumbnail endpoints.
			secured.PUT("/thumbnail", controllers.UploadThumbnail())
			secured.DELETE("/thumbnail", controllers.DeleteThumbnail())
			// user address endpoints.
			secured.POST("/address", controllers.CreateUserAddress())
			secured.PUT("/address", controllers.UpdateUserAddress())
			// email notification.
			secured.POST("/send-verify-email", controllers.SendVerifyEmail())
			// Login histories.
			secured.GET("/:userId/login-history", controllers.GetLoginHistories())
			secured.DELETE("/:userId/login-history", controllers.DeleteLoginHistories())
		}

	}
}
