package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Shop struct {
	AnnouncementModifiedAt time.Time            `bson:"announcement_modified_at" json:"announcementModifiedAt" validate:"omitempty"`
	ModifiedAt             time.Time            `bson:"modified_at" json:"modifiedAt" validate:"required"`
	CreatedAt              time.Time            `bson:"created_at" json:"createdAt" validate:"required"`
	User                   ListingUserExcept    `bson:"user" json:"user"`
	Policy                 ShopPolicy           `bson:"policy" json:"policy" validate:"required"`
	BannerURL              string               `bson:"banner_url" json:"bannerUrl"`
	Status                 ShopStatus           `bson:"status" json:"status" validate:"required,oneof=inactive active banned suspended warning pendingreview"`
	Name                   string               `bson:"name" json:"name" validate:"required"`
	Announcement           string               `bson:"announcement" json:"announcement" validate:"omitempty"`
	SalesMessage           string               `bson:"sales_message" json:"salesMessage"`
	Description            string               `bson:"description" json:"description" validate:"required"`
	VacationMessage        string               `bson:"vacation_message" json:"vacationMessage" validate:"omitempty"`
	Slug                   string               `bson:"slug" json:"slug" validate:"required"`
	LogoURL                string               `bson:"logo_url" json:"logoUrl"`
	Username               string               `bson:"username" json:"username" validate:"required"`
	RecentReviews          []ShopReview         `bson:"recent_reviews" json:"recentReviews"`
	Categories             []ShopCategory       `bson:"-" json:"categories"`
	Followers              []ShopFollower       `bson:"followers" json:"followers"`
	Links                  []Link               `bson:"-" json:"links"`
	Gallery                []string             `bson:"gallery" json:"gallery"`
	About                  ShopAbout            `bson:"about" json:"about"`
	Address                UserAddress          `bson:"address" json:"address"`
	Rating                 Rating               `bson:"rating" json:"rating"`
	ReviewsCount           int                  `bson:"reviews_count" json:"reviewsCount"`
	FinancialInformation   FinancialInformation `bson:"financial_information" json:"financialInformation"`
	ListingActiveCount     int                  `bson:"listing_active_count" json:"listing_active_count" validate:"required"`
	FollowerCount          int                  `bson:"follower_count" json:"followerCount" validate:"required"`
	ID                     primitive.ObjectID   `bson:"_id" json:"_id" validate:"required"`
	UserID                 primitive.ObjectID   `bson:"user_id" json:"userId"`
	UserAddressId          primitive.ObjectID   `bson:"user_address_id" json:"userAddressId"`
	Location               primitive.ObjectID   `bson:"location" json:"location"`
	IsVacation             bool                 `bson:"is_vacation" json:"is_vacation"`
	IsLive                 bool                 `bson:"is_live" json:"isLive"`
}

type ShopExcerpt struct {
	ID                 primitive.ObjectID `bson:"_id"                   json:"id"`
	Name               string             `bson:"name"                  json:"name"`
	Slug               string             `bson:"slug"                  json:"slug"`
	Username           string             `bson:"username"              json:"username"`
	LogoURL            string             `bson:"logo_url"              json:"logoUrl"`
	BannerURL          string             `bson:"banner_url"            json:"bannerUrl"`
	Status             ShopStatus         `bson:"status"                json:"status"`
	CreatedAt          time.Time          `bson:"created_at"            json:"createdAt"`
	ListingActiveCount int                `bson:"listing_active_count"  json:"listingActiveCount"`
	FollowerCount      int                `bson:"follower_count"        json:"followerCount"`
	Rating             Rating             `bson:"rating"                json:"rating"`
	ReviewsCount       int                `bson:"reviews_count"         json:"reviewsCount"`
}

type ShopCategory struct {
	Name          string `bson:"name" json:"name"`
	Path          string `bson:"path" json:"path"`
	ListingsCount int    `bson:"count" json:"count"`
}

type ShopFollower struct {
	JoinedAt  time.Time          `bson:"joined_at" json:"joinedAt" validate:"required"`
	FirstName string             `bson:"first_name" json:"firstName" validate:"required"`
	LastName  string             `bson:"last_name" json:"lastName" validate:"required"`
	LoginName string             `bson:"login_name" json:"loginName" validate:"required"`
	Thumbnail string             `bson:"thumbnail" json:"thumbnail"`
	Id        primitive.ObjectID `bson:"_id" json:"_id"`
	UserId    primitive.ObjectID `bson:"user_id" json:"userId"`
	ShopId    primitive.ObjectID `bson:"shop_id" json:"shopId"`
	IsOwner   bool               `bson:"is_owner" json:"isOwner" validate:"required"`
}

type ShopFollowerExcerpt struct {
	FirstName string             `bson:"first_name" json:"firstName" validate:"required"`
	LastName  string             `bson:"last_name" json:"lastName" validate:"required"`
	LoginName string             `bson:"login_name" json:"loginName"`
	Thumbnail string             `bson:"thumbnail" json:"thumbnail"`
	Id        primitive.ObjectID `bson:"follower_id" json:"followerId"`
	UserId    primitive.ObjectID `bson:"user_id" json:"userId"`
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

type ShopAbout struct {
	Status    ShopAboutStatus    `bson:"status" json:"status" validate:"required"`
	Headline  string             `bson:"headline" json:"headline"`
	Story     string             `bson:"story" json:"story" validate:"required"`
	X         string             `bson:"x" json:"x" validate:"required"`
	Facebook  string             `bson:"facebook" json:"facebook" validate:"required"`
	Instagram string             `bson:"instagram" json:"instagram" validate:"required"`
	ID        primitive.ObjectID `bson:"_id" json:"_id" validate:"required"`
	ShopID    primitive.ObjectID `bson:"shop_id" json:"shopId" validate:"required"`
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
