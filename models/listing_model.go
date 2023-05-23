package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type ListingStateType string

const (
	// ListingStateActive The Listing is currently for sale.
	ListingStateActive ListingStateType = "active"
	// ListingStateRemoved The Listing has been removed by its owner.
	ListingStateRemoved ListingStateType = "remove"
	// ListingStateSoldOut The Listing has sold out. *Sold out listings can be edited, but active=true will only be honored if renew=true is also passed. Users will be billed for this action. Otherwise, the listing will remain in the sold_out state. Note that when editing a sold out listing, you will need to update the quantity to a value greater than zero.
	ListingStateSoldOut ListingStateType = "soldout"
	// ListingStateExpired The Listing has expired. **Expired listings can be edited, but active=true will only be honored if renew=true is also passed. Users will be billed for this action. Otherwise, the listing will remain in the expired state.
	ListingStateExpired ListingStateType = "expired"
	// ListingStateEdit The Listing is inactive. (For legacy reasons, this displays as "edit".)
	ListingStateEdit ListingStateType = "edit"
	// ListingStateDraft Draft listings are listings that have been saved prior to their first activation. The API can create draft listings, and also make draft listings active, but note that a listing in any other state cannot be moved to draft, nor can a draft listing be moved to any // state other than active.
	ListingStateDraft ListingStateType = "draft"
	// ListingStatePrivate The owner of the Listing has requested that it not appear in API results.
	ListingStatePrivate ListingStateType = "private"
	// ListingStateUnavailable The Listing has been removed by Etsy admin for unspecified reasons. Listings in this state may be missing some information which is // normally required.
	ListingStateUnavailable ListingStateType = "unavailable"
)

type ListingDateMeta struct {
	//  Creation time.
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	//  The listing's expiration date and time, in epoch seconds.
	EndingAt time.Time `bson:"ending_at" json:"ending_at"`
	//  The date and time the listing was originally posted, in epoch seconds.
	OriginallyCreatedAt time.Time `bson:"originally_created_at" json:"originally_created_at"`
	//  The date and time the listing was updated, in epoch seconds.
	ModifiedAt time.Time `bson:"modified_at" json:"modified_at"`
}

type ListingCategory struct {
	CategoryId            string `bson:"category_id" json:"category_id"`
	CategoryName          string `bson:"category_name" json:"category_name"`
	CategoryPath          string `bson:"category_path" json:"category_path"`
	SuggestedCategoryId   string `bson:"suggested_category_id" json:"suggested_category_id"`
	SuggestedCategoryPath string `bson:"suggested_category_path" json:"suggested_category_path"`
}

type ListingState struct {
	//  One of active, removed, sold_out, expired, alchemy, edit, create, private, or unavailable.
	State ListingStateType `bson:"state" json:"state"`
	//  The time at which the listing last changed state.
	StateUpdatedAt time.Time `bson:"state_updated_at" json:"state_updated_at"`
}

type ListingProcessing struct {
	ProcessingMin int `bson:"processing_min" json:"processing_min"`
	ProcessingMax int `bson:"processing_max" json:"processing_max"`
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
	ItemWeight int `bson:"item_weight" json:"item_weight"`
	// The units used to represent the weight of this item.
	ItemWeightUnit WeightUnit `bson:"item_weight_unit" json:"item_weight_unit" validate:"oneof=oz g lb kg"`
	//  How long the item is.
	ItemLength int `bson:"item_length" json:"item_length"`
	//  How wide the item is.
	ItemWidth int `bson:"item_width" json:"item_width"`
	//  How tall the item is.
	ItemHeight int `bson:"item_height" json:"item_height"`
	// The units used to represent the dimensions of this item.
	ItemDimensionUnit DimensionUnit `bson:"item_dimension_unit" json:"item_dimension_unit" validate:"oneof=inc ft mm cm m"`
}

type WhoMade string

const (
	WhoMadeIDid        = "i_did"
	WhoMadeCollective  = "collective"
	WhoMadeSomeoneElse = "someone_else"
)

type WhenMade string

const (
	WhenMadeIn2020_2022 = "in2020_2022"
	WhenMadeIn2010_2019 = "in2010_2019"
	WhenMadeIn2003_2009 = "in2003_2009"
	WhenMadebefore_2003 = "before_2003"
	WhenMadeIn2000_2002 = "in2000_2002"
	WhenMadeIn1990s     = "in1990s"
	WhenMadeIn1980s     = "in1980s"
	WhenMadeIn1970s     = "in1970s"
	WhenMadeIn1960s     = "in1960s"
	WhenMadeIn1950s     = "in1950s"
	WhenMadeIn1940s     = "in1940s"
	WhenMadeIn1930s     = "in1930s"
	WhenMadeIn1920s     = "in1920s"
	WhenMadeIn1910s     = "in1910s"
	WhenMadeIn1900s     = "in1900s"
	WhenMadeIn1800s     = "in1800s"
	WhenMadeIn1700s     = "in1700s"
	WhenMadebefore_1700 = "before_1700"
)

type Listing struct {
	// The listing's ObjectId
	ID primitive.ObjectID `bson:"_id" json:"_id"`
	//  Current state of this listing.
	State ListingState `bson:"state" json:"state"`
	//  The Object ID of the user who posted the item.
	UserId primitive.ObjectID `bson:"user_id" json:"user_id"`
	//  The listing's title. This string is valid if it does not match the following pattern: /[^\p{L}\p{Nd}\p{P}\p{Sm}\p{Zs}™©®]/u. The characters %, :, & and + can only be used once each.
	Title primitive.ObjectID `bson:"title" json:"title"`
	//  A description of the item.
	Description string `bson:"description" json:"description"`
	// The main image associated with this Listing
	MainImage string `bson:"main_image" json:"main_image"`
	// An array of images for the listing, can include up to 10 images.
	Images []string `bson:"images" json:"images"`
	// All date related data for this listing.
	Date ListingDateMeta `bson:"date" json:"date"`
	//  The item's price (will be treated as private for sold listings).
	Price float64 `bson:"price" json:"price"`
	//  The ISO (alphabetic) code for the item's currency.
	CurrencyCode string `bson:"currency_code" json:"currency_code"`
	//  The quantity of this item available for sale.
	Quantity int `bson:"quantity" json:"quantity"`
	//  A list of distinct SKUs applied to a listing.
	SKU string `bson:"sku" json:"sku"`
	//  A list of tags for the item. A tag is valid if it does not match the pattern: /[^\p{L}\p{Nd}\p{Zs}\-'™©®]/u
	Tags string `bson:"tags" json:"tags"`
	// Category attached to this listing.
	Category ListingCategory `bson:"category" json:"category"`
	// A list of materials used in the item. A material is valid if it does not match the pattern: /[^\p{L}\p{Nd}\p{Zs}]/u
	Materials []string `bson:"materials" json:"materials"`
	//  The full URL to the listing's page on Khoomi.
	Slug string `bson:"slug" json:"slug"`
	// The number of times the listing has been viewed on Khoomi.com (does not include API views).
	Views int `bson:"views" json:"views"`
	//  The number of members who've marked this Listing as a favorite
	FavorersCount int `bson:"favorers_count" json:"favorers_count"`
	// The Object ID of the shipping template associated with the listing.
	ShippingProfileId primitive.ObjectID `bson:"shipping_profile_id" json:"shipping_profile_id"`
	// Days for processing this listing.
	Processing ListingProcessing `bson:"processing" json:"processing"`
	// True if the listing is a supply.
	IsSupply bool `bson:"is_supply" json:"is_supply"`
	//  Who made the item being listed.
	WhoMade bool `bson:"who_made" json:"who_made" validate:"oneof=i_did collective someone_else"`
	// True if the listing is a supply.
	WhenMade bool `bson:"when_made" json:"when_made"  validate:"oneof=in2020_2023 in2010_2019 in2003_2009 before_2003 in2000_2002 in1990s in1980s in1970s in1960s in1950s in1940s in1930s in1920s in1910s in1900s in1800s in1700s before_1700"`
	// wight, len, etc of this listing.
	Measurement ListingMeasurement `bson:"measurement" json:"measurement"`
	//  If this flag is true, any applicable shop tax rates will not be applied to this listing on checkout.
	NonTaxable bool `bson:"non_taxable" json:"non_taxable"`
	//  If this flag is true, a buyer may contact the seller for a customized order. Can only be set when the shop accepts custom orders and defaults to true in that case.
	IsCustomizable bool `bson:"is_customizable" json:"is_customizable"`
	//  True if variations are available for this Listing.
	HasVariations bool `bson:"has_variations" json:"has_variations"`
	//  True if this listing has been set to automatic renewals.
	ShouldAutoRenew bool `bson:"should_auto_renew" json:"should_auto_renew"`
}
