package models

import "strings"

// ParseDimensionUnit converts string dimension unit to DimensionUnit enum
func ParseDimensionUnit(unit string) DimensionUnit {
	switch strings.ToLower(unit) {
	case string(DimensionUnitINC):
		return DimensionUnitINC
	case string(DimensionUnitFT):
		return DimensionUnitFT
	case string(DimensionUnitMM):
		return DimensionUnitMM
	case string(DimensionUnitCM):
		return DimensionUnitCM
	case string(DimensionUnitM):
		return DimensionUnitM
	default:
		return DimensionUnitCM
	}
}

// ParseWeightUnit converts string weight unit to WeightUnit enum
func ParseWeightUnit(unit string) WeightUnit {
	switch strings.ToLower(unit) {
	case string(WeightUnitG):
		return WeightUnitG
	case string(WeightUnitKG):
		return WeightUnitKG
	case string(WeightUnitLB):
		return WeightUnitLB
	case string(WeightUnitOZ):
		return WeightUnitOZ
	default:
		return WeightUnitKG
	}
}