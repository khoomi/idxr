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

		// Fetch listing data
		var listing models.Listing
		err = common.ListingCollection.FindOne(ctx, bson.M{"_id": cartReq.ListingId}).Decode(&listing)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, fmt.Errorf("listing not found"))
			return
		}

		// Validate listing is active
		if listing.State.State != models.ListingStateActive {
			util.HandleError(c, http.StatusBadRequest, fmt.Errorf("listing is not available"))
			return
		}

		// Validate sufficient inventory
		if listing.Inventory.Quantity < cartReq.Quantity {
			util.HandleError(c, http.StatusBadRequest, fmt.Errorf("insufficient inventory. Available: %d, Requested: %d", listing.Inventory.Quantity, cartReq.Quantity))
			return
		}

		// Fetch shop data separately using the shop_id from the request
		var shop models.Shop
		err = common.ShopCollection.FindOne(ctx, bson.M{"_id": cartReq.ShopId}).Decode(&shop)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, fmt.Errorf("shop not found"))
			return
		}

		unitPrice := listing.Inventory.Price
		totalPrice := float64(cartReq.Quantity) * unitPrice

		cartDoc := models.CartItem{
			Id:              primitive.NewObjectID(),
			UserId:          myId,
			ListingId:       cartReq.ListingId,
			ShopId:          cartReq.ShopId,
			Title:           listing.Details.Title,
			Thumbnail:       listing.MainImage,
			Quantity:        cartReq.Quantity,
			UnitPrice:       unitPrice,
			TotalPrice:      totalPrice,
			Variant:         cartReq.Variant,
			DynamicType:     listing.Details.DynamicType,
			Personalization: cartReq.Personalization,

			ShopName:     shop.Name,
			ShopUsername: shop.Username,
			ShopSlug:     shop.Slug,

			AvailableQuantity: listing.Inventory.Quantity,
			ListingState:      listing.State,

			OriginalPrice:  unitPrice,
			PriceUpdatedAt: now,

			ShippingProfileId: listing.ShippingProfileId,

			AddedAt:    now,
			ModifiedAt: now,
			ExpiresAt:  now.Add(common.CART_ITEM_EXPIRATION_TIME),
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

		// Use aggregation to join with current listing data for validation
		pipeline := []bson.M{
			{"$match": bson.M{"userId": myId}},
			{
				"$lookup": bson.M{
					"from":         "Listing",
					"localField":   "listing_id",
					"foreignField": "_id",
					"as":           "current_listing",
				},
			},
			{"$unwind": bson.M{"path": "$current_listing", "preserveNullAndEmptyArrays": true}},
			{
				"$addFields": bson.M{
					"is_listing_available": bson.M{
						"$cond": bson.M{
							"if":   bson.M{"$eq": []interface{}{"$current_listing.state", "active"}},
							"then": true,
							"else": false,
						},
					},
					"current_price":    "$current_listing.inventory.price",
					"current_quantity": "$current_listing.inventory.quantity",
					"price_changed": bson.M{
						"$ne": []interface{}{"$unit_price", "$current_listing.inventory.price"},
					},
					"insufficient_stock": bson.M{
						"$gt": []interface{}{"$quantity", "$current_listing.inventory.quantity"},
					},
				},
			},
			{"$sort": util.GetLoginHistorySortBson(paginationArgs.Sort)},
			{"$skip": int64(paginationArgs.Skip)},
			{"$limit": int64(paginationArgs.Limit)},
		}

		cursor, err := common.UserCartCollection.Aggregate(ctx, pipeline)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}
		defer cursor.Close(ctx)

		var cartItemsWithValidation []struct {
			models.CartItem    `bson:",inline"`
			IsListingAvailable bool    `bson:"is_listing_available"`
			CurrentPrice       float64 `bson:"current_price"`
			CurrentQuantity    int     `bson:"current_quantity"`
			PriceChanged       bool    `bson:"price_changed"`
			InsufficientStock  bool    `bson:"insufficient_stock"`
		}

		if err = cursor.All(ctx, &cartItemsWithValidation); err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		// Convert back to regular cart items and add validation flags
		var cartItems []models.CartItemJson
		for _, item := range cartItemsWithValidation {
			cartItemResponse := models.CartItemJson{
				Id:                item.Id,
				ListingId:         item.ListingId,
				ShopId:            item.ShopId,
				UserId:            item.UserId,
				Title:             item.Title,
				Thumbnail:         item.Thumbnail,
				Quantity:          item.Quantity,
				UnitPrice:         item.UnitPrice,
				TotalPrice:        item.TotalPrice,
				Variant:           item.Variant,
				DynamicType:       item.DynamicType,
				Personalization:   item.Personalization,
				ShopName:          item.ShopName,
				ShopUsername:      item.ShopUsername,
				ShopSlug:          item.ShopSlug,
				AvailableQuantity: item.AvailableQuantity,
				ListingState:      item.ListingState,
				OriginalPrice:     item.OriginalPrice,
				PriceUpdatedAt:    item.PriceUpdatedAt,
				ShippingProfileId: item.ShippingProfileId,
				ExpiresAt:         item.ExpiresAt,
				AddedAt:           item.AddedAt,
				ModifiedAt:        item.ModifiedAt,

				// Validation flags
				IsAvailable:       item.IsListingAvailable,
				PriceChanged:      item.PriceChanged,
				CurrentPrice:      item.CurrentPrice,
				InsufficientStock: item.InsufficientStock,
				CurrentQuantity:   item.CurrentQuantity,
			}
			cartItems = append(cartItems, cartItemResponse)
		}

		count, err := common.UserCartCollection.CountDocuments(ctx, bson.M{"userId": myId})
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

// IncreaseCartItemQuantity
func IncreaseCartItemQuantity() gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		cartItemId := c.Param("cartId")
		cartItemObjectID, err := primitive.ObjectIDFromHex(cartItemId)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, errors.New("invalid cart item id"))
			return
		}
		myId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		// Find existing cart item
		var cartItem models.CartItem
		filter := bson.M{"_id": cartItemObjectID, "userId": myId}
		err = common.UserCartCollection.FindOne(ctx, filter).Decode(&cartItem)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, fmt.Errorf("cart item not found"))
			return
		}

		// Fetch current listing data to validate inventory
		var listing models.Listing
		err = common.ListingCollection.FindOne(ctx, bson.M{"_id": cartItem.ListingId}).Decode(&listing)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, fmt.Errorf("listing not found"))
			return
		}

		// Validate listing is still active
		if listing.State.State != models.ListingStateActive {
			util.HandleError(c, http.StatusBadRequest, fmt.Errorf("listing is not available"))
			return
		}

		// Validate sufficient inventory for increase
		if listing.Inventory.Quantity <= cartItem.Quantity {
			util.HandleError(c, http.StatusBadRequest, fmt.Errorf("insufficient inventory. Available: %d, Current: %d", listing.Inventory.Quantity, cartItem.Quantity))
			return
		}

		unitPrice := listing.Inventory.Price

		// For time-series collections, use delete and insert pattern
		newQuantity := cartItem.Quantity + 1
		newTotalPrice := float64(newQuantity) * unitPrice

		// Update the cart item fields
		cartItem.Quantity = newQuantity
		cartItem.UnitPrice = unitPrice
		cartItem.TotalPrice = newTotalPrice
		cartItem.AvailableQuantity = listing.Inventory.Quantity
		cartItem.ListingState = listing.State
		cartItem.OriginalPrice = unitPrice
		cartItem.PriceUpdatedAt = now
		cartItem.ModifiedAt = now
		cartItem.ExpiresAt = now.Add(common.CART_ITEM_EXPIRATION_TIME)

		// Delete the old document
		_, err = common.UserCartCollection.DeleteOne(ctx, filter)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		// Insert the updated document
		_, err = common.UserCartCollection.InsertOne(ctx, cartItem)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Cart item quantity increased", gin.H{
			"quantity":   newQuantity,
			"unitPrice":  unitPrice,
			"totalPrice": newTotalPrice,
		})
	}
}

// DecreaseCartItemQuantity
func DecreaseCartItemQuantity() gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		cartItemId := c.Param("cartId")
		cartItemObjectID, err := primitive.ObjectIDFromHex(cartItemId)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, errors.New("invalid cart item id"))
			return
		}
		myId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		// Find existing cart item
		var cartItem models.CartItem
		filter := bson.M{"_id": cartItemObjectID, "userId": myId}
		err = common.UserCartCollection.FindOne(ctx, filter).Decode(&cartItem)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, fmt.Errorf("cart item not found"))
			return
		}

		// Validate minimum quantity
		if cartItem.Quantity <= 1 {
			util.HandleError(c, http.StatusBadRequest, fmt.Errorf("cannot decrease quantity below 1. Use delete endpoint to remove item"))
			return
		}

		// Fetch current listing data for price validation
		var listing models.Listing
		err = common.ListingCollection.FindOne(ctx, bson.M{"_id": cartItem.ListingId}).Decode(&listing)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, fmt.Errorf("listing not found"))
			return
		}

		unitPrice := listing.Inventory.Price

		// For time-series collections, use delete and insert pattern
		newQuantity := cartItem.Quantity - 1
		newTotalPrice := float64(newQuantity) * unitPrice

		// Update the cart item fields
		cartItem.Quantity = newQuantity
		cartItem.UnitPrice = unitPrice
		cartItem.TotalPrice = newTotalPrice
		cartItem.AvailableQuantity = listing.Inventory.Quantity
		cartItem.ListingState = listing.State
		cartItem.OriginalPrice = unitPrice
		cartItem.PriceUpdatedAt = now
		cartItem.ModifiedAt = now
		cartItem.ExpiresAt = now.Add(common.CART_ITEM_EXPIRATION_TIME)

		// Delete the old document
		_, err = common.UserCartCollection.DeleteOne(ctx, filter)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		// Insert the updated document
		_, err = common.UserCartCollection.InsertOne(ctx, cartItem)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Cart item quantity decreased", gin.H{
			"quantity":   newQuantity,
			"unitPrice":  unitPrice,
			"totalPrice": newTotalPrice,
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
