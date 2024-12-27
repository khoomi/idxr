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
	StateUpdatedAt time.Time        `bson:"state_updated_at" json:"stateUpdatedAt"`
}

type ListingDateMeta struct {
	CreatedAt  time.Time `bson:"created_at" json:"createdAt"`
	EndingAt   time.Time `bson:"ending_at" json:"endingAt"`
	ModifiedAt time.Time `bson:"modified_at" json:"modifiedAt"`
}

type ListingCategory struct {
	CategoryId   string `bson:"category_id" json:"categoryId"`
	CategoryName string `bson:"category_name" json:"categoryName"`
	CategoryPath string `bson:"category_path" json:"categoryPath"`
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
	ItemWeight        float64       `bson:"item_weight" json:"itemWeight"`
	ItemWeightUnit    WeightUnit    `bson:"item_weight_unit" json:"itemWeightUnit" validate:"oneof=oz g lb kg"`
	ItemLength        float64       `bson:"item_length" json:"itemLength"`
	ItemWidth         float64       `bson:"item_width" json:"itemWidth"`
	ItemHeight        float64       `bson:"item_height" json:"itemHeight"`
	ItemDimensionUnit DimensionUnit `bson:"item_dimension_unit" json:"itemDimensionUnit" validate:"oneof=inc ft mm cm m"`
}

type WhoMade string

const (
	WhoMadeIDid        = "i_did"
	WhoMadeCollective  = "collective"
	WhoMadeSomeoneElse = "someone_else"
)

type Listing struct {
	ID                   primitive.ObjectID          `bson:"_id" json:"_id"`
	Code                 string                      `bson:"code" json:"code"`
	State                ListingState                `bson:"state" json:"state"`
	UserId               primitive.ObjectID          `bson:"user_id" json:"userId"`
	ShopId               primitive.ObjectID          `bson:"shop_id" json:"shopId"`
	MainImage            string                      `bson:"main_image" json:"mainImage"`
	Images               []string                    `bson:"images" json:"images"`
	ListingDetails       ListingDetails              `bson:"details" json:"details"`
	Date                 ListingDateMeta             `bson:"date" json:"date"`
	Slug                 string                      `bson:"slug" json:"slug"`
	Views                int                         `bson:"views" json:"views"`
	FavorersCount        int                         `bson:"favorers_count" json:"favorersCount"`
	ShippingProfileId    primitive.ObjectID          `bson:"shipping_profile_id" json:"shippingProfileId"`
	NonTaxable           bool                        `bson:"non_taxable" json:"nonTaxable"`
	Variations           []ListingVariation          `bson:"variations" json:"variations"`
	ShouldAutoRenew      bool                        `bson:"should_auto_renew" json:"shouldAutoRenew"`
	Inventory            Inventory                   `bson:"inventory" json:"inventory"`
	RecentReviews        []ListingReview             `bson:"recent_reviews" json:"recentReviews"`
	Rating               ListingRating               `bson:"rating" json:"rating"`
	Measurements         ListingMeasurement          `bson:"measurements" json:"measurements"`
	FinancialInformation ListingFinancialInformation `bson:"financial_information" json:"financialInformation"`
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
	UserId               primitive.ObjectID          `bson:"user_id" json:"userId"`
	ShopId               primitive.ObjectID          `bson:"shop_id" json:"shopId"`
	MainImage            string                      `bson:"main_image" json:"mainImage"`
	Images               []string                    `bson:"images" json:"images"`
	ListingDetails       ListingDetails              `bson:"details" json:"details"`
	Date                 ListingDateMeta             `bson:"date" json:"date"`
	Slug                 string                      `bson:"slug" json:"slug"`
	Views                int                         `bson:"views" json:"views"`
	FavorersCount        int                         `bson:"favorers_count" json:"favorersCount"`
	ShippingProfileId    primitive.ObjectID          `bson:"shipping_profile_id" json:"shippingProfileId"`
	NonTaxable           bool                        `bson:"non_taxable" json:"nonTaxable"`
	Variations           []ListingVariation          `bson:"variations" json:"variations"`
	ShouldAutoRenew      bool                        `bson:"should_auto_renew" json:"shouldAutoRenew"`
	Inventory            Inventory                   `bson:"inventory" json:"inventory"`
	RecentReviews        []ListingReview             `bson:"recent_reviews" json:"recentReviews"`
	Rating               ListingRating               `bson:"rating" json:"rating"`
	TotalOrders          int                         `bson:"total_orders" json:"totalOrders"`
	Sales                float64                     `bson:"sales" json:"sales"`
	User                 ListingUserExcept           `bson:"user" json:"user"`
	Shop                 ListingShopExcept           `bson:"shop" json:"shop"`
	Measurements         ListingMeasurement          `bson:"measurements" json:"measurements"`
	FinancialInformation ListingFinancialInformation `bson:"financial_information" json:"financialInformation"`
}

type ListingUserExcept struct {
	LoginName string `bson:"login_name" json:"loginName" validate:"required"`
	FirstName string `bson:"first_name" json:"firstName"`
	LastName  string `bson:"last_name" json:"lastName"`
	Thumbnail string `bson:"thumbnail" json:"thumbnail"`
}

type ListingShopExcept struct {
	Name         string `bson:"name" json:"name" validate:"required"`
	Description  string `bson:"description" json:"description" validate:"required"`
	Username     string `bson:"username" json:"username" validate:"required"`
	Location     string `bson:"location" json:"location"`
	Slug         string `bson:"slug" json:"slug" validate:"required"`
	LogoURL      string `bson:"logo_url" json:"logoUrl"`
	ReviewsCount int    `bson:"reviews_count" json:"reviewsCount"`
}

type ListingDetails struct {
	Title                       string                 `bson:"title" json:"title"`
	Description                 string                 `bson:"description" json:"description"`
	Condition                   string                 `bson:"condition" json:"condition" validate:"oneof=new used refurbished"`
	Category                    ListingCategory        `bson:"category" json:"category"`
	WhoMade                     string                 `bson:"who_made" json:"whoMade" validate:"oneof=i_did collective someone_else"`
	WhenMade                    string                 `bson:"when_made" json:"whenMade"  validate:"oneof=in2020_2023 in2010_2019 in2003_2009 before_2003 in2000_2002 in1990s in1980s in1970s in1960s in1950s in1940s in1930s in1920s in1910s in1900s in1800s in1700s before_1700"`
	Type                        string                 `bson:"type" json:"type"`
	Keywords                    []string               `bson:"keywords" json:"keywords"`
	Tags                        []string               `bson:"tags" json:"tags"`
	Color                       string                 `bson:"color" json:"color"`
	Dynamic                     map[string]interface{} `bson:"dynamic" json:"dynamic"`
	DynamicType                 string                 `bson:"dynamic_type" json:"dynamicType" validate:"oneof=accessories-and-jewelry art clothing furniture gifts home general"`
	HasVariations               bool                   `bson:"has_variations" json:"hasVariations"`
	Sustainability              string                 `bson:"sustainability" json:"sustainability"`
	Personalization             bool                   `bson:"personalization" json:"personalization"`
	PersonalizationText         string                 `bson:"personalization_text" json:"personalizationText"`
	PersonalizationTextChars    int                    `bson:"personalization_text_chars" json:"personalizationTextChars"`
	PersonalizationTextOptional bool                   `bson:"personalization_text_optional" json:"personalizationTextOptional"`
}

type ListingReview struct {
	Id           primitive.ObjectID `bson:"_id" json:"_id"`
	UserId       primitive.ObjectID `bson:"user_id" json:"userId"`
	ShopId       primitive.ObjectID `bson:"shop_id" json:"shopId"`
	Review       string             `bson:"review" json:"review"`
	ReviewAuthor string             `bson:"review_author" json:"reviewAuthor"`
	Thumbnail    string             `bson:"thumbnail" json:"thumbnail"`
	CreatedAt    time.Time          `bson:"created_at" json:"createdAt"`
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
	DiscountPercentage int    `json:"discountPercentage"`
	PromotionType      string `json:"promotionType"`
	PromotionCode      string `json:"promotionCode"`
	ValidUntil         string `json:"validUntil"`
}

type Inventory struct {
	DomesticPricing bool      `bson:"domestic_pricing" json:"domesticPricing"`
	DomesticPrice   float64   `bson:"domestic_price" json:"domesticPrice"`
	Price           float64   `bson:"price" json:"price" validate:"required"`
	InitialQuantity int       `bson:"initial_quantity" json:"initialQuantity"`
	Quantity        int       `bson:"quantity" json:"quantity" validate:"required"`
	SKU             string    `bson:"sku" json:"sku" validate:"required"`
	CurrencyCode    string    `bson:"currency_code" json:"currencyCode"`
	ModifiedAt      time.Time `bson:"modified_at" json:"modifiedAt"`
}

type ListingRating struct {
	Rating          float64 `json:"rating"`
	ReviewCount     int     `json:"review_count"`
	PositiveReviews int     `bson:"positive_reviews" json:"positiveReviews"`
	NegativeReviews int     `bson:"negative_reviews" json:"negativeReviews"`
}

type ListingFinancialInformation struct {
	TotalOrders     int     `bson:"total_orders" json:"totalOrders"`
	Sales           float64 `bson:"sales" json:"sales"`
	OrdersCompleted int     `bson:"orders_completed" json:"ordersCompleted"`
	OrdersPending   int     `bson:"orders_pending" json:"ordersPending"`
	OrdersCanceled  int     `bson:"orders_canceled" json:"ordersCanceled"`
	Revenue         float64 `bson:"revenue" json:"revenue"`
	Profit          float64 `bson:"profit" json:"profit"`
	ShippingRevenue float64 `bson:"shipping_revenue" json:"shippingRevenue"`
}

type NewListing struct {
	Inventory      Inventory          `json:"inventory" validate:"required"`
	Variations     []ListingVariation `json:"variations"`
	ListingDetails NewListingDetails  `json:"details"`
	Measurements   ListingMeasurement `json:"measurements"`
}

type NewListingDetails struct {
	Title                       string                 `json:"title" validate:"required,min=10,max=50"`
	Category                    ListingCategory        `json:"category" validate:"required"`
	Description                 string                 `json:"description" validate:"required,min=50,max=500"`
	WhoMade                     string                 `json:"whoMade" validate:"oneof=i_did collective someone_else"`
	WhenMade                    string                 `json:"whenMade"  validate:"oneof=in2020_2023 in2010_2019 in2003_2009 before_2003 in2000_2002 in1990s in1980s in1970s in1960s in1950s in1940s in1930s in1920s in1910s in1900s in1800s in1700s before_1700"`
	Type                        string                 `json:"type"validate:"required"`
	Keywords                    []string               `json:"keywords"`
	Tags                        []string               `json:"tags"`
	Personalization             bool                   `json:"personalization"`
	PersonalizationText         string                 `json:"personalizationText"`
	PersonalizationTextChars    int                    `json:"personalizationTextChars"`
	PersonalizationTextOptional bool                   `json:"personalizationTextOptional"`
	Dynamic                     map[string]interface{} `json:"dynamic"`
	DynamicType                 string                 `json:"dynamic_type"`
	HasVariations               bool                   `json:"has_variations"`
	Condition                   string                 `json:"condition" validate:"oneof=new used refurbished"`
	Color                       string                 `json:"color"`
	OtherColor                  string                 `json:"otherColor"`
	ShippingProfileId           string                 `json:"shippingProfileId"`
	Sustainability              string                 `json:"sustainability"`
}

type ClothListing struct {
	Fabric    string   `json:"fabric"`
	Size      string   `json:"size"`
	Scale     string   `json:"scale" validate:"oneof=EU US/CA"`
	Materials []string `json:"materials"`
}
