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
	ListingId       primitive.ObjectID `bson:"listing_id" json:"listingId"`
	ShopId          primitive.ObjectID `bson:"shopId" json:"shopId"`
	UserId          primitive.ObjectID `bson:"userId" json:"userId"`
	Title           string             `bson:"title" json:"title"`
	Thumbnail       string             `bson:"thumbnail" json:"thumbnail"`
	Quantity        int                `bson:"quantity" json:"quantity"`
	UnitPrice       float64            `bson:"unit_price" json:"unitPrice"`
	TotalPrice      float64            `bson:"total_price" json:"totalPrice"`
	Variant         *VariantSelection  `bson:"variant,omitempty" json:"variant,omitempty"`
	DynamicType     DynamicType        `bson:"dynamic_type" json:"dynamicType"`
	Personalization *Personalization   `bson:"personalization,omitempty" json:"personalization,omitempty"`

	// Shop information
	ShopName     string `bson:"shop_name" json:"shopName"`
	ShopUsername string `bson:"shop_username" json:"shopUsername"`
	ShopSlug     string `bson:"shop_slug" json:"shopSlug"`

	// Inventory and availability
	AvailableQuantity int          `bson:"available_quantity" json:"availableQuantity"`
	ListingState      ListingState `bson:"listing_state" json:"listingState"`

	// Pricing validation
	OriginalPrice  float64   `bson:"original_price" json:"originalPrice"`
	PriceUpdatedAt time.Time `bson:"price_updated_at" json:"priceUpdatedAt"`

	// Shipping
	ShippingProfileId primitive.ObjectID `bson:"shipping_profile_id" json:"shippingProfileId"`

	ExpiresAt  time.Time `bson:"expiresAt" json:"expiresAt"`
	AddedAt    time.Time `bson:"addedAt" json:"addedAt"`
	ModifiedAt time.Time `bson:"modifiedAt" json:"modifiedAt"`
}

type CartItemRequest struct {
	ListingId       primitive.ObjectID `json:"listingId"`
	ShopId          primitive.ObjectID `json:"shopId"`
	UserId          primitive.ObjectID `json:"userId"`
	Quantity        int                `json:"quantity" validate:"gte=1"`
	Variant         *VariantSelection  `json:"variant,omitempty"`
	Personalization *Personalization   `json:"personalization,omitempty"`
}

type CartItemJson struct {
	Id              primitive.ObjectID `json:"_id"`
	ListingId       primitive.ObjectID `json:"listingId"`
	ShopId          primitive.ObjectID `json:"shopId"`
	UserId          primitive.ObjectID `json:"userId"`
	Title           string             `json:"title"`
	Thumbnail       string             `json:"thumbnail"`
	Quantity        int                `json:"quantity"`
	UnitPrice       float64            `json:"unitPrice"`
	TotalPrice      float64            `json:"totalPrice"`
	Variant         *VariantSelection  `json:"variant,omitempty"`
	DynamicType     DynamicType        `json:"dynamicType"`
	Personalization *Personalization   `json:"personalization,omitempty"`

	// Shop information
	ShopName     string `json:"shopName"`
	ShopUsername string `json:"shopUsername"`
	ShopSlug     string `json:"shopSlug"`

	// Inventory and availability
	AvailableQuantity int          `json:"availableQuantity"`
	ListingState      ListingState `json:"listingState"`

	// Pricing validation
	OriginalPrice  float64   `json:"originalPrice"`
	PriceUpdatedAt time.Time `json:"priceUpdatedAt"`

	// Shipping
	ShippingProfileId primitive.ObjectID `json:"shippingProfileId"`

	// Timestamps
	ExpiresAt  time.Time `json:"expiresAt"`
	AddedAt    time.Time `json:"addedAt"`
	ModifiedAt time.Time `json:"modifiedAt"`

	// Validation flags
	IsAvailable       bool    `json:"isAvailable"`
	PriceChanged      bool    `json:"priceChanged"`
	CurrentPrice      float64 `json:"currentPrice"`
	InsufficientStock bool    `json:"insufficientStock"`
	CurrentQuantity   int     `json:"currentQuantity"`
}
