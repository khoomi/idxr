package internal

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"context"
	"fmt"
	"log"
	"time"

	"khoomi-api-io/api/pkg/util"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := ExtractToken(c)
		if tokenString == "" {
			util.HandleError(c, 401, errors.New("request does not contain an access token"), "request does not contain an access token")
			c.Abort()
			return
		}
		_, err := ValidateToken(tokenString)
		if err != nil {
			util.HandleError(c, 401, err, err.Error())
			c.Abort()
			return
		}

		res := IsTokenValid(util.REDIS, tokenString)
		if !res {
			util.HandleError(c, 401, errors.New("why are you trying to act with a blacklisted token? huh? please login again"), "why are you trying to act with a blacklisted token? huh? please login again")
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

// Validate param userid again session userid.
func ValidateUserID(c *gin.Context) (primitive.ObjectID, error) {
	auth, err := InitJwtClaim(c)
	if err != nil {
		errMsg := fmt.Sprintf("unauthorized: User ID not found in authentication token - %v", err.Error())
		return primitive.NilObjectID, errors.New(errMsg)
	}

	userId := c.Param("userid")
	res, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return primitive.NilObjectID, err
	}

	if userId != auth.Id {
		errMsg := fmt.Sprintln("unauthorized: User ID in the URL path doesn't match with currently authenticated user")
		return primitive.NilObjectID, errors.New(errMsg)
	}

	return res, nil
}

func InvalidateToken(db *redis.Client, tokenString string) error {
	// Add the token to the blacklist with an expiration time of 24 hours
	_, err := db.Set(context.Background(), tokenString, true, 24*time.Hour).Result()
	if err != nil {
		return err
	}

	return nil
}

// Check if token is in the blacklisst
func IsTokenValid(db *redis.Client, tokenString string) bool {
	_, err := db.Get(context.Background(), tokenString).Result()
	if err == redis.Nil {
		return true
	}
	if err != nil {
		log.Printf("Error while checking blacklist: %s", err)
		return false
	}

	// Token is in the blacklist, so it's invalid
	return false
}
