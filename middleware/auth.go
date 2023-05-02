package middleware

import (
	"github.com/gin-gonic/gin"
	"khoomi-api-io/khoomi_api/auth"
)

func Auth() gin.HandlerFunc {
	return func(context *gin.Context) {
		tokenString := auth.ExtractToken(context)

		if tokenString == "" {
			context.JSON(401, gin.H{"error": "request does not contain an access token"})
			context.Abort()
			return
		}
		err := auth.ValidateToken(tokenString)
		if err != nil {
			context.JSON(401, gin.H{"error": err.Error()})
			context.Abort()
			return
		}
		context.Next()
	}
}
