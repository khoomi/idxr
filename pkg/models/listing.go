package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ListingStateType string

const (
	ListingStateActive      ListingStateType = "active"
	ListingStateRemoved     ListingStateType = "remove"
	ListingStateSoldOut     ListingStateType = "soldout"
	ListingStateExpired     ListingStateType = "expired"
	ListingStateEdit        ListingStateType = "edit"
	ListingStateDraft       ListingStateType = "draft"
	ListingStatePrivate     ListingStateType = "private"
	ListingStateUnavailable ListingStateType = "unavailable"
	ListingStateDeactivated ListingStateType = "deactivated"
)

type ListingState struct {
	State          ListingStateType `bson:"state" json:"state"`
	StateUpdatedAt time.Time        `bson:"state_updated_at" json:"state_updated_at"`
}

type ListingDateMeta struct {
	CreatedAt  time.Time `bson:"created_at" json:"created_at"`
	EndingAt   time.Time `bson:"ending_at" json:"ending_at"`
	ModifiedAt time.Time `bson:"modified_at" json:"modified_at"`
}

type ListingCategory struct {
	CategoryId   string `bson:"category_id" json:"category_id"`
	CategoryName string `bson:"category_name" json:"category_name"`
	CategoryPath string `bson:"category_path" json:"category_path"`
}

type ListingProcessing struct {
	ProcessingMin     int    `bson:"processing_min" json:"processing_min"`
	ProcessingMinUnit string `bson:"processing_min_unit" json:"processing_min_unit"`
	ProcessingMax     int    `bson:"processing_max" json:"processing_max"`
	ProcessingMaxUnit string `bson:"processing_max_unit" json:"processing_max_unit"`
}

type WeightUnit string

const (
	WeightUnitOZ = "oz"
	WeightUnitG  = "g"
	WeightUnitLB = "lb"
	WeightUnitKG = "kg"
)

type DimensionUnit string

const (
	DimensionUnitINC = "inc"
	DimensionUnitFT  = "ft"
	DimensionUnitMM  = "mm"
	DimensionUnitCM  = "cm"
	DimensionUnitM   = "m"
)

type ListingMeasurement struct {
	// How much the item weighs.
	ItemWeight float64 `bson:"item_weight" json:"item_weight"`
	// The units used to represent the weight of this item.
	ItemWeightUnit WeightUnit `bson:"item_weight_unit" json:"item_weight_unit" validate:"oneof=oz g lb kg"`
	//  How long the item is.
	ItemLength float64 `bson:"item_length" json:"item_length"`
	//  How wide the item is.
	ItemWidth float64 `bson:"item_width" json:"item_width"`
	//  How tall the item is.
	ItemHeight float64 `bson:"item_height" json:"item_height"`
	// The units used to represent the dimensions of this item.
	ItemDimensionUnit DimensionUnit `bson:"item_dimension_unit" json:"item_dimension_unit" validate:"oneof=inc ft mm cm m"`
}

type WhoMade string

const (
	WhoMadeIDid        = "i_did"
	WhoMadeCollective  = "collective"
	WhoMadeSomeoneElse = "someone_else"
)

type Listing struct {
	ID        primitive.ObjectID `bson:"_id" json:"_id"`
	Code      string             `bson:"code" json:"code"`
	State     ListingState       `bson:"state" json:"state"`
	UserId    primitive.ObjectID `bson:"user_id" json:"user_id"`
	ShopId    primitive.ObjectID `bson:"shop_id" json:"shop_id"`
	MainImage string             `bson:"main_image" json:"main_image"`
	Images    []string           `bson:"images" json:"images"`
	// Video                string                      `bson:"video" json:"video"`
	ListingDetails       ListingDetails              `bson:"details" json:"details"`
	Date                 ListingDateMeta             `bson:"date" json:"date"`
	Slug                 string                      `bson:"slug" json:"slug"`
	Views                int                         `bson:"views" json:"views"`
	FavorersCount        int                         `bson:"favorers_count" json:"favorers_count"`
	ShippingProfileId    primitive.ObjectID          `bson:"shipping_profile_id" json:"shipping_profile_id"`
	Processing           ListingProcessing           `bson:"processing" json:"processing"`
	NonTaxable           bool                        `bson:"non_taxable" json:"non_taxable"`
	Variations           []ListingVariation          `bson:"variations" json:"variations"`
	ShouldAutoRenew      bool                        `bson:"should_auto_renew" json:"should_auto_renew"`
	Inventory            Inventory                   `bson:"inventory" json:"inventory"`
	RecentReviews        []ListingReview             `bson:"recent_reviews" json:"recent_reviews"`
	Rating               ListingRating               `bson:"rating" json:"rating"`
	Measurements         ListingMeasurement          `bson:"measurements" json:"measurements"`
	FinancialInformation ListingFinancialInformation `bson:"financial_information" json:"financial_information"`
}

type ListingWithAnalytics struct {
	Listing
	Sales float64 `bson:"sales" json:"sales"`
}

type ListingsSummary struct {
	ID          primitive.ObjectID    `bson:"_id" json:"_id"`
	Code        string                `bson:"code" json:"code"`
	State       ListingState          `bson:"state" json:"state"`
	UserId      primitive.ObjectID    `bson:"user_id" json:"userId"`
	ShopId      primitive.ObjectID    `bson:"shop_id" json:"shopId"`
	MainImage   string                `bson:"main_image" json:"mainImage"`
	Images      []string              `bson:"images" json:"images"`
	Date        ListingDateMeta       `bson:"date" json:"date"`
	Slug        string                `bson:"slug" json:"slug"`
	Sales       float64               `bson:"sales" json:"sales"`
	TotalOrders int                   `bson:"total_orders" json:"totalOrders"`
	Inventory   InventorySummary      `bson:"inventory" json:"inventory"`
	Details     ListingDetailsSummary `bson:"details" json:"details"`
}

type ListingDetailsSummary struct {
	Title    string          `bson:"title" json:"title"`
	Category ListingCategory `bson:"category" json:"category"`
}

type InventorySummary struct {
	Quantity int    `bson:"quantity" json:"quantity" validate:"required"`
	Price    string `bson:"price" json:"price" validate:"required"`
}

type ListingExtra struct {
	ID                   primitive.ObjectID          `bson:"_id" json:"_id"`
	State                ListingState                `bson:"state" json:"state"`
	UserId               primitive.ObjectID          `bson:"user_id" json:"user_id"`
	ShopId               primitive.ObjectID          `bson:"shop_id" json:"shop_id"`
	MainImage            string                      `bson:"main_image" json:"main_image"`
	Images               []string                    `bson:"images" json:"images"`
	ListingDetails       ListingDetails              `bson:"details" json:"details"`
	Date                 ListingDateMeta             `bson:"date" json:"date"`
	Slug                 string                      `bson:"slug" json:"slug"`
	Views                int                         `bson:"views" json:"views"`
	FavorersCount        int                         `bson:"favorers_count" json:"favorers_count"`
	ShippingProfileId    primitive.ObjectID          `bson:"shipping_profile_id" json:"shipping_profile_id"`
	Processing           ListingProcessing           `bson:"processing" json:"processing"`
	NonTaxable           bool                        `bson:"non_taxable" json:"non_taxable"`
	Variations           []ListingVariation          `bson:"variations" json:"variations"`
	ShouldAutoRenew      bool                        `bson:"should_auto_renew" json:"should_auto_renew"`
	Inventory            Inventory                   `bson:"inventory" json:"inventory"`
	RecentReviews        []ListingReview             `bson:"recent_reviews" json:"recent_reviews"`
	Rating               ListingRating               `bson:"rating" json:"rating"`
	TotalOrders          int                         `bson:"total_orders" json:"total_orders"`
	Sales                float64                     `bson:"sales" json:"sales"`
	User                 ListingUserExcept           `bson:"user" json:"user"`
	Shop                 ListingShopExcept           `bson:"shop" json:"shop"`
	Measurements         ListingMeasurement          `bson:"measurements" json:"measurements"`
	FinancialInformation ListingFinancialInformation `bson:"financial_information" json:"financial_information"`
}

type ListingUserExcept struct {
	LoginName string `bson:"login_name" json:"login_name" validate:"required"`
	FirstName string `bson:"first_name" json:"first_name"`
	LastName  string `bson:"last_name" json:"last_name"`
	Thumbnail string `bson:"thumbnail" json:"thumbnail"`
}

type ListingShopExcept struct {
	Name         string `bson:"name" json:"name" validate:"required"`
	Description  string `bson:"description" json:"description" validate:"required"`
	Username     string `bson:"username" json:"username" validate:"required"`
	Location     string `bson:"location" json:"location"`
	Slug         string `bson:"slug" json:"slug" validate:"required"`
	LogoURL      string `bson:"logo_url" json:"logo_url"`
	ReviewsCount int    `bson:"reviews_count" json:"reviews_count"`
}

type ListingDetails struct {
	Title                       string                 `bson:"title" json:"title"`
	Description                 string                 `bson:"description" json:"description"`
	Condition                   string                 `bson:"condition" json:"condition" validate:"oneof=new used refurbished"`
	Category                    ListingCategory        `bson:"category" json:"category"`
	WhoMade                     string                 `bson:"who_made" json:"who_made" validate:"oneof=i_did collective someone_else"`
	WhenMade                    string                 `bson:"when_made" json:"when_made"  validate:"oneof=in2020_2023 in2010_2019 in2003_2009 before_2003 in2000_2002 in1990s in1980s in1970s in1960s in1950s in1940s in1930s in1920s in1910s in1900s in1800s in1700s before_1700"`
	Type                        string                 `bson:"type" json:"type"`
	Keywords                    []string               `bson:"keywords" json:"keywords"`
	Tags                        []string               `bson:"tags" json:"tags"`
	Color                       string                 `bson:"color" json:"color"`
	Dynamic                     map[string]interface{} `bson:"dynamic" json:"dynamic"`
	DynamicType                 string                 `bson:"dynamic_type" json:"dynamic_type" validate:"oneof=accessories-and-jewelry art clothing furniture gifts home general"`
	HasVariations               bool                   `bson:"has_variations" json:"has_variations"`
	Sustainability              string                 `bson:"sustainability" json:"sustainability"`
	Personalization             bool                   `bson:"personalization" json:"personalization"`
	PersonalizationText         string                 `bson:"personalization_text" json:"personalization_text"`
	PersonalizationTextChars    int                    `bson:"personalization_text_chars" json:"personalization_text_chars"`
	PersonalizationTextOptional bool                   `bson:"personalization_text_optional" json:"personalization_text_optional"`
}

type ListingReview struct {
	Id           primitive.ObjectID `bson:"_id" json:"_id"`
	UserId       primitive.ObjectID `bson:"user_id" json:"user_id"`
	ShopId       primitive.ObjectID `bson:"shop_id" json:"shop_id"`
	Review       string             `bson:"review" json:"review"`
	ReviewAuthor string             `bson:"review_author" json:"review_author"`
	Thumbnail    string             `bson:"thumbnail" json:"thumbnail"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	Status       ShopReviewStatus   `bson:"status" json:"status" validate:"required,oneof=approved pending spam"`
}

type ListingVariation struct {
	VariationType string `json:"type"`
	Value         string `json:"value"`
	Price         string `json:"price"`
	Quantity      int    `json:"quantity"`
	SKU           string `json:"sku"`
	Unit          string `json:"unit"`
}

type ListingDiscountsPromotions struct {
	DiscountPercentage int    `json:"discount_percentage"`
	PromotionType      string `json:"promotion_type"`
	PromotionCode      string `json:"promotion_code"`
	ValidUntil         string `json:"valid_until"`
}

type Inventory struct {
	DomesticPricing bool      `bson:"domestic_pricing" json:"domestic_pricing"`
	DomesticPrice   float64   `bson:"domestic_price" json:"domestic_price"`
	Price           float64   `bson:"price" json:"price" validate:"required"`
	InitialQuantity int       `bson:"initial_quantity" json:"initial_quantity"`
	Quantity        int       `bson:"quantity" json:"quantity" validate:"required"`
	SKU             string    `bson:"sku" json:"sku" validate:"required"`
	CurrencyCode    string    `bson:"currency_code" json:"currency_code"`
	ModifiedAt      time.Time `bson:"modified_at" json:"modified_at"`
}

type ListingRating struct {
	Rating          float64 `json:"rating"`
	ReviewCount     int     `json:"review_count"`
	PositiveReviews int     `bson:"positive_reviews" json:"positive_reviews"`
	NegativeReviews int     `bson:"negative_reviews" json:"negative_reviews"`
}

type ListingFinancialInformation struct {
	TotalOrders     int     `bson:"total_orders" json:"total_orders"`
	Sales           float64 `bson:"sales" json:"sales"`
	OrdersCompleted int     `bson:"orders_completed" json:"orders_completed"`
	OrdersPending   int     `bson:"orders_pending" json:"orders_pending"`
	OrdersCanceled  int     `bson:"orders_canceled" json:"orders_canceled"`
	Revenue         float64 `bson:"revenue" json:"revenue"`
	Profit          float64 `bson:"profit" json:"profit"`
	ShippingRevenue float64 `bson:"shipping_revenue" json:"shipping_revenue"`
}

type NewListing struct {
	Inventory      Inventory          `json:"inventory" validate:"required"`
	Variations     []ListingVariation `json:"variations"`
	Processing     ListingProcessing  `json:"processing" validate:"required"`
	ListingDetails NewListingDetails  `json:"details"`
	Measurements   ListingMeasurement `json:"measurements"`
}

type NewListingDetails struct {
	Title                       string                 `json:"title" validate:"required,min=10,max=50"`
	Category                    ListingCategory        `json:"category" validate:"required"`
	Description                 string                 `json:"description" validate:"required,min=50,max=500"`
	WhoMade                     string                 `json:"who_made" validate:"oneof=i_did collective someone_else"`
	WhenMade                    string                 `json:"when_made"  validate:"oneof=in2020_2023 in2010_2019 in2003_2009 before_2003 in2000_2002 in1990s in1980s in1970s in1960s in1950s in1940s in1930s in1920s in1910s in1900s in1800s in1700s before_1700"`
	Type                        string                 `json:"type"validate:"required"`
	Keywords                    []string               `json:"keywords"`
	Tags                        []string               `json:"tags"`
	Personalization             bool                   `json:"personalization"`
	PersonalizationText         string                 `json:"personalization_text"`
	PersonalizationTextChars    int                    `json:"personalization_text_chars"`
	PersonalizationTextOptional bool                   `json:"personalization_text_optional"`
	Dynamic                     map[string]interface{} `json:"dynamic"`
	DynamicType                 string                 `json:"dynamic_type"`
	HasVariations               bool                   `json:"has_variations"`
	Condition                   string                 `json:"condition" validate:"oneof=new used refurbished"`
	Color                       string                 `json:"color"`
	OtherColor                  string                 `json:"otherColor"`
	ShippingProfileId           string                 `json:"shipping_profile_id"`
	Sustainability              string                 `json:"sustainability"`
}

type ClothListing struct {
	Fabric    string   `json:"fabric"`
	Size      string   `json:"size"`
	Scale     string   `json:"scale" validate:"oneof=EU US/CA"`
	Materials []string `json:"materials"`
}
