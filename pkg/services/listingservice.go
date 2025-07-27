package services

import (
	"context"
	"errors"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"khoomi-api-io/api/internal/common"
	"khoomi-api-io/api/pkg/util"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

type listingService struct{}

func NewListingService() ListingService {
	return &listingService{}
}

// VerifyListingOwnership verifies if a user owns a given listing using its listingId
func (s *listingService) VerifyListingOwnership(ctx context.Context, userID, listingID primitive.ObjectID) error {
	// Use FindOne with projection to only fetch _id field - most efficient approach
	var result struct {
		ID primitive.ObjectID `bson:"_id"`
	}
	err := common.ListingCollection.FindOne(ctx, bson.M{"_id": listingID, "user_id": userID}).Decode(&result)
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
			err := common.ListingCollection.FindOne(sessionCtx, bson.M{"_id": listingID, "user_id": userID}).Decode(&listing)
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
			deleteReviewsResult, err := common.ListingReviewCollection.DeleteMany(sessionCtx, reviewFilter)
			if err != nil {
				return nil, err
			}
			totalDeletedReviews += deleteReviewsResult.DeletedCount

			deleteResult, err := common.ListingCollection.DeleteOne(sessionCtx, bson.M{"_id": listingID, "user_id": userID})
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
			updateResult, err := common.ShopCollection.UpdateOne(sessionCtx, shopFilter, shopUpdate)
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
				_, err = common.ShopCollection.UpdateOne(sessionCtx, shopFilter, shopRatingUpdate)
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

	wc := writeconcern.New(writeconcern.WMajority())
	txnOptions := options.Transaction().SetWriteConcern(wc)
	session, err := util.DB.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, callback, txnOptions)
	if err != nil {
		return nil, err
	}

	return result, nil
}
