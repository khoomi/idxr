package services

import (
	"context"
	"fmt"
	"time"

	"khoomi-api-io/api/internal/common"
	"khoomi-api-io/api/pkg/models"
	"khoomi-api-io/api/pkg/util"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// CartServiceImpl implements the CartService interface
type CartServiceImpl struct {
	userCartCollection *mongo.Collection
	listingCollection  *mongo.Collection
	shopCollection     *mongo.Collection
}

// NewCartService creates a new instance of CartService
func NewCartService() CartService {
	return &CartServiceImpl{
		userCartCollection: util.GetCollection(util.DB, "UserCart"),
		listingCollection:  util.GetCollection(util.DB, "Listing"),
		shopCollection:     util.GetCollection(util.DB, "Shop"),
	}
}

// SaveCartItem adds a new item to the user's cart with validation
func (cs *CartServiceImpl) SaveCartItem(ctx context.Context, userID primitive.ObjectID, req models.CartItemRequest) (primitive.ObjectID, error) {
	now := time.Now()

	if err := common.Validate.Struct(&req); err != nil {
		return primitive.NilObjectID, err
	}

	listing, err := cs.validateListing(ctx, req.ListingId)
	if err != nil {
		return primitive.NilObjectID, err
	}

	if listing.Inventory.Quantity < req.Quantity {
		return primitive.NilObjectID, fmt.Errorf("insufficient inventory. Available: %d, Requested: %d",
			listing.Inventory.Quantity, req.Quantity)
	}

	shop, err := cs.getShop(ctx, req.ShopId)
	if err != nil {
		return primitive.NilObjectID, err
	}

	cartItem := cs.buildCartItem(userID, req, listing, shop, now)

	res, err := cs.userCartCollection.InsertOne(ctx, cartItem)
	if err != nil {
		return primitive.NilObjectID, err
	}

	if insertedID, ok := res.InsertedID.(primitive.ObjectID); ok {
		return insertedID, nil
	}

	return primitive.NilObjectID, fmt.Errorf("failed to get inserted ID")
}

// GetCartItems retrieves user's cart items with validation flags
func (cs *CartServiceImpl) GetCartItems(ctx context.Context, userID primitive.ObjectID, pagination util.PaginationArgs) ([]models.CartItemJson, int64, error) {
	pipeline := []bson.M{
		{"$match": bson.M{"userId": userID}},
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
						"if":   bson.M{"$eq": []any{"$current_listing.state", "active"}},
						"then": true,
						"else": false,
					},
				},
				"current_price":    "$current_listing.inventory.price",
				"current_quantity": "$current_listing.inventory.quantity",
				"price_changed": bson.M{
					"$ne": []any{"$unit_price", "$current_listing.inventory.price"},
				},
				"insufficient_stock": bson.M{
					"$gt": []any{"$quantity", "$current_listing.inventory.quantity"},
				},
			},
		},
		{"$sort": util.GetLoginHistorySortBson(pagination.Sort)},
		{"$skip": int64(pagination.Skip)},
		{"$limit": int64(pagination.Limit)},
	}

	cursor, err := cs.userCartCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, err
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
		return nil, 0, err
	}

	cartItems := make([]models.CartItemJson, 0, len(cartItemsWithValidation))
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
			IsAvailable:       item.IsListingAvailable,
			PriceChanged:      item.PriceChanged,
			CurrentPrice:      item.CurrentPrice,
			InsufficientStock: item.InsufficientStock,
			CurrentQuantity:   item.CurrentQuantity,
		}
		cartItems = append(cartItems, cartItemResponse)
	}

	count, err := cs.userCartCollection.CountDocuments(ctx, bson.M{"userId": userID})
	if err != nil {
		return nil, 0, err
	}

	return cartItems, count, nil
}

// IncreaseCartItemQuantity increases the quantity of a cart item
func (cs *CartServiceImpl) IncreaseCartItemQuantity(ctx context.Context, userID, cartItemID primitive.ObjectID) (*CartQuantityResponse, error) {
	return cs.updateCartItemQuantity(ctx, userID, cartItemID, 1)
}

// DecreaseCartItemQuantity decreases the quantity of a cart item
func (cs *CartServiceImpl) DecreaseCartItemQuantity(ctx context.Context, userID, cartItemID primitive.ObjectID) (*CartQuantityResponse, error) {
	return cs.updateCartItemQuantity(ctx, userID, cartItemID, -1)
}

// DeleteCartItem removes a single cart item
func (cs *CartServiceImpl) DeleteCartItem(ctx context.Context, userID, cartItemID primitive.ObjectID) (int64, error) {
	filter := bson.M{"_id": cartItemID, "userId": userID}
	result, err := cs.userCartCollection.DeleteOne(ctx, filter)
	if err != nil {
		return 0, err
	}

	if result.DeletedCount == 0 {
		return 0, errors.New("cart item not found")
	}

	return result.DeletedCount, nil
}

// DeleteCartItems removes multiple cart items
func (cs *CartServiceImpl) DeleteCartItems(ctx context.Context, userID primitive.ObjectID, cartItemIDs []primitive.ObjectID) (int64, error) {
	if len(cartItemIDs) == 0 {
		return 0, errors.New("no cart item IDs provided")
	}

	filter := bson.M{
		"_id":    bson.M{"$in": cartItemIDs},
		"userId": userID,
	}

	result, err := cs.userCartCollection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}

	if result.DeletedCount == 0 {
		return 0, errors.New("no cart items deleted")
	}

	return result.DeletedCount, nil
}

// ClearCartItems removes all cart items for a user
func (cs *CartServiceImpl) ClearCartItems(ctx context.Context, userID primitive.ObjectID) (int64, error) {
	filter := bson.M{"userId": userID}
	result, err := cs.userCartCollection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}

	return result.DeletedCount, nil
}

// ValidateCartItems validates all cart items and returns validation results
func (cs *CartServiceImpl) ValidateCartItems(ctx context.Context, userID primitive.ObjectID) (*CartValidationResult, error) {
	cartItems, _, err := cs.GetCartItems(ctx, userID, util.PaginationArgs{Limit: 1000, Skip: 0})
	if err != nil {
		return nil, err
	}

	validItems := make([]models.CartItemJson, 0)
	invalidItems := make([]models.CartItemJson, 0)

	for _, item := range cartItems {
		if item.IsAvailable && !item.PriceChanged && !item.InsufficientStock {
			validItems = append(validItems, item)
		} else {
			invalidItems = append(invalidItems, item)
		}
	}

	return &CartValidationResult{
		ValidItems:      validItems,
		InvalidItems:    invalidItems,
		TotalItems:      len(cartItems),
		TotalValid:      len(validItems),
		TotalInvalid:    len(invalidItems),
		HasInvalidItems: len(invalidItems) > 0,
	}, nil
}

// Helper methods

func (cs *CartServiceImpl) validateListing(ctx context.Context, listingID primitive.ObjectID) (*models.Listing, error) {
	var listing models.Listing
	err := cs.listingCollection.FindOne(ctx, bson.M{"_id": listingID}).Decode(&listing)
	if err != nil {
		return nil, fmt.Errorf("listing not found")
	}

	if listing.State.State != models.ListingStateActive {
		return nil, fmt.Errorf("listing is not available")
	}

	return &listing, nil
}

func (cs *CartServiceImpl) getShop(ctx context.Context, shopID primitive.ObjectID) (*models.Shop, error) {
	var shop models.Shop
	err := cs.shopCollection.FindOne(ctx, bson.M{"_id": shopID}).Decode(&shop)
	if err != nil {
		return nil, fmt.Errorf("shop not found")
	}
	return &shop, nil
}

func (cs *CartServiceImpl) buildCartItem(userID primitive.ObjectID, req models.CartItemRequest, listing *models.Listing, shop *models.Shop, now time.Time) models.CartItem {
	unitPrice := listing.Inventory.Price
	totalPrice := float64(req.Quantity) * unitPrice

	return models.CartItem{
		Id:                primitive.NewObjectID(),
		UserId:            userID,
		ListingId:         req.ListingId,
		ShopId:            req.ShopId,
		Title:             listing.Details.Title,
		Thumbnail:         listing.MainImage,
		Quantity:          req.Quantity,
		UnitPrice:         unitPrice,
		TotalPrice:        totalPrice,
		Variant:           req.Variant,
		DynamicType:       listing.Details.DynamicType,
		Personalization:   req.Personalization,
		ShopName:          shop.Name,
		ShopUsername:      shop.Username,
		ShopSlug:          shop.Slug,
		AvailableQuantity: listing.Inventory.Quantity,
		ListingState:      listing.State,
		OriginalPrice:     unitPrice,
		PriceUpdatedAt:    now,
		ShippingProfileId: listing.ShippingProfileId,
		AddedAt:           now,
		ModifiedAt:        now,
		ExpiresAt:         now.Add(common.CART_ITEM_EXPIRATION_TIME),
	}
}

func (cs *CartServiceImpl) updateCartItemQuantity(ctx context.Context, userID, cartItemID primitive.ObjectID, delta int) (*CartQuantityResponse, error) {
	now := time.Now()

	var cartItem models.CartItem
	filter := bson.M{"_id": cartItemID, "userId": userID}
	err := cs.userCartCollection.FindOne(ctx, filter).Decode(&cartItem)
	if err != nil {
		return nil, fmt.Errorf("cart item not found")
	}

	newQuantity := cartItem.Quantity + delta
	if newQuantity <= 0 {
		return nil, fmt.Errorf("cannot decrease quantity below 1. Use delete endpoint to remove item")
	}

	listing, err := cs.validateListing(ctx, cartItem.ListingId)
	if err != nil {
		return nil, err
	}

	if delta > 0 && listing.Inventory.Quantity < newQuantity {
		return nil, fmt.Errorf("insufficient inventory. Available: %d, Requested: %d",
			listing.Inventory.Quantity, newQuantity)
	}

	unitPrice := listing.Inventory.Price
	newTotalPrice := float64(newQuantity) * unitPrice

	cartItem.Quantity = newQuantity
	cartItem.UnitPrice = unitPrice
	cartItem.TotalPrice = newTotalPrice
	cartItem.AvailableQuantity = listing.Inventory.Quantity
	cartItem.ListingState = listing.State
	cartItem.OriginalPrice = unitPrice
	cartItem.PriceUpdatedAt = now
	cartItem.ModifiedAt = now
	cartItem.ExpiresAt = now.Add(common.CART_ITEM_EXPIRATION_TIME)

	_, err = cs.userCartCollection.DeleteOne(ctx, filter)
	if err != nil {
		return nil, err
	}

	_, err = cs.userCartCollection.InsertOne(ctx, cartItem)
	if err != nil {
		return nil, err
	}

	return &CartQuantityResponse{
		Quantity:   newQuantity,
		UnitPrice:  unitPrice,
		TotalPrice: newTotalPrice,
	}, nil
}
