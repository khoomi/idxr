package models

import (
	"errors"
	"fmt"
	"time"
)

type ListingStateType string

const (
	ListingStateActive      ListingStateType = "active"
	ListingStateRemoved     ListingStateType = "removed"
	ListingStateSoldOut     ListingStateType = "soldout"
	ListingStateExpired     ListingStateType = "expired"
	ListingStateEdit        ListingStateType = "edit"
	ListingStateDraft       ListingStateType = "draft"
	ListingStatePrivate     ListingStateType = "private"
	ListingStateUnavailable ListingStateType = "unavailable"
	ListingStateDeactivated ListingStateType = "deactivated"
)

func (ListingStateType) ParseListingStateType(state string) (ListingStateType, error) {
	switch state {
	case "active":
		return ListingStateActive, nil
	case "removed":
		return ListingStateRemoved, nil
	case "soldout":
		return ListingStateSoldOut, nil
	case "expired":
		return ListingStateExpired, nil
	case "edit":
		return ListingStateEdit, nil
	case "private":
		return ListingStatePrivate, nil
	case "unavailable":
		return ListingStateUnavailable, nil
	case "deactivated":
		return ListingStateDeactivated, nil
	}

	err := fmt.Sprintf("Invalid listing state from request: %v", state)

	return ListingStateActive, errors.New(err)
}

type ListingState struct {
	StateUpdatedAt time.Time        `bson:"state_updated_at" json:"stateUpdatedAt"`
	State          ListingStateType `bson:"state" json:"state"`
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

type Measurement struct {
	ItemWeightUnit    WeightUnit    `bson:"item_weight_unit" json:"itemWeightUnit" validate:"oneof=oz g lb kg"`
	ItemDimensionUnit DimensionUnit `bson:"item_dimension_unit" json:"itemDimensionUnit" validate:"oneof=inc ft mm cm m"`
	ItemWeight        float64       `bson:"item_weight" json:"itemWeight"`
	ItemLength        float64       `bson:"item_length" json:"itemLength"`
	ItemWidth         float64       `bson:"item_width" json:"itemWidth"`
	ItemHeight        float64       `bson:"item_height" json:"itemHeight"`
}

type WhoMade string

const (
	WhoMadeIDid        = "i_did"
	WhoMadeCollective  = "collective"
	WhoMadeSomeoneElse = "someone_else"
)
