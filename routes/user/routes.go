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
		secured := api.Group("/secured").Use(middleware.Auth())
		{
			secured.GET("/ping", controllers.Ping)
			// Login histories
			user.GET("/:userId/login-history", controllers.GetLoginHistories())
			user.DELETE("/:userId/login-history", controllers.DeleteLoginHistories())
		}

	}
}
