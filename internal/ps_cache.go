package internal

import (
	"context"
	"fmt"
	"khoomi-api-io/api/pkg/util"
	"log"
)

var CHANNEL_GLOBAL_CACHE = "GLOBAL_CACHE"

type CacheMessageType string

const (
	CacheRevalidateUser CacheMessageType = "revalidateUser"

	CacheRevalidateListing               CacheMessageType = "revalidateListing"
	CacheRevalidateSingleListing         CacheMessageType = "revalidateSingleListing"
	CacheRevalidateSingleListingReviews  CacheMessageType = "revalidateSingleListing"
	CacheRevalidateFavoriteListingToggle CacheMessageType = "revalidateToggleFavoriteListing"

	CacheRevalidateShop       CacheMessageType = "revalidateShop"
	CacheRevalidateSingleShop CacheMessageType = "revalidateSingleShop"
	CacheRevalidateShopReview CacheMessageType = "revalidateShopReviews"
)

type CacheMessage struct {
	Message CacheMessageType
	Payload string
}

// Helper functions to publish a message to pub sub.
func PublishCacheMessage(ctx context.Context, message CacheMessageType, payload string) error {
	cacheMessage := CacheMessage{
		Message: message,
		Payload: payload,
	}
	err := util.REDIS.Publish(ctx, CHANNEL_GLOBAL_CACHE, fmt.Sprintf("%v", cacheMessage)).Err()
	if err != nil {
		log.Println(err)
		return err
	}

	log.Println("Published cache message")

	return nil
}
