package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SellerVerification struct {
	CreatedAt          time.Time          `bson:"created_at" json:"createdAt" validate:"required"`
	VerifiedAt         time.Time          `bson:"verified_at" json:"verifiedAt" validate:"required"`
	ModifiedAt         time.Time          `bson:"modified_at" json:"modifiedAt" validate:"required"`
	DOB                string             `bson:"dob" json:"dob"`
	FirstName          string             `bson:"first_name" json:"firstName"`
	LastName           string             `bson:"last_name" json:"lastName"`
	CountryOfResidence string             `bson:"country_of_residence" json:"countryOfResidence"`
	Card               string             `bson:"card" json:"card" validate:"oneof=NIN DL PAP"`
	Nationality        string             `bson:"nationality" json:"nationality"`
	VerifyAs           string             `bson:"verify_as" json:"verifyAs" validate:"oneof=IND ORG"`
	CardNumber         int                `bson:"card_number" json:"cardNumber" validate:"required"`
	ID                 primitive.ObjectID `bson:"_id" json:"_id"`
	ShopId             primitive.ObjectID `bson:"shop_id" json:"shopId"`
	IsVerified         bool               `bson:"is_verified" json:"isVerified" validate:"required"`
}

type CreateSellerVerificationRequest struct {
	VerifyAs           string `json:"verifyAs" validate:"oneof=IND ORG"`
	Nationality        string `json:"nationality"`
	FirstName          string `json:"firstName"`
	LastName           string `json:"lastName"`
	DOB                string `json:"dob"`
	CountryOfResidence string `json:"countryOfResidence"`
	Card               string `json:"card" validate:"oneof=NIN DL PAP"`
	CardNumber         int    `json:"cardNumber" validate:"required"`
	IsOnboarding       bool   `json:"isOnboarding"`
}
