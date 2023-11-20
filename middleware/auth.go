package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"khoomi-api-io/khoomi_api/configs"

	"github.com/gin-gonic/gin"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := configs.ExtractToken(c)

		if tokenString == "" {
			c.JSON(401, gin.H{"error": "request does not contain an access token"})
			c.Abort()
			return
		}
		err := configs.ValidateToken(tokenString)
		if err != nil {
			c.JSON(401, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		res := configs.IsTokenValid(configs.REDIS, tokenString)
		if !res {
			c.JSON(401, gin.H{"error": "why are you trying to act with a blacklisted token? huh? please login again"})
			c.Abort()
		}

		c.Next()
	}
}

func GenerateSecureToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}

	return hex.EncodeToString(b)
}
