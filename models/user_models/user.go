package user_models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User Khoomi user_models basic data
type User struct {
	Id             primitive.ObjectID `bson:"_id"`
	LoginName      string             `bson:"login_name"`
	PrimaryEmail   string             `bson:"primary_email"`
	FirstName      string             `bson:"first_name"`
	LastName       string             `bson:"last_name"`
	Auth           UserAuthData       `bson:"auth"`
	Thumbnail      string             `bson:"thumbnail"`
	ProfileUid     primitive.ObjectID `bson:"profile_uid"`
	LoginCounts    int                `bson:"login_counts"`
	LastLogin      time.Time          `bson:"last_login"`
	CreatedAt      time.Time          `bson:"created_at"`
	ModifiedAt     time.Time          `bson:"modified_at"`
	ReferredByUser string             `bson:"referred_by_user"`
	Role           UserRole           `bson:"role"`
	Status         UserStatus         `bson:"status"`
	Shops          []string           `bson:"shops"`
	FavoriteShops  []string           `bson:"favorite_shops"`
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
	PasswordDigest string    `bson:"password_digest"`
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
