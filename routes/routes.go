package routes

import (
	"github.com/gin-gonic/gin"
	"khoomi-api-io/khoomi_api/configs"
	"khoomi-api-io/khoomi_api/controllers"
	"khoomi-api-io/khoomi_api/middleware"
	"khoomi-api-io/khoomi_api/routes/user"
)

func InitRoute() *gin.Engine {
	router := gin.Default()
	api := router.Group("/api", configs.KhoomiRateLimiter())
	{
		api.POST("/auth", controllers.AuthenticateUser())
		api.DELETE("/logout", controllers.Logout()).Use(middleware.Auth())
		api.POST("/send-password-reset", controllers.PasswordResetEmail())
		api.POST("/password-reset", controllers.PasswordReset())
		api.GET("/verify-email", controllers.VerifyEmail())

		user.Routes(api)
	}

	return router
}
