package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"khoomi-api-io/api/internal"
	"khoomi-api-io/api/internal/common"
	"khoomi-api-io/api/pkg/models"
	"khoomi-api-io/api/pkg/util"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type NotificationServiceImpl struct {
	userNotificationCollection *mongo.Collection
	shopNotificationCollection *mongo.Collection
}

func NewNotificationService() NotificationService {
	return &NotificationServiceImpl{
		userNotificationCollection: util.GetCollection(util.DB, "UserNotification"),
		shopNotificationCollection: util.GetCollection(util.DB, "ShopNotification"),
	}
}

// SendReviewNotificationAsync sends review-related notifications asynchronously
func (ns *NotificationServiceImpl) SendReviewNotificationAsync(ctx context.Context, reviewID primitive.ObjectID) error {
	go func() {
		notifCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var review models.ListingReview
		err := common.ListingReviewCollection.FindOne(notifCtx, bson.M{"_id": reviewID}).Decode(&review)
		if err != nil {
			log.Printf("Failed to fetch review %s: %v", reviewID.Hex(), err)
			return
		}
		var shop models.Shop
		err = common.ShopCollection.FindOne(notifCtx, bson.M{"_id": review.ShopId}).Decode(&shop)
		if err != nil {
			log.Printf("Failed to fetch shop %s: %v", review.ShopId.Hex(), err)
			return
		}

		var listing models.Listing
		err = common.ListingCollection.FindOne(notifCtx, bson.M{"_id": review.ListingId}).Decode(&listing)
		if err != nil {
			log.Printf("Failed to fetch listing %s: %v", review.ListingId.Hex(), err)
			return
		}

		notification := models.UserNotification{
			ID:       primitive.NewObjectID(),
			UserID:   shop.UserID,
			Type:     models.NotificationReviewReceived,
			Title:    "New Review Received",
			Message:  fmt.Sprintf("You received a %d-star review for %s from %s", review.Rating, listing.Details.Title, review.ReviewAuthor),
			Priority: models.NotificationPriorityMedium,
			IsRead:   false,
			Data: map[string]any{
				"reviewId":   review.Id.Hex(),
				"listingId":  review.ListingId.Hex(),
				"shopId":     review.ShopId.Hex(),
				"rating":     review.Rating,
				"reviewText": review.Review,
			},
			RelatedEntityID:   &review.ListingId,
			RelatedEntityType: "listing",
			ActionURL:         fmt.Sprintf("/shops/%s/reviews", shop.ID.Hex()),
			ImageURL:          listing.Images[0],
			CreatedAt:         time.Now(),
		}

		expiry := notification.Type.GetExpiryDuration()
		if expiry > 0 {
			expiresAt := time.Now().Add(expiry)
			notification.ExpiresAt = &expiresAt
		}

		_, err = ns.userNotificationCollection.InsertOne(notifCtx, notification)
		if err != nil {
			log.Printf("Failed to create review notification: %v", err)
			return
		}

		var shopNotifSettings models.ShopNotificationSettings
		err = util.GetCollection(util.DB, "ShopNotificationSettings").FindOne(
			notifCtx,
			bson.M{"shop_id": shop.ID},
		).Decode(&shopNotifSettings)

		if err == nil && shopNotifSettings.EmailEnabled && shopNotifSettings.CustomerNotifications {
			// TODO: Send email notification when email service is available
			log.Printf("Would send email notification for review to shop %s", shop.Name)
		}

		log.Printf("Successfully created review notification for shop %s", shop.ID.Hex())
	}()

	return nil
}

// SendCartAbandonmentNotificationAsync sends cart abandonment notifications
func (ns *NotificationServiceImpl) SendCartAbandonmentNotificationAsync(ctx context.Context, userID primitive.ObjectID) error {
	go func() {
		notifCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		filter := bson.M{
			"user_id": userID,
		}

		cursor, err := common.UserCartCollection.Find(notifCtx, filter)
		if err != nil {
			log.Printf("Failed to fetch abandoned cart items for user %s: %v", userID.Hex(), err)
			return
		}
		defer cursor.Close(notifCtx)

		var cartItems []models.CartItem
		if err := cursor.All(notifCtx, &cartItems); err != nil {
			log.Printf("Failed to decode cart items: %v", err)
			return
		}

		if len(cartItems) == 0 {
			return // No abandoned items
		}

		var user models.User
		err = common.UserCollection.FindOne(notifCtx, bson.M{"_id": userID}).Decode(&user)
		if err != nil {
			log.Printf("Failed to fetch user %s: %v", userID.Hex(), err)
			return
		}

		sevenDaysAgo := time.Now().Add(-7 * 24 * time.Hour)
		existingNotif := bson.M{
			"user_id":    userID,
			"type":       models.NotificationCartAbandonment,
			"created_at": bson.M{"$gte": sevenDaysAgo},
		}

		count, err := ns.userNotificationCollection.CountDocuments(notifCtx, existingNotif)
		if err == nil && count > 0 {
			return // Already sent recent abandonment notification
		}

		notification := models.UserNotification{
			ID:       primitive.NewObjectID(),
			UserID:   userID,
			Type:     models.NotificationCartAbandonment,
			Title:    "Items in Your Cart Are Waiting!",
			Message:  fmt.Sprintf("Hi %s, you have %d items in your cart. Complete your purchase before they're gone!", user.FirstName, len(cartItems)),
			Priority: models.NotificationPriorityLow,
			IsRead:   false,
			Data: map[string]any{
				"cartItemCount": len(cartItems),
			},
			RelatedEntityType: "cart",
			ActionURL:         "/cart",
			CreatedAt:         time.Now(),
		}

		expiry := notification.Type.GetExpiryDuration()
		if expiry > 0 {
			expiresAt := time.Now().Add(expiry)
			notification.ExpiresAt = &expiresAt
		}

		_, err = ns.userNotificationCollection.InsertOne(notifCtx, notification)
		if err != nil {
			log.Printf("Failed to create cart abandonment notification: %v", err)
			return
		}

		var userNotifSettings models.UserNotificationSettings
		err = util.GetCollection(util.DB, "UserNotificationSettings").FindOne(
			notifCtx,
			bson.M{"user_id": userID},
		).Decode(&userNotifSettings)

		if err == nil && userNotifSettings.EmailEnabled {
			// TODO: Send email reminder when email service is available
			log.Printf("Would send cart abandonment email to user %s", user.PrimaryEmail)
		}

		log.Printf("Successfully created cart abandonment notification for user %s", userID.Hex())
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
func (ns *NotificationServiceImpl) CreateNotification(ctx context.Context, notification models.UserNotification) (primitive.ObjectID, error) {
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

	result, err := ns.userNotificationCollection.InsertOne(ctx, notification)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return result.InsertedID.(primitive.ObjectID), nil
}

// GetNotification retrieves the most recent notification for a user
func (ns *NotificationServiceImpl) GetNotification(ctx context.Context, userID primitive.ObjectID) (*models.UserNotification, error) {
	var notification models.UserNotification

	opts := options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}})
	err := ns.userNotificationCollection.FindOne(ctx, bson.M{"user_id": userID}, opts).Decode(&notification)
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

// UpdateNotification updates a specific notification
func (ns *NotificationServiceImpl) UpdateNotification(ctx context.Context, userID primitive.ObjectID, notification models.UserNotification) error {
	var filter bson.M
	if !notification.ID.IsZero() {
		filter = bson.M{"_id": notification.ID, "user_id": userID}
	} else {
		filter = bson.M{"user_id": userID}
	}

	update := bson.M{"$set": notification}
	_, err := ns.userNotificationCollection.UpdateOne(ctx, filter, update)
	return err
}

// GetUserNotifications retrieves paginated notifications for a user with filters
func (ns *NotificationServiceImpl) GetUserNotifications(ctx context.Context, userID primitive.ObjectID, filters models.NotificationFilters, pagination util.PaginationArgs) ([]models.UserNotification, int64, error) {
	filter := bson.M{"user_id": userID}

	// Apply filters
	if len(filters.Types) > 0 {
		filter["type"] = bson.M{"$in": filters.Types}
	}
	if len(filters.Priorities) > 0 {
		filter["priority"] = bson.M{"$in": filters.Priorities}
	}
	if filters.IsRead != nil {
		filter["is_read"] = *filters.IsRead
	}
	if filters.StartDate != nil {
		filter["created_at"] = bson.M{"$gte": *filters.StartDate}
	}
	if filters.EndDate != nil {
		if _, exists := filter["created_at"]; exists {
			filter["created_at"].(bson.M)["$lte"] = *filters.EndDate
		} else {
			filter["created_at"] = bson.M{"$lte": *filters.EndDate}
		}
	}

	// Count total documents
	count, err := ns.userNotificationCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Find notifications with pagination
	opts := options.Find().
		SetSort(bson.D{{Key: "priority", Value: -1}, {Key: "created_at", Value: -1}}).
		SetSkip(int64(pagination.Skip)).
		SetLimit(int64(pagination.Limit))

	cursor, err := ns.userNotificationCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var notifications []models.UserNotification
	if err := cursor.All(ctx, &notifications); err != nil {
		return nil, 0, err
	}

	if notifications == nil {
		notifications = []models.UserNotification{}
	}

	return notifications, count, nil
}

// GetUnreadNotifications retrieves unread notifications for a user
func (ns *NotificationServiceImpl) GetUnreadNotifications(ctx context.Context, userID primitive.ObjectID, pagination util.PaginationArgs) ([]models.UserNotification, int64, error) {
	filter := bson.M{
		"user_id": userID,
		"is_read": false,
	}

	count, err := ns.userNotificationCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "priority", Value: -1}, {Key: "created_at", Value: -1}}).
		SetSkip(int64(pagination.Skip)).
		SetLimit(int64(pagination.Limit))

	cursor, err := ns.userNotificationCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var notifications []models.UserNotification
	if err := cursor.All(ctx, &notifications); err != nil {
		return nil, 0, err
	}

	if notifications == nil {
		notifications = []models.UserNotification{}
	}

	return notifications, count, nil
}

// GetNotificationByID retrieves a specific notification by ID
func (ns *NotificationServiceImpl) GetNotificationByID(ctx context.Context, userID, notificationID primitive.ObjectID) (*models.UserNotification, error) {
	var notification models.UserNotification

	filter := bson.M{
		"_id":     notificationID,
		"user_id": userID,
	}

	err := ns.userNotificationCollection.FindOne(ctx, filter).Decode(&notification)
	if err != nil {
		return nil, err
	}

	return &notification, nil
}

// MarkNotificationAsRead marks a single notification as read
func (ns *NotificationServiceImpl) MarkNotificationAsRead(ctx context.Context, userID, notificationID primitive.ObjectID) error {
	filter := bson.M{
		"_id":     notificationID,
		"user_id": userID,
	}

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"is_read": true,
			"read_at": now,
		},
	}

	result, err := ns.userNotificationCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

// MarkAllNotificationsAsRead marks all notifications as read for a user
func (ns *NotificationServiceImpl) MarkAllNotificationsAsRead(ctx context.Context, userID primitive.ObjectID) (int64, error) {
	filter := bson.M{
		"user_id": userID,
		"is_read": false,
	}

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"is_read": true,
			"read_at": now,
		},
	}

	result, err := ns.userNotificationCollection.UpdateMany(ctx, filter, update)
	if err != nil {
		return 0, err
	}

	return result.ModifiedCount, nil
}

// MarkNotificationsAsRead marks multiple notifications as read
func (ns *NotificationServiceImpl) MarkNotificationsAsRead(ctx context.Context, userID primitive.ObjectID, notificationIDs []primitive.ObjectID) (int64, error) {
	filter := bson.M{
		"_id":     bson.M{"$in": notificationIDs},
		"user_id": userID,
		"is_read": false,
	}

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"is_read": true,
			"read_at": now,
		},
	}

	result, err := ns.userNotificationCollection.UpdateMany(ctx, filter, update)
	if err != nil {
		return 0, err
	}

	return result.ModifiedCount, nil
}

// DeleteNotification deletes a specific notification
func (ns *NotificationServiceImpl) DeleteNotification(ctx context.Context, userID, notificationID primitive.ObjectID) error {
	filter := bson.M{
		"_id":     notificationID,
		"user_id": userID,
	}

	result, err := ns.userNotificationCollection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

// DeleteExpiredNotifications removes all expired notifications
func (ns *NotificationServiceImpl) DeleteExpiredNotifications(ctx context.Context) (int64, error) {
	filter := bson.M{
		"expires_at": bson.M{
			"$ne":  nil,
			"$lte": time.Now(),
		},
	}

	result, err := ns.userNotificationCollection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}

	if result.DeletedCount > 0 {
		log.Printf("Deleted %d expired notifications", result.DeletedCount)
	}

	return result.DeletedCount, nil
}

// DeleteAllUserNotifications deletes all notifications for a user
func (ns *NotificationServiceImpl) DeleteAllUserNotifications(ctx context.Context, userID primitive.ObjectID) (int64, error) {
	filter := bson.M{"user_id": userID}

	result, err := ns.userNotificationCollection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}

	return result.DeletedCount, nil
}

// GetUnreadNotificationCount returns the count of unread notifications for a user
func (ns *NotificationServiceImpl) GetUnreadNotificationCount(ctx context.Context, userID primitive.ObjectID) (int64, error) {
	filter := bson.M{
		"user_id": userID,
		"is_read": false,
	}

	count, err := ns.userNotificationCollection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetNotificationStats returns statistics about user notifications
func (ns *NotificationServiceImpl) GetNotificationStats(ctx context.Context, userID primitive.ObjectID) (map[string]int64, error) {
	stats := make(map[string]int64)

	totalFilter := bson.M{"user_id": userID}
	total, err := ns.userNotificationCollection.CountDocuments(ctx, totalFilter)
	if err != nil {
		return nil, err
	}
	stats["total"] = total

	unreadFilter := bson.M{"user_id": userID, "is_read": false}
	unread, err := ns.userNotificationCollection.CountDocuments(ctx, unreadFilter)
	if err != nil {
		return nil, err
	}
	stats["unread"] = unread
	stats["read"] = total - unread

	highPriorityFilter := bson.M{
		"user_id":  userID,
		"is_read":  false,
		"priority": bson.M{"$in": []string{"high", "urgent"}},
	}
	highPriority, err := ns.userNotificationCollection.CountDocuments(ctx, highPriorityFilter)
	if err != nil {
		return nil, err
	}
	stats["highPriorityUnread"] = highPriority

	return stats, nil
}

// Helper methods for creating common notifications

// CreateOrderNotification creates a notification for order-related events
func (ns *NotificationServiceImpl) CreateOrderNotification(ctx context.Context, userID primitive.ObjectID, orderID primitive.ObjectID, notifType models.NotificationType, title, message string) error {
	notification := models.UserNotification{
		ID:                primitive.NewObjectID(),
		UserID:            userID,
		Type:              notifType,
		Title:             title,
		Message:           message,
		Priority:          models.NotificationPriorityHigh,
		IsRead:            false,
		RelatedEntityID:   &orderID,
		RelatedEntityType: "order",
		ActionURL:         fmt.Sprintf("/orders/%s", orderID.Hex()),
		CreatedAt:         time.Now(),
	}

	expiry := notifType.GetExpiryDuration()
	if expiry > 0 {
		expiresAt := time.Now().Add(expiry)
		notification.ExpiresAt = &expiresAt
	}

	_, err := ns.userNotificationCollection.InsertOne(ctx, notification)
	return err
}

// CreatePaymentNotification creates a notification for payment-related events
func (ns *NotificationServiceImpl) CreatePaymentNotification(ctx context.Context, userID primitive.ObjectID, paymentID primitive.ObjectID, notifType models.NotificationType, title, message string, amount float64) error {
	notification := models.UserNotification{
		ID:       primitive.NewObjectID(),
		UserID:   userID,
		Type:     notifType,
		Title:    title,
		Message:  message,
		Priority: models.NotificationPriorityUrgent,
		IsRead:   false,
		Data: map[string]any{
			"paymentId": paymentID.Hex(),
			"amount":    amount,
		},
		RelatedEntityID:   &paymentID,
		RelatedEntityType: "payment",
		CreatedAt:         time.Now(),
	}

	if notifType != models.NotificationPaymentFailed {
		expiry := notifType.GetExpiryDuration()
		if expiry > 0 {
			expiresAt := time.Now().Add(expiry)
			notification.ExpiresAt = &expiresAt
		}
	}

	_, err := ns.userNotificationCollection.InsertOne(ctx, notification)
	return err
}

// CreateInventoryNotification creates a notification for inventory-related events
func (ns *NotificationServiceImpl) CreateInventoryNotification(ctx context.Context, shopID primitive.ObjectID, listingID primitive.ObjectID, notifType models.NotificationType, title, message string, stockLevel int) error {
	var shop models.Shop
	err := common.ShopCollection.FindOne(ctx, bson.M{"_id": shopID}).Decode(&shop)
	if err != nil {
		return err
	}

	priority := models.NotificationPriorityMedium
	if notifType == models.NotificationOutOfStock {
		priority = models.NotificationPriorityHigh
	}

	notification := models.UserNotification{
		ID:       primitive.NewObjectID(),
		UserID:   shop.UserID,
		Type:     notifType,
		Title:    title,
		Message:  message,
		Priority: priority,
		IsRead:   false,
		Data: map[string]any{
			"shopId":     shopID.Hex(),
			"listingId":  listingID.Hex(),
			"stockLevel": stockLevel,
		},
		RelatedEntityID:   &listingID,
		RelatedEntityType: "listing",
		ActionURL:         fmt.Sprintf("/shops/%s/inventory", shopID.Hex()),
		CreatedAt:         time.Now(),
	}

	expiry := notifType.GetExpiryDuration()
	if expiry > 0 {
		expiresAt := time.Now().Add(expiry)
		notification.ExpiresAt = &expiresAt
	}

	_, err = ns.userNotificationCollection.InsertOne(ctx, notification)
	return err
}

// CreateSystemNotification creates a system-wide notification
func (ns *NotificationServiceImpl) CreateSystemNotification(ctx context.Context, userID primitive.ObjectID, notifType models.NotificationType, title, message string, priority models.NotificationPriority) error {
	notification := models.UserNotification{
		ID:                primitive.NewObjectID(),
		UserID:            userID,
		Type:              notifType,
		Title:             title,
		Message:           message,
		Priority:          priority,
		IsRead:            false,
		RelatedEntityType: "system",
		CreatedAt:         time.Now(),
	}

	if notifType != models.NotificationSecurityAlert {
		expiry := notifType.GetExpiryDuration()
		if expiry > 0 {
			expiresAt := time.Now().Add(expiry)
			notification.ExpiresAt = &expiresAt
		}
	}

	_, err := ns.userNotificationCollection.InsertOne(ctx, notification)
	return err
}
