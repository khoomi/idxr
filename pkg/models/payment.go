package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type PaymentInformation struct {
	ID            primitive.ObjectID `bson:"_id" json:"_id" validate:"required"`
	UserID        primitive.ObjectID `bson:"user_id" json:"userId" validate:"required"`
	AccountName   string             `bson:"account_name" json:"accountName" validate:"required"`
	AccountNumber string             `bson:"account_number" json:"accountNumber" validate:"required"`
	BankName      string             `bson:"bank_name" json:"bankName" validate:"required"`
	IsDefault     bool               `bson:"is_default" json:"isDefault" validate:"required"`
}

type PaymentInformationRequest struct {
	AccountName   string `bson:"account_name" json:"accountName" validate:"required"`
	AccountNumber string `bson:"account_number" json:"accountNumber" validate:"required,min=10,max=10"`
	BankName      string `bson:"bank_name" json:"bankName" validate:"required"`
	IsDefault     bool   `bson:"is_default" json:"isDefault" validate:"required"`
}
