package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"khoomi-api-io/api/internal/auth"
	"khoomi-api-io/api/internal/common"
	"khoomi-api-io/api/pkg/models"
	"khoomi-api-io/api/pkg/util"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SaveCartItem: save cart listing.
func SaveCartItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		myId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		var cartReq models.CartItemRequest
		if err := c.ShouldBindJSON(&cartReq); err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		if err := common.Validate.Struct(&cartReq); err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		// Fetch listing snapshot (e.g. title, price, image, dynamic_type)
		var listing models.Listing
		err = common.ListingCollection.FindOne(ctx, bson.M{"_id": cartReq.ListingId}).Decode(&listing)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, fmt.Errorf("listing not found"))
			return
		}

		unitPrice := listing.Inventory.Price
		totalPrice := float64(cartReq.Quantity) * unitPrice

		cartDoc := models.CartItem{
			UserId:          myId,
			ListingId:       cartReq.ListingId,
			Title:           listing.Details.Title,
			Thumbnail:       listing.MainImage,
			Quantity:        cartReq.Quantity,
			UnitPrice:       unitPrice,
			TotalPrice:      totalPrice,
			Variant:         cartReq.Variant,
			DynamicType:     listing.Details.DynamicType,
			Personalization: cartReq.Personalization,
			ModifiedAt:      now,
			ExpiresAt:       now.Add(common.CART_ITEM_EXPIRATION_TIME),
		}

		// You can also Upsert instead of InsertOne if you want to overwrite
		res, err := common.UserCartCollection.InsertOne(ctx, cartDoc)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Item added to cart", res.InsertedID)
	}
}

// GetCartItems(): get all cart listing.
func GetCartItems() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		myId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		paginationArgs := common.GetPaginationArgs(c)
		filter := bson.M{"userId": myId}
		findOptions := options.Find().
			SetLimit(int64(paginationArgs.Limit)).
			SetSkip(int64(paginationArgs.Skip)).
			SetSort(util.GetLoginHistorySortBson(paginationArgs.Sort))
		cursor, err := common.UserCartCollection.Find(ctx, filter, findOptions)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}
		defer cursor.Close(ctx)

		var cartItems []models.CartItem
		if err = cursor.All(ctx, &cartItems); err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}

		count, err := common.UserCartCollection.CountDocuments(ctx, filter)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		util.HandleSuccessMeta(c, http.StatusOK, "success", cartItems, gin.H{
			"pagination": util.Pagination{
				Limit: paginationArgs.Limit,
				Skip:  paginationArgs.Skip,
				Count: count,
			},
		})
	}
}

// DeleteCartItem(): get all cart listing.
func DeleteCartItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		cartItemId := c.Param("cartId")
		cartItemObjectID, err := primitive.ObjectIDFromHex(cartItemId)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, errors.New("bad payment id"))
			return
		}
		myId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		res, err := common.IsSeller(c, myId)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}
		if !res {
			util.HandleError(c, http.StatusUnauthorized, errors.New("Can't perform this action"))
			return
		}
		filter := bson.M{"_id": cartItemObjectID, "userId": myId}
		result, err := common.UserCartCollection.DeleteOne(ctx, filter)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}

		if result.DeletedCount == 0 {
			util.HandleError(c, http.StatusNotFound, errors.New("No records deleted."))
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Cart item deleted successfully", result.DeletedCount)
	}
}

// deleteManyCartItems: delete multiple cart items by their IDs
func DeleteCartItems() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		idStrings := c.QueryArray("id")

		if len(idStrings) == 0 {
			util.HandleError(c, http.StatusBadRequest, errors.New("no cart item IDs provided"))
			return
		}

		myId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err)
			return
		}

		isSeller, err := common.IsSeller(c, myId)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}
		if !isSeller {
			util.HandleError(c, http.StatusUnauthorized, errors.New("unauthorized"))
			return
		}

		var objectIDs []primitive.ObjectID
		for _, idStr := range idStrings {
			objectID, err := primitive.ObjectIDFromHex(idStr)
			if err != nil {
				util.HandleError(c, http.StatusBadRequest, fmt.Errorf("invalid cart item ID: %s", idStr))
				return
			}
			objectIDs = append(objectIDs, objectID)
		}

		filter := bson.M{
			"_id":    bson.M{"$in": objectIDs},
			"userId": myId,
		}

		result, err := common.UserCartCollection.DeleteMany(ctx, filter)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		if result.DeletedCount == 0 {
			util.HandleError(c, http.StatusNotFound, errors.New("no cart items deleted"))
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Cart items deleted successfully", result.DeletedCount)
	}
}

// ClearCartItems: clear all cart items
func ClearCartItems() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		myId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err)
			return
		}

		isSeller, err := common.IsSeller(c, myId)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}
		if !isSeller {
			util.HandleError(c, http.StatusUnauthorized, errors.New("unauthorized"))
			return
		}

		filter := bson.M{"userId": myId}
		result, err := common.UserCartCollection.DeleteMany(ctx, filter)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		if result.DeletedCount == 0 {
			util.HandleSuccess(c, http.StatusOK, "Cart is already empty", result.DeletedCount)
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Cart cleared successfully", result.DeletedCount)
	}
}
