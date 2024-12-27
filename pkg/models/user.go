package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User Khoomi user_models basic data
type User struct {
	LastLogin                time.Time          `bson:"last_login" json:"lastLogin"`
	ModifiedAt               time.Time          `bson:"modified_at" json:"modifiedAt"`
	CreatedAt                time.Time          `bson:"created_at" json:"created_at"`
	Auth                     UserAuthData       `bson:"auth,omitempty" json:"auth,omitempty" validate:"required"`
	Thumbnail                string             `bson:"thumbnail" json:"thumbnail"`
	LoginName                string             `bson:"login_name" json:"loginName" validate:"required"`
	LastLoginIp              string             `bson:"last_login_ip" json:"-"`
	Bio                      string             `bson:"bio" json:"bio"`
	Phone                    string             `bson:"phone" json:"phone"`
	LastName                 string             `bson:"last_name" json:"lastName"`
	PrimaryEmail             string             `bson:"primary_email" json:"PrimaryEmail" validate:"required"`
	FirstName                string             `bson:"first_name" json:"firstName"`
	Status                   UserStatus         `bson:"status" json:"status"`
	ReferredByUser           string             `bson:"referred_by_user" json:"referredByUser"`
	Role                     UserRole           `bson:"role" json:"role"`
	FavoriteShops            []string           `bson:"favorite_shops" json:"favoriteShops"`
	Links                    []Link             `bson:"-" json:"links"`
	Birthdate                UserBirthdate      `bson:"birthdate" json:"birthdate"`
	TransactionSoldCount     int                `bson:"transaction_sold_count" json:"transactionSoldCount"`
	TransactionBuyCount      int                `bson:"transaction_buy_count" json:"transactionBuyCount"`
	LoginCounts              int                `bson:"login_counts" json:"-"`
	ShopID                   primitive.ObjectID `bson:"shop_id" json:"shopId"`
	Id                       primitive.ObjectID `bson:"_id" json:"_id" validate:"required"`
	IsSeller                 bool               `bson:"is_seller" json:"isSeller"`
	AllowLoginIpNotification bool               `bson:"allow_login_ip_notification" json:"allowLoginIpNotification"`
}

// UserRegistrationBody -> expected data for signup process
type UserRegistrationBody struct {
	FirstName string `json:"firstName,omitempty" validate:"required,min=3"`
	LastName  string `json:"lastName,omitempty" validate:"required,min=3"`
	Email     string `json:"email,omitempty" validate:"required,email"`
	Password  string `json:"password,omitempty" validate:"required"`
}

type NewPasswordRequest struct {
	CurrentPassword string `form:"currentPassword" validate:"required"`
	NewPassword     string `form:"newPassword" validate:"required"`
}

type FirstLastName struct {
	FirstName string `bson:"firstName" json:"first_name" validate:"required"`
	LastName  string `bson:"lastName" json:"last_name" validate:"required"`
}

// UserRole -> contains different roles that can be assigned to users
type UserRole string

const (
	Super   UserRole = "Super"
	Mod     UserRole = "Mod"
	Regular UserRole = "User"
)

type UserStatus string

const (
	UserStatusActive UserStatus = "Active"
	Inactive         UserStatus = "Inactive"
	Suspended        UserStatus = "Suspended"
	Deleted          UserStatus = "Deleted"
	Banned           UserStatus = "Banned"
)

// UserAuthData -> authentication data
type UserAuthData struct {
	ModifiedAt     time.Time `bson:"modifiedAt"`
	PasswordDigest string    `bson:"passwordDigest,omitempty" json:"-"`
	EmailVerified  bool      `bson:"emailVerified"`
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
	// add more countries as needed
)

type UserBirthdate struct {
	Day   int `bson:"day" json:"day" validate:"required"`
	Month int `bson:"month" json:"month" validate:"required"`
	Year  int `bson:"year" json:"year" validate:"required"`
}

type UserWishlist struct {
	ID        primitive.ObjectID `bson:"_id" json:"_id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"userId"`
	ListingId primitive.ObjectID `bson:"listing_id" json:"listingId"`
	CreatedAt time.Time          `bson:"created_at" json:"createdAt"`
}

type AccountDeletionRequested struct {
	ID     primitive.ObjectID `bson:"_id" json:"_id"`
	UserID primitive.ObjectID `bson:"user_id" json:"userId"`
}

type RefreshTokenPayload struct {
	Token string `json:"token" validate:"required"`
}

// USER NOTIFICATION
type Notification struct {
	// ID of the shop.
	ID               primitive.ObjectID `bson:"_id" json:"_id" validate:"required"`
	UserID           primitive.ObjectID `bson:"user_id" json:"userId" validate:"required"`
	NewMessage       bool               `bson:"new_message" json:"newMessage" validate:"required"`
	NewFollower      bool               `bson:"new_follower" json:"newFollower" validate:"required"`
	ListingExpNotice bool               `bson:"listing_exp_notice" json:"listingExpNotice" validate:"required"`
	SellerActivity   bool               `bson:"seller_activity" json:"sellerActivity" validate:"required"`
	NewsAndFeatures  bool               `bson:"news_and_features" json:"newsAndFeatures" validate:"required"`
}

type NotificationRequest struct {
	// ID of the shop.
	NewMessage       bool `bson:"new_message" json:"newMessage" validate:"required"`
	NewFollower      bool `bson:"new_follower" json:"newFollower" validate:"required"`
	ListingExpNotice bool `bson:"listing_exp_notice" json:"listingExpNotice" validate:"required"`
	SellerActivity   bool `bson:"seller_activity" json:"sellerActivity" validate:"required"`
	NewsAndFeatures  bool `bson:"news_and_features" json:"newsAndFeatures" validate:"required"`
}
