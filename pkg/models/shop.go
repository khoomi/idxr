package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Shop struct {
	AnnouncementModifiedAt time.Time            `bson:"announcement_modified_at" json:"announcementModifiedAt" validate:"omitempty"`
	ModifiedAt             time.Time            `bson:"modified_at" json:"modifiedAt" validate:"required"`
	CreatedAt              time.Time            `bson:"created_at" json:"createdAt" validate:"required"`
	User                   ListingUserExcept    `bson:"user" json:"user"`
	Policy                 ShopPolicy           `bson:"policy" json:"policy" validate:"required"`
	BannerURL              string               `bson:"banner_url" json:"bannerUrl"`
	Status                 ShopStatus           `bson:"status" json:"status" validate:"required,oneof=inactive active banned suspended warning pendingreview"`
	Name                   string               `bson:"name" json:"name" validate:"required"`
	Announcement           string               `bson:"announcement" json:"announcement"`
	SalesMessage           string               `bson:"sales_message" json:"salesMessage"`
	Description            string               `bson:"description" json:"description" validate:"required"`
	VacationMessage        string               `bson:"vacation_message" json:"vacationMessage"`
	Slug                   string               `bson:"slug" json:"slug" validate:"required"`
	LogoURL                string               `bson:"logo_url" json:"logoUrl"`
	Username               string               `bson:"username" json:"username" validate:"required"`
	Categories             []ShopCategory       `bson:"-" json:"categories"`
	Followers              []ShopFollower       `bson:"followers" json:"followers"`
	Links                  []Link               `bson:"-" json:"links"`
	Gallery                []string             `bson:"gallery" json:"gallery"`
	About                  ShopAbout            `bson:"about" json:"about"`
	Address                ShopAddress          `bson:"address" json:"address"`
	Rating                 Rating               `bson:"rating" json:"rating"`
	ReviewsCount           int                  `bson:"reviews_count" json:"reviewsCount"`
	FinancialInformation   FinancialInformation `bson:"financial_information" json:"financialInformation"`
	ListingActiveCount     int64                `bson:"listing_active_count" json:"listingActiveCount" validate:"required"`
	FollowerCount          int                  `bson:"follower_count" json:"followerCount" validate:"required"`
	ID                     primitive.ObjectID   `bson:"_id" json:"_id" validate:"required"`
	UserID                 primitive.ObjectID   `bson:"user_id" json:"userId"`
	IsVacation             bool                 `bson:"is_vacation" json:"isVacation"`
	IsLive                 bool                 `bson:"is_live" json:"isLive"`
}

type ShopAddress struct {
	City       string    `bson:"city" json:"city"`
	State      string    `bson:"state" json:"state"`
	Street     string    `bson:"street" json:"street"`
	PostalCode string    `bson:"postal_code" json:"postalCode"`
	Country    Country   `bson:"country" json:"country"`
	ModifiedAt time.Time `bson:"modified_at" json:"modifiedAt"`
}

type ShopExcerpt struct {
	ID                 primitive.ObjectID `bson:"_id"                   json:"id"`
	Name               string             `bson:"name"                  json:"name"`
	Slug               string             `bson:"slug"                  json:"slug"`
	Username           string             `bson:"username"              json:"username"`
	LogoURL            string             `bson:"logo_url"              json:"logoUrl"`
	BannerURL          string             `bson:"banner_url"            json:"bannerUrl"`
	Status             ShopStatus         `bson:"status"                json:"status"`
	CreatedAt          time.Time          `bson:"created_at"            json:"createdAt"`
	ListingActiveCount int64              `bson:"listing_active_count"  json:"listingActiveCount"`
	FollowerCount      int                `bson:"follower_count"        json:"followerCount"`
	Rating             Rating             `bson:"rating"                json:"rating"`
	ReviewsCount       int                `bson:"reviews_count"         json:"reviewsCount"`
}

type ShopCategory struct {
	Name          string `bson:"name" json:"name"`
	Path          string `bson:"path" json:"path"`
	ListingsCount int    `bson:"count" json:"count"`
}

type ShopFollower struct {
	JoinedAt  time.Time          `bson:"joined_at" json:"joinedAt" validate:"required"`
	FirstName string             `bson:"first_name" json:"firstName" validate:"required"`
	LastName  string             `bson:"last_name" json:"lastName" validate:"required"`
	LoginName string             `bson:"login_name" json:"loginName" validate:"required"`
	Thumbnail string             `bson:"thumbnail" json:"thumbnail"`
	Id        primitive.ObjectID `bson:"_id" json:"_id"`
	UserId    primitive.ObjectID `bson:"user_id" json:"userId"`
	ShopId    primitive.ObjectID `bson:"shop_id" json:"shopId"`
	IsOwner   bool               `bson:"is_owner" json:"isOwner" validate:"required"`
}

type ShopFollowerExcerpt struct {
	FirstName string             `bson:"first_name" json:"firstName" validate:"required"`
	LastName  string             `bson:"last_name" json:"lastName" validate:"required"`
	LoginName string             `bson:"login_name" json:"loginName"`
	Thumbnail string             `bson:"thumbnail" json:"thumbnail"`
	Id        primitive.ObjectID `bson:"follower_id" json:"followerId"`
	UserId    primitive.ObjectID `bson:"user_id" json:"userId"`
	IsOwner   bool               `bson:"is_owner" json:"isOwner" validate:"required"`
}

type ShopStatus string

const (
	ShopStatusInactive      ShopStatus = "inactive"
	ShopStatusActive        ShopStatus = "active"
	ShopStatusBanned        ShopStatus = "banned"
	ShopStatusSuspended     ShopStatus = "suspended"
	ShopStatusWarning       ShopStatus = "warning"
	ShopStatusPendingReview ShopStatus = "pendingreview"
)

type ShopPolicy struct {
	PaymentPolicy  string `bson:"payment_policy" json:"paymentPolicy"`
	ShippingPolicy string `bson:"shipping_policy" json:"shippingPolicy"`
	RefundPolicy   string `bson:"refund_policy" json:"refundPolicy"`
	AdditionalInfo string `bson:"additional_info" json:"additionalInfo"`
}

type ShopAnnouncementRequest struct {
	Announcement string `bson:"announcement" json:"announcement"`
}

type ShopVacationRequest struct {
	Message    string `json:"vacationMessage"`
	IsVacation bool   `json:"isVacation"`
}

type ShopBasicInformationRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	IsLive       bool   `json:"isLive"`
	Announcement string `json:"announcement"`
	SalesMessage string `json:"salesMessage"`
}

type ShopAbout struct {
	Headline  string `bson:"headline" json:"headline" validate:"required"`
	Story     string `bson:"story" json:"story" validate:"required"`
	X         string `bson:"x" json:"x" validate:"required"`
	Facebook  string `bson:"facebook" json:"facebook" validate:"required"`
	Instagram string `bson:"instagram" json:"instagram" validate:"required"`
}

type ShopReturnPolicies struct {
	ID               primitive.ObjectID `bson:"_id" json:"_id" validate:"omitempty"`
	ShopId           primitive.ObjectID `bson:"shop_id" json:"shopId" validate:"omitempty"`
	AcceptsReturn    bool               `bson:"accepts_return" json:"acceptsReturn"`
	AcceptsExchanges bool               `bson:"accepts_exchanges" json:"acceptsExchanges"`
	Deadline         int                `bson:"deadline" json:"deadline" validate:"oneof=7 14 21 30 45 60 90"`
}

type UpdateShopStatusReq struct {
	Status bool `json:"status" validate:"required"`
}

type ComplianceInformation struct {
	ID                   primitive.ObjectID `bson:"_id" json:"_id"`
	ShopID               primitive.ObjectID `bson:"shop_id" json:"shopId"`
	TermsOfUse           bool               `bson:"terms_of_use" json:"termsOfUse"`
	SellerPolicie        bool               `bson:"seller_policies" json:"sellerPolicies"`
	IntellectualProperty bool               `bson:"intellectual_property" json:"intellectualProperty"`
}

type ComplianceInformationRequest struct {
	TermsOfUse           bool `bson:"terms_of_use" json:"termsOfUse"`
	SellerPolicie        bool `bson:"seller_policies" json:"sellerPolicies"`
	IntellectualProperty bool `bson:"intellectual_property" json:"intellectualProperty"`
}

type ShopNotificationType string

const (
	ShopNotificationNewOrder             ShopNotificationType = "new_order"
	ShopNotificationPaymentConfirmed     ShopNotificationType = "payment_confirmed"
	ShopNotificationPaymentFailed        ShopNotificationType = "payment_failed"
	ShopNotificationOrderCancelled       ShopNotificationType = "order_cancelled"
	ShopNotificationLowStock             ShopNotificationType = "low_stock"
	ShopNotificationOutOfStock           ShopNotificationType = "out_of_stock"
	ShopNotificationInventoryRestocked   ShopNotificationType = "inventory_restocked"
	ShopNotificationNewReview            ShopNotificationType = "new_review"
	ShopNotificationCustomerMessage      ShopNotificationType = "customer_message"
	ShopNotificationReturnRequest        ShopNotificationType = "return_request"
	ShopNotificationSalesSummary         ShopNotificationType = "sales_summary"
	ShopNotificationRevenueMilestone     ShopNotificationType = "revenue_milestone"
	ShopNotificationPopularProduct       ShopNotificationType = "popular_product"
	ShopNotificationAccountVerification  ShopNotificationType = "account_verification"
	ShopNotificationPolicyUpdate         ShopNotificationType = "policy_update"
	ShopNotificationSecurityAlert        ShopNotificationType = "security_alert"
	ShopNotificationSubscriptionReminder ShopNotificationType = "subscription_reminder"
)

type ShopNotificationPriority string

const (
	ShopNotificationPriorityLow      ShopNotificationPriority = "low"
	ShopNotificationPriorityMedium   ShopNotificationPriority = "medium"
	ShopNotificationPriorityHigh     ShopNotificationPriority = "high"
	ShopNotificationPriorityCritical ShopNotificationPriority = "critical"
)

type ShopNotificationSettings struct {
	ID                     primitive.ObjectID `bson:"_id" json:"_id"`
	ShopID                 primitive.ObjectID `bson:"shop_id" json:"shopId"`
	EmailEnabled           bool               `bson:"email_enabled" json:"emailEnabled"`
	SMSEnabled             bool               `bson:"sms_enabled" json:"smsEnabled"`
	PushEnabled            bool               `bson:"push_enabled" json:"pushEnabled"`
	OrderNotifications     bool               `bson:"order_notifications" json:"orderNotifications"`
	PaymentNotifications   bool               `bson:"payment_notifications" json:"paymentNotifications"`
	InventoryNotifications bool               `bson:"inventory_notifications" json:"inventoryNotifications"`
	CustomerNotifications  bool               `bson:"customer_notifications" json:"customerNotifications"`
	AnalyticsNotifications bool               `bson:"analytics_notifications" json:"analyticsNotifications"`
	SystemNotifications    bool               `bson:"system_notifications" json:"systemNotifications"`
	CreatedAt              time.Time          `bson:"created_at" json:"createdAt"`
	ModifiedAt             time.Time          `bson:"modified_at" json:"modifiedAt"`
}

type ShopNotification struct {
	ID        primitive.ObjectID       `bson:"_id" json:"_id"`
	ShopID    primitive.ObjectID       `bson:"shop_id" json:"shopId"`
	Type      ShopNotificationType     `bson:"type" json:"type"`
	Title     string                   `bson:"title" json:"title"`
	Message   string                   `bson:"message" json:"message"`
	Priority  ShopNotificationPriority `bson:"priority" json:"priority"`
	IsRead    bool                     `bson:"is_read" json:"isRead"`
	Data      map[string]any           `bson:"data" json:"data"`
	CreatedAt time.Time                `bson:"created_at" json:"createdAt"`
	ReadAt    *time.Time               `bson:"read_at" json:"readAt"`
	ExpiresAt *time.Time               `bson:"expires_at" json:"expiresAt"`
}

type ShopNotificationRequest struct {
	Type     ShopNotificationType     `json:"type" validate:"required"`
	Title    string                   `json:"title" validate:"required"`
	Message  string                   `json:"message" validate:"required"`
	Priority ShopNotificationPriority `json:"priority" validate:"required"`
	Data     map[string]any           `json:"data"`
}

type UpdateShopNotificationSettingsRequest struct {
	ID                     primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	EmailEnabled           *bool              `json:"emailEnabled"`
	SMSEnabled             *bool              `json:"smsEnabled"`
	PushEnabled            *bool              `json:"pushEnabled"`
	OrderNotifications     *bool              `json:"orderNotifications"`
	PaymentNotifications   *bool              `json:"paymentNotifications"`
	InventoryNotifications *bool              `json:"inventoryNotifications"`
	CustomerNotifications  *bool              `json:"customerNotifications"`
	AnalyticsNotifications *bool              `json:"analyticsNotifications"`
	SystemNotifications    *bool              `json:"systemNotifications"`
}
