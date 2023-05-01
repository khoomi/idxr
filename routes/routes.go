package routes

import (
	"github.com/gin-gonic/gin"
	"khoomi-api-io/khoomi_api/controllers/user_controllers"
	"khoomi-api-io/khoomi_api/routes/user"
)

func InitRoute() *gin.Engine {
	router := gin.Default()
	api := router.Group("/api")
	{
		api.POST("/auth", user_controllers.AuthenticateUser())
		user.UserRoutes(api)
	}

	return router
}
