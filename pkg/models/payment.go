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
	BankName      string `json:"bankName" validate:"required"`
	AccountName   string `json:"accountName" validate:"required"`
	AccountNumber string `json:"accountNumber" validate:"required"`
	IsDefault     bool   `json:"isDefault"`
	IsOnboarding  bool   `json:"isOnboarding"`
}

// Used for Payment Card details.
type PaymentCardInformation struct {
	CreatedAt      time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt      time.Time          `bson:"updatedAt" json:"updatedAt"`
	CardHolderName string             `bson:"cardHolderName" json:"cardHolderName,omitempty" validate:"required"`
	CardNumber     string             `bson:"cardNumber" json:"_" validate:"required"`
	LastFourDigits string             `bson:"lastFourDigits" json:"lastFourDigits,omitempty" validate:"required"`
	ExpiryMonth    string             `bson:"expiryMonth" json:"expiryMonth,omitempty" validate:"required"`
	ExpiryYear     string             `bson:"expiryYear" json:"expiryYear,omitempty" validate:"required"`
	CVV            string             `bson:"cvv" json:"cvv,omitempty" validate:"required"`
	Company        string             `bson:"company" json:"company,omitempty" validate:"required"`
	ID             primitive.ObjectID `bson:"_id" json:"_id" validate:"required"`
	UserID         primitive.ObjectID `bson:"userId" json:"userId" validate:"required"`
	IsDefault      bool               `bson:"isDefault" json:"isDefault"`
}

type PaymentCardInformationRequest struct {
	CardHolderName string `json:"cardHolderName,omitempty" validate:"required"`
	CardNumber     string `json:"cardNumber,omitempty" validate:"required"`
	ExpiryMonth    string `json:"expiryMonth,omitempty" validate:"required"`
	ExpiryYear     string `json:"expiryYear,omitempty" validate:"required"`
	CVV            string `json:"cvv,omitempty" validate:"required"`
	IsDefault      bool   `json:"isDefault"`
}
