package user

import (
	"github.com/gin-gonic/gin"
	"khoomi-api-io/khoomi_api2/controllers/user_controllers"
)

func UserRoutes(api *gin.RouterGroup) {
	user := api.Group("/user_models")
	{
		user.POST("/user", user_controllers.CreateUser())
		user.GET("/user/:userId", user_controllers.GetUser())

	}
}
