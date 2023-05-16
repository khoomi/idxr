package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type UserAddress struct {
	Id                       primitive.ObjectID `bson:"_id" json:"_id"`
	City                     string             `bson:"city" json:"city" validate:"required"`
	State                    string             `bson:"state" json:"state" validate:"required"`
	Street                   string             `bson:"street" json:"street" validate:"required"`
	PostalCode               string             `bson:"postal_code" json:"postal_code" validate:"required"`
	Country                  Country            `bson:"country" json:"country" validate:"required"`
	UserId                   primitive.ObjectID `bson:"user_id" json:"user_id"`
	IsDefaultShippingAddress bool               `bson:"is_default_shipping_address" json:"is_default_shipping_address"`
}

type ShopShippingProfile struct {
	Id        primitive.ObjectID `bson:"_id" json:"_id" validate:"omitempty"`
	ListingId primitive.ObjectID `bson:"listing_id" json:"listing_id"`
	// The name string of this shipping profile.
	Title string `bson:"title" json:"title" validate:"required"`
	// The minimum time required to process to ship listings with this shipping profile.
	MinProcessingTime int `bson:"min_processing_time" json:"min_processing_time" validate:"required"`
	// The maximum required to process to ship listings with this shipping profile.
	MaxProcessingTime int `bson:"max_processing_time" json:"max_processing_time" validate:"required"`
	// Default: "business_days"
	// Enum: "business_days" "weeks"
	//The unit used to represent how long a processing time is. A week is equivalent to 5 business days. If none is provided, the unit is set to "business_days".
	ProcessingTimeUnit string `bson:"processing_time_unit" json:"processing_time_unit" validate:"required,oneof=business_days weeks, default=business_days"`
	// Default: "all"
	DestinationStates []string `bson:"destination_region" json:"destination_region" validate:"required, default=all"`
	// The minimum number of business days a buyer can expect to wait to receive their purchased item once it has shipped.
	MinDeliveryDays int `bson:"min_delivery_days" json:"min_delivery_days"`
	// The maximum number of business days a buyer can expect to wait to receive their purchased item once it has shipped
	MaxDeliveryDays int `bson:"max_delivery_days" json:"max_delivery_days"`
	// The state string for the location from which the listing ships.
	OriginState string `bson:"origin_state" json:"origin_state"`
	// The postal code for the location from which the listing ships.
	OriginPostalCode int `bson:"origin_postal_code" json:"origin_postal_code"`
	// The cost of shipping to this region alone, measured in the store's default currency.
	PrimaryCost float64 `bson:"primary_cost" json:"primary_cost"`
	// The cost of shipping to this region with another item, measured in the store's default currency.
	SecondaryCost       float64 `bson:"secondary_cost" json:"secondary_cost"`
	DomesticHandlingFee float32 `bson:"domestic_handling_fee" json:"domestic_handling_fee"`
}

type ShopShippingProfileRequest struct {
	Title               string             `bson:"title" json:"title" validate:"required"`
	ListingId           primitive.ObjectID `bson:"listing_id" json:"listing_id"`
	OriginState         string             `bson:"origin_state" json:"origin_state" validate:"required"`
	OriginPostalCode    int                `bson:"origin_postal_code" json:"origin_postal_code"`
	MinProcessingTime   int                `bson:"min_processing_time" json:"min_processing_time" validate:"required,default=0"`
	MaxProcessingTime   int                `bson:"max_processing_time" json:"max_processing_time" validate:"required,default=0"`
	ProcessingTimeUnit  string             `bson:"processing_time_unit" json:"processing_time_unit" validate:"required,oneof=business_days weeks, default=business_days"`
	MinDeliveryDays     int                `bson:"min_delivery_days" json:"min_delivery_days" validate:"required,default=0"`
	MaxDeliveryDays     int                `bson:"max_delivery_days" json:"max_delivery_days" validate:"required,default=0"`
	PrimaryCost         float64            `bson:"primary_cost" json:"primary_cost"`
	SecondaryCost       float64            `bson:"secondary_cost" json:"secondary_cost"`
	DomesticHandlingFee float32            `bson:"domestic_handling_fee" json:"domestic_handling_fee"`
	DestinationRegion   []string           `bson:"destination_region" json:"destination_region" validate:"required,oneof=nc ne nw ss se sw none, default=[sw]"`
}
