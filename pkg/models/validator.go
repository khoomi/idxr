package models

import (
	"encoding/json"
	"fmt"
	"log"
)

// Validate if submitted listing is a dynamic type.
func IsValidDynamicType(dt DynamicType) bool {
	switch dt {
	case FurnitureType,
		GiftsAndOccasionsType,
		ClothingType,
		ArtAndCollectiblesType,
		AceessoriesAndJewelryType,
		HomeAndLivingType:
		return true
	default:
		return false
	}
}

func (n *NewListingDetails) ParseDynamicData() (any, error) {
	if n.Dynamic == nil {
		return nil, fmt.Errorf("dynamic field is nil")
	}

	raw, err := json.Marshal(n.Dynamic)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal dynamic map: %w", err)
	}
	log.Println(n.DynamicType)

	switch n.DynamicType {
	case FurnitureType:
		var v Furniture
		if err := json.Unmarshal(raw, &v); err != nil {
			return nil, err
		}
		return &v, nil

	case ClothingType:
		var v Clothing
		if err := json.Unmarshal(raw, &v); err != nil {
			return nil, err
		}
		return &v, nil

	case GiftsAndOccasionsType:
		var v GiftsAndOccasions
		if err := json.Unmarshal(raw, &v); err != nil {
			return nil, err
		}
		return &v, nil

	case ArtAndCollectiblesType:
		var v ArtAndCollectibles
		if err := json.Unmarshal(raw, &v); err != nil {
			return nil, err
		}
		return &v, nil

	case AceessoriesAndJewelryType:
		var v AceessoriesAndJewelry
		if err := json.Unmarshal(raw, &v); err != nil {
			return nil, err
		}
		return &v, nil

	case HomeAndLivingType:
		var v HomeAndLiving
		if err := json.Unmarshal(raw, &v); err != nil {
			return nil, err
		}
		return &v, nil

	default:
		return nil, fmt.Errorf("unknown dynamic type: %s", n.DynamicType)
	}
}

func (n *NewListingDetails) SetDynamicToTypedField() error {
	typed, err := n.ParseDynamicData()
	if err != nil {
		return err
	}

	switch v := typed.(type) {
	case *Clothing:
		n.ClothingData = v
	case *Furniture:
		n.FurnitureData = v
	case *GiftsAndOccasions:
		n.GiftsAndOccasionsData = v
	case *ArtAndCollectibles:
		n.ArtAndCollectiblesData = v
	case *AceessoriesAndJewelry:
		n.AceessoriesAndJewelryData = v
	case *HomeAndLiving:
		n.HomeAndLivingData = v
	default:
		return fmt.Errorf("unsupported dynamic type: %T", v)
	}

	return nil
}
