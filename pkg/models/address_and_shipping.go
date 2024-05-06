package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type UserAddress struct {
	Id                       primitive.ObjectID `bson:"_id" json:"_id"`
	UserId                   primitive.ObjectID `bson:"user_id" json:"user_id"`
	City                     string             `bson:"city" json:"city" validate:"required"`
	State                    string             `bson:"state" json:"state" validate:"required"`
	Street                   string             `bson:"street" json:"street" validate:"required"`
	PostalCode               string             `bson:"postal_code" json:"postal_code" validate:"required"`
	Country                  Country            `bson:"country" json:"country"`
	IsDefaultShippingAddress bool               `bson:"is_default_shipping_address" json:"is_default_shipping_address"`
}

type UserAddressExcerpt struct {
	City                     string  `bson:"city" json:"city" validate:"required"`
	State                    string  `bson:"state" json:"state" validate:"required"`
	Street                   string  `bson:"street" json:"street" validate:"required"`
	PostalCode               string  `bson:"postal_code" json:"postal_code" validate:"required"`
	Country                  Country `bson:"country" json:"country"`
	IsDefaultShippingAddress bool    `bson:"is_default_shipping_address" json:"is_default_shipping_address"`
}

type UserAddressUpdateRequest struct {
	City                     string  `bson:"city" json:"city" validate:"required"`
	State                    string  `bson:"state" json:"state" validate:"required"`
	Street                   string  `bson:"street" json:"street" validate:"required"`
	PostalCode               string  `bson:"postal_code" json:"postal_code" validate:"required"`
	Country                  Country `bson:"country" json:"country" validate:"required"`
	IsDefaultShippingAddress bool    `bson:"is_default_shipping_address" json:"is_default_shipping_address"`
}

type ShopShippingProfile struct {
	ID                       primitive.ObjectID `bson:"_id" json:"_id" validate:"omitempty"`
	ShopID                   primitive.ObjectID `bson:"shop_id" json:"shop_id" validate:"required"`
	Title                    string             `bson:"title" json:"title" validate:"required"`
	MinProcessingTime        int                `bson:"min_processing_time" json:"min_processing_time" validate:"required"`
	MaxProcessingTime        int                `bson:"max_processing_time" json:"max_processing_time" validate:"required"`
	ProcessingTimeUnit       string             `bson:"processing_time_unit" json:"processing_time_unit" validate:"required,oneof=days weeks"`
	DestinationBy            string             `bson:"destination_by" json:"destination_by"`
	Destinations             []string           `bson:"destinations" json:"destinations"`
	MinDeliveryDays          int                `bson:"min_delivery_days" json:"min_delivery_days"`
	MaxDeliveryDays          int                `bson:"max_delivery_days" json:"max_delivery_days"`
	OriginState              string             `bson:"origin_state" json:"origin_state"`
	OriginPostalCode         int                `bson:"origin_postal_code" json:"origin_postal_code"`
	PrimaryPrice             string             `bson:"primary_price" json:"primary_price"`
	SecondaryPrice           string             `bson:"secondary_price" json:"secondary_price"`
	HandlingFee              string             `bson:"handling_fee" json:"handling_fee"`
	ShippingMethods          []string           `bson:"shipping_methods" json:"shipping_methods" validate:"oneof=standard express next-day"`
	IsDefaultShippingProfile bool               `bson:"is_default_profile" json:"is_default_profile"`
	OffersFreeShipping       bool               `bson:"offers_free_shipping" json:"offers_free_shipping"`
	AutoCalculatePrice       bool               `bson:"auto_calculate_price" json:"auto_calculate_price"`
	ShippingService          bool               `bson:"service" json:"service"`
	Policy                   ShippingPolicy     `bson:"policy" json:"policy"`
	CreatedAt                primitive.DateTime `bson:"created_at" json:"created_at"`
	ModifiedAt               primitive.DateTime `bson:"modified_at" json:"modified_at"`
}

type ShippingProfileForListing struct {
	Title                    string         `bson:"title" json:"title" validate:"required"`
	MinProcessingTime        int            `bson:"min_processing_time" json:"min_processing_time" validate:"required"`
	MaxProcessingTime        int            `bson:"max_processing_time" json:"max_processing_time" validate:"required"`
	ProcessingTimeUnit       string         `bson:"processing_time_unit" json:"processing_time_unit" validate:"required,oneof=days weeks"`
	DestinationBy            string         `bson:"destination_by" json:"destination_by"`
	Destinations             []string       `bson:"destinations" json:"destinations"`
	MinDeliveryDays          int            `bson:"min_delivery_days" json:"min_delivery_days"`
	MaxDeliveryDays          int            `bson:"max_delivery_days" json:"max_delivery_days"`
	OriginState              string         `bson:"origin_state" json:"origin_state"`
	OriginPostalCode         int            `bson:"origin_postal_code" json:"origin_postal_code"`
	PrimaryPrice             string         `bson:"primary_price" json:"primary_price"`
	SecondaryPrice           string         `bson:"secondary_price" json:"secondary_price"`
	HandlingFee              string         `bson:"handling_fee" json:"handling_fee"`
	ShippingMethods          []string       `bson:"shipping_methods" json:"shipping_methods" validate:"oneof=standard express next-day"`
	IsDefaultShippingProfile bool           `bson:"is_default_profile" json:"is_default_profile"`
	OffersFreeShipping       bool           `bson:"offers_free_shipping" json:"offers_free_shipping"`
	AutoCalculatePrice       bool           `bson:"auto_calculate_price" json:"auto_calculate_price"`
	ShippingService          bool           `bson:"service" json:"service"`
	Policy                   ShippingPolicy `bson:"policy" json:"policy"`
}

type ShopShippingProfileRequest struct {
	ID                       primitive.ObjectID `bson:"_id" json:"_id" validate:"omitempty"`
	Title                    string             `bson:"title" json:"title" validate:"required"`
	MinProcessingTime        int                `bson:"min_processing_time" json:"min_processing_time" validate:"required"`
	MaxProcessingTime        int                `bson:"max_processing_time" json:"max_processing_time" validate:"required"`
	ProcessingTimeUnit       string             `bson:"processing_time_unit" json:"processing_time_unit" validate:"required,oneof=days weeks"`
	DestinationBy            string             `bson:"destination_by" json:"destination_by"`
	Destinations             []string           `bson:"destination" json:"destination"`
	MinDeliveryDays          int                `bson:"min_delivery_days" json:"min_delivery_days"`
	MaxDeliveryDays          int                `bson:"max_delivery_days" json:"max_delivery_days"`
	OriginState              string             `bson:"origin_state" json:"origin_state"`
	OriginPostalCode         int                `bson:"origin_postal_code" json:"origin_postal_code"`
	PrimaryPrice             string             `bson:"primary_price" json:"primary_price"`
	SecondaryPrice           string             `bson:"secondary_price" json:"secondary_price"`
	HandlingFee              string             `bson:"handling_fee" json:"handling_fee"`
	IsDefaultShippingProfile bool               `bson:"is_default_profile" json:"is_default_profile"`
	OffersFreeShipping       bool               `bson:"offers_free_shipping" json:"offers_free_shipping"`
	AutoCalculatePrice       bool               `bson:"auto_calculate_price" json:"auto_calculate_price"`
	ShippingService          bool               `bson:"service" json:"service"`
	Policy                   ShippingPolicy     `bson:"policy" json:"policy"`
}

type ShippingPolicy struct {
	ReturnPeriod   int      `bson:"return_period" json:"return_period" validate:"omitempty"`
	ReturnUnit     string   `bson:"return_unit" json:"return_unit" validate:"oneof=days weeks"`
	AcceptReturns  bool     `bson:"accept_returns" json:"accept_returns" validate:"omitempty"`
	AcceptExchange bool     `bson:"accept_exchange" json:"accept_exchange" validate:"omitempty"`
	Conditions     []string `bson:"conditons" json:"conditons" validate:"omitempty"`
}
