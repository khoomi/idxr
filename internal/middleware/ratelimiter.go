package middleware

import (
	"time"

	"khoomi-api-io/api/pkg/util"

	ratelimit "github.com/JGLTechnologies/gin-rate-limit"
	"github.com/gin-gonic/gin"
)

func KhoomiRateLimiter() gin.HandlerFunc {
	store := ratelimit.RedisStore(&ratelimit.RedisOptions{
		RedisClient: util.REDIS,
		Rate:        time.Second,
		Limit:       5,
	})

	return ratelimit.RateLimiter(store, &ratelimit.Options{
		ErrorHandler: func(c *gin.Context, info ratelimit.Info) {
			c.String(429, "Too many requests. Try again in "+time.Until(info.ResetTime).String())
		},
	})
}
