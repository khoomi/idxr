package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type VariantSelection struct {
	Size  string `bson:"size,omitempty" json:"size,omitempty"`
	Color string `bson:"color,omitempty" json:"color,omitempty"`
	Style string `bson:"style,omitempty" json:"style,omitempty"`
}

type CartItem struct {
	Id              primitive.ObjectID `bson:"_id" json:"_id"`
	ListingId       primitive.ObjectID `bson:"listing_id" json:"listing_id"`
	ShopId          primitive.ObjectID `bson:"shopId" json:"shopId"`
	UserId          primitive.ObjectID `bson:"userId" json:"userId"`
	Title           string             `bson:"title" json:"title"`
	Thumbnail       string             `bson:"thumbnail" json:"thumbnail"`
	Quantity        int                `bson:"quantity" json:"quantity"`
	UnitPrice       float64            `bson:"unit_price" json:"unit_price"`
	TotalPrice      float64            `bson:"total_price" json:"total_price"`
	Variant         *VariantSelection  `bson:"variant,omitempty" json:"variant,omitempty"`
	DynamicType     DynamicType        `bson:"dynamic_type" json:"dynamic_type"`
	Personalization *Personalization   `bson:"personalization,omitempty" json:"personalization,omitempty"`

	// Shop information
	ShopName     string `bson:"shop_name" json:"shop_name"`
	ShopUsername string `bson:"shop_username" json:"shop_username"`
	ShopSlug     string `bson:"shop_slug" json:"shop_slug"`

	// Inventory and availability
	AvailableQuantity int          `bson:"available_quantity" json:"available_quantity"`
	ListingState      ListingState `bson:"listing_state" json:"listing_state"`

	// Pricing validation
	OriginalPrice  float64   `bson:"original_price" json:"original_price"`
	PriceUpdatedAt time.Time `bson:"price_updated_at" json:"price_updated_at"`

	// Shipping
	ShippingProfileId primitive.ObjectID `bson:"shipping_profile_id" json:"shipping_profile_id"`

	ExpiresAt  time.Time `bson:"expires_at" json:"expiresAt"`
	AddedAt    time.Time `bson:"added_at" json:"addedAt"`
	ModifiedAt time.Time `bson:"modified_at" json:"modifiedAt"`
}

type CartItemRequest struct {
	ListingId       primitive.ObjectID `bson:"listing_id" json:"listing_id"`
	ShopId          primitive.ObjectID `bson:"shopId" json:"shopId"`
	UserId          primitive.ObjectID `bson:"userId" json:"userId"`
	Quantity        int                `json:"quantity" validate:"gte=1"`
	Variant         *VariantSelection  `json:"variant,omitempty"`
	Personalization *Personalization   `json:"personalization,omitempty"`
}
