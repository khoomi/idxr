package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"khoomi-api-io/khoomi_api/auth"
	"khoomi-api-io/khoomi_api/configs"
	"khoomi-api-io/khoomi_api/helper"
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

		res := helper.IsTokenValid(context, configs.REDIS, tokenString)
		if res == false {
			context.JSON(401, gin.H{"error": "why are you trying to act with a blacklisted token? huh? please login again"})
			context.Abort()
		}

		context.Next()
	}
}

func GenerateSecureToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
