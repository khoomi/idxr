package models

import (
	"fmt"
	"mime/multipart"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type File struct {
	File multipart.File `json:"file,omitempty" validate:"required"`
}

type Url struct {
	Url string `json:"url,omitempty" validate:"required"`
}

type Link struct {
	Href string `json:"href" validate:"required"`
	Rel  string `json:"rel" validate:"required"`
}

type UserSession struct {
	RefreshToken string             `bson:"refreshToken" json:"refreshToken"`
	UserAgent    string             `bson:"useragent" json:"userAgent"`
	UserIP       string             `bson:"userip" json:"userip"`
	ExpiresAt    primitive.DateTime `bson:"expires_at" json:"expiresAt"`
	ID           primitive.ObjectID `bson:"_id" json:"_id"`
	UserID       primitive.ObjectID `bson:"user_id" json:"userId"`
	IsBlocked    bool               `bson:"is_blocked" json:"isBlocked"`
	CreatedAt    bool               `bson:"created_at" json:"createdAt"`
}

type Rating struct {
	AverageRating  float64 `bson:"average_rating" json:"averageRating"`
	ReviewCount    int     `bson:"review_count" json:"reviewCount"`
	FiveStarCount  int     `bson:"five_star_count" json:"fiveStarCount"`
	FourStarCount  int     `bson:"four_star_count" json:"fourStarCount"`
	ThreeStarCount int     `bson:"three_star_count" json:"threeStarCount"`
	TwoStarCount   int     `bson:"two_star_count" json:"twoStarCount"`
	OneStarCount   int     `bson:"one_star_count" json:"oneStarCount"`
}

type ReviewRequest struct {
	Review string `bson:"review" json:"review" validate:"required"`
	Rating int    `bson:"rating" json:"rating" validate:"required,min=1,max=5"`
}

type ReviewStatus string

const (
	ReviewStatusApproved ReviewStatus = "approved"
	ReviewStatusPending  ReviewStatus = "pending"
	ReviewStatusSpam     ReviewStatus = "spam"
)

type EmbeddedReview struct {
	Id           primitive.ObjectID `json:"Id"`
	Review       string             `json:"review"`
	ReviewAuthor string             `json:"reviewAuthor"`
	Thumbnail    string             `json:"thumbnail"`
	Rating       int                `json:"rating"`
	UserId       primitive.ObjectID `json:"userId"`
	DataId       primitive.ObjectID `json:"dataId"`
}

type ListingReview struct {
	CreatedAt    time.Time          `bson:"created_at" json:"createdAt"`
	Review       string             `bson:"review" json:"review"`
	ReviewAuthor string             `bson:"review_author" json:"reviewAuthor"`
	Thumbnail    string             `bson:"thumbnail" json:"thumbnail"`
	Status       ReviewStatus       `bson:"status" json:"status" validate:"required,oneof=approved pending spam"`
	Rating       int                `bson:"rating" json:"rating" validate:"required,min=1,max=5"`
	Id           primitive.ObjectID `bson:"_id" json:"_id"`
	UserId       primitive.ObjectID `bson:"user_id" json:"userId"`
	ListingId    primitive.ObjectID `bson:"listing_id" json:"listingId"`
}

type ShopReview struct {
	CreatedAt    time.Time          `bson:"created_at" json:"createdAt"`
	Review       string             `bson:"review" json:"review"`
	ReviewAuthor string             `bson:"review_author" json:"reviewAuthor"`
	Thumbnail    string             `bson:"thumbnail" json:"thumbnail"`
	Status       ReviewStatus       `bson:"status" json:"status" validate:"required,oneof=approved pending spam"`
	Rating       int                `bson:"rating" json:"rating" validate:"required,min=1,max=5"`
	Id           primitive.ObjectID `bson:"_id" json:"_id"`
	UserId       primitive.ObjectID `bson:"user_id" json:"userId"`
	ShopId       primitive.ObjectID `bson:"shop_id" json:"shopId"`
}

func GenLink(rel, href string) Link {
	return Link{Href: href, Rel: rel}
}

func (user *User) ConstructUserLinks() {
	tempUser := user
	_id := tempUser.Id.Hex()
	// self
	user.Links = append(tempUser.Links, GenLink("self", fmt.Sprintf("users/%v", _id)))
	if !tempUser.ShopID.IsZero() {
		// shops
		user.Links = append(tempUser.Links, GenLink("shops", fmt.Sprintf("shops/%v", tempUser.ShopID.Hex())))
	}
	// deletion
	user.Links = append(tempUser.Links, GenLink("deletion", fmt.Sprintf("users/%v/deletion", _id)))
	// notification-setting
	user.Links = append(tempUser.Links, GenLink("notificationSettings", fmt.Sprintf("users/%v/notification-settings", _id)))
	// thumbnail
	user.Links = append(tempUser.Links, GenLink("thumbnail", fmt.Sprintf("users/%v/thumbnail", _id)))
	// addresses
	user.Links = append(tempUser.Links, GenLink("addresses", fmt.Sprintf("users/%v/addresses", _id)))
	// birthdate
	user.Links = append(tempUser.Links, GenLink("birthdate", fmt.Sprintf("users/%v/birthdate", _id)))
	// login-history
	user.Links = append(tempUser.Links, GenLink("loginHistory", fmt.Sprintf("users/%v/login-history", _id)))
	// login-notification
	user.Links = append(tempUser.Links, GenLink("loginNotification", fmt.Sprintf("users/%v/login-notification", _id)))
	// favorite-shop
	user.Links = append(tempUser.Links, GenLink("favoriteShop", fmt.Sprintf("users/%v/favorite-shop", _id)))
	// wishlist
	user.Links = append(tempUser.Links, GenLink("wishlist", fmt.Sprintf("users/%v/wishlist", _id)))
	// payment-information
	user.Links = append(tempUser.Links, GenLink("paymentInformation", fmt.Sprintf("users/%v/payment-information", _id)))
	// change-password
	user.Links = append(tempUser.Links, GenLink("changePassword", fmt.Sprintf("users/%v/change-password", _id)))
	// send-verify-email
	user.Links = append(tempUser.Links, GenLink("sendVerifyEmail", fmt.Sprintf("users/%v/send-verify-email", _id)))
}

func (shop *Shop) ConstructShopLinks() {
	tempShop := shop
	shop_id := tempShop.ID.Hex()
	// self
	shop.Links = append(tempShop.Links, GenLink("self", fmt.Sprintf("shops/%v", shop_id)))
	// about
	shop.Links = append(tempShop.Links, GenLink("about", fmt.Sprintf("shops/%v/about", shop_id)))
	// gallery
	shop.Links = append(tempShop.Links, GenLink("gallery", fmt.Sprintf("shops/%v/gallery", shop_id)))
	// followers
	shop.Links = append(tempShop.Links, GenLink("followers", fmt.Sprintf("shops/%v/followers", shop_id)))
	// reviews
	shop.Links = append(tempShop.Links, GenLink("reviews", fmt.Sprintf("shops/%v/reviews", shop_id)))
	// policies
	shop.Links = append(tempShop.Links, GenLink("policies", fmt.Sprintf("shops/%v/policies", shop_id)))
	// verification
	shop.Links = append(tempShop.Links, GenLink("verification", fmt.Sprintf("shops/%v/verification", shop_id)))
	// compliance
	shop.Links = append(tempShop.Links, GenLink("compliance", fmt.Sprintf("shops/%v/compliance", shop_id)))
	// shipping
	shop.Links = append(tempShop.Links, GenLink("shipping", fmt.Sprintf("shops/%v/shipping", shop_id)))
	// status
	shop.Links = append(tempShop.Links, GenLink("status", fmt.Sprintf("shops/%v/status", shop_id)))
	// logo
	shop.Links = append(tempShop.Links, GenLink("logo", fmt.Sprintf("shops/%v/logo", shop_id)))
	// banner
	shop.Links = append(tempShop.Links, GenLink("banner", fmt.Sprintf("shops/%v/banner", shop_id)))
	// information
	shop.Links = append(tempShop.Links, GenLink("information", fmt.Sprintf("shops/%v/information", shop_id)))
	// vacation
	shop.Links = append(tempShop.Links, GenLink("vacation", fmt.Sprintf("shops/%v/vacation", shop_id)))
}
