package models

import (
	"encoding/json"
	"fmt"
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
	Details              Details              `bson:"details" json:"details"`
	Inventory            Inventory            `bson:"inventory" json:"inventory"`
	FinancialInformation FinancialInformation `bson:"financial_information" json:"financialInformation"`
	Rating               Rating               `bson:"rating" json:"rating"`
	Views                int                  `bson:"views" json:"views"`
	FavorersCount        int                  `bson:"favorers_count" json:"favorersCount"`
	ShippingProfileId    primitive.ObjectID   `bson:"shipping_profile_id" json:"shippingProfileId"`
	ID                   primitive.ObjectID   `bson:"_id" json:"_id"`
	ShopId               primitive.ObjectID   `bson:"shop_id" json:"shopId"`
	UserId               primitive.ObjectID   `bson:"user_id" json:"userId"`
	NonTaxable           bool                 `bson:"non_taxable" json:"nonTaxable"`
	ShouldAutoRenew      bool                 `bson:"should_auto_renew" json:"shouldAutoRenew"`
	HasVariations        bool                 `bson:"has_variations" json:"hasVariations"`
}

type Personalization struct {
	Text       string `bson:"text" json:"text"`
	Characters int    `bson:"characters" json:"characters"`
	Optional   bool   `bson:"optional" json:"optional"`
}

type Details struct {
	Dynamic            map[string]any  `bson:"dynamic" json:"dynamic"`
	DynamicType        DynamicType     `bson:"dynamic_type" json:"dynamicType" validate:"oneof=accessories-and-jewelry art clothing furniture gifts home general"`
	Category           Category        `bson:"category" json:"category"`
	Sustainability     string          `bson:"sustainability" json:"sustainability"`
	Description        string          `bson:"description" json:"description"`
	Condition          string          `bson:"condition" json:"condition" validate:"oneof=new used refurbished"`
	WhoMade            string          `bson:"who_made" json:"whoMade" validate:"oneof=i_did collective someone_else"`
	WhenMade           string          `bson:"when_made" json:"whenMade"  validate:"oneof=made_to_order in2020_2023 in2010_2019 in2003_2009 before_2003 in2000_2002 in1990s in1980s in1970s in1960s in1950s in1940s in1930s in1920s in1910s in1900s in1800s in1700s before_1700"`
	Type               string          `bson:"type" json:"type"`
	Title              string          `bson:"title" json:"title"`
	Tags               []string        `bson:"tags" json:"tags"`
	Keywords           []string        `bson:"keywords" json:"keywords"`
	HasPersonalization bool            `bson:"has_personalization" json:"hasPersonalization"`
	Personalization    Personalization `bson:"personalization" json:"personalization"`

	AceessoriesAndJewelryData *AceessoriesAndJewelry `json:"-" bson:"accessories_and_jewelry_data,omitempty"`
	ClothingData              *Clothing              `json:"-" bson:"clothing_data,omitempty"`
	FurnitureData             *Furniture             `json:"-" bson:"furniture_data,omitempty"`
	GiftsAndOccasionsData     *GiftsAndOccasions     `json:"-" bson:"gifts_and_occasions_data,omitempty"`
	ArtAndCollectiblesData    *ArtAndCollectibles    `json:"-" bson:"art_and_collectibles_data,omitempty"`
	HomeAndLivingData         *HomeAndLiving         `json:"-" bson:"home_and_living_data,omitempty"`
}

type Variation struct {
	ID       string   `json:"id" bson:"id"`
	Name     string   `json:"name" bson:"name"`
	Value    string   `json:"value" bson:"value"`
	Quantity int      `json:"quantity" bson:"quantity"`
	Price    *float64 `json:"price,omitempty" bson:"price,omitempty"`
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

type ListingSummary struct {
	Date      ListingDateMeta    `bson:"date" json:"date"`
	State     ListingState       `bson:"state" json:"state"`
	Details   DetailsSummary     `bson:"details" json:"details"`
	Inventory InventorySummary   `bson:"inventory" json:"inventory"`
	Slug      string             `bson:"slug" json:"slug"`
	Code      string             `bson:"code" json:"code"`
	MainImage string             `bson:"main_image" json:"mainImage"`
	Images    []string           `bson:"images" json:"images"`
	ID        primitive.ObjectID `bson:"_id" json:"_id"`
	ShopId    primitive.ObjectID `bson:"shop_id" json:"shopId"`
	UserId    primitive.ObjectID `bson:"user_id" json:"userId"`
	Rating    Rating             `bson:"rating" json:"rating"`
	Views     int                `bson:"views" json:"reviews"`
	Shipping  *ShippingSummary   `bson:"shipping" json:"shipping"`
}

type ShopListingSummary struct {
	Date      ListingDateMeta    `bson:"date" json:"date"`
	State     ListingState       `bson:"state" json:"state"`
	Details   DetailsSummary     `bson:"details" json:"details"`
	Inventory InventorySummary   `bson:"inventory" json:"inventory"`
	Slug      string             `bson:"slug" json:"slug"`
	Code      string             `bson:"code" json:"code"`
	MainImage string             `bson:"main_image" json:"mainImage"`
	Images    []string           `bson:"images" json:"images"`
	ID        primitive.ObjectID `bson:"_id" json:"_id"`
	ShopId    primitive.ObjectID `bson:"shop_id" json:"shopId"`
	UserId    primitive.ObjectID `bson:"user_id" json:"userId"`
	Rating    Rating             `bson:"rating" json:"rating"`
	Views     int                `bson:"views" json:"reviews"`
	Shipping  *ShippingSummary   `bson:"shipping" json:"shipping"`
	Siblings  []Listing          `bson:"siblings" json:"siblings"`
	Shop      ShopExcerpt        `bson:"shop" json:"shop"`
	User      ListingUserExcept  `bson:"user" json:"user"`
}

type DetailsSummary struct {
	Title    string   `bson:"title" json:"title"`
	Category Category `bson:"category" json:"category"`
}

type InventorySummary struct {
	Price    float64 `bson:"price" json:"price" validate:"required"`
	Quantity int     `bson:"quantity" json:"quantity" validate:"required"`
}

type ShippingSummary struct {
	Processing         ShippingProcessing `bson:"processing" json:"processing"`
	OffersFreeShipping bool               `bson:"offers_free_shipping" json:"offers_free_shipping"`
	Destinations       []string           `bson:"destinations" json:"destinations"`
	MaaxDeliveryDays   int                `bson:"max_delivery_days" json:"max_delivery_days"`
}

type ListingExtra struct {
	Date                 ListingDateMeta           `bson:"date" json:"date"`
	User                 ListingUserExcept         `bson:"user" json:"user"`
	State                ListingState              `bson:"state" json:"state"`
	MainImage            string                    `bson:"main_image" json:"mainImage"`
	Slug                 string                    `bson:"slug" json:"slug"`
	Shop                 ListingShopExcept         `bson:"shop" json:"shop"`
	Images               []string                  `bson:"images" json:"images"`
	Variations           []Variation               `bson:"variations" json:"variations"`
	Details              Details                   `bson:"details" json:"details"`
	Inventory            Inventory                 `bson:"inventory" json:"inventory"`
	FinancialInformation FinancialInformation      `bson:"financial_information" json:"financialInformation"`
	Rating               Rating                    `bson:"rating" json:"rating"`
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
	Name               string    `bson:"name" json:"name" validate:"required"`
	Description        string    `bson:"description" json:"description" validate:"required"`
	Username           string    `bson:"username" json:"username" validate:"required"`
	Location           string    `bson:"location" json:"location"`
	Slug               string    `bson:"slug" json:"slug" validate:"required"`
	LogoURL            string    `bson:"logo_url" json:"logoUrl"`
	Rating             Rating    `bson:"rating" json:"rating"`
	CreatedAt          time.Time `bson:"created_at" json:"createdAt" validate:"required"`
	ListingActiveCount int64     `bson:"listing_active_count"  json:"listingActiveCount"`
}

type NewListing struct {
	Details      NewListingDetails `json:"details"`
	Variations   []Variation       `json:"variations"`
	Inventory    Inventory         `json:"inventory" validate:"required"`
	IsOnboarding bool              `json:"isOnboarding"`
}

type UpdateListing struct {
	Details    *UpdateListingDetails `json:"details,omitempty"`
	Variations []Variation           `json:"variations,omitempty"`
	Inventory  *UpdateInventory      `json:"inventory,omitempty"`
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

type UpdateListingDetails struct {
	Dynamic            map[string]any   `json:"dynamic,omitempty"`
	DynamicType        *DynamicType     `json:"dynamicType,omitempty"`
	Category           *Category        `json:"category,omitempty"`
	Condition          *string          `json:"condition,omitempty" validate:"omitempty,oneof=new used refurbished"`
	Title              *string          `json:"title,omitempty" validate:"omitempty,min=10,max=80"`
	WhenMade           *string          `json:"whenMade,omitempty"`
	Type               *string          `json:"type,omitempty"`
	Sustainability     *string          `json:"sustainability,omitempty"`
	WhoMade            *string          `json:"whoMade,omitempty" validate:"omitempty,oneof=i_did collective someone_else"`
	ShippingProfileId  *string          `json:"shippingProfileId,omitempty"`
	OtherColor         *string          `json:"otherColor,omitempty"`
	Color              *string          `json:"color,omitempty"`
	Description        *string          `json:"description,omitempty" validate:"omitempty,min=50,max=500"`
	Tags               []string         `json:"tags,omitempty"`
	Keywords           []string         `json:"keywords,omitempty"`
	HasPersonalization *bool            `json:"hasPersonalization,omitempty"`
	Personalization    *Personalization `json:"personalization,omitempty"`

	// Dynamic category data
	AceessoriesAndJewelryData *AceessoriesAndJewelry `json:"-" bson:"accessories_and_jewelry_data,omitempty"`
	ClothingData              *Clothing              `json:"-" bson:"clothing_data,omitempty"`
	FurnitureData             *Furniture             `json:"-" bson:"furniture_data,omitempty"`
	GiftsAndOccasionsData     *GiftsAndOccasions     `json:"-" bson:"gifts_and_occasions_data,omitempty"`
	ArtAndCollectiblesData    *ArtAndCollectibles    `json:"-" bson:"art_and_collectibles_data,omitempty"`
	HomeAndLivingData         *HomeAndLiving         `json:"-" bson:"home_and_living_data,omitempty"`
}

// UpdateInventory represents partial updates to listing inventory
type UpdateInventory struct {
	DomesticPricing *bool    `json:"domesticPricing,omitempty"`
	DomesticPrice   *float64 `json:"domesticPrice,omitempty"`
	Price           *float64 `json:"price,omitempty"`
	Quantity        *int     `json:"quantity,omitempty"`
	SKU             *string  `json:"sku,omitempty"`
}

func (uld *UpdateListingDetails) SetDynamicToTypedField() error {
	if uld.DynamicType == nil || uld.Dynamic == nil {
		return nil
	}

	switch *uld.DynamicType {
	case ClothingType:
		var data Clothing
		if err := convertMapToStruct(uld.Dynamic, &data); err != nil {
			return err
		}
		uld.ClothingData = &data
	case FurnitureType:
		var data Furniture
		if err := convertMapToStruct(uld.Dynamic, &data); err != nil {
			return err
		}
		uld.FurnitureData = &data
	case AccessoriesAndJewelryType:
		var data AceessoriesAndJewelry
		if err := convertMapToStruct(uld.Dynamic, &data); err != nil {
			return err
		}
		uld.AceessoriesAndJewelryData = &data
	case GiftsAndOccasionsType:
		var data GiftsAndOccasions
		if err := convertMapToStruct(uld.Dynamic, &data); err != nil {
			return err
		}
		uld.GiftsAndOccasionsData = &data
	case ArtAndCollectiblesType:
		var data ArtAndCollectibles
		if err := convertMapToStruct(uld.Dynamic, &data); err != nil {
			return err
		}
		uld.ArtAndCollectiblesData = &data
	case HomeAndLivingType:
		var data HomeAndLiving
		if err := convertMapToStruct(uld.Dynamic, &data); err != nil {
			return err
		}
		uld.HomeAndLivingData = &data
	}

	return nil
}

func convertMapToStruct(m map[string]any, v any) error {
	data, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to marshal map: %w", err)
	}

	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal to struct: %w", err)
	}

	return nil
}
