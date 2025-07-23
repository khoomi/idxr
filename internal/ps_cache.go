package internal

import (
	"context"
	"encoding/json"
	"khoomi-api-io/api/pkg/util"
	"log"
	"time"
)

var CHANNEL_GLOBAL_CACHE = "GLOBAL_CACHE"

type CacheMessageType string

const (
	CacheInvalidateUser               CacheMessageType = "user.invalidate"
	CacheInvalidateUserAddress        CacheMessageType = "user.addresses.invalidate"
	CacheInvalidateUserDeletion       CacheMessageType = "user.deletion.invalidate"
	CacheInvalidateUserWishlist       CacheMessageType = "user.wishlist.invalidate"
	CacheInvalidateUserPaymentCard    CacheMessageType = "user.payment.cards.invalidate"
	CacheInvalidateUserNotifications  CacheMessageType = "user.notifications.invalidate"
	CacheInvalidateUserLoginHistories CacheMessageType = "user.logins.invalidate"

	CacheInvalidateListing               CacheMessageType = "listing.invalidate"
	CacheInvalidateListings              CacheMessageType = "listings.invalidate"
	CacheInvalidateShopListings          CacheMessageType = "shop.listings.invalidate"
	CacheInvalidateListingReviews        CacheMessageType = "listing.reviews.invalidate"
	CacheInvalidateListingFavoriteToggle CacheMessageType = "listing.favorite.toggle"

	CacheInvalidateShop               CacheMessageType = "shop.invalidate"
	CacheInvalidateShops              CacheMessageType = "shops.invalidate"
	CacheInvalidateShopAbout          CacheMessageType = "shop.about.invalidate"
	CacheInvalidateShopPolicy         CacheMessageType = "shop.policy.invalidate"
	CacheInvalidateShopShipping       CacheMessageType = "shop.shipping.invalidate"
	CacheInvalidateShopCompliance     CacheMessageType = "shop.compliance.invalidate"
	CacheInvalidateShopFavoriteToggle CacheMessageType = "shop.favorite.toggle"

	CacheInvalidateCart CacheMessageType = "cart.invalidate"

	CacheInvalidatePayment CacheMessageType = "payment.invalidate"
)

type CacheMessage struct {
	Type      CacheMessageType `json:"type"`
	Payload   string           `json:"payload"`
	Timestamp int64            `json:"timestamp"`
}

// PublishCacheMessage publishes a cache invalidation message to Redis pub/sub as JSON
func PublishCacheMessage(ctx context.Context, messageType CacheMessageType, payload string) error {
	cacheMessage := CacheMessage{
		Type:      messageType,
		Payload:   payload,
		Timestamp: time.Now().Unix(),
	}

	messageJSON, err := json.Marshal(cacheMessage)
	if err != nil {
		log.Printf("Failed to marshal cache message: %v", err)
		return err
	}

	err = util.REDIS.Publish(ctx, CHANNEL_GLOBAL_CACHE, string(messageJSON)).Err()
	if err != nil {
		log.Printf("Failed to publish cache message: %v", err)
		return err
	}

	log.Printf("Published cache message: %s", messageJSON)
	return nil
}

// PublishCacheMessageDirect publishes a cache invalidation message directly without context
func PublishCacheMessageDirect(messageType CacheMessageType, payload string) error {
	return PublishCacheMessage(context.Background(), messageType, payload)
}
