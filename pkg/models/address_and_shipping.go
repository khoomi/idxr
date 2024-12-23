package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type UserAddress struct {
	Id         primitive.ObjectID `bson:"_id" json:"_id"`
	UserId     primitive.ObjectID `bson:"user_id" json:"user_id"`
	City       string             `bson:"city" json:"city" validate:"required"`
	State      string             `bson:"state" json:"state" validate:"required"`
	Street     string             `bson:"street" json:"street" validate:"required"`
	PostalCode string             `bson:"postal_code" json:"postalCode" validate:"required"`
	Country    Country            `bson:"country" json:"country"`
	IsDefault  bool               `bson:"is_default_shipping_address" json:"isDefault"`
}

type UserAddressExcerpt struct {
	City       string  `bson:"city" json:"city" validate:"required"`
	State      string  `bson:"state" json:"state" validate:"required"`
	Street     string  `bson:"street" json:"street" validate:"required"`
	PostalCode string  `bson:"postal_code" json:"postalCode" validate:"required"`
	Country    Country `bson:"country" json:"country"`
	IsDefault  bool    `bson:"is_default_shipping_address" json:"isDefault"`
}

type UserAddressUpdateRequest struct {
	City       string  `bson:"city" json:"city" validate:"required"`
	State      string  `bson:"state" json:"state" validate:"required"`
	Street     string  `bson:"street" json:"street" validate:"required"`
	PostalCode string  `bson:"postal_code" json:"postalCode" validate:"required"`
	Country    Country `bson:"country" json:"country" validate:"required"`
	IsDefault  bool    `bson:"is_default_shipping_address" json:"isDefault"`
}

type ShopShippingProfile struct {
	ID                 primitive.ObjectID `bson:"_id" json:"_id" validate:"omitempty"`
	ShopID             primitive.ObjectID `bson:"shop_id" json:"shopId" validate:"required"`
	Title              string             `bson:"title" json:"title" validate:"required"`
	DestinationBy      string             `bson:"destination_by" json:"destinationBy"`
	Destinations       []string           `bson:"destinations" json:"destinations"`
	MinDeliveryDays    int                `bson:"min_delivery_days" json:"minDeliveryDays"`
	MaxDeliveryDays    int                `bson:"max_delivery_days" json:"maxDeliveryDays"`
	OriginState        string             `bson:"origin_state" json:"originState"`
	OriginPostalCode   int                `bson:"origin_postal_code" json:"originPostalCode"`
	PrimaryPrice       string             `bson:"primary_price" json:"primaryPrice"`
	SecondaryPrice     string             `bson:"secondary_price" json:"secondaryPrice"`
	HandlingFee        string             `bson:"handling_fee" json:"handlingFee"`
	IsDefault          bool               `bson:"is_default_profile" json:"isDefault"`
	OffersFreeShipping bool               `bson:"offers_free_shipping" json:"offersFreeShipping"`
	ShippingMethods    []string           `bson:"shipping_methods" json:"methods" validate:"oneof=standard express next-day"`
	ShippingService    []string           `bson:"service" json:"service"`
	Policy             ShippingPolicy     `bson:"policy" json:"policy"`
	CreatedAt          primitive.DateTime `bson:"created_at" json:"createdAt"`
	ModifiedAt         primitive.DateTime `bson:"modified_at" json:"modifiedAt"`
	Processing         ShippingProcessing `bson:"processing" json:"processing"`
}

type ShippingProfileForListing struct {
	Title                    string             `bson:"title" json:"title" validate:"required"`
	DestinationBy            string             `bson:"destination_by" json:"destinationBy"`
	Destinations             []string           `bson:"destinations" json:"destinations"`
	MinDeliveryDays          int                `bson:"min_delivery_days" json:"minDeliveryDays"`
	MaxDeliveryDays          int                `bson:"max_delivery_days" json:"maxDeliveryDays"`
	OriginState              string             `bson:"origin_state" json:"originState"`
	OriginPostalCode         int                `bson:"origin_postal_code" json:"originPostalCode"`
	PrimaryPrice             string             `bson:"primary_price" json:"primaryPrice"`
	SecondaryPrice           string             `bson:"secondary_price" json:"secondaryPrice"`
	HandlingFee              string             `bson:"handling_fee" json:"handlingFee"`
	ShippingService          []string           `bson:"service" json:"service"`
	ShippingMethods          []string           `bson:"shipping_methods" json:"methods" validate:"oneof=standard express next-day"`
	IsDefaultShippingProfile bool               `bson:"is_default_profile" json:"isDefault"`
	OffersFreeShipping       bool               `bson:"offers_free_shipping" json:"offersFreeShipping"`
	Policy                   ShippingPolicy     `bson:"policy" json:"policy"`
	Processing               ShippingProcessing `bson:"processing" json:"processing"`
}

type ShopShippingProfileRequest struct {
	ID                 primitive.ObjectID `bson:"_id" json:"_id" validate:"omitempty"`
	Title              string             `bson:"title" json:"title" validate:"required"`
	DestinationBy      string             `bson:"destination_by" json:"destinationBy"`
	Destinations       []string           `bson:"destination" json:"destinations"`
	MinDeliveryDays    int                `bson:"min_delivery_days" json:"minDeliveryDays"`
	MaxDeliveryDays    int                `bson:"max_delivery_days" json:"maxDeliveryDays"`
	OriginState        string             `bson:"origin_state" json:"originState" validate:"required"`
	OriginPostalCode   int                `bson:"origin_postal_code" json:"originPostalCode"`
	PrimaryPrice       string             `bson:"primary_price" json:"primaryPrice" validate:"required"`
	SecondaryPrice     string             `bson:"secondary_price" json:"secondaryPrice"`
	HandlingFee        string             `bson:"handling_fee" json:"handlingFee"`
	IsDefault          bool               `bson:"is_default" json:"isDefault"`
	OffersFreeShipping bool               `bson:"offers_free_shipping" json:"offersFreeShipping"`
	ShippingMethod     []string           `bson:"methods" json:"methods"`
	Policy             ShippingPolicy     `bson:"policy" json:"policy"`
	Processing         ShippingProcessing `bson:"processing" json:"processing"`
}

type ShippingPolicy struct {
	ReturnPeriod   int      `bson:"return_period" json:"returnPeriod" validate:"omitempty"`
	ReturnUnit     string   `bson:"return_unit" json:"returnUnit" validate:"oneof=days weeks"`
	AcceptReturns  bool     `bson:"accept_returns" json:"acceptReturns" validate:"omitempty"`
	AcceptExchange bool     `bson:"accept_exchange" json:"acceptExchange" validate:"omitempty"`
	Conditions     []string `bson:"conditons" json:"conditons" validate:"omitempty"`
}

type ShippingProcessing struct {
	Min     int    `bson:"processing_min" json:"min"`
	MinUnit string `bson:"processing_min_unit" json:"minUnit"`
	Max     int    `bson:"processing_max" json:"max"`
	MaxUnit string `bson:"processing_max_unit" json:"maxUnit"`
}
