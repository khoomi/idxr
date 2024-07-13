package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User Khoomi user_models basic data
type User struct {
	// Id uniquely identifies the user in the database.
	Id primitive.ObjectID `bson:"_id" json:"_id" validate:"required"`
	// LoginName represents the username used for signing in.
	LoginName string `bson:"login_name" json:"login_name" validate:"required"`
	// PrimaryEmail is the main email address associated with the user account.
	PrimaryEmail string `bson:"primary_email" json:"primary_email" validate:"required"`
	// FirstName of the user.
	FirstName string `bson:"first_name" json:"first_name"`
	// LastName of the user.
	LastName string `bson:"last_name" json:"last_name"`
	// Auth contains authentication data such as password hashes.
	Auth UserAuthData `bson:"auth,omitempty" json:"auth,omitempty" validate:"required"`
	// Thumbnail is the URL to the user's profile picture.
	Thumbnail string `bson:"thumbnail" json:"thumbnail"`
	// Bio is a short biography or description of the user.
	Bio string `bson:"bio" json:"bio"`
	// Phone number of the user.
	Phone string `bson:"phone" json:"phone"`
	// Birthdate represents the user's date of birth.
	Birthdate UserBirthdate `bson:"birthdate" json:"birthdate"`
	// IsSeller indicates whether the user has seller privileges.
	IsSeller bool `bson:"is_seller" json:"is_seller"`
	// TransactionBuyCount is the total number of purchases made by the user.
	TransactionBuyCount int `bson:"transaction_buy_count" json:"transaction_buy_count"`
	// TransactionSoldCount is the total number of sales made by the user.
	TransactionSoldCount int `bson:"transaction_sold_count" json:"transaction_sold_count"`
	// ReferredByUser indicates the user ID of the person who referred this user.
	ReferredByUser string `bson:"referred_by_user" json:"referred_by_user"`
	// Role defines the user's role within the platform (e.g., admin, regular user).
	Role UserRole `bson:"role" json:"role"`
	// Status indicates the current state of the user's account (e.g., active, suspended).
	Status UserStatus `bson:"status" json:"status"`
	// ShopID links the user to a specific shop if they are a seller.
	ShopID primitive.ObjectID `bson:"shop_id" json:"shop_id"`
	// FavoriteShops contains a list of shop IDs that the user has marked as favorite.
	FavoriteShops []string `bson:"favorite_shops" json:"favorite_shops"`
	// CreatedAt is the timestamp when the user account was created.
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	// ModifiedAt is the timestamp of the last modification to the user's account.
	ModifiedAt time.Time `bson:"modified_at" json:"modified_at"`
	// LastLogin is the timestamp of the user's last login.
	LastLogin time.Time `bson:"last_login" json:"last_login"`
	// LoginCounts tracks the number of times the user has logged into the account.
	LoginCounts int `bson:"login_counts" json:"-"`
	// LastLoginIp stores the IP address from which the user last accessed the account.
	LastLoginIp string `bson:"last_login_ip" json:"-"`
	// AllowLoginIpNotification indicates if the user opts in to receive notifications for new IP logins.
	AllowLoginIpNotification bool `bson:"allow_login_ip_notification" json:"allow_login_ip_notification"`
	// Links is a collection of hyperlinks related to the user, not stored in the database.
	Links []Link `bson:"-" json:"links"`
}

// UserRegistrationBody -> expected data for signup process
type UserRegistrationBody struct {
	FirstName string `json:"first_name,omitempty" validate:"required,min=3"`
	LastName  string `json:"last_name,omitempty" validate:"required,min=3"`
	Email     string `json:"email,omitempty" validate:"required,email"`
	Password  string `json:"password,omitempty" validate:"required"`
}

type NewPasswordRequest struct {
	CurrentPassword string `form:"current_password" validate:"required"`
	NewPassword     string `form:"new_password" validate:"required"`
}

type FirstLastName struct {
	FirstName string `bson:"first_name" json:"first_name" validate:"required"`
	LastName  string `bson:"last_name" json:"last_name" validate:"required"`
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
	EmailVerified  bool      `bson:"email_verified"`
	ModifiedAt     time.Time `bson:"modified_at"`
	PasswordDigest string    `bson:"password_digest,omitempty" json:"-"`
}

// UserLoginBody -> expected data for login process
type UserLoginBody struct {
	Email    string `json:"email,omitempty" validate:"required,email"`
	Password string `json:"password,omitempty" validate:"required"`
}

// LoginHistory -> User login history
type LoginHistory struct {
	Id        primitive.ObjectID `bson:"_id" json:"_id"`
	UserUid   primitive.ObjectID `bson:"user_uid" json:"user_uid"`
	Date      time.Time          `bson:"date" json:"date"`
	UserAgent string             `bson:"user_agent" json:"user_agent"`
	IpAddr    string             `bson:"ip_addr" json:"ip_addr"`
}

type LoginHistoryIds struct {
	IDs []string `json:"ids"`
}

type UserPasswordResetToken struct {
	UserId      primitive.ObjectID `bson:"user_uid" json:"user_uid"`
	TokenDigest string             `bson:"token_digest" json:"token_digest"`
	CreatedAt   primitive.DateTime `bson:"created_at" json:"created_at"`
	ExpiresAt   primitive.DateTime `bson:"expired_at" json:"expires_at"`
}

type UserVerifyEmailToken struct {
	UserId      primitive.ObjectID `bson:"user_uid" json:"user_uid"`
	TokenDigest string             `bson:"token_digest" json:"token_digest"`
	CreatedAt   primitive.DateTime `bson:"created_at" json:"created_at"`
	ExpiresAt   primitive.DateTime `bson:"expired_at" json:"expires_at"`
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
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	ListingId primitive.ObjectID `bson:"listing_id" json:"listing_id"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

type AccountDeletionRequested struct {
	ID     primitive.ObjectID `bson:"_id" json:"_id"`
	UserID primitive.ObjectID `bson:"user_id" json:"user_id"`
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
	NewMessage       bool `bson:"new_message" json:"new_message" validate:"required"`
	NewFollower      bool `bson:"new_follower" json:"new_follower" validate:"required"`
	ListingExpNotice bool `bson:"listing_exp_notice" json:"listing_exp_notice" validate:"required"`
	SellerActivity   bool `bson:"seller_activity" json:"seller_activity" validate:"required"`
	NewsAndFeatures  bool `bson:"news_and_features" json:"news_and_features" validate:"required"`
}
