package routes

import (
	"github.com/gin-gonic/gin"
	"khoomi-api-io/khoomi_api/configs"
	"khoomi-api-io/khoomi_api/controllers"
	"khoomi-api-io/khoomi_api/routes/user"
)

func InitRoute() *gin.Engine {
	router := gin.Default()
	api := router.Group("/api", configs.KhoomiRateLimiter())
	{
		api.POST("/auth", controllers.AuthenticateUser())
		api.GET("/send-password-reset", controllers.PasswordResetEmail())
		api.GET("/password-reset/:userid", controllers.PasswordReset())

		user.Routes(api)
	}

	return router
}
