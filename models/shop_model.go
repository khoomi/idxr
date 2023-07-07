package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Shop struct {
	// ID of the shop.
	ID primitive.ObjectID `bson:"_id" json:"_id" validate:"required"`
	// The name of the shop. and must follow the pattern specified.
	Name string `bson:"name" json:"name" validate:"required"`
	// Description of the shop.
	Description string `bson:"description" json:"description" validate:"required"`
	// The login name for the shop.
	Username string `bson:"username" json:"username" validate:"required"`
	// ID of the user that owns the shop.
	UserID primitive.ObjectID `bson:"user_id" json:"user_id" validate:""`
	// Number of active listings in the shop.
	ListingActiveCount int `bson:"listing_active_count" json:"listing_active_count" validate:"required"`
	// Announcement fo the shop. This field is optional.
	Announcement string `bson:"announcement" json:"announcement" validate:"omitempty"`
	// Indicates whether the shop is on vacation.
	IsVacation bool `bson:"is_vacation" json:"is_vacation" validate:"default=false"`
	// Message displayed when the shop is on vacation. This field is optional.
	VacationMessage string `bson:"vacation_message" json:"vacation_message" validate:"omitempty"`
	// Slug for the shop.
	Slug string `bson:"slug" json:"slug" validate:"required"`
	// URL for the shop's logo. This field is optional.
	LogoURL string `bson:"logo_url" json:"logo_url"`
	// URL for the shop's banner. This field is optional.
	BannerURL string `bson:"banner_url" json:"banner_url"`
	// List of image URLs for the shop's gallery. This field is optional.
	Gallery []string `bson:"gallery" json:"gallery"`
	// List of user IDs that have favorited the shop. This field is optional.
	Favorers []string `bson:"favorers" json:"favorers"`
	// Number of users that have favorited the shop.
	FavorerCount int `bson:"favorer_count" json:"favorer_count" validate:"required"`
	// List of members of the shop.
	Members []ShopMember `bson:"members" json:"members"`
	// Status of the shop. and must be one of the specified values.
	Status ShopStatus `bson:"status" json:"status" validate:"required,oneof=inactive active banned suspended warning pendingreview"`
	// Date and time when the shop was created.
	CreatedAt time.Time `bson:"created_at" json:"created_at" validate:"required"`
	// Date and time when the shop was last modified.
	ModifiedAt time.Time `bson:"modified_at" json:"modified_at" validate:"required"`
	// Policy for the shop.
	Policy ShopPolicy `bson:"policy" json:"policy" validate:"required"`
	// List of recent reviews for the shop.
	RecentReviews []ShopReview `bson:"recent_reviews" json:"recent_reviews"`
	// Numbers of reviews for the shop.
	ReviewsCount int `bson:"reviews_count" json:"reviews_count"`
	// A message string sent to users who complete a purchase from this shop.
	SalesMessage              []ShopReview `bson:"sales_message" json:"sales_message" validate:""`
	IsDirectCheckoutOnboarded bool         `bson:"is_direct_checkout_onboarded" json:"is_direct_checkout_onboarded" validate:"default=false"`
	IsKhoomiPaymentOnboarded  bool         `bson:"is_khoomi_payment_onboarded" json:"is_khoomi_payment_onboarded" validate:"default=false"`
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
	Review string `bson:"review" json:"review" validate:"required"`
}

type ShopAbout struct {
	ID                    primitive.ObjectID `bson:"_id" json:"_id" validate:"required"`
	ShopID                primitive.ObjectID `bson:"shop_id" json:"shop_id" validate:"required"`
	Status                ShopAboutStatus    `bson:"status" json:"status" validate:"required"`
	RelatedLinks          string             `bson:"related_links" json:"related_links"`
	StoryLeadingParagraph string             `bson:"story_leading_paragraph" json:"story_leading_paragraph"`
	StoryHeadline         string             `bson:"story_headline" json:"story_headline"`
}

type ShopAboutStatus string

//const (
//	ShopAboutStatusDraft  ShopAboutStatus = "draft"
//	ShopAboutStatusActive ShopAboutStatus = "active"
//)

type ShopAboutRequest struct {
	Status                ShopAboutStatus `bson:"status" json:"status" validate:"required,oneof=draft active"`
	RelatedLinks          string          `bson:"related_links" json:"related_links"`
	StoryLeadingParagraph string          `bson:"story_leading_paragraph" json:"story_leading_paragraph" validate:"required"`
	StoryHeadline         string          `bson:"story_headline" json:"story_headline" validate:"required"`
}

type ShopReturnPolicies struct {
	ID               primitive.ObjectID `bson:"_id" json:"_id" validate:"omitempty"`
	ShopId           primitive.ObjectID `bson:"shop_id" json:"shop_id" validate:"omitempty"`
	AcceptsReturn    bool               `bson:"accepts_return" json:"accepts_return"  validate:"default=false"`
	AcceptsExchanges bool               `bson:"accepts_exchanges" json:"accepts_exchanges"  validate:"default=false"`
	Deadline         int                `bson:"deadline" json:"deadline" validate:"oneof=7 14 21 30 45 60 90, default=7"`
}
