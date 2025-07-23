package services

import (
	"context"
	"errors"

	"khoomi-api-io/api/internal/common"
	"khoomi-api-io/api/pkg/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// UserAddressServiceImpl implements the UserAddressService interface
type UserAddressServiceImpl struct{}

// NewUserAddressService creates a new instance of UserAddressService
func NewUserAddressService() UserAddressService {
	return &UserAddressServiceImpl{}
}

// CreateUserAddress creates a new user address
func (uas *UserAddressServiceImpl) CreateUserAddress(ctx context.Context, userID primitive.ObjectID, address models.UserAddressExcerpt) (primitive.ObjectID, error) {
	// Check if user has reached maximum address limit
	err := CheckRecordLimit(ctx, common.UserAddressCollection, "user_id", userID, 10, "maximum address limit reached")
	if err != nil {
		return primitive.NilObjectID, err
	}

	addressID := primitive.NewObjectID()
	userAddress := models.UserAddress{
		Id:         addressID,
		UserId:     userID,
		City:       address.City,
		State:      address.State,
		Street:     address.Street,
		PostalCode: address.PostalCode,
		Country:    models.CountryNigeria,
		IsDefault:  address.IsDefault,
	}

	callback := func(ctx mongo.SessionContext) (any, error) {
		// If this is set as default, update other addresses
		if address.IsDefault {
			err = SetOtherRecordsToFalse(ctx, common.UserAddressCollection, "user_id", userID, addressID, "is_default_shipping_address")
			if err != nil {
				return nil, err
			}
		}

		insertResult, err := common.UserAddressCollection.InsertOne(ctx, userAddress)
		if err != nil {
			return nil, err
		}

		return insertResult, nil
	}

	_, err = ExecuteTransaction(ctx, callback)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return addressID, nil
}

// GetUserAddresses retrieves user addresses with authorization check
func (uas *UserAddressServiceImpl) GetUserAddresses(ctx context.Context, authenticatedUserID, targetUserID primitive.ObjectID) ([]models.UserAddress, error) {
	// Authorization check: ensure authenticated user can only access their own addresses
	if authenticatedUserID != targetUserID {
		return nil, errors.New("unauthorized to access other user's addresses")
	}

	filter := bson.M{"user_id": targetUserID}
	cursor, err := common.UserAddressCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var userAddresses []models.UserAddress
	if err := cursor.All(ctx, &userAddresses); err != nil {
		return nil, err
	}

	return userAddresses, nil
}

// UpdateUserAddress updates an existing user address
func (uas *UserAddressServiceImpl) UpdateUserAddress(ctx context.Context, userID, addressID primitive.ObjectID, address models.UserAddressExcerpt) error {
	callback := func(ctx mongo.SessionContext) (any, error) {
		// If this is set as default, update other addresses
		if address.IsDefault {
			err := SetOtherRecordsToFalse(ctx, common.UserAddressCollection, "user_id", userID, addressID, "is_default_shipping_address")
			if err != nil {
				return nil, err
			}
		}

		filter := bson.M{"user_id": userID, "_id": addressID}
		update := bson.M{
			"$set": bson.M{
				"city":                        address.City,
				"state":                       address.State,
				"street":                      address.Street,
				"postal_code":                 address.PostalCode,
				"country":                     models.CountryNigeria,
				"is_default_shipping_address": address.IsDefault,
			},
		}

		result, err := common.UserAddressCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			return nil, err
		}

		if result.ModifiedCount == 0 {
			return nil, errors.New("address not found or no changes made")
		}

		return result, nil
	}

	_, err := ExecuteTransaction(ctx, callback)
	return err
}

// ChangeDefaultAddress changes the default address for a user
func (uas *UserAddressServiceImpl) ChangeDefaultAddress(ctx context.Context, userID, addressID primitive.ObjectID) error {
	callback := func(ctx mongo.SessionContext) (any, error) {
		// Set all other addresses to non-default
		err := SetOtherRecordsToFalse(ctx, common.UserAddressCollection, "user_id", userID, addressID, "is_default_shipping_address")
		if err != nil {
			return nil, err
		}

		// Set the specified address as default
		filter := bson.M{"user_id": userID, "_id": addressID}
		result, err := common.UserAddressCollection.UpdateOne(
			ctx,
			filter,
			bson.M{"$set": bson.M{"is_default_shipping_address": true}},
		)
		if err != nil {
			return nil, err
		}

		if result.ModifiedCount == 0 {
			return nil, errors.New("address not found")
		}

		return result, nil
	}

	_, err := ExecuteTransaction(ctx, callback)
	return err
}

// DeleteUserAddress deletes a user address
func (uas *UserAddressServiceImpl) DeleteUserAddress(ctx context.Context, userID, addressID primitive.ObjectID) error {
	filter := bson.M{"user_id": userID, "_id": addressID}
	result, err := common.UserAddressCollection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("user address not found")
	}

	return nil
}

