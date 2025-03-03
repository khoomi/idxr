package middleware

import (
	"time"

	ratelimit "github.com/JGLTechnologies/gin-rate-limit"
	"github.com/gin-gonic/gin"
	"khoomi-api-io/api/pkg/util"
)

func keyFunc(c *gin.Context) string {
	return c.ClientIP()
}

func errorHandler(c *gin.Context, info ratelimit.Info) {
	c.String(429, "Too many requests. Try again in "+time.Until(info.ResetTime).String())
}

func KhoomiRateLimiter() gin.HandlerFunc {
	// This makes it so each ip can only make 5 requests per second
	store := ratelimit.RedisStore(&ratelimit.RedisOptions{
		RedisClient: util.REDIS,
		Rate:        time.Second,
		Limit:       5,
	})

	return ratelimit.RateLimiter(store, &ratelimit.Options{
		ErrorHandler: errorHandler,
		KeyFunc:      keyFunc,
	})
}
