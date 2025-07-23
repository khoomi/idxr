package services

import (
	"context"
	"errors"
	
	"khoomi-api-io/api/internal/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// verifyListingOwnership is a temporary helper method to verify listing ownership
// TODO: This should be refactored to use ListingService dependency injection
func (rs *ReviewServiceImpl) verifyListingOwnership(ctx context.Context, userID, listingID primitive.ObjectID) error {
	var listing struct {
		UserID primitive.ObjectID `bson:"user_id"`
	}
	
	err := common.ListingCollection.FindOne(ctx, bson.M{"_id": listingID}).Decode(&listing)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("listing not found")
		}
		return err
	}
	
	if listing.UserID != userID {
		return errors.New("unauthorized: you don't own this listing")
	}
	
	return nil
}