package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SellerVerification struct {
	ID                 primitive.ObjectID `bson:"_id" json:"_id"`
	ShopId             primitive.ObjectID `bson:"shop_id" json:"shop_id"`
	VerifyAs           string             `bson:"verify_as" json:"verify_as" validate:"oneof=IND ORG"`
	Nationality        string             `bson:"nationality" json:"nationality"`
	FirstName          string             `bson:"first_name" json:"first_name"`
	LastName           string             `bson:"last_name" json:"last_name"`
	DOB                string             `bson:"dob" json:"dob"`
	CountryOfResidence string             `bson:"country_of_residence" json:"country_of_residence"`
	Card               string             `bson:"card" json:"card" validate:"oneof=NIN DL PAP"`
	CardNumber         int                `bson:"card_number" json:"card_number" validate:"required"`
	CreatedAt          time.Time          `bson:"created_at" json:"created_at" validate:"required"`
	ModifiedAt         time.Time          `bson:"modified_at" json:"modified_at" validate:"required"`
	VerifiedAt         time.Time          `bson:"verified_at" json:"verified_at" validate:"required"`
	IsVerified         bool               `bson:"is_verified" json:"is_verified" validate:"required"`
}

type CreateSellerVerificationRequest struct {
	VerifyAs           string `bson:"verify_as" json:"verify_as" validate:"oneof=IND ORG"`
	Nationality        string `bson:"nationality" json:"nationality"`
	FirstName          string `bson:"first_name" json:"first_name"`
	LastName           string `bson:"last_name" json:"last_name"`
	DOB                string `bson:"dob" json:"dob"`
	CountryOfResidence string `bson:"country_of_residence" json:"country_of_residence"`
	Card               string `bson:"card" json:"card" validate:"oneof=NIN DL PAP"`
	CardNumber         int    `bson:"card_number" json:"card_number" validate:"required"`
}
