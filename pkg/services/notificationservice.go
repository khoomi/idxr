package services

import (
	"context"
	"log"
	"time"

	"khoomi-api-io/api/internal"
	"khoomi-api-io/api/pkg/models"
	"khoomi-api-io/api/pkg/util"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type NotificationServiceImpl struct {
	notificationCollection     *mongo.Collection // For UserNotification (backward compatibility)
	userNotificationCollection *mongo.Collection
	shopNotificationCollection *mongo.Collection
}

func NewNotificationService() NotificationService {
	userNotificationCol := util.GetCollection(util.DB, "UserNotification")
	return &NotificationServiceImpl{
		notificationCollection:     userNotificationCol, // Same collection for backward compatibility
		userNotificationCollection: userNotificationCol,
		shopNotificationCollection: util.GetCollection(util.DB, "ShopNotification"),
	}
}

// SendReviewNotificationAsync sends review-related notifications asynchronously
func (ns *NotificationServiceImpl) SendReviewNotificationAsync(ctx context.Context, reviewID primitive.ObjectID) error {
	// This would typically be implemented with a job queue like Redis Queue, RabbitMQ, etc.
	// For now, we'll use a simple goroutine approach
	go func() {
		// TODO: Implement actual notification logic
		// - Send email to shop owner about new review
		// - Send push notification to relevant users
		// - Update notification counters
		log.Printf("Async: Processing review notification for review ID: %s", reviewID.Hex())
		// Example notification processing:
		// 1. Get review details
		// 2. Get shop owner details
		// 3. Send email notification
		// 4. Update notification preferences
	}()

	return nil
}

// SendCartAbandonmentNotificationAsync sends cart abandonment notifications
func (ns *NotificationServiceImpl) SendCartAbandonmentNotificationAsync(ctx context.Context, userID primitive.ObjectID) error {
	go func() {
		// TODO: Implement cart abandonment notification
		// - Check if cart has items older than X hours
		// - Send reminder email
		// - Track abandonment metrics
		log.Printf("Async: Processing cart abandonment notification for user ID: %s", userID.Hex())
	}()

	return nil
}

// InvalidateReviewCache invalidates review-related cache entries
func (ns *NotificationServiceImpl) InvalidateReviewCache(ctx context.Context, listingID primitive.ObjectID) error {
	return internal.PublishCacheMessageDirect(internal.CacheInvalidateListingReviews, listingID.Hex())
}

// InvalidateCartCache invalidates cart-related cache entries
func (ns *NotificationServiceImpl) InvalidateCartCache(ctx context.Context, userID primitive.ObjectID) error {
	return internal.PublishCacheMessageDirect(internal.CacheInvalidateCart, userID.Hex())
}

// CreateNotification creates a new user notification
func (ns *NotificationServiceImpl) CreateNotification(ctx context.Context, notification models.Notification) (primitive.ObjectID, error) {
	if notification.ID.IsZero() {
		notification.ID = primitive.NewObjectID()
	}
	if notification.CreatedAt.IsZero() {
		notification.CreatedAt = time.Now()
	}
	if !notification.IsRead {
		notification.IsRead = false
	}

	if notification.ExpiresAt == nil && notification.Type != "" {
		expiry := notification.Type.GetExpiryDuration()
		if expiry > 0 {
			expiresAt := time.Now().Add(expiry)
			notification.ExpiresAt = &expiresAt
		}
	}

	result, err := ns.notificationCollection.InsertOne(ctx, notification)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return result.InsertedID.(primitive.ObjectID), nil
}

// GetNotification retrieves the most recent notification for a user
func (ns *NotificationServiceImpl) GetNotification(ctx context.Context, userID primitive.ObjectID) (*models.Notification, error) {
	var notification models.Notification

	opts := options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}})
	err := ns.notificationCollection.FindOne(ctx, bson.M{"user_id": userID}, opts).Decode(&notification)
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

// UpdateNotification updates a specific notification
func (ns *NotificationServiceImpl) UpdateNotification(ctx context.Context, userID primitive.ObjectID, notification models.Notification) error {
	var filter bson.M
	if !notification.ID.IsZero() {
		filter = bson.M{"_id": notification.ID, "user_id": userID}
	} else {
		filter = bson.M{"user_id": userID}
	}

	update := bson.M{"$set": notification}
	_, err := ns.notificationCollection.UpdateOne(ctx, filter, update)
	return err
}
