package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SellerPaymentInformation struct {
	CreatedAt     time.Time          `bson:"created_at" json:"createdAt,omitempty"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updatedAt,omitempty"`
	AccountName   string             `bson:"account_name" json:"accountName" validate:"required"`
	AccountNumber string             `bson:"account_number" json:"accountNumber" validate:"required,min=10,max=10"`
	BankName      string             `bson:"bank_name" json:"bankName" validate:"required"`
	ID            primitive.ObjectID `bson:"_id" json:"_id" validate:"required"`
	UserID        primitive.ObjectID `bson:"user_id" json:"userId" validate:"required"`
	IsDefault     bool               `bson:"is_default" json:"isDefault"`
}

// Used when a seller submits new or updated bank account info.
type SellerPaymentInformationRequest struct {
	AccountName   string `bson:"account_name" json:"accountName" validate:"required"`
	AccountNumber string `bson:"account_number" json:"accountNumber" validate:"required,min=10,max=10"`
	BankName      string `bson:"bank_name" json:"bankName" validate:"required"`
	IsDefault     bool   `bson:"is_default" json:"isDefault"`
}

type BuyerPaymentInformation struct {
	CreatedAt      time.Time          `bson:"created_at" json:"createdAt,omitempty"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updatedAt,omitempty"`
	CardHolderName string             `bson:"card_holder_name" json:"cardHolderName,omitempty" validate:"required"`
	CardNumber     string             `bson:"card_number" json:"cardNumber,omitempty" validate:"required,credit_card"`
	ExpiryMonth    string             `bson:"expiry_month" json:"expiryMonth,omitempty" validate:"required,len=2,numeric"`
	ExpiryYear     string             `bson:"expiry_year" json:"expiryYear,omitempty" validate:"required,len=4,numeric"`
	CVV            string             `bson:"cvv" json:"cvv,omitempty" validate:"required,min=3,max=4,numeric"`
	ID             primitive.ObjectID `bson:"_id" json:"_id" validate:"required"`
	UserID         primitive.ObjectID `bson:"user_id" json:"userId" validate:"required"`
	IsDefault      bool               `bson:"is_default" json:"isDefault"`
}

type BuyerPaymentInformationRequest struct {
	CardHolderName string `bson:"card_holder_name" json:"cardHolderName,omitempty" validate:"required"`
	CardNumber     string `bson:"card_number" json:"cardNumber,omitempty" validate:"required,credit_card"`
	ExpiryMonth    string `bson:"expiry_month" json:"expiryMonth,omitempty" validate:"required,len=2,numeric"`
	ExpiryYear     string `bson:"expiry_year" json:"expiryYear,omitempty" validate:"required,len=4,numeric"`
	CVV            string `bson:"cvv" json:"cvv,omitempty" validate:"required,min=3,max=4,numeric"`
	IsDefault      bool   `bson:"is_default" json:"isDefault"`
}
