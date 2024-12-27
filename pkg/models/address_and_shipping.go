package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type UserAddress struct {
	City       string             `bson:"city" json:"city" validate:"required"`
	State      string             `bson:"state" json:"state" validate:"required"`
	Street     string             `bson:"street" json:"street" validate:"required"`
	PostalCode string             `bson:"postal_code" json:"postalCode" validate:"required"`
	Country    Country            `bson:"country" json:"country"`
	Id         primitive.ObjectID `bson:"_id" json:"_id"`
	UserId     primitive.ObjectID `bson:"user_id" json:"userId"`
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
	Title              string             `bson:"title" json:"title" validate:"required"`
	DestinationBy      string             `bson:"destination_by" json:"destinationBy"`
	OriginState        string             `bson:"origin_state" json:"originState"`
	ShippingMethods    []string           `bson:"shipping_methods" json:"methods" validate:"oneof=standard express next-day"`
	Destinations       []string           `bson:"destinations" json:"destinations"`
	ShippingService    []string           `bson:"service" json:"service"`
	Processing         ShippingProcessing `bson:"processing" json:"processing"`
	Policy             ShippingPolicy     `bson:"policy" json:"policy"`
	SecondaryPrice     float64            `bson:"secondary_price" json:"secondaryPrice"`
	PrimaryPrice       float64            `bson:"primary_price" json:"primaryPrice"`
	MinDeliveryDays    int                `bson:"min_delivery_days" json:"minDeliveryDays"`
	HandlingFee        float64            `bson:"handling_fee" json:"handlingFee"`
	OriginPostalCode   int                `bson:"origin_postal_code" json:"originPostalCode"`
	MaxDeliveryDays    int                `bson:"max_delivery_days" json:"maxDeliveryDays"`
	CreatedAt          primitive.DateTime `bson:"created_at" json:"createdAt"`
	ModifiedAt         primitive.DateTime `bson:"modified_at" json:"modifiedAt"`
	ID                 primitive.ObjectID `bson:"_id" json:"_id" validate:"omitempty"`
	ShopID             primitive.ObjectID `bson:"shop_id" json:"shopId" validate:"required"`
	IsDefault          bool               `bson:"is_default_profile" json:"isDefault"`
	OffersFreeShipping bool               `bson:"offers_free_shipping" json:"offersFreeShipping"`
}

type ShippingProfileForListing struct {
	Title                    string             `bson:"title" json:"title" validate:"required"`
	DestinationBy            string             `bson:"destination_by" json:"destinationBy"`
	OriginState              string             `bson:"origin_state" json:"originState"`
	Destinations             []string           `bson:"destinations" json:"destinations"`
	ShippingMethods          []string           `bson:"shipping_methods" json:"methods" validate:"oneof=standard express next-day"`
	ShippingService          []string           `bson:"service" json:"service"`
	Processing               ShippingProcessing `bson:"processing" json:"processing"`
	Policy                   ShippingPolicy     `bson:"policy" json:"policy"`
	SecondaryPrice           float64            `bson:"secondary_price" json:"secondaryPrice"`
	HandlingFee              float64            `bson:"handling_fee" json:"handlingFee"`
	PrimaryPrice             float64            `bson:"primary_price" json:"primaryPrice"`
	OriginPostalCode         int                `bson:"origin_postal_code" json:"originPostalCode"`
	MaxDeliveryDays          int                `bson:"max_delivery_days" json:"maxDeliveryDays"`
	MinDeliveryDays          int                `bson:"min_delivery_days" json:"minDeliveryDays"`
	IsDefaultShippingProfile bool               `bson:"is_default_profile" json:"isDefault"`
	OffersFreeShipping       bool               `bson:"offers_free_shipping" json:"offersFreeShipping"`
}

type ShopShippingProfileRequest struct {
	Title              string             `bson:"title" json:"title" validate:"required"`
	DestinationBy      string             `bson:"destination_by" json:"destinationBy"`
	OriginState        string             `bson:"origin_state" json:"originState" validate:"required"`
	ShippingMethod     []string           `bson:"methods" json:"methods"`
	Destinations       []string           `bson:"destination" json:"destinations"`
	Processing         ShippingProcessing `bson:"processing" json:"processing"`
	Policy             ShippingPolicy     `bson:"policy" json:"policy"`
	SecondaryPrice     float64            `bson:"secondary_price" json:"secondaryPrice"`
	PrimaryPrice       float64            `bson:"primary_price" json:"primaryPrice" validate:"required"`
	HandlingFee        float64            `bson:"handling_fee" json:"handlingFee"`
	OriginPostalCode   int                `bson:"origin_postal_code" json:"originPostalCode"`
	MaxDeliveryDays    int                `bson:"max_delivery_days" json:"maxDeliveryDays"`
	MinDeliveryDays    int                `bson:"min_delivery_days" json:"minDeliveryDays"`
	ID                 primitive.ObjectID `bson:"_id" json:"_id" validate:"omitempty"`
	IsDefault          bool               `bson:"is_default" json:"isDefault"`
	OffersFreeShipping bool               `bson:"offers_free_shipping" json:"offersFreeShipping"`
}

type ShippingPolicy struct {
	ReturnUnit     string   `bson:"return_unit" json:"returnUnit" validate:"oneof=days weeks"`
	Conditions     []string `bson:"conditons" json:"conditons" validate:"omitempty"`
	ReturnPeriod   int      `bson:"return_period" json:"returnPeriod" validate:"omitempty"`
	AcceptReturns  bool     `bson:"accept_returns" json:"acceptReturns" validate:"omitempty"`
	AcceptExchange bool     `bson:"accept_exchange" json:"acceptExchange" validate:"omitempty"`
}

type ShippingProcessing struct {
	MinUnit string `bson:"processing_min_unit" json:"minUnit"`
	MaxUnit string `bson:"processing_max_unit" json:"maxUnit"`
	Min     int    `bson:"processing_min" json:"min"`
	Max     int    `bson:"processing_max" json:"max"`
}
