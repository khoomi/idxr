package auth

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"khoomi-api-io/api/pkg/util"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var SESSION_NAME = "____kh"

type UserSession struct {
	ExpiresAt time.Time          `json:"expiresAt"`
	UserId    primitive.ObjectID `json:"userId"`
	Email     string             `json:"email"`
	LoginName string             `json:"loginName"`
}

func (s UserSession) MarshalBinary() ([]byte, error) {
	return json.Marshal(s)
}

func (s *UserSession) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, s)
}

// Checks if user session is expired.
func (s UserSession) Expired() bool {
	expired := s.ExpiresAt.Before(time.Now())
	return expired
}

// Set new user login session
func SetSession(ctx *gin.Context, userId primitive.ObjectID, email, loginName string) (string, error) {
	log.Println("setting cookie")
	key := GenerateSecureToken(20)
	ttl := time.Hour * (24 * 7)
	sessExpTime := time.Now().Add(ttl)
	value := UserSession{
		UserId:    userId,
		Email:     email,
		LoginName: loginName,
		ExpiresAt: sessExpTime,
	}

	domain := getDomainFromRequest(ctx)
	secure := isHTTPS(ctx)

	log.Println(domain, secure)

	ctx.SetCookie(SESSION_NAME, key, int(ttl.Seconds()), "/", domain, secure, true)
	return key, util.REDIS.Set(ctx, key, value, ttl).Err()
}

func getDomainFromRequest(ctx *gin.Context) string {
	host := ctx.Request.Host

	// Remove port
	if colonIndex := strings.LastIndex(host, ":"); colonIndex != -1 {
		host = host[:colonIndex]
	}

	if host == "localhost" || host == "127.0.0.1" {
		return "localhost"
	}

	// For production domains (khoomi.com, api.khoomi.com, etc.)
	parts := strings.Split(host, ".")
	if len(parts) >= 2 {
		return "." + strings.Join(parts[len(parts)-2:], ".")
	}

	return host
}

func isHTTPS(ctx *gin.Context) bool {
	if ctx.Request.TLS != nil {
		return true
	}

	if ctx.GetHeader("X-Forwarded-Proto") == "https" {
		return true
	}

	if ctx.GetHeader("X-Forwarded-Ssl") == "on" {
		return true
	}

	return false
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
	key := ctx.Request.Header.Get("Authorization")
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
