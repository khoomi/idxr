package user

import (
	"github.com/gin-gonic/gin"
	"khoomi-api-io/khoomi_api/controllers"
	"khoomi-api-io/khoomi_api/middleware"
)

func Routes(api *gin.RouterGroup) {
	user := api.Group("/users")
	{
		user.POST("/", controllers.CreateUser())
		user.GET("/:userId", controllers.GetUser())
		secured := api.Group("/user").Use(middleware.Auth())
		{
			secured.GET("/ping", controllers.Ping)
			secured.GET("/me", controllers.CurrentUser)
			secured.PUT("/me", controllers.UpdateFirstLastName())
			secured.PUT("/thumbnail", controllers.UploadThumbnail())
			secured.DELETE("/thumbnail", controllers.DeleteThumbnail())
			secured.POST("/send-verify-email", controllers.SendVerifyEmail())
			// Login histories
			secured.GET("/:userId/login-history", controllers.GetLoginHistories())
			secured.DELETE("/:userId/login-history", controllers.DeleteLoginHistories())
		}

	}
}
