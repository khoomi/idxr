package services

import (
	"context"
	"errors"

	"khoomi-api-io/api/pkg/util"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

// TransactionCallback defines the callback function for database transactions
type TransactionCallback func(ctx mongo.SessionContext) (any, error)

// ExecuteTransaction executes a database transaction with proper error handling
func ExecuteTransaction(ctx context.Context, callback TransactionCallback) (any, error) {
	wc := writeconcern.New(writeconcern.WMajority())
	txnOptions := options.Transaction().SetWriteConcern(wc)
	session, err := util.DB.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)

	result, err := session.WithTransaction(ctx, callback, txnOptions)
	if err != nil {
		return nil, err
	}

	if err := session.CommitTransaction(ctx); err != nil {
		return nil, err
	}

	return result, nil
}

// SetOtherRecordsToFalse sets a boolean field to false for other records belonging to a user
// This is commonly used for default settings (addresses, payment methods, etc.)
func SetOtherRecordsToFalse(ctx context.Context, collection *mongo.Collection, userFieldName string, userID primitive.ObjectID, recordID primitive.ObjectID, boolFieldName string) error {
	filter := bson.M{
		userFieldName: userID,
		"_id":         bson.M{"$ne": recordID},
		boolFieldName: true,
	}

	update := bson.M{
		"$set": bson.M{boolFieldName: false},
	}

	_, err := collection.UpdateMany(ctx, filter, update)
	return err
}

// CheckRecordLimit checks if a user has reached a specified limit for a collection
func CheckRecordLimit(ctx context.Context, collection *mongo.Collection, userFieldName string, userID primitive.ObjectID, limit int64, errorMessage string) error {
	filter := bson.M{userFieldName: userID}
	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return err
	}

	if count >= limit {
		return errors.New(errorMessage)
	}

	return nil
}

// CommonCollectionLimits defines common limits used across services
const (
	MaxUserAddresses     = 5
	MaxPaymentMethods    = 5
	MaxPaymentCards      = 5
)

// Common error messages
const (
	ErrMaxAddressesReached     = "max allowed addresses reached. please delete other address to accommodate a new one"
	ErrMaxPaymentInfoReached   = "Max allowed payment information reached. Please delete other payment information to accommodate a new one."
	ErrMaxPaymentCardsReached  = "Max allowed payment cards reached. Please delete other cards to accommodate a new one."
)

// Helper functions for common operations

// SetOtherAddressesToFalse sets other addresses' default shipping flag to false
func SetOtherAddressesToFalse(ctx context.Context, collection *mongo.Collection, userID, addressID primitive.ObjectID) error {
	return SetOtherRecordsToFalse(ctx, collection, "user_id", userID, addressID, "is_default_shipping_address")
}

// SetOtherPaymentMethodsToFalse sets other payment methods' default flag to false
func SetOtherPaymentMethodsToFalse(ctx context.Context, collection *mongo.Collection, userID, paymentID primitive.ObjectID, userFieldName string) error {
	return SetOtherRecordsToFalse(ctx, collection, userFieldName, userID, paymentID, "is_default")
}

// CheckUserAddressLimit checks if user has reached address limit
func CheckUserAddressLimit(ctx context.Context, collection *mongo.Collection, userID primitive.ObjectID) error {
	return CheckRecordLimit(ctx, collection, "user_id", userID, MaxUserAddresses, ErrMaxAddressesReached)
}

// CheckPaymentMethodLimit checks if user has reached payment method limit
func CheckPaymentMethodLimit(ctx context.Context, collection *mongo.Collection, userFieldName string, userID primitive.ObjectID, errorMessage string) error {
	return CheckRecordLimit(ctx, collection, userFieldName, userID, MaxPaymentMethods, errorMessage)
}