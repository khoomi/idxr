package services

import (
	"context"
	"errors"

	"khoomi-api-io/api/internal/common"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// UserFavoriteServiceImpl implements the UserFavoriteService interface
type UserFavoriteServiceImpl struct{}

// NewUserFavoriteService creates a new instance of UserFavoriteService
func NewUserFavoriteService() UserFavoriteService {
	return &UserFavoriteServiceImpl{}
}

// ToggleFavoriteShop toggles a shop's favorite status for a user
func (ufs *UserFavoriteServiceImpl) ToggleFavoriteShop(ctx context.Context, userID, shopID primitive.ObjectID, action string) error {

	callback := func(ctx mongo.SessionContext) (interface{}, error) {
		filter := bson.M{"_id": userID}

		switch action {
		case "add":
			// Add to user's favorite shops array
			update := bson.M{"$push": bson.M{"favorite_shops": shopID}}
			_, err := common.UserCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				return nil, err
			}

			// Insert into favorite shop collection
			_, err = common.UserFavoriteShopCollection.InsertOne(ctx, bson.M{"shopId": shopID, "userId": userID})
			if err != nil {
				return nil, err
			}
			return nil, nil

		case "remove":
			// Remove from user's favorite shops array
			update := bson.M{"$pull": bson.M{"favorite_shops": shopID}}
			_, err := common.UserCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				return nil, err
			}

			// Delete from favorite shop collection
			_, err = common.UserFavoriteShopCollection.DeleteOne(ctx, bson.M{"shopId": shopID, "userId": userID})
			if err != nil {
				return nil, err
			}
			return nil, nil

		default:
			return nil, errors.New("action query is missing from url")
		}
	}

	_, err := ExecuteTransaction(ctx, callback)
	return err
}

// IsShopFavorited checks if a shop is favorited by a user
func (ufs *UserFavoriteServiceImpl) IsShopFavorited(ctx context.Context, userID, shopID primitive.ObjectID) (bool, error) {
	result := common.UserFavoriteShopCollection.FindOne(ctx, bson.M{"shopId": shopID, "userId": userID})
	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, result.Err()
	}
	return true, nil
}

// ToggleFavoriteListing toggles a listing's favorite status for a user
func (ufs *UserFavoriteServiceImpl) ToggleFavoriteListing(ctx context.Context, userID, listingID primitive.ObjectID, action string) error {

	callback := func(ctx mongo.SessionContext) (interface{}, error) {
		filter := bson.M{"_id": userID}

		switch action {
		case "add":
			// Add to user's favorite listings array (store as string for compatibility)
			update := bson.M{"$push": bson.M{"favorite_listings": listingID.Hex()}}
			_, err := common.UserCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				return nil, err
			}

			// Insert into favorite listing collection
			_, err = common.UserFavoriteListingCollection.InsertOne(ctx, bson.M{"listingId": listingID, "userId": userID})
			if err != nil {
				return nil, err
			}
			return nil, nil

		case "remove":
			// Remove from user's favorite listings array (store as string for compatibility)
			update := bson.M{"$pull": bson.M{"favorite_listings": listingID.Hex()}}
			_, err := common.UserCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				return nil, err
			}

			// Delete from favorite listing collection
			_, err = common.UserFavoriteListingCollection.DeleteOne(ctx, bson.M{"listingId": listingID, "userId": userID})
			if err != nil {
				return nil, err
			}
			return nil, nil

		default:
			return nil, errors.New("action query is missing from url")
		}
	}

	_, err := ExecuteTransaction(ctx, callback)
	return err
}

// IsListingFavorited checks if a listing is favorited by a user
func (ufs *UserFavoriteServiceImpl) IsListingFavorited(ctx context.Context, userID, listingID primitive.ObjectID) (bool, error) {
	result := common.UserFavoriteListingCollection.FindOne(ctx, bson.M{"listingId": listingID, "userId": userID})
	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, result.Err()
	}
	return true, nil
}
