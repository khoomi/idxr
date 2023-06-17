package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type UserAddress struct {
	Id                       primitive.ObjectID `bson:"_id" json:"_id"`
	UserId                   primitive.ObjectID `bson:"user_id" json:"user_id"`
	City                     string             `bson:"city" json:"city" validate:"required"`
	State                    string             `bson:"state" json:"state" validate:"required"`
	Street                   string             `bson:"street" json:"street" validate:"required"`
	PostalCode               string             `bson:"postal_code" json:"postal_code" validate:"required"`
	Country                  Country            `bson:"country" json:"country" validate:"required"`
	IsDefaultShippingAddress bool               `bson:"is_default_shipping_address" json:"is_default_shipping_address"`
}

type UserAddressUpdateequest struct {
	City                     string  `bson:"city" json:"city" validate:"required"`
	State                    string  `bson:"state" json:"state" validate:"required"`
	Street                   string  `bson:"street" json:"street" validate:"required"`
	PostalCode               string  `bson:"postal_code" json:"postal_code" validate:"required"`
	Country                  Country `bson:"country" json:"country" validate:"required"`
	IsDefaultShippingAddress bool    `bson:"is_default_shipping_address" json:"is_default_shipping_address"`
}

type ShopShippingProfile struct {
	// The unique identifier of the shipping profile.
	ID primitive.ObjectID `bson:"_id" json:"_id" validate:"omitempty"`
	// The identifier of the associated shop.
	ShopID primitive.ObjectID `bson:"shop_id" json:"shop_id" validate:"required"`
	// The name of this shipping profile.
	Title string `bson:"title" json:"title" validate:"required"`
	// The minimum time required to process and ship listings with this shipping profile.
	MinProcessingTime int `bson:"min_processing_time" json:"min_processing_time" validate:"required"`
	// The maximum time required to process and ship listings with this shipping profile.
	MaxProcessingTime int `bson:"max_processing_time" json:"max_processing_time" validate:"required"`
	// The unit used to represent the processing time. Valid values: "business_days", "weeks". Default: "business_days".
	ProcessingTimeUnit string `bson:"processing_time_unit" json:"processing_time_unit" validate:"required,oneof=business_days weeks,default=business_days"`
	// The states/regions where this shipping profile applies. Default: ["all"].
	DestinationStates []string `bson:"destination_states" json:"destination_states" validate:"required,default=all"`
	// The regions where this shipping profile applies. Default: ["all"].
	DestinationRegion []string `bson:"destination_region" json:"destination_region" validate:"required,default=all"`
	// The minimum number of days for delivery.
	MinDeliveryDays int `bson:"min_delivery_days" json:"min_delivery_days"`
	// The maximum number of days for delivery.
	MaxDeliveryDays int `bson:"max_delivery_days" json:"max_delivery_days"`
	// The state from which the listing ships.
	OriginState string `bson:"origin_state" json:"origin_state"`
	// The postal code from which the listing ships.
	OriginPostalCode int `bson:"origin_postal_code" json:"origin_postal_code"`
	// The cost of shipping to this region alone.
	PrimaryCost float64 `bson:"primary_cost" json:"primary_cost"`
	// The domestic handling fee.
	DomesticHandlingFee float32 `bson:"domestic_handling_fee" json:"domestic_handling_fee"`
	// The available shipping methods. Valid values: "standard", "express", "next-day".
	ShippingMethods []string `bson:"shipping_methods" json:"shipping_methods" validate:"dive,oneof=standard express next-day"`
}

type ShopShippingProfileRequest struct {
	Title               string   `json:"title" validate:"required"`
	MinProcessingTime   int      `json:"min_processing_time" validate:"required"`
	MaxProcessingTime   int      `json:"max_processing_time" validate:"required"`
	ProcessingTimeUnit  string   `json:"processing_time_unit" validate:"required,oneof=business_days weeks,default=business_days"`
	DestinationStates   []string `json:"destination_states" validate:"required,default=all"`
	DestinationRegion   []string `json:"destination_region" validate:"required,default=all"`
	MinDeliveryDays     int      `json:"min_delivery_days"`
	MaxDeliveryDays     int      `json:"max_delivery_days"`
	OriginState         string   `json:"origin_state"`
	OriginPostalCode    int      `json:"origin_postal_code"`
	PrimaryCost         float64  `json:"primary_cost"`
	DomesticHandlingFee float32  `json:"domestic_handling_fee"`
	ShippingMethods     []string `json:"shipping_methods" validate:"required,gt=0,dive,oneof=standard express next-day"`
}
