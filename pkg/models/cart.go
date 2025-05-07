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
