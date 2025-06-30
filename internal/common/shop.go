package common

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	auth "khoomi-api-io/api/internal/auth"
	"khoomi-api-io/api/pkg/util"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// VerifyShopOwnership verifies if a user owns a given shop using it's shopId.
func VerifyShopOwnership(ctx context.Context, userId, shopId primitive.ObjectID) error {
	// Use FindOne with projection to only fetch _id field - most efficient approach
	var result struct {
		ID primitive.ObjectID `bson:"_id"`
	}
	err := ShopCollection.FindOne(ctx, bson.M{"_id": shopId, "user_id": userId}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("user does not own the shop")
		}
		return err
	}
	return nil
}

// VerifyListingOwnership verifies if a user owns a given listing using it's listingId.
func VerifyListingOwnership(ctx context.Context, userId, listingId primitive.ObjectID) error {
	// Use FindOne with projection to only fetch _id field - most efficient approach
	var result struct {
		ID primitive.ObjectID `bson:"_id"`
	}
	err := ListingCollection.FindOne(ctx, bson.M{"_id": listingId, "user_id": userId}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("user does not own the listing")
		}
		return err
	}
	return nil
}
func MyShopIdAndMyId(c *gin.Context) (primitive.ObjectID, primitive.ObjectID, error) {
	nilObjectId := primitive.NilObjectID

	shopId := c.Param("shopid")
	fmt.Println(shopId)
	shopOBjectID, err := primitive.ObjectIDFromHex(shopId)
	if err != nil {
		return nilObjectId, nilObjectId, err
	}

	session, err := auth.GetSessionAuto(c)
	if err != nil {
		util.HandleError(c, http.StatusUnauthorized, err)
		return nilObjectId, nilObjectId, err
	}

	return shopOBjectID, session.UserId, nil
}
