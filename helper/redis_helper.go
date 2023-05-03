package helper

import (
	"context"
	"github.com/redis/go-redis/v9"
	"log"
	"time"
)

func InvalidateToken(c context.Context, db *redis.Client, tokenString string) error {
	// Add the token to the blacklist with an expiration time of 24 hours
	_, err := db.Set(c, tokenString, true, 24*time.Hour).Result()
	if err != nil {
		return err
	}

	return nil
}

func IsTokenValid(c context.Context, db *redis.Client, tokenString string) bool {
	// Check if the token is in the blacklist
	_, err := db.Get(c, tokenString).Result()
	if err == redis.Nil {
		// Token is not in the blacklist, so it's valid
		return true
	}
	if err != nil {
		// Error while checking the blacklist
		log.Printf("Error while checking blacklist: %s", err)
		return false
	}

	// Token is in the blacklist, so it's invalid
	return false
}
