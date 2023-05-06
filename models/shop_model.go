package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Shop struct {
	ID                 primitive.ObjectID `bson:"_id" json:"_id" validate:"required"`
	ShopName           string             `bson:"shop_name" json:"shop_name" validate:"required,pattern=^\p{L}+[\p{L}\p{Pd}\p{Zs}']*\p{L}+$|^\p{L}+$"`
	Description        string             `bson:"description" json:"description" validate:"required"`
	LoginName          string             `bson:"login_name" json:"login_name" validate:"required"`
	UserID             primitive.ObjectID `bson:"user_id" json:"user_id" validate:""`
	ListingActiveCount int                `bson:"listing_active_count" json:"listing_active_count" validate:"required"`
	Announcement       string             `bson:"announcement" json:"announcement" validate:"omitempty"`
	IsVacation         bool               `bson:"is_vacation" json:"is_vacation"`
	VacationMessage    string             `bson:"vacation_message" json:"vacation_message" validate:"omitempty"`
	Slug               string             `bson:"slug" json:"slug" validate:"required"`
	LogoURL            string             `bson:"logo_url" json:"logo_url"`
	BannerURL          string             `bson:"banner_url" json:"banner_url"`
	Gallery            []string           `bson:"gallery" json:"gallery" `
	Favorers           []string           `bson:"favorers" json:"favorers" `
	FavorerCount       int                `bson:"favorer_count" json:"favorer_count" validate:"required"`
	Members            []ShopMember       `bson:"members" json:"members" `
	Status             ShopStatus         `bson:"status" json:"status" validate:"required,oneof=inactive active banned suspended warning pendingreview"`
	CreatedAt          time.Time          `bson:"created_at" json:"created_at" validate:"required"`
	ModifiedAt         time.Time          `bson:"modified_at" json:"modified_at" validate:"required"`
	Policy             ShopPolicy         `bson:"policy" json:"policy" validate:"required"`
	RecentReviews      []ShopReview       `bson:"recent_reviews" json:"recent_reviews" validate:"required"`
}

type ShopMember struct {
	Id        primitive.ObjectID `bson:"_id" json:"_id"`
	MemberId  primitive.ObjectID `bson:"member_id" json:"member_id"`
	ShopId    primitive.ObjectID `bson:"shop_id" json:"shop_id"`
	LoginName string             `bson:"login_name" json:"login_name" validate:"required"`
	Thumbnail string             `bson:"thumbnail" json:"thumbnail"`
	IsOwner   bool               `bson:"is_owner" json:"is_owner" validate:"required"`
	OwnerId   primitive.ObjectID `bson:"owner_id" json:"owner_id"`
	JoinedAt  time.Time          `bson:"joined_at" json:"joined_at" validate:"required"`
}

type ShopMemberEmbedded struct {
	MemberId  primitive.ObjectID `bson:"member_id" json:"member_id"`
	LoginName string             `bson:"login_name" json:"login_name"`
	Thumbnail string             `bson:"thumbnail" json:"thumbnail"`
	IsOwner   bool               `bson:"is_owner" json:"is_owner"`
}

type ShopMemberFromRequest struct {
	MemberId  string `bson:"member_id" json:"member_id"`
	Thumbnail string `bson:"thumbnail" json:"thumbnail"`
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
	PaymentPolicy  string `bson:"payment_policy" json:"payment_policy"`
	ShippingPolicy string `bson:"shipping_policy" json:"shipping_policy"`
	RefundPolicy   string `bson:"refund_policy" json:"refund_policy"`
	AdditionalInfo string `bson:"additional_info" json:"additional_info"`
}

type NewShopRequest struct {
	ShopName    string `bson:"shop_name" json:"shop_name" validate:"required,pattern=^(?!s)(?!.*s$)(?=.*[a-zA-Z0-9])[a-zA-Z0-9 '~?!]{2,}$"`
	Description string `bson:"description" json:"description" validate:"required"`
}

type ShopAnnouncementRequest struct {
	Announcement string `bson:"announcement" json:"announcement"`
}

type ShopVacationRequest struct {
	Message    string `bson:"message" json:"message"`
	IsVacation bool   `bson:"is_vacation" json:"is_vacation"`
}

type ShopReviewStatus string

const (
	ShopReviewStatusApproved ShopReviewStatus = "approved"
	ShopReviewStatusPending  ShopReviewStatus = "pending"
	ShopReviewStatusSpam     ShopReviewStatus = "spam"
)

type EmbeddedShopReview struct {
	UserId       primitive.ObjectID `bson:"user_id" json:"user_id"`
	ShopId       primitive.ObjectID `bson:"shop_id" json:"shop_id"`
	Review       string             `bson:"review" json:"review"`
	ReviewAuthor string             `bson:"review_author" json:"review_author"`
	Thumbnail    string             `bson:"thumbnail" json:"thumbnail"`
}
type ShopReview struct {
	Id           primitive.ObjectID `bson:"_id" json:"_id"`
	UserId       primitive.ObjectID `bson:"user_id" json:"user_id"`
	ShopId       primitive.ObjectID `bson:"shop_id" json:"shop_id"`
	Review       string             `bson:"review" json:"review"`
	ReviewAuthor string             `bson:"review_author" json:"review_author"`
	Thumbnail    string             `bson:"thumbnail" json:"thumbnail"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	Status       ShopReviewStatus   `bson:"status" json:"status" validate:"required,oneof=approved pending spam"`
}

type ShopReviewRequest struct {
	Review string `bson:"review" json:"review" validate:"required,pattern=^[A-Za-z][^\.:]*[\.:]$"`
}
