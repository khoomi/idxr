package auth

import (
	"encoding/json"
	"fmt"
	"khoomi-api-io/api/pkg/util"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var SESSION_NAME = "____kh"

type UserSession struct {
	UserId    string    `json:"userId"`
	Email     string    `json:"email"`
	LoginName string    `json:"loginName"`
	ExpiresAt time.Time `json:"expiresAt"`
}

func (s UserSession) MarshalBinary() ([]byte, error) {
	return json.Marshal(s)
}

func (s *UserSession) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, s)
}

// Get user objectId
func (s UserSession) GetUserObjectId() (primitive.ObjectID, error) {
	id, err := primitive.ObjectIDFromHex(s.UserId)
	return id, err
}

// Checks if user session is expired.
func (s UserSession) Expired() bool {
	expired := s.ExpiresAt.Before(time.Now())
	return expired
}

// Set new user login session
func SetSession(ctx *gin.Context, userId, email, loginName string) (string, error) {
	key := GenerateSecureToken(20)
	ttl := time.Hour * 24
	sessExpTime := time.Now().Add(ttl)
	value := UserSession{
		UserId:    userId,
		Email:     email,
		LoginName: loginName,
		ExpiresAt: sessExpTime,
	}
	ctx.SetCookie(SESSION_NAME, key, int(ttl.Seconds()), "/", "localhost", false, true)

	return key, util.REDIS.Set(ctx, key, value, ttl).Err()
}

// Get new user login session
func GetSession(ctx *gin.Context, key string) (UserSession, error) {
	value, err := util.REDIS.Get(ctx, key).Result()
	if err != nil {
		return UserSession{}, err
	}

	var session UserSession
	err = json.Unmarshal([]byte(value), &session)
	if err != nil {
		return UserSession{}, err
	}

	return session, nil
}

// Get new user session
func GetSessionAuto(ctx *gin.Context) (UserSession, error) {
	key, err := ExtractSessionKey(ctx)
	if err != nil {
		return UserSession{}, err
	}
	return GetSession(ctx, key)
}

// Delete user session
func DeleteSession(ctx *gin.Context) {
	key, err := ctx.Cookie(SESSION_NAME)
	if err != nil {
		log.Println(err)
	}

	err = util.REDIS.Del(ctx, key).Err()
	if err != nil {
		log.Println(err)
	}

	ctx.SetCookie(SESSION_NAME, "", 0, "/", "localhost", false, true)
}

// Extract session token from request header.
func ExtractSessionKey(ctx *gin.Context) (string, error) {
  key := ctx.Request.Header.Get("Authorization");
  value, err := ExtractBearerToken(key)
	if err != nil {
		return "", err
	}

	return value, nil
}


// ExtractBearerToken extracts the Bearer token from the Authorization header
func ExtractBearerToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", fmt.Errorf("authorization header is empty")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", fmt.Errorf("authorization header does not start with 'Bearer '")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return "", fmt.Errorf("token is empty")
	}

	return token, nil
}
