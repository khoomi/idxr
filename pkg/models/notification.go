package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	// Order related notifications
	NotificationOrderCreated   NotificationType = "order_created"
	NotificationOrderShipped   NotificationType = "order_shipped"
	NotificationOrderDelivered NotificationType = "order_delivered"
	NotificationOrderCancelled NotificationType = "order_cancelled"
	NotificationOrderRefunded  NotificationType = "order_refunded"

	// Payment related notifications
	NotificationPaymentSuccess NotificationType = "payment_success"
	NotificationPaymentFailed  NotificationType = "payment_failed"
	NotificationPaymentPending NotificationType = "payment_pending"
	NotificationRefundIssued   NotificationType = "refund_issued"

	// Review related notifications
	NotificationReviewReceived NotificationType = "review_received"
	NotificationReviewApproved NotificationType = "review_approved"
	NotificationReviewFlagged  NotificationType = "review_flagged"

	// Cart related notifications
	NotificationCartAbandonment NotificationType = "cart_abandonment"
	NotificationItemBackInStock NotificationType = "item_back_in_stock"
	NotificationPriceDropped   NotificationType = "price_dropped"

	// Inventory related notifications
	NotificationLowStock      NotificationType = "low_stock"
	NotificationOutOfStock    NotificationType = "out_of_stock"
	NotificationRestocked     NotificationType = "restocked"
	NotificationPopularItem   NotificationType = "popular_item"

	// Shop related notifications
	NotificationNewFollower       NotificationType = "new_follower"
	NotificationShopAnnouncement  NotificationType = "shop_announcement"
	NotificationShopVacationMode  NotificationType = "shop_vacation_mode"
	NotificationShopVerified      NotificationType = "shop_verified"

	// System notifications
	NotificationSystemAlert      NotificationType = "system_alert"
	NotificationSecurityAlert    NotificationType = "security_alert"
	NotificationAccountUpdate    NotificationType = "account_update"
	NotificationPolicyUpdate     NotificationType = "policy_update"
	NotificationPromotionalOffer NotificationType = "promotional_offer"
)

// NotificationPriority represents the priority level of a notification
type NotificationPriority string

const (
	NotificationPriorityLow    NotificationPriority = "low"
	NotificationPriorityMedium NotificationPriority = "medium"
	NotificationPriorityHigh   NotificationPriority = "high"
	NotificationPriorityUrgent NotificationPriority = "urgent"
)

// UserNotification represents a notification for a user
type UserNotification struct {
	ID                primitive.ObjectID    `bson:"_id" json:"_id"`
	UserID            primitive.ObjectID    `bson:"user_id" json:"userId" validate:"required"`
	Type              NotificationType      `bson:"type" json:"type" validate:"required"`
	Title             string                `bson:"title" json:"title" validate:"required"`
	Message           string                `bson:"message" json:"message" validate:"required"`
	Priority          NotificationPriority  `bson:"priority" json:"priority" validate:"required"`
	IsRead            bool                  `bson:"is_read" json:"isRead"`
	ReadAt            *time.Time            `bson:"read_at,omitempty" json:"readAt,omitempty"`
	Data              map[string]any        `bson:"data,omitempty" json:"data,omitempty"`
	RelatedEntityID   *primitive.ObjectID   `bson:"related_entity_id,omitempty" json:"relatedEntityId,omitempty"`
	RelatedEntityType string                `bson:"related_entity_type,omitempty" json:"relatedEntityType,omitempty"`
	ActionURL         string                `bson:"action_url,omitempty" json:"actionUrl,omitempty"`
	ImageURL          string                `bson:"image_url,omitempty" json:"imageUrl,omitempty"`
	CreatedAt         time.Time             `bson:"created_at" json:"createdAt"`
	ExpiresAt         *time.Time            `bson:"expires_at,omitempty" json:"expiresAt,omitempty"`
}

// Notification is an alias for UserNotification to maintain backward compatibility
type Notification = UserNotification

// UserNotificationRequest represents a request to create a user notification
type UserNotificationRequest struct {
	Type              NotificationType      `json:"type" validate:"required"`
	Title             string                `json:"title" validate:"required"`
	Message           string                `json:"message" validate:"required"`
	Priority          NotificationPriority  `json:"priority" validate:"required"`
	Data              map[string]any        `json:"data,omitempty"`
	RelatedEntityID   *primitive.ObjectID   `json:"relatedEntityId,omitempty"`
	RelatedEntityType string                `json:"relatedEntityType,omitempty"`
	ActionURL         string                `json:"actionUrl,omitempty"`
	ImageURL          string                `json:"imageUrl,omitempty"`
	ExpiresAt         *time.Time            `json:"expiresAt,omitempty"`
}

// MarkAsReadRequest represents a request to mark notifications as read
type MarkAsReadRequest struct {
	NotificationIDs []primitive.ObjectID `json:"notificationIds" validate:"required,min=1"`
}

// NotificationFilters represents filters for querying notifications
type NotificationFilters struct {
	Types      []NotificationType   `json:"types,omitempty"`
	Priorities []NotificationPriority `json:"priorities,omitempty"`
	IsRead     *bool                `json:"isRead,omitempty"`
	StartDate  *time.Time           `json:"startDate,omitempty"`
	EndDate    *time.Time           `json:"endDate,omitempty"`
}

// GetExpiryDuration returns the default expiry duration for a notification type
func (nt NotificationType) GetExpiryDuration() time.Duration {
	switch nt {
	case NotificationPromotionalOffer:
		return 7 * 24 * time.Hour // 7 days
	case NotificationCartAbandonment, NotificationPriceDropped:
		return 3 * 24 * time.Hour // 3 days
	case NotificationSystemAlert, NotificationPolicyUpdate:
		return 30 * 24 * time.Hour // 30 days
	case NotificationOrderCreated, NotificationOrderShipped, NotificationOrderDelivered:
		return 90 * 24 * time.Hour // 90 days
	case NotificationSecurityAlert:
		return 0 // Never expire
	default:
		return 30 * 24 * time.Hour // Default to 30 days
	}
}

// ShouldSendEmail determines if this notification type should trigger an email
func (nt NotificationType) ShouldSendEmail() bool {
	switch nt {
	case NotificationOrderCreated, NotificationOrderShipped, NotificationOrderDelivered,
		NotificationPaymentSuccess, NotificationPaymentFailed,
		NotificationSecurityAlert, NotificationAccountUpdate:
		return true
	default:
		return false
	}
}

// GetPriorityWeight returns a numeric weight for sorting notifications by priority
func (np NotificationPriority) GetPriorityWeight() int {
	switch np {
	case NotificationPriorityUrgent:
		return 4
	case NotificationPriorityHigh:
		return 3
	case NotificationPriorityMedium:
		return 2
	case NotificationPriorityLow:
		return 1
	default:
		return 0
	}
}