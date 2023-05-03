package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User Khoomi user_models basic data
type User struct {
	Id                   primitive.ObjectID `bson:"_id" json:"_id" validate:"required"`
	LoginName            string             `bson:"login_name" json:"login_name" validate:"required"`
	PrimaryEmail         string             `bson:"primary_email" json:"primary_email" validate:"required"`
	FirstName            string             `bson:"first_name" json:"first_name"`
	LastName             string             `bson:"last_name" json:"last_name"`
	Auth                 UserAuthData       `bson:"auth,omitempty" json:"auth,omitempty" validate:"required"`
	Thumbnail            string             `bson:"thumbnail" json:"thumbnail"`
	Bio                  string             `bson:"bio" json:"bio"`
	Phone                string             `bson:"phone" json:"phone"`
	Birthdate            UserBirthday       `bson:"birthdate" json:"birthdate"`
	IsSeller             bool               `bson:"is_seller" json:"is_seller"`
	TransactionBuyCount  int                `bson:"transaction_buy_count" json:"transaction_buy_count"`
	TransactionSoldCount int                `bson:"transaction_sold_count" json:"transaction_sold_count"`
	AddressId            primitive.ObjectID `bson:"address_id" json:"address_id"`
	ReferredByUser       string             `bson:"referred_by_user" json:"referred_by_user"`
	Role                 UserRole           `bson:"role" json:"role"`
	Status               UserStatus         `bson:"status" json:"status"`
	Shops                []string           `bson:"shops" json:"shops"`
	FavoriteShops        []string           `bson:"favorite_shops" json:"favorite_shops"`
	CreatedAt            time.Time          `bson:"created_at" json:"created_at" validate:"required"`
	ModifiedAt           time.Time          `bson:"modified_at" json:"modified_at" validate:"required"`
	LastLogin            time.Time          `bson:"last_login" json:"last_login"`
	LoginCounts          int                `bson:"login_counts" json:"login_counts" validate:"required"`
	LastLoginIp          string             `bson:"last_login_ip" json:"last_login_ip"`
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
	Active    UserStatus = "Active"
	Inactive  UserStatus = "Inactive"
	Suspended UserStatus = "Suspended"
	Deleted   UserStatus = "Deleted"
	Banned    UserStatus = "Banned"
)

// UserAuthData -> authentication data
type UserAuthData struct {
	EmailVerified  bool      `bson:"email_verified"`
	ModifiedAt     time.Time `bson:"modified_at"`
	PasswordDigest string    `bson:"password_digest,omitempty" json:"password_digest,omitempty"`
}

// UserRegistrationBody -> expected data for signup process
type UserRegistrationBody struct {
	LoginName string `json:"login_name,omitempty" validate:"required,min=6"`
	Email     string `json:"email,omitempty" validate:"required,email"`
	Password  string `json:"password,omitempty" validate:"required"`
}

// UserLoginBody -> expected data for login process
type UserLoginBody struct {
	Email    string `json:"email,omitempty" validate:"required,email"`
	Password string `json:"password,omitempty" validate:"required"`
}

// LoginHistory -> User login  history
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

type UserAddress struct {
	Id         primitive.ObjectID `bson:"_id" json:"_id"`
	City       string             `bson:"city" json:"city" validate:"required"`
	State      string             `bson:"state" json:"street_address" validate:"required"`
	Street     string             `bson:"street" json:"street" validate:"required"`
	PostalCode string             `bson:"postal_code" json:"postal_code" validate:"required"`
	Country    Country            `bson:"country" json:"country" validate:"required"`
	UserId     primitive.ObjectID `bson:"user_id" json:"user_id"`
}

type Country string

const (
	CountryNigeria Country = "Nigeria"
	// add more countries as needed
)

type UserBirthday struct {
	Day   int `bson:"day" json:"day"`
	Month int `bson:"month" json:"month"`
	Year  int `bson:"year" json:"year"`
}
