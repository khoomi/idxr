package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User Khoomi user_models basic data
type User struct {
	Id           primitive.ObjectID `bson:"_id" json:"_id"`
	LoginName    string             `bson:"login_name" json:"login_name"`
	PrimaryEmail string             `bson:"primary_email" json:"primary_email"`
	FirstLastName
	Auth           UserAuthData       `bson:"auth,omitempty" json:"auth,omitempty"`
	Thumbnail      string             `bson:"thumbnail" json:"thumbnail"`
	ProfileUid     primitive.ObjectID `bson:"profile_uid" json:"profile_uid"`
	LoginCounts    int                `bson:"login_counts" json:"login_counts"`
	LastLogin      time.Time          `bson:"last_login" json:"last_login"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	ModifiedAt     time.Time          `bson:"modified_at" json:"modified_at"`
	ReferredByUser string             `bson:"referred_by_user" json:"referred_by_user"`
	Role           UserRole           `bson:"role" json:"role"`
	Status         UserStatus         `bson:"status" json:"status"`
	Shops          []string           `bson:"shops" json:"shops"`
	FavoriteShops  []string           `bson:"favorite_shops" json:"favorite_shops"`
}

type FirstLastName struct {
	FirstName string `bson:"first_name" json:"first_name" validate:"required"`
	LastName  string `bson:"last_name" json:"last_name" validate:"required""`
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
