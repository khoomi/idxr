package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"

	"khoomi-api-io/api/pkg/util"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		key, err := ExtractSessionKey(c)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err)
			c.Abort()
			return
		}

		session, err := GetSession(c, key)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err)
			c.Abort()
			return
		}

		if session.Expired() {
			err = util.REDIS.Del(c, key).Err()
			if err != nil {
				fmt.Println(err)
			}
			util.HandleError(c, 401, errors.New("session expired"))
			c.Abort()
			return
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
	key, err := ExtractSessionKey(c)
	if err != nil {
		return primitive.NilObjectID, err
	}
	session, err := GetSession(c, key)
	if err != nil {
		return primitive.NilObjectID, err
	}

	userId := c.Param("userid")
	res, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return primitive.NilObjectID, err
	}

	if userId != session.UserId.Hex() {
		errMsg := fmt.Sprintln("unauthorized")
		return primitive.NilObjectID, errors.New(errMsg)
	}

	return res, nil
}

func GetSessionUserID(c *gin.Context) (primitive.ObjectID, error) {
	key, err := ExtractSessionKey(c)
	if err != nil {
		return primitive.NilObjectID, err
	}

	session, err := GetSession(c, key)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return session.UserId, nil
}
