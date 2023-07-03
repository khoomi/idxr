package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type PaymentInformation struct {
	ID            primitive.ObjectID `bson:"_id" json:"_id" validate:"required"`
	UserID        primitive.ObjectID `bson:"user_id" json:"user_id" validate:"required"`
	AccountName   string             `bson:"account_name" json:"account_name" validate:"required"`
	AccountNumber string             `bson:"account_number" json:"account_number" validate:"required"`
	BankName      string             `bson:"bank_name" json:"bank_name" validate:"required"`
	IsDefault     string             `bson:"is_default" json:"is_default" validate:"required"`
}

type PaymentInformationRequest struct {
	AccountName   string `bson:"account_name" json:"account_name" validate:"required"`
	AccountNumber string `bson:"account_number" json:"account_number" validate:"required,min=10,max=10"`
	BankName      string `bson:"bank_name" json:"bank_name" validate:"required"`
	IsDefault     string `bson:"is_default" json:"is_default" validate:"required"`
}
