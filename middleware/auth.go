package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"errors"

	"khoomi-api-io/khoomi_api/config"
	"khoomi-api-io/khoomi_api/helper"

	"github.com/gin-gonic/gin"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := config.ExtractToken(c)
		if tokenString == "" {
			helper.HandleError(c, 401, errors.New("request does not contain an access token"), "request does not contain an access token")
			c.Abort()
			return
		}
		_, err := config.ValidateToken(tokenString)
		if err != nil {
			helper.HandleError(c, 401, err, err.Error())
			c.Abort()
			return
		}

		res := config.IsTokenValid(config.REDIS, tokenString)
		if !res {
			helper.HandleError(c, 401, errors.New("why are you trying to act with a blacklisted token? huh? please login again"), "why are you trying to act with a blacklisted token? huh? please login again")
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
