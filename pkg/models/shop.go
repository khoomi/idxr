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
	UserID                 primitive.ObjectID `bson:"user_id" json:"userId"`
	UserAddressId          primitive.ObjectID `bson:"user_address_id" json:"userAddressId"`
	Location               primitive.ObjectID `bson:"location" json:"location"`
	ListingActiveCount     int                `bson:"listing_active_count" json:"listing_active_count" validate:"required"`
	Announcement           string             `bson:"announcement" json:"announcement" validate:"omitempty"`
	AnnouncementModifiedAt time.Time          `bson:"announcement_modified_at" json:"announcementModifiedAt" validate:"omitempty"`
	IsVacation             bool               `bson:"is_vacation" json:"is_vacation"`
	VacationMessage        string             `bson:"vacation_message" json:"vacationMessage" validate:"omitempty"`
	Slug                   string             `bson:"slug" json:"slug" validate:"required"`
	LogoURL                string             `bson:"logo_url" json:"logoUrl"`
	BannerURL              string             `bson:"banner_url" json:"bannerUrl"`
	Gallery                []string           `bson:"gallery" json:"gallery"`
	FollowerCount          int                `bson:"follower_count" json:"followerCount" validate:"required"`
	Followers              []ShopFollower     `bson:"followers" json:"followers"`
	Status                 ShopStatus         `bson:"status" json:"status" validate:"required,oneof=inactive active banned suspended warning pendingreview"`
	IsLive                 bool               `bson:"is_live" json:"isLive"`
	CreatedAt              time.Time          `bson:"created_at" json:"createdAt" validate:"required"`
	ModifiedAt             time.Time          `bson:"modified_at" json:"modifiedAt" validate:"required"`
	Policy                 ShopPolicy         `bson:"policy" json:"policy" validate:"required"`
	RecentReviews          []ShopReview       `bson:"recent_reviews" json:"recentReviews"`
	ReviewsCount           int                `bson:"reviews_count" json:"reviewsCount"`
	SalesMessage           string             `bson:"sales_message" json:"salesMessage"`
	User                   ListingUserExcept  `bson:"user" json:"user"`
	About                  ShopAbout          `bson:"about" json:"about"`
	Address                UserAddress        `bson:"address" json:"address"`
	Links                  []Link             `bson:"-" json:"links"`
	Categories             []ShopCategory     `bson:"-" json:"categories"`
}

type ShopCategory struct {
	Name          string `bson:"name" json:"name"`
	Path          string `bson:"path" json:"path"`
	ListingsCount int    `bson:"count" json:"count"`
}

type ShopFollower struct {
	Id        primitive.ObjectID `bson:"_id" json:"_id"`
	UserId    primitive.ObjectID `bson:"user_id" json:"userId"`
	ShopId    primitive.ObjectID `bson:"shop_id" json:"shopId"`
	FirstName string             `bson:"first_name" json:"firstName" validate:"required"`
	LastName  string             `bson:"last_name" json:"lastName" validate:"required"`
	LoginName string             `bson:"login_name" json:"loginName" validate:"required"`
	Thumbnail string             `bson:"thumbnail" json:"thumbnail"`
	IsOwner   bool               `bson:"is_owner" json:"isOwner" validate:"required"`
	JoinedAt  time.Time          `bson:"joined_at" json:"joinedAt" validate:"required"`
}

type ShopFollowerExcerpt struct {
	Id        primitive.ObjectID `bson:"follower_id" json:"followerId"`
	UserId    primitive.ObjectID `bson:"user_id" json:"userId"`
	FirstName string             `bson:"first_name" json:"firstName" validate:"required"`
	LastName  string             `bson:"last_name" json:"lastName" validate:"required"`
	LoginName string             `bson:"login_name" json:"loginName"`
	Thumbnail string             `bson:"thumbnail" json:"thumbnail"`
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
	Message    string `bson:"message" json:"message"`
	IsVacation bool   `bson:"is_vacation" json:"isVacation"`
}

type ShopReviewStatus string

const (
	ShopReviewStatusApproved ShopReviewStatus = "approved"
	ShopReviewStatusPending  ShopReviewStatus = "pending"
	ShopReviewStatusSpam     ShopReviewStatus = "spam"
)

type EmbeddedShopReview struct {
	UserId       primitive.ObjectID `bson:"user_id" json:"userId"`
	ShopId       primitive.ObjectID `bson:"shop_id" json:"shopId"`
	Review       string             `bson:"review" json:"review"`
	ReviewAuthor string             `bson:"review_author" json:"reviewAuthor"`
	Thumbnail    string             `bson:"thumbnail" json:"thumbnail"`
}

type ShopReview struct {
	Id           primitive.ObjectID `bson:"_id" json:"_id"`
	UserId       primitive.ObjectID `bson:"user_id" json:"userId"`
	ShopId       primitive.ObjectID `bson:"shop_id" json:"shopId"`
	Review       string             `bson:"review" json:"review"`
	ReviewAuthor string             `bson:"review_author" json:"reviewAuthor"`
	Thumbnail    string             `bson:"thumbnail" json:"thumbnail"`
	CreatedAt    time.Time          `bson:"created_at" json:"createdAt"`
	Status       ShopReviewStatus   `bson:"status" json:"status" validate:"required,oneof=approved pending spam"`
}

type ShopReviewRequest struct {
	Review string `bson:"review" json:"review" validate:"required"`
}

type ShopAbout struct {
	ID        primitive.ObjectID `bson:"_id" json:"_id" validate:"required"`
	ShopID    primitive.ObjectID `bson:"shop_id" json:"shopId" validate:"required"`
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
