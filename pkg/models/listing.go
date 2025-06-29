package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Listing struct {
	Date                 ListingDateMeta      `bson:"date" json:"date"`
	State                ListingState         `bson:"state" json:"state"`
	MainImage            string               `bson:"main_image" json:"mainImage"`
	Code                 string               `bson:"code" json:"code"`
	Slug                 string               `bson:"slug" json:"slug"`
	Images               []string             `bson:"images" json:"images"`
	Variations           []Variation          `bson:"variations" json:"variations"`
	RecentReviews        []Review             `bson:"recent_reviews" json:"recentReviews"`
	Details              Details              `bson:"details" json:"details"`
	Measurements         Measurement          `bson:"measurements" json:"measurements"`
	Inventory            Inventory            `bson:"inventory" json:"inventory"`
	FinancialInformation FinancialInformation `bson:"financial_information" json:"financialInformation"`
	Rating               ListingRating        `bson:"rating" json:"rating"`
	Views                int                  `bson:"views" json:"views"`
	FavorersCount        int                  `bson:"favorers_count" json:"favorersCount"`
	ShippingProfileId    primitive.ObjectID   `bson:"shipping_profile_id" json:"shippingProfileId"`
	ID                   primitive.ObjectID   `bson:"_id" json:"_id"`
	ShopId               primitive.ObjectID   `bson:"shop_id" json:"shopId"`
	UserId               primitive.ObjectID   `bson:"user_id" json:"userId"`
	NonTaxable           bool                 `bson:"non_taxable" json:"nonTaxable"`
	ShouldAutoRenew      bool                 `bson:"should_auto_renew" json:"shouldAutoRenew"`
}

type Personalization struct {
	Text       string `bson:"text" json:"text"`
	Characters int    `bson:"characters" json:"characters"`
	Optional   bool   `bson:"optional" json:"optional"`
}

type Details struct {
	Dynamic        map[string]any `bson:"dynamic" json:"dynamic"`
	DynamicType    DynamicType    `bson:"dynamic_type" json:"dynamicType" validate:"oneof=accessories-and-jewelry art clothing furniture gifts home general"`
	Category       Category       `bson:"category" json:"category"`
	Sustainability string         `bson:"sustainability" json:"sustainability"`
	Description    string         `bson:"description" json:"description"`
	Condition      string         `bson:"condition" json:"condition" validate:"oneof=new used refurbished"`
	WhoMade        string         `bson:"who_made" json:"whoMade" validate:"oneof=i_did collective someone_else"`
	WhenMade       string         `bson:"when_made" json:"whenMade"  validate:"oneof=in2020_2023 in2010_2019 in2003_2009 before_2003 in2000_2002 in1990s in1980s in1970s in1960s in1950s in1940s in1930s in1920s in1910s in1900s in1800s in1700s before_1700"`
	Type           string         `bson:"type" json:"type"`
	Title          string         `bson:"title" json:"title"`
	Tags           []string       `bson:"tags" json:"tags"`
	Keywords       []string       `bson:"keywords" json:"keywords"`

	HasPersonalization bool            `bson:"has_personalization" json:"has_personalization"`
	Personalization    Personalization `json:"personalization"`

	AceessoriesAndJewelryData *AceessoriesAndJewelry `json:"-" bson:"accessories_and_jewelry_data,omitempty"`
	ClothingData              *Clothing              `json:"-" bson:"clothing_data,omitempty"`
	FurnitureData             *Furniture             `json:"-" bson:"furniture_data,omitempty"`
	GiftsAndOccasionsData     *GiftsAndOccasions     `json:"-" bson:"gifts_and_occasions_data,omitempty"`
	ArtAndCollectiblesData    *ArtAndCollectibles    `json:"-" bson:"art_and_collectibles_data,omitempty"`
	HomeAndLivingData         *HomeAndLiving         `json:"-" bson:"home_and_living_data,omitempty"`
}

type Review struct {
	CreatedAt    time.Time          `bson:"created_at" json:"createdAt"`
	Review       string             `bson:"review" json:"review"`
	ReviewAuthor string             `bson:"review_author" json:"reviewAuthor"`
	Thumbnail    string             `bson:"thumbnail" json:"thumbnail"`
	Status       ShopReviewStatus   `bson:"status" json:"status" validate:"required,oneof=approved pending spam"`
	Id           primitive.ObjectID `bson:"_id" json:"_id"`
	UserId       primitive.ObjectID `bson:"user_id" json:"userId"`
	ShopId       primitive.ObjectID `bson:"shop_id" json:"shopId"`
}

type Variation struct {
	VariationType string `json:"type"`
	Value         string `json:"value"`
	Price         string `json:"price"`
	Unit          string `json:"unit"`
	Quantity      int    `json:"quantity"`
}

type ListingDiscountsPromotions struct {
	PromotionType      string `json:"promotionType"`
	PromotionCode      string `json:"promotionCode"`
	ValidUntil         string `json:"validUntil"`
	DiscountPercentage int    `json:"discountPercentage"`
}

type Inventory struct {
	ModifiedAt      time.Time `bson:"modified_at" json:"modifiedAt"`
	SKU             string    `bson:"sku" json:"sku"`
	CurrencyCode    string    `bson:"currency_code" json:"currencyCode"`
	DomesticPrice   float64   `bson:"domestic_price" json:"domesticPrice"`
	Price           float64   `bson:"price" json:"price" validate:"required"`
	InitialQuantity int       `bson:"initial_quantity" json:"initialQuantity"`
	Quantity        int       `bson:"quantity" json:"quantity" validate:"required"`
	DomesticPricing bool      `bson:"domestic_pricing" json:"domesticPricing"`
}

type ListingRating struct {
	Rating          float64 `json:"rating"`
	ReviewCount     int     `json:"review_count"`
	PositiveReviews int     `bson:"positive_reviews" json:"positiveReviews"`
	NegativeReviews int     `bson:"negative_reviews" json:"negativeReviews"`
}

type FinancialInformation struct {
	TotalOrders     int     `bson:"total_orders" json:"totalOrders"`
	Sales           float64 `bson:"sales" json:"sales"`
	OrdersCompleted int     `bson:"orders_completed" json:"ordersCompleted"`
	OrdersPending   int     `bson:"orders_pending" json:"ordersPending"`
	OrdersCanceled  int     `bson:"orders_canceled" json:"ordersCanceled"`
	Revenue         float64 `bson:"revenue" json:"revenue"`
	Profit          float64 `bson:"profit" json:"profit"`
	ShippingRevenue float64 `bson:"shipping_revenue" json:"shippingRevenue"`
}

type ListingWithAnalytics struct {
	Listing
	Sales float64 `bson:"sales" json:"sales"`
}

type ListingsSummary struct {
	Date        ListingDateMeta    `bson:"date" json:"date"`
	State       ListingState       `bson:"state" json:"state"`
	Details     DetailsSummary     `bson:"details" json:"details"`
	Inventory   InventorySummary   `bson:"inventory" json:"inventory"`
	Slug        string             `bson:"slug" json:"slug"`
	Code        string             `bson:"code" json:"code"`
	MainImage   string             `bson:"main_image" json:"mainImage"`
	Images      []string           `bson:"images" json:"images"`
	Sales       float64            `bson:"sales" json:"sales"`
	TotalOrders int                `bson:"total_orders" json:"totalOrders"`
	ID          primitive.ObjectID `bson:"_id" json:"_id"`
	ShopId      primitive.ObjectID `bson:"shop_id" json:"shopId"`
	UserId      primitive.ObjectID `bson:"user_id" json:"userId"`
}

type DetailsSummary struct {
	Title    string   `bson:"title" json:"title"`
	Category Category `bson:"category" json:"category"`
}

type InventorySummary struct {
	Price    string `bson:"price" json:"price" validate:"required"`
	Quantity int    `bson:"quantity" json:"quantity" validate:"required"`
}

type ListingExtra struct {
	Date                 ListingDateMeta           `bson:"date" json:"date"`
	User                 ListingUserExcept         `bson:"user" json:"user"`
	State                ListingState              `bson:"state" json:"state"`
	MainImage            string                    `bson:"main_image" json:"mainImage"`
	Slug                 string                    `bson:"slug" json:"slug"`
	Shop                 ListingShopExcept         `bson:"shop" json:"shop"`
	Images               []string                  `bson:"images" json:"images"`
	RecentReviews        []Review                  `bson:"recent_reviews" json:"recentReviews"`
	Variations           []Variation               `bson:"variations" json:"variations"`
	Details              Details                   `bson:"details" json:"details"`
	Measurements         Measurement               `bson:"measurements" json:"measurements"`
	Inventory            Inventory                 `bson:"inventory" json:"inventory"`
	FinancialInformation FinancialInformation      `bson:"financial_information" json:"financialInformation"`
	Rating               ListingRating             `bson:"rating" json:"rating"`
	TotalOrders          int                       `bson:"total_orders" json:"totalOrders"`
	Sales                float64                   `bson:"sales" json:"sales"`
	FavorersCount        int                       `bson:"favorers_count" json:"favorersCount"`
	Views                int                       `bson:"views" json:"views"`
	ShippingProfileId    primitive.ObjectID        `bson:"shipping_profile_id" json:"shippingProfileId"`
	ID                   primitive.ObjectID        `bson:"_id" json:"_id"`
	ShopId               primitive.ObjectID        `bson:"shop_id" json:"shopId"`
	UserId               primitive.ObjectID        `bson:"user_id" json:"userId"`
	ShouldAutoRenew      bool                      `bson:"should_auto_renew" json:"shouldAutoRenew"`
	NonTaxable           bool                      `bson:"non_taxable" json:"nonTaxable"`
	Shipping             ShippingProfileForListing `bson:"shipping" json:"shipping"`
	Siblings             []Listing                 `json:"siblings"`
}

type ListingUserExcept struct {
	LoginName string `bson:"login_name" json:"loginName" validate:"required"`
	FirstName string `bson:"first_name" json:"firstName"`
	LastName  string `bson:"last_name" json:"lastName"`
	Thumbnail string `bson:"thumbnail" json:"thumbnail"`
}

type ListingShopExcept struct {
	Name         string     `bson:"name" json:"name" validate:"required"`
	Description  string     `bson:"description" json:"description" validate:"required"`
	Username     string     `bson:"username" json:"username" validate:"required"`
	Location     string     `bson:"location" json:"location"`
	Slug         string     `bson:"slug" json:"slug" validate:"required"`
	LogoURL      string     `bson:"logo_url" json:"logoUrl"`
	Rating       ShopRating `bson:"rating" json:"rating"`
	ReviewsCount int        `bson:"reviews_count" json:"reviewsCount"`
	CreatedAt    time.Time  `bson:"created_at" json:"createdAt" validate:"required"`
}

type NewListing struct {
	Details      NewListingDetails `json:"details"`
	Measurements Measurement       `json:"measurements"`
	Variations   []Variation       `json:"variations"`
	Inventory    Inventory         `json:"inventory" validate:"required"`
}

type NewListingDetails struct {
	Dynamic            map[string]any  `json:"dynamic"`
	DynamicType        DynamicType     `json:"dynamicType"`
	Category           Category        `json:"category" validate:"required"`
	Condition          string          `json:"condition" validate:"oneof=new used refurbished"`
	Title              string          `json:"title" validate:"required,min=10,max=50"`
	WhenMade           string          `json:"whenMade"`
	Type               string          `json:"type" validate:"required"`
	Sustainability     string          `json:"sustainability"`
	WhoMade            string          `json:"whoMade" validate:"oneof=i_did collective someone_else"`
	ShippingProfileId  string          `json:"shippingProfileId"`
	OtherColor         string          `json:"otherColor"`
	Color              string          `json:"color"`
	Description        string          `json:"description" validate:"required,min=50,max=500"`
	Tags               []string        `json:"tags"`
	Keywords           []string        `json:"keywords"`
	HasPersonalization bool            `json:"hasPersonalization"`
	Personalization    Personalization `json:"personalization"`

	AceessoriesAndJewelryData *AceessoriesAndJewelry `json:"-" bson:"accessories_and_jewelry_data,omitempty"`
	ClothingData              *Clothing              `json:"-" bson:"clothing_data,omitempty"`
	FurnitureData             *Furniture             `json:"-" bson:"furniture_data,omitempty"`
	GiftsAndOccasionsData     *GiftsAndOccasions     `json:"-" bson:"gifts_and_occasions_data,omitempty"`
	ArtAndCollectiblesData    *ArtAndCollectibles    `json:"-" bson:"art_and_collectibles_data,omitempty"`
	HomeAndLivingData         *HomeAndLiving         `json:"-" bson:"home_and_living_data,omitempty"`
}
