package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"khoomi-api-io/api/internal/common"
	"khoomi-api-io/api/pkg/models"
	"khoomi-api-io/api/pkg/util"

	"github.com/cloudinary/cloudinary-go/api/uploader"
	"github.com/gin-gonic/gin"
	slug2 "github.com/gosimple/slug"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type listingService struct {
	listingCollection       *mongo.Collection
	shopCollection          *mongo.Collection
	listingReviewCollection *mongo.Collection
}

func NewListingService() ListingService {
	return &listingService{
		listingCollection:       util.GetCollection(util.DB, "Listing"),
		shopCollection:          util.GetCollection(util.DB, "Shop"),
		listingReviewCollection: util.GetCollection(util.DB, "ListingReview"),
	}
}

// VerifyListingOwnership verifies if a user owns a given listing using its listingId
func (s *listingService) VerifyListingOwnership(ctx context.Context, userID, listingID primitive.ObjectID) error {
	// Use FindOne with projection to only fetch _id field - most efficient approach
	var result struct {
		ID primitive.ObjectID `bson:"_id"`
	}
	err := s.listingCollection.FindOne(ctx, bson.M{"_id": listingID, "user_id": userID}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("user does not own the listing")
		}
		return err
	}
	return nil
}

// GenerateListingBson builds bson.M from listingid param
func (s *listingService) GenerateListingBson(listingID string) (bson.M, error) {
	if primitive.IsValidObjectID(listingID) {
		listingObjectID, e := primitive.ObjectIDFromHex(listingID)
		if e != nil {
			return nil, e
		}

		return bson.M{"_id": listingObjectID}, nil
	} else {
		return bson.M{"slug": strings.TrimSpace(listingID)}, nil
	}
}

// GenerateListingCode generates a listing code
func (s *listingService) GenerateListingCode() string {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	letterChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numberChars := "0123456789"
	letters := make([]byte, 4)
	for i := range letters {
		letters[i] = letterChars[rand.Intn(len(letterChars))]
	}

	numbers := make([]byte, 4)
	for i := range numbers {
		numbers[i] = numberChars[rand.Intn(len(numberChars))]
	}

	productCode := string(letters) + "-" + string(numbers)
	return productCode
}

// GetListingSortingBson returns bson for sorting listings
func (s *listingService) GetListingSortingBson(sort string) bson.D {
	value := -1
	var key string

	switch sort {
	case "created_at_asc":
		key = "date.created_at"
	case "created_at_desc":
		key = "date.created_at"
	case "modified_at_asc":
		key = "date.modified_at"
	case "modified_at_desc":
		key = "date.modified_at"
	case "state_updated_at_asc":
		key = "state.updated_at"
	case "state_updated_at_desc":
		key = "state.updated_at"
	case "views_asc":
		key = "views"
	case "views_desc":
		key = "views"
	case "sales_asc":
		key = "financial_information.sales"
	case "sales_desc":
		key = "financial_information.sales"
	case "price_asc":
		key = "inventory.price"
	case "price_desc":
		key = "inventory.price"
	case "rating_desc":
		key = "rating.rating.positive_reviews"
	default:
		key = "date.created_at"
	}

	if strings.Contains(sort, "asc") {
		value = 1
	}
	return bson.D{{Key: key, Value: value}}
}

// GetListingFilters returns bson.M for filtering listings based on query parameters
func (s *listingService) GetListingFilters(c *gin.Context) bson.M {
	match := bson.M{}

	if minPrice := c.Query("min_price"); minPrice != "" {
		if price, err := strconv.ParseFloat(minPrice, 64); err == nil {
			match["inventory.price"] = bson.M{"$gte": price}
		}
	}
	if maxPrice := c.Query("max_price"); maxPrice != "" {
		if price, err := strconv.ParseFloat(maxPrice, 64); err == nil {
			if val, ok := match["inventory.price"].(bson.M); ok {
				val["$lte"] = price
			} else {
				match["inventory.price"] = bson.M{"$lte": price}
			}
		}
	}
	if category := c.Query("category"); category != "" && category != "All" {
		match["details.category.id"] = category
	}

	if state := c.Query("status"); state != "" {
		match["state.state"] = state
	}

	if userID := c.Query("user_id"); userID != "" {
		if oid, err := primitive.ObjectIDFromHex(userID); err == nil {
			match["user_id"] = oid
		}
	}

	if shopID := c.Query("shop_id"); shopID != "" {
		if oid, err := primitive.ObjectIDFromHex(shopID); err == nil {
			match["shop_id"] = oid
		}
	}

	if days := c.Query("recent_days"); days != "" {
		if d, err := strconv.Atoi(days); err == nil {
			from := time.Now().AddDate(0, 0, -d)
			match["date.created_at"] = bson.M{"$gte": from}
		}
	}

	if tags := c.QueryArray("tags"); len(tags) > 0 {
		match["details.tags"] = bson.M{"$in": tags}
	}

	if color := c.Query("color"); color != "" {
		match["details.color"] = color
	}

	if search := c.Query("search"); search != "" {
		match["$text"] = bson.M{"$search": search}
	}

	if hp := c.Query("has_personalization"); hp == "true" {
		match["details.has_personalization"] = true
	}

	if hv := c.Query("has_variations"); hv == "true" {
		match["details.has_variations"] = true
	}

	if wm := c.Query("who_made"); wm != "" {
		match["details.who_made"] = wm
	}

	if wm := c.Query("when_made"); wm != "" {
		match["details.when_made"] = wm
	}

	if c := c.Query("condition"); c != "" {
		match["details.condition"] = c
	}

	if c := c.Query("sustainability"); c != "" {
		match["details.sustainability"] = c
	}

	if rating := c.Query("min_rating"); rating != "" {
		if r, err := strconv.ParseFloat(rating, 64); err == nil {
			match["rating.rating"] = bson.M{"$gte": r}
		}
	}

	return match
}

// DeleteListings deletes multiple listings and handles all related data cleanup
func (s *listingService) DeleteListings(ctx context.Context, userID, shopID primitive.ObjectID, listingIDs []primitive.ObjectID, reviewService ReviewService) (*DeleteListingsResult, error) {
	result := &DeleteListingsResult{
		DeletedListings:    []primitive.ObjectID{},
		NotDeletedListings: []primitive.ObjectID{},
		DeletedReviews:     0,
		UpdatedShop:        false,
	}

	if len(listingIDs) == 0 {
		return result, errors.New("no listing IDs provided")
	}

	for _, listingID := range listingIDs {
		cartFilter := bson.M{"listing_id": listingID}
		common.UserCartCollection.DeleteMany(ctx, cartFilter)

		favoriteFilter := bson.M{"listing_id": listingID}
		common.UserFavoriteListingCollection.DeleteMany(ctx, favoriteFilter)
	}

	callback := func(sessionCtx mongo.SessionContext) (any, error) {
		var deletedCount int64 = 0
		var totalDeletedReviews int64 = 0
		var deletedListings []primitive.ObjectID
		var notDeletedListings []primitive.ObjectID

		for _, listingID := range listingIDs {
			var listing struct {
				ShopID primitive.ObjectID `bson:"shop_id"`
			}
			err := s.listingCollection.FindOne(sessionCtx, bson.M{"_id": listingID, "user_id": userID}).Decode(&listing)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					notDeletedListings = append(notDeletedListings, listingID)
					continue
				}
				return nil, err
			}

			if listing.ShopID != shopID {
				notDeletedListings = append(notDeletedListings, listingID)
				continue
			}

			reviewFilter := bson.M{"listing_id": listingID}
			deleteReviewsResult, err := s.listingReviewCollection.DeleteMany(sessionCtx, reviewFilter)
			if err != nil {
				return nil, err
			}
			totalDeletedReviews += deleteReviewsResult.DeletedCount

			deleteResult, err := s.listingCollection.DeleteOne(sessionCtx, bson.M{"_id": listingID, "user_id": userID})
			if err != nil {
				return nil, err
			}
			if deleteResult.DeletedCount == 0 {
				notDeletedListings = append(notDeletedListings, listingID)
				continue
			}

			deletedListings = append(deletedListings, listingID)
			deletedCount++
		}

		if deletedCount > 0 {
			shopFilter := bson.M{"_id": shopID}
			shopUpdate := bson.M{"$inc": bson.M{"listing_active_count": -deletedCount}}
			updateResult, err := s.shopCollection.UpdateOne(sessionCtx, shopFilter, shopUpdate)
			if err != nil {
				return nil, err
			}
			if updateResult.ModifiedCount > 0 {
				result.UpdatedShop = true
			}

			if totalDeletedReviews > 0 && reviewService != nil {
				newShopRating, err := reviewService.CalculateShopRating(sessionCtx, shopID)
				if err != nil {
					return nil, err
				}
				shopRatingUpdate := bson.M{"$set": bson.M{"rating": newShopRating}}
				_, err = s.shopCollection.UpdateOne(sessionCtx, shopFilter, shopRatingUpdate)
				if err != nil {
					return nil, err
				}
			}
		}

		result.DeletedListings = deletedListings
		result.NotDeletedListings = notDeletedListings
		result.DeletedReviews = totalDeletedReviews

		return result, nil
	}

	txResult, err := ExecuteTransaction(ctx, callback)
	if err != nil {
		return nil, err
	}

	if deleteResult, ok := txResult.(*DeleteListingsResult); ok {
		return deleteResult, nil
	}
	return nil, errors.New("failed to get delete result from transaction")
}

// CreateListing creates a new listing
func (s *listingService) CreateListing(ctx context.Context, req CreateListingRequest) (primitive.ObjectID, error) {
	tempShopService := NewShopService()

	newListingID := primitive.NewObjectID()
	callback := func(ctx mongo.SessionContext) (any, error) {
		err := tempShopService.VerifyShopOwnership(ctx, req.UserID, req.ShopID)
		if err != nil {
			return primitive.NilObjectID, errors.New("only shop owners can create listings")
		}

		if err := req.NewListing.Details.SetDynamicToTypedField(); err != nil {
			return primitive.NilObjectID, fmt.Errorf("invalid dynamic data: %v", err)
		}

		var shippingId primitive.ObjectID
		shippingObj, err := primitive.ObjectIDFromHex(req.NewListing.Details.ShippingProfileId)
		if err != nil {
			var shipping models.ShopShippingProfile
			err := common.ShippingProfileCollection.FindOne(ctx, bson.M{"shop_id": req.ShopID, "is_default_profile": true}).Decode(&shipping)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					shippingId = primitive.NilObjectID
				} else {
					return primitive.NilObjectID, fmt.Errorf("failed to fetch default shipping profile: %v", err)
				}
			} else {
				shippingId = shipping.ID
			}
		} else {
			shippingId = shippingObj
		}

		now := time.Now()
		listingDetails := models.Details{
			Type:               req.NewListing.Details.Type,
			Tags:               req.NewListing.Details.Tags,
			Title:              req.NewListing.Details.Title,
			Dynamic:            req.NewListing.Details.Dynamic,
			DynamicType:        req.NewListing.Details.DynamicType,
			WhoMade:            req.NewListing.Details.WhoMade,
			Keywords:           req.NewListing.Details.Keywords,
			WhenMade:           req.NewListing.Details.WhenMade,
			Category:           req.NewListing.Details.Category,
			Condition:          req.NewListing.Details.Condition,
			Description:        req.NewListing.Details.Description,
			Sustainability:     req.NewListing.Details.Sustainability,
			HasPersonalization: req.NewListing.Details.HasPersonalization,
			Personalization:    req.NewListing.Details.Personalization,

			ClothingData:              req.NewListing.Details.ClothingData,
			FurnitureData:             req.NewListing.Details.FurnitureData,
			GiftsAndOccasionsData:     req.NewListing.Details.GiftsAndOccasionsData,
			ArtAndCollectiblesData:    req.NewListing.Details.ArtAndCollectiblesData,
			AceessoriesAndJewelryData: req.NewListing.Details.AceessoriesAndJewelryData,
			HomeAndLivingData:         req.NewListing.Details.HomeAndLivingData,
		}

		listingInventory := models.Inventory{
			DomesticPricing: req.NewListing.Inventory.DomesticPricing,
			DomesticPrice:   req.NewListing.Inventory.DomesticPrice,
			Price:           req.NewListing.Inventory.Price,
			InitialQuantity: req.NewListing.Inventory.Quantity,
			Quantity:        req.NewListing.Inventory.Quantity,
			SKU:             req.NewListing.Inventory.SKU,
			CurrencyCode:    "NGN",
			ModifiedAt:      now,
		}

		listingDate := models.ListingDateMeta{
			CreatedAt:  now,
			EndingAt:   now,
			ModifiedAt: now,
		}

		listingRating := models.Rating{
			AverageRating:  0.0,
			ReviewCount:    0,
			FiveStarCount:  0,
			FourStarCount:  0,
			ThreeStarCount: 0,
			TwoStarCount:   0,
			OneStarCount:   0,
		}

		listingFinancialInformation := models.FinancialInformation{
			TotalOrders:     0,
			Sales:           0,
			OrdersPending:   0,
			OrdersCanceled:  0,
			OrdersCompleted: 0,
			Revenue:         0.0,
			Profit:          0.0,
			ShippingRevenue: 0.0,
		}

		listing := models.Listing{
			ID:                   newListingID,
			Code:                 s.GenerateListingCode(),
			UserId:               req.UserID,
			ShopId:               req.ShopID,
			MainImage:            req.MainImageURL,
			Images:               req.ImagesURLs,
			Details:              listingDetails,
			Slug:                 slug2.Make(req.NewListing.Details.Title),
			Date:                 listingDate,
			State:                models.ListingState{State: models.ListingStateActive, StateUpdatedAt: now},
			ShippingProfileId:    shippingId,
			NonTaxable:           true,
			ShouldAutoRenew:      false,
			Variations:           req.NewListing.Variations,
			Inventory:            listingInventory,
			Rating:               listingRating,
			FinancialInformation: listingFinancialInformation,
			Views:                0,
			FavorersCount:        0,
		}

		_, err = s.listingCollection.InsertOne(ctx, listing)

		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				return nil, errors.New("listing with similar data already exists")
			}
			return nil, fmt.Errorf("database error while creating listing: %v", err)
		}

		// Update shop listing count
		filter := bson.M{"_id": req.ShopID}
		update := bson.M{"$inc": bson.M{"listing_active_count": 1}}
		updateResult, err := s.shopCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			log.Printf("Failed to update shop listing count for shop %s: %v", req.ShopID.Hex(), err)
		} else if updateResult.MatchedCount == 0 {
			log.Printf("Warning: Shop %s not found when updating listing count", req.ShopID.Hex())
		}

		if req.IsOnboarding {
			filter := bson.M{"_id": req.UserID}
			update := bson.M{"$set": bson.M{"modified_at": now, "seller_onboarding_level": models.OnboardingLevelListing}}
			_, err = common.UserCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				return nil, err
			}
		}

		return newListingID, nil
	}

	result, err := ExecuteTransaction(ctx, callback)
	if err != nil {
		log.Printf("Transaction failed for listing creation: %v", err)
		// Cleanup uploaded images on error
		if mainResult, ok := req.MainImageResult.(uploader.UploadResult); ok && mainResult.PublicID != "" {
			if _, destroyErr := util.DestroyMedia(mainResult.PublicID); destroyErr != nil {
				log.Printf("Failed to cleanup main image %s: %v", mainResult.PublicID, destroyErr)
			}
		}
		for _, file := range req.ImagesResults {
			if result, ok := file.(uploader.UploadResult); ok {
				if _, destroyErr := util.DestroyMedia(result.PublicID); destroyErr != nil {
					log.Printf("Failed to cleanup image %s: %v", result.PublicID, destroyErr)
				}
			}
		}
		return primitive.NilObjectID, err
	}

	if listingID, ok := result.(primitive.ObjectID); ok {
		return listingID, nil
	}
	return primitive.NilObjectID, errors.New("failed to get listing ID from transaction")
}

// GetListing retrieves a single listing with all related data
func (s *listingService) GetListing(ctx context.Context, listingID string) (*models.ListingExtra, error) {
	listingIdentifier, err := s.GenerateListingBson(listingID)
	if err != nil {
		return nil, err
	}

	pipeline := []bson.M{
		{"$match": listingIdentifier},
		{
			"$lookup": bson.M{
				"from":         "Listing",
				"localField":   "shop_id",
				"foreignField": "shop_id",
				"as":           "siblings",
				"pipeline": []bson.M{
					{"$limit": 6},
				},
			},
		},
		{
			"$lookup": bson.M{
				"from":         "Shop",
				"localField":   "shop_id",
				"foreignField": "_id",
				"as":           "shop",
			},
		},
		{"$unwind": "$shop"},
		{
			"$lookup": bson.M{
				"from":         "User",
				"localField":   "user_id",
				"foreignField": "_id",
				"as":           "user",
			},
		},
		{"$unwind": "$user"},
		{
			"$lookup": bson.M{
				"from":         "ShopShippingProfile",
				"localField":   "shipping_profile_id",
				"foreignField": "_id",
				"as":           "shipping",
			},
		},
		{"$unwind": bson.M{"path": "$shipping", "preserveNullAndEmptyArrays": true}},
		{
			"$project": bson.M{
				"_id":                 1,
				"state":               1,
				"user_id":             1,
				"shop_id":             1,
				"main_image":          1,
				"images":              1,
				"details":             1,
				"date":                1,
				"slug":                1,
				"views":               1,
				"favorers_count":      1,
				"shipping_profile_id": 1,
				"processing":          1,
				"non_taxable":         1,
				"variations":          1,
				"should_auto_renew":   1,
				"inventory":           1,
				"recent_reviews":      1,
				"reviews_count":       1,
				"rating":              1,
				"measurements":        1,
				"user": bson.M{
					"login_name":             "$user.login_name",
					"first_name":             "$user.first_name",
					"last_name":              "$user.last_name",
					"thumbnail":              "$user.thumbnail",
					"transaction_buy_count":  "$user.transaction_buy_count",
					"transaction_sold_count": "$user.transaction_sold_count",
				},
				"shop": bson.M{
					"name":          "$shop.username",
					"username":      "$shop.username",
					"slug":          "$shop.slug",
					"logo_url":      "$shop.logo_url",
					"location":      "$shop.location",
					"description":   "$shop.description",
					"reviews_count": "$shop.reviews_count",
					"rating":        "$shop.rating",
					"is_live":       "$shop.is_live",
					"created_at":    "$shop.created_at",
				},
				"shipping": bson.M{
					"title":                "$shipping.title",
					"destination_by":       "$shipping.destination_by",
					"destinations":         "$shipping.destinations",
					"min_delivery_days":    "$shipping.min_delivery_days",
					"max_delivery_days":    "$shipping.max_delivery_days",
					"origin_state":         "$shipping.origin_state",
					"origin_postal_code":   "$shipping.origin_postal_code",
					"primary_price":        "$shipping.primary_price",
					"secondary_price":      "$shipping.secondary_price",
					"handling_fee":         "$shipping.handling_fee",
					"shipping_methods":     "$shipping.shipping_methods",
					"is_default_profile":   "$shipping.is_default_profile",
					"offers_free_shipping": "$shipping.offers_free_shipping",
					"processing":           "$shipping.processing",
					"service":              "$shop.service",
					"policy":               "$shipping.policy",
				},
				"siblings": 1,
			},
		},
	}

	cursor, err := s.listingCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var listing models.ListingExtra
	if cursor.Next(ctx) {
		if err := cursor.Decode(&listing); err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("no listing found")
	}

	return &listing, nil
}

// GetListings retrieves multiple listings with filters and pagination
func (s *listingService) GetListings(ctx context.Context, pagination util.PaginationArgs, filters bson.M, sort bson.D) ([]models.ListingExtra, int64, error) {
	pipeline := []bson.M{
		{"$match": filters},
		{
			"$lookup": bson.M{
				"from":         "Shop",
				"localField":   "shop_id",
				"foreignField": "_id",
				"as":           "shop",
			},
		},
		{"$unwind": "$shop"},
		{
			"$lookup": bson.M{
				"from":         "User",
				"localField":   "user_id",
				"foreignField": "_id",
				"as":           "user",
			},
		},
		{"$unwind": "$user"},
		{
			"$project": bson.M{
				"_id":                 1,
				"state":               1,
				"user_id":             1,
				"shop_id":             1,
				"main_image":          1,
				"images":              1,
				"details":             1,
				"date":                1,
				"slug":                1,
				"views":               1,
				"favorers_count":      1,
				"shipping_profile_id": 1,
				"processing":          1,
				"non_taxable":         1,
				"variations":          1,
				"should_auto_renew":   1,
				"inventory":           1,
				"recent_reviews":      1,
				"reviews_count":       1,
				"total_orders":        1,
				"sales":               1,
				"rating":              1,
				"user": bson.M{
					"login_name": "$user.login_name",
					"first_name": "$user.first_name",
					"last_name":  "$user.last_name",
					"thumbnail":  "$user.thumbnail",
				},
				"shop": bson.M{
					"name":          "$shop.name",
					"username":      "$shop.username",
					"slug":          "$shop.slug",
					"logo_url":      "$shop.logo_url",
					"location":      "$shop.location",
					"description":   "$shop.description",
					"reviews_count": "$shop.reviews_count",
				},
			},
		},
		{"$sort": sort},
		{"$skip": int64(pagination.Skip)},
		{"$limit": int64(pagination.Limit)},
	}

	cursor, err := s.listingCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var listings []models.ListingExtra
	if err := cursor.All(ctx, &listings); err != nil {
		return nil, 0, err
	}

	// Count total documents
	countPipeline := []bson.M{
		{"$match": filters},
		{"$count": "total"},
	}
	countCursor, err := s.listingCollection.Aggregate(ctx, countPipeline)
	if err != nil {
		return nil, 0, err
	}
	defer countCursor.Close(ctx)

	var countResult struct {
		Total int64 `bson:"total"`
	}
	if countCursor.Next(ctx) {
		if err := countCursor.Decode(&countResult); err != nil {
			return nil, 0, err
		}
	}

	return listings, countResult.Total, nil
}

// GetMyListingsSummary gets listing summaries for a specific shop and user
func (s *listingService) GetMyListingsSummary(ctx context.Context, shopID, userID primitive.ObjectID, pagination util.PaginationArgs, sort bson.D) ([]models.ListingsSummary, int64, error) {
	findOptions := options.Find().
		SetLimit(int64(pagination.Limit)).
		SetSkip(int64(pagination.Skip)).
		SetSort(sort)

	filter := bson.M{"shop_id": shopID, "user_id": userID}
	cursor, err := s.listingCollection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var listings []models.ListingsSummary
	if err := cursor.All(ctx, &listings); err != nil {
		return nil, 0, err
	}

	count, err := s.listingCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return listings, count, nil
}

// GetShopListings gets all listings for a specific shop
func (s *listingService) GetShopListings(ctx context.Context, shopID primitive.ObjectID, pagination util.PaginationArgs, filters bson.M, sort bson.D) ([]models.Listing, int64, error) {
	filters["shop_id"] = shopID

	findOptions := options.Find().
		SetLimit(int64(pagination.Limit)).
		SetSkip(int64(pagination.Skip)).
		SetSort(sort)

	cursor, err := s.listingCollection.Find(ctx, filters, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	count, err := s.listingCollection.CountDocuments(ctx, filters)
	if err != nil {
		return nil, 0, err
	}

	var listings []models.Listing
	if err = cursor.All(ctx, &listings); err != nil {
		return nil, 0, err
	}

	return listings, count, nil
}

// UpdateListingState updates the state of multiple listings
func (s *listingService) UpdateListingState(ctx context.Context, userID primitive.ObjectID, listingIDs []string, newState models.ListingStateType) (*UpdateListingStateResult, error) {
	now := time.Now()
	result := &UpdateListingStateResult{
		UpdatedListings:    []primitive.ObjectID{},
		NotUpdatedListings: []primitive.ObjectID{},
	}

	for _, id := range listingIDs {
		idObjectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			result.NotUpdatedListings = append(result.NotUpdatedListings, idObjectID)
			continue
		}

		_, err = s.listingCollection.UpdateOne(ctx,
			bson.M{"_id": idObjectID, "user_id": userID},
			bson.M{"$set": bson.M{
				"state.state":            newState,
				"state.state_updated_at": now,
				"date.modified_at":       now,
			}},
		)
		if err != nil {
			result.NotUpdatedListings = append(result.NotUpdatedListings, idObjectID)
			continue
		}

		result.UpdatedListings = append(result.UpdatedListings, idObjectID)
	}

	return result, nil
}

// HasUserCreatedListing checks if a user has created at least one listing
func (s *listingService) HasUserCreatedListing(ctx context.Context, userID primitive.ObjectID) (bool, error) {
	filter := bson.M{"user_id": userID}
	findOptions := options.Find().SetLimit(1)

	cursor, err := s.listingCollection.Find(ctx, filter, findOptions)
	if err != nil {
		return false, err
	}
	defer cursor.Close(ctx)

	var listing []models.Listing
	if err := cursor.All(ctx, &listing); err != nil {
		return false, err
	}

	return len(listing) > 0, nil
}

// UpdateListing updates an existing listing with partial updates including image handling
func (s *listingService) UpdateListing(ctx context.Context, req UpdateListingRequest) error {
	err := s.VerifyListingOwnership(ctx, req.UserID, req.ListingID)
	if err != nil {
		return err
	}

	var currentListing models.Listing
	err = s.listingCollection.FindOne(ctx, bson.M{"_id": req.ListingID}).Decode(&currentListing)
	if err != nil {
		return fmt.Errorf("failed to fetch current listing: %v", err)
	}

	updateDoc := bson.M{}
	now := time.Now()

	if req.NewMainImageURL != nil {
		updateDoc["main_image"] = *req.NewMainImageURL
		s.scheduleImageCleanup(currentListing.MainImage, append(currentListing.Images, *req.NewMainImageURL))
	} else if !req.KeepMainImage {
		updateDoc["main_image"] = common.DEFAULT_THUMBNAIL
		s.scheduleImageCleanup(currentListing.MainImage, currentListing.Images)
	}

	updatedImages := s.processImageUpdates(
		currentListing.Images,
		req.ImagesToAdd,
		req.ImagesToRemove,
		req.ImageOrder,
	)
	updateDoc["images"] = updatedImages

	if req.UpdatedListing.Details != nil {
		if err := req.UpdatedListing.Details.SetDynamicToTypedField(); err != nil {
			return fmt.Errorf("invalid dynamic data: %v", err)
		}

		s.addDetailsUpdates(updateDoc, req.UpdatedListing.Details)
	}

	if req.UpdatedListing.Inventory != nil {
		s.addInventoryUpdates(updateDoc, req.UpdatedListing.Inventory)
	}

	if req.UpdatedListing.Variations != nil {
		updateDoc["variations"] = req.UpdatedListing.Variations
	}

	updateDoc["date.modified_at"] = now

	result, err := s.listingCollection.UpdateOne(
		ctx,
		bson.M{"_id": req.ListingID, "user_id": req.UserID},
		bson.M{"$set": updateDoc},
	)
	if err != nil {
		s.cleanupUploadedImages(req.MainImageResult, req.NewImageResults)
		return fmt.Errorf("failed to update listing: %v", err)
	}

	if result.MatchedCount == 0 {
		s.cleanupUploadedImages(req.MainImageResult, req.NewImageResults)
		return errors.New("listing not found or unauthorized")
	}

	for _, imageURL := range req.ImagesToRemove {
		if imageURL != common.DEFAULT_THUMBNAIL {
			go s.cleanupImageByURL(imageURL)
		}
	}

	return nil
}

// processImageUpdates handles the logic for updating the images array
func (s *listingService) processImageUpdates(current []string, toAdd []string, toRemove []string, order []string) []string {
	if len(order) > 0 {
		return order
	}

	imageMap := make(map[string]bool)
	for _, img := range current {
		imageMap[img] = true
	}

	for _, img := range toRemove {
		delete(imageMap, img)
	}

	var result []string
	for _, img := range current {
		if imageMap[img] {
			result = append(result, img)
		}
	}

	result = append(result, toAdd...)

	return result
}

// addDetailsUpdates adds listing details updates to the update document
func (s *listingService) addDetailsUpdates(updateDoc bson.M, details *models.UpdateListingDetails) {
	if details.Title != nil {
		updateDoc["details.title"] = *details.Title
	}
	if details.Description != nil {
		updateDoc["details.description"] = *details.Description
	}
	if details.Category != nil {
		updateDoc["details.category"] = *details.Category
	}
	if details.Condition != nil {
		updateDoc["details.condition"] = *details.Condition
	}
	if details.Type != nil {
		updateDoc["details.type"] = *details.Type
	}
	if details.WhoMade != nil {
		updateDoc["details.who_made"] = *details.WhoMade
	}
	if details.WhenMade != nil {
		updateDoc["details.when_made"] = *details.WhenMade
	}
	if details.Sustainability != nil {
		updateDoc["details.sustainability"] = *details.Sustainability
	}
	if details.ShippingProfileId != nil {
		if shippingID, err := primitive.ObjectIDFromHex(*details.ShippingProfileId); err == nil {
			updateDoc["shipping_profile_id"] = shippingID
		}
	}
	if details.Tags != nil {
		updateDoc["details.tags"] = details.Tags
	}
	if details.Keywords != nil {
		updateDoc["details.keywords"] = details.Keywords
	}
	if details.HasPersonalization != nil {
		updateDoc["details.has_personalization"] = *details.HasPersonalization
	}
	if details.Personalization != nil {
		updateDoc["details.personalization"] = *details.Personalization
	}

	if details.ClothingData != nil {
		updateDoc["details.clothing_data"] = details.ClothingData
	}
	if details.FurnitureData != nil {
		updateDoc["details.furniture_data"] = details.FurnitureData
	}
	if details.AceessoriesAndJewelryData != nil {
		updateDoc["details.accessories_and_jewelry_data"] = details.AceessoriesAndJewelryData
	}
	if details.GiftsAndOccasionsData != nil {
		updateDoc["details.gifts_and_occasions_data"] = details.GiftsAndOccasionsData
	}
	if details.ArtAndCollectiblesData != nil {
		updateDoc["details.art_and_collectibles_data"] = details.ArtAndCollectiblesData
	}
	if details.HomeAndLivingData != nil {
		updateDoc["details.home_and_living_data"] = details.HomeAndLivingData
	}
}

// addInventoryUpdates adds inventory updates to the update document
func (s *listingService) addInventoryUpdates(updateDoc bson.M, inventory *models.UpdateInventory) {
	updateDoc["inventory.modified_at"] = time.Now()

	if inventory.Price != nil {
		updateDoc["inventory.price"] = *inventory.Price
	}
	if inventory.Quantity != nil {
		updateDoc["inventory.quantity"] = *inventory.Quantity
	}
	if inventory.SKU != nil {
		updateDoc["inventory.sku"] = *inventory.SKU
	}
	if inventory.DomesticPricing != nil {
		updateDoc["inventory.domestic_pricing"] = *inventory.DomesticPricing
	}
	if inventory.DomesticPrice != nil {
		updateDoc["inventory.domestic_price"] = *inventory.DomesticPrice
	}
}

// scheduleImageCleanup checks if an image should be cleaned up
func (s *listingService) scheduleImageCleanup(imageURL string, keepImages []string) {
	// Don't cleanup default thumbnail
	if imageURL == common.DEFAULT_THUMBNAIL {
		return
	}

	for _, img := range keepImages {
		if img == imageURL {
			return
		}
	}

	go s.cleanupImageByURL(imageURL)
}

// cleanupImageByURL extracts public ID from Cloudinary URL and deletes it
func (s *listingService) cleanupImageByURL(imageURL string) {
	// Extract public ID from Cloudinary URL
	// Example: https://res.cloudinary.com/kh-oo-mi/image/upload/v1234567890/folder/publicid.jpg
	parts := strings.Split(imageURL, "/")
	if len(parts) < 2 {
		return
	}

	// Get everything after "upload/"
	uploadIndex := -1
	for i, part := range parts {
		if part == "upload" {
			uploadIndex = i
			break
		}
	}

	if uploadIndex == -1 || uploadIndex+2 >= len(parts) {
		return
	}

	publicIDParts := parts[uploadIndex+2:]
	if len(publicIDParts) > 0 {
		lastPart := publicIDParts[len(publicIDParts)-1]
		extIndex := strings.LastIndex(lastPart, ".")
		if extIndex > 0 {
			publicIDParts[len(publicIDParts)-1] = lastPart[:extIndex]
		}

		publicID := strings.Join(publicIDParts, "/")
		if publicID != "" {
			if _, err := util.DestroyMedia(publicID); err != nil {
				log.Printf("Failed to cleanup image %s: %v", publicID, err)
			}
		}
	}
}

// cleanupUploadedImages cleans up images that were uploaded but update failed
func (s *listingService) cleanupUploadedImages(mainImageResult any, newImageResults []any) {
	if mainResult, ok := mainImageResult.(uploader.UploadResult); ok && mainResult.PublicID != "" {
		if _, err := util.DestroyMedia(mainResult.PublicID); err != nil {
			log.Printf("Failed to cleanup main image %s: %v", mainResult.PublicID, err)
		}
	}

	for _, result := range newImageResults {
		if imgResult, ok := result.(uploader.UploadResult); ok && imgResult.PublicID != "" {
			if _, err := util.DestroyMedia(imgResult.PublicID); err != nil {
				log.Printf("Failed to cleanup image %s: %v", imgResult.PublicID, err)
			}
		}
	}
}
