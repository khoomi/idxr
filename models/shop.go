package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Shop struct {
	ID                     primitive.ObjectID `bson:"_id" json:"_id" validate:"required"`
	Name                   string             `bson:"name" json:"name" validate:"required"`
	Description            string             `bson:"description" json:"description" validate:"required"`
	Username               string             `bson:"username" json:"username" validate:"required"`
	UserID                 primitive.ObjectID `bson:"user_id" json:"user_id"`
	UserAddressId          primitive.ObjectID `bson:"user_address_id" json:"user_address_id"`
	Location               primitive.ObjectID `bson:"location" json:"location"`
	ListingActiveCount     int                `bson:"listing_active_count" json:"listing_active_count" validate:"required"`
	Announcement           string             `bson:"announcement" json:"announcement" validate:"omitempty"`
	AnnouncementModifiedAt time.Time          `bson:"announcement_modified_at" json:"announcement_modified_at" validate:"omitempty"`
	IsVacation             bool               `bson:"is_vacation" json:"is_vacation"`
	VacationMessage        string             `bson:"vacation_message" json:"vacation_message" validate:"omitempty"`
	Slug                   string             `bson:"slug" json:"slug" validate:"required"`
	LogoURL                string             `bson:"logo_url" json:"logo_url"`
	BannerURL              string             `bson:"banner_url" json:"banner_url"`
	Gallery                []string           `bson:"gallery" json:"gallery"`
	FollowerCount          int                `bson:"follower_count" json:"follower_count" validate:"required"`
	Followers              []ShopFollower     `bson:"followers" json:"followers"`
	Status                 ShopStatus         `bson:"status" json:"status" validate:"required,oneof=inactive active banned suspended warning pendingreview"`
	IsLive                 bool               `bson:"is_live" json:"is_live"`
	CreatedAt              time.Time          `bson:"created_at" json:"created_at" validate:"required"`
	ModifiedAt             time.Time          `bson:"modified_at" json:"modified_at" validate:"required"`
	Policy                 ShopPolicy         `bson:"policy" json:"policy" validate:"required"`
	RecentReviews          []ShopReview       `bson:"recent_reviews" json:"recent_reviews"`
	ReviewsCount           int                `bson:"reviews_count" json:"reviews_count"`
	SalesMessage           string             `bson:"sales_message" json:"sales_message"`
	User                   ListingUserExcept  `bson:"user" json:"user"`
	Address                UserAddress        `bson:"address" json:"address"`
}

type ShopFollower struct {
	Id        primitive.ObjectID `bson:"_id" json:"_id"`
	UserId    primitive.ObjectID `bson:"user_id" json:"user_id"`
	ShopId    primitive.ObjectID `bson:"shop_id" json:"shop_id"`
	FirstName string             `bson:"first_name" json:"first_name" validate:"required"`
	LastName  string             `bson:"last_name" json:"last_name" validate:"required"`
	LoginName string             `bson:"login_name" json:"login_name" validate:"required"`
	Thumbnail string             `bson:"thumbnail" json:"thumbnail"`
	IsOwner   bool               `bson:"is_owner" json:"is_owner" validate:"required"`
	JoinedAt  time.Time          `bson:"joined_at" json:"joined_at" validate:"required"`
}

type ShopFollowerExcerpt struct {
	Id        primitive.ObjectID `bson:"follower_id" json:"follower_id"`
	UserId    primitive.ObjectID `bson:"user_id" json:"user_id"`
	FirstName string             `bson:"first_name" json:"first_name" validate:"required"`
	LastName  string             `bson:"last_name" json:"last_name" validate:"required"`
	LoginName string             `bson:"login_name" json:"login_name"`
	Thumbnail string             `bson:"thumbnail" json:"thumbnail"`
	IsOwner   bool               `bson:"is_owner" json:"is_owner" validate:"required"`
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
	Review string `bson:"review" json:"review" validate:"required"`
}

type ShopAbout struct {
	ID        primitive.ObjectID `bson:"_id" json:"_id" validate:"required"`
	ShopID    primitive.ObjectID `bson:"shop_id" json:"shop_id" validate:"required"`
	Status    ShopAboutStatus    `bson:"status" json:"status" validate:"required"`
	Headline  string             `bson:"headline" json:"headline"`
	Story     string             `bson:"story" json:"story" validate:"required"`
	X         string             `bson:"x" json:"x" validate:"required"`
	Facebook  string             `bson:"facebook" json:"facebook" validate:"required"`
	Instagram string             `bson:"instagram" json:"instagram" validate:"required"`
}

type ShopAboutStatus string

const (
	ShopAboutStatusDraft  ShopAboutStatus = "draft"
	ShopAboutStatusActive ShopAboutStatus = "active"
)

type ShopAboutRequest struct {
	Status    ShopAboutStatus `bson:"status" json:"status" validate:"required,oneof=draft active"`
	Headline  string          `bson:"headline" json:"headline"`
	Story     string          `bson:"story" json:"story" validate:"required"`
	X         string          `bson:"x" json:"x" validate:"required"`
	Facebook  string          `bson:"facebook" json:"facebook" validate:"required"`
	Instagram string          `bson:"instagram" json:"instagram" validate:"required"`
}

type ShopReturnPolicies struct {
	ID               primitive.ObjectID `bson:"_id" json:"_id" validate:"omitempty"`
	ShopId           primitive.ObjectID `bson:"shop_id" json:"shop_id" validate:"omitempty"`
	AcceptsReturn    bool               `bson:"accepts_return" json:"accepts_return"`
	AcceptsExchanges bool               `bson:"accepts_exchanges" json:"accepts_exchanges"`
	Deadline         int                `bson:"deadline" json:"deadline" validate:"oneof=7 14 21 30 45 60 90"`
}

type UpdateShopStatusReq struct {
	Status bool `json:"status" validate:"required"`
}

type ComplianceInformation struct {
	ID                   primitive.ObjectID `bson:"_id" json:"_id"`
	ShopID               primitive.ObjectID `bson:"shop_id" json:"shop_id"`
	TermsOfUse           bool               `bson:"terms_of_use" json:"terms_of_use"`
	SellerPolicie        bool               `bson:"seller_policies" json:"seller_policies"`
	IntellectualProperty bool               `bson:"intellectual_property" json:"intellectual_property"`
}

type ComplianceInformationRequest struct {
	TermsOfUse           bool `bson:"terms_of_use" json:"terms_of_use"`
	SellerPolicie        bool `bson:"seller_policies" json:"seller_policies"`
	IntellectualProperty bool `bson:"intellectual_property" json:"intellectual_property"`
}
