package services

import (
	"context"
	"log"

	"khoomi-api-io/api/internal"
	"khoomi-api-io/api/pkg/models"
	"khoomi-api-io/api/pkg/util"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type NotificationServiceImpl struct {
	notificationCollection *mongo.Collection
}

func NewNotificationService() NotificationService {
	return &NotificationServiceImpl{
		notificationCollection: util.GetCollection(util.DB, "UserNotification"),
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

// CreateNotification creates a new notification
func (ns *NotificationServiceImpl) CreateNotification(ctx context.Context, notification models.Notification) (primitive.ObjectID, error) {
	result, err := ns.notificationCollection.InsertOne(ctx, notification)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return result.InsertedID.(primitive.ObjectID), nil
}

// GetNotification retrieves a notification by user ID
func (ns *NotificationServiceImpl) GetNotification(ctx context.Context, userID primitive.ObjectID) (*models.Notification, error) {
	var notification models.Notification
	err := ns.notificationCollection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&notification)
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

// UpdateNotification updates a notification for a specific user
func (ns *NotificationServiceImpl) UpdateNotification(ctx context.Context, userID primitive.ObjectID, notification models.Notification) error {
	filter := bson.M{"user_id": userID}
	update := bson.M{"$set": notification}
	_, err := ns.notificationCollection.UpdateOne(ctx, filter, update)
	return err
}
