package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserRole string

const (
	Super   UserRole = "Super"
	Mod     UserRole = "Mod"
	Regular UserRole = "User"
)

type Gender string

const (
	Male           Gender = "male"
	Female         Gender = "female"
	Other          Gender = "Other"
	PreferNotToSay Gender = "PreferNotToSay"
)

type UserStatus string

const (
	UserStatusActive UserStatus = "Active"
	Inactive         UserStatus = "Inactive"
	Suspended        UserStatus = "Suspended"
	Deleted          UserStatus = "Deleted"
	Banned           UserStatus = "Banned"
)

// Khoomi User user_models basic data
type User struct {
	LastLogin                time.Time             `bson:"last_login" json:"lastLogin"`
	ModifiedAt               time.Time             `bson:"modified_at" json:"modifiedAt"`
	CreatedAt                time.Time             `bson:"created_at" json:"createdAt"`
	Auth                     UserAuthData          `bson:"auth,omitempty" json:"auth,omitempty" validate:"required"`
	Thumbnail                string                `bson:"thumbnail" json:"thumbnail"`
	LoginName                string                `bson:"login_name" json:"loginName" validate:"required"`
	LastLoginIp              string                `bson:"last_login_ip" json:"-"`
	Bio                      string                `bson:"bio" json:"bio"`
	Phone                    string                `bson:"phone" json:"phone"`
	LastName                 string                `bson:"last_name" json:"lastName"`
	PrimaryEmail             string                `bson:"primary_email" json:"primaryEmail" validate:"required"`
	FirstName                string                `bson:"first_name" json:"firstName"`
	Gender                   Gender                `bson:"gender" json:"gender"`
	Status                   UserStatus            `bson:"status" json:"status"`
	ReferredByUser           string                `bson:"referred_by_user" json:"referredByUser"`
	Role                     UserRole              `bson:"role" json:"role"`
	FavoriteShops            []string              `bson:"favorite_shops" json:"favoriteShops"`
	FavoriteListings         []string              `bson:"favorite_listings" json:"favoriteListings"`
	Links                    []Link                `bson:"-" json:"links"`
	Birthdate                *time.Time            `bson:"birthdate,omitempty" json:"birthdate,omitempty"`
	TransactionSoldCount     int                   `bson:"transaction_sold_count" json:"transactionSoldCount"`
	TransactionBuyCount      int                   `bson:"transaction_buy_count" json:"transactionBuyCount"`
	LoginCounts              int                   `bson:"login_counts" json:"-"`
	ShopID                   primitive.ObjectID    `bson:"shop_id" json:"shopId"`
	Id                       primitive.ObjectID    `bson:"_id" json:"_id" validate:"required"`
	IsSeller                 bool                  `bson:"is_seller" json:"isSeller"`
	AllowLoginIpNotification bool                  `bson:"allow_login_ip_notification" json:"allowLoginIpNotification"`
	ReviewCount              int                   `bson:"review_count" json:"reviewCount"`
	Shop                     *ShopExcerpt          `bson:"shop" json:"shop"`
	SellerOnboardingLevel    SellerOnboardingLevel `bson:"seller_onboarding_level" json:"sellerOnboardingLevel"`
}

type SellerOnboardingLevel string

const (
	OnboardingLevelBuyer        SellerOnboardingLevel = "buyer"
	OnboardingLevelCreatedShop  SellerOnboardingLevel = "created_shop"
	OnboardingLevelListing      SellerOnboardingLevel = "listing"
	OnboardingLevelPayment      SellerOnboardingLevel = "payment"
	OnboardingLevelShipping     SellerOnboardingLevel = "shipping"
	OnboardingLevelCompliance   SellerOnboardingLevel = "compliance"
	OnboardingLevelVerification SellerOnboardingLevel = "verification"
)

type FirstLastName struct {
	FirstName string `bson:"firstName" json:"first_name" validate:"required"`
	LastName  string `bson:"lastName" json:"last_name" validate:"required"`
}

type CreateUserRequest struct {
	FirstName string `json:"firstName,omitempty" validate:"required,min=3"`
	LastName  string `json:"lastName,omitempty"`
	Email     string `json:"email,omitempty" validate:"required,email"`
	Password  string `json:"password,omitempty" validate:"required"`
}

type UpdateUserProfileRequest struct {
	FirstName string     `json:"firstName,omitempty" validate:"required,min=3"`
	LastName  string     `json:"lastName,omitempty"`
	Email     string     `json:"email,omitempty" validate:"required,email"`
	Gender    Gender     `json:"gender,omitempty"`
	Dob       *time.Time `json:"dob,omitempty"`
	Phone     string     `json:"phone,omitempty"`
}

type PasswordChangeRequest struct {
	CurrentPassword string `form:"currentPassword" validate:"required"`
	NewPassword     string `form:"newPassword" validate:"required"`
}

type UserAuthRequest struct {
	Email    string
	Password string
}

// UserAuthData -> authentication data
type UserAuthData struct {
	ModifiedAt         time.Time `bson:"modified_at" json:"modifiedAt"`
	PasswordDigest     string    `bson:"password_digest,omitempty" json:"-"`
	EmailVerified      bool      `bson:"email_verified" json:"emailVerified"`
	AuthenticationType bool      `bson:"authentication_type" json:"authenticationType"`
}

// UserLoginBody -> expected data for login process
type UserLoginBody struct {
	Email    string `json:"email,omitempty" validate:"required,email"`
	Password string `json:"password,omitempty" validate:"required"`
}

// LoginHistory -> User login history
type LoginHistory struct {
	Date      time.Time          `bson:"date" json:"date"`
	UserAgent string             `bson:"user_agent" json:"userAgent"`
	IpAddr    string             `bson:"ip_addr" json:"ipAddr"`
	Id        primitive.ObjectID `bson:"_id" json:"_id"`
	UserUid   primitive.ObjectID `bson:"user_uid" json:"userId"`
}

type LoginHistoryIds struct {
	IDs []string `json:"ids"`
}

type UserPasswordResetToken struct {
	TokenDigest string             `bson:"token_digest" json:"tokenDigest"`
	CreatedAt   primitive.DateTime `bson:"created_at" json:"createdAt"`
	ExpiresAt   primitive.DateTime `bson:"expired_at" json:"expiresAt"`
	UserId      primitive.ObjectID `bson:"user_uid" json:"userId"`
}

type UserVerifyEmailToken struct {
	TokenDigest string             `bson:"token_digest" json:"tokenDigest"`
	CreatedAt   primitive.DateTime `bson:"created_at" json:"createdAt"`
	ExpiresAt   primitive.DateTime `bson:"expired_at" json:"expiresAt"`
	UserId      primitive.ObjectID `bson:"user_uid" json:"userId"`
}

type Country string

const (
	CountryNigeria Country = "Nigeria"
)

type UserWishlist struct {
	CreatedAt time.Time          `bson:"created_at" json:"createdAt"`
	ID        primitive.ObjectID `bson:"_id" json:"_id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"userId"`
	ListingId primitive.ObjectID `bson:"listing_id" json:"listingId"`
}

type AccountDeletionRequested struct {
	ID     primitive.ObjectID `bson:"_id" json:"_id"`
	UserID primitive.ObjectID `bson:"user_id" json:"userId"`
}

type RefreshTokenPayload struct {
	Token string `json:"token" validate:"required"`
}

type UserNotificationSettings struct {
	ID                   primitive.ObjectID `bson:"_id" json:"_id" validate:"required"`
	UserID               primitive.ObjectID `bson:"user_id" json:"userId" validate:"required"`
	EmailEnabled         bool               `bson:"email_enabled" json:"emailEnabled"`
	SMSEnabled           bool               `bson:"sms_enabled" json:"smsEnabled"`
	PushEnabled          bool               `bson:"push_enabled" json:"pushEnabled"`
	Promotional          bool               `bson:"promotional" json:"promotional" validate:"required"`
	SupportMessage       bool               `bson:"support_message" json:"supportMessage" validate:"required"`
	NewMessage           bool               `bson:"new_message" json:"newMessage" validate:"required"`
	NewFollower          bool               `bson:"new_follower" json:"newFollower" validate:"required"`
	NewsAndFeatures      bool               `bson:"news_and_features" json:"newsAndFeatures" validate:"required"`
	OrderUpdates         bool               `bson:"order_updates" json:"orderUpdates" validate:"required"`
	PaymentConfirmations bool               `bson:"payment_confirmations" json:"paymentConfirmations" validate:"required"`
	DeliveryUpdates      bool               `bson:"delivery_updates" json:"deliveryUpdates" validate:"required"`
	CreatedAt            time.Time          `bson:"created_at" json:"createdAt"`
	ModifiedAt           time.Time          `bson:"modified_at" json:"modifiedAt"`
}

type UserNotificationSettingsRequest struct {
	EmailEnabled         bool `bson:"email_enabled" json:"emailEnabled"`
	SMSEnabled           bool `bson:"sms_enabled" json:"smsEnabled"`
	PushEnabled          bool `bson:"push_enabled" json:"pushEnabled"`
	NewMessage           bool `bson:"new_message" json:"newMessage" validate:"required"`
	NewFollower          bool `bson:"new_follower" json:"newFollower" validate:"required"`
	NewsAndFeatures      bool `bson:"news_and_features" json:"newsAndFeatures" validate:"required"`
	OrderUpdates         bool `bson:"order_updates" json:"orderUpdates" validate:"required"`
	PaymentConfirmations bool `bson:"paymentConfirmations" json:"paymentConfirmations" validate:"required"`
	DeliveryUpdates      bool `bson:"deliveryUpdates" json:"deliveryUpdates" validate:"required"`
	Promotional          bool `bson:"promotional" json:"promotional" validate:"required"`
	SupportMessage       bool `bson:"support_message" json:"supportMessage" validate:"required"`
}
