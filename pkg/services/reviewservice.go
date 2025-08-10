package services

import (
	"context"
	"log"
	"strings"
	"time"

	"khoomi-api-io/api/internal/common"
	"khoomi-api-io/api/pkg/models"
	"khoomi-api-io/api/pkg/util"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ReviewServiceImpl struct{}

func NewReviewService() ReviewService {
	return &ReviewServiceImpl{}
}

func (rs *ReviewServiceImpl) CalculateListingRating(ctx context.Context, listingID primitive.ObjectID) (models.Rating, error) {
	pipeline := []bson.M{
		{"$match": bson.M{"listing_id": listingID}},
		{
			"$group": bson.M{
				"_id":           nil,
				"averageRating": bson.M{"$avg": "$rating"},
				"reviewCount":   bson.M{"$sum": 1},
				"fiveStarCount": bson.M{
					"$sum": bson.M{
						"$cond": bson.A{bson.M{"$eq": bson.A{"$rating", 5}}, 1, 0},
					},
				},
				"fourStarCount": bson.M{
					"$sum": bson.M{
						"$cond": bson.A{bson.M{"$eq": bson.A{"$rating", 4}}, 1, 0},
					},
				},
				"threeStarCount": bson.M{
					"$sum": bson.M{
						"$cond": bson.A{bson.M{"$eq": bson.A{"$rating", 3}}, 1, 0},
					},
				},
				"twoStarCount": bson.M{
					"$sum": bson.M{
						"$cond": bson.A{bson.M{"$eq": bson.A{"$rating", 2}}, 1, 0},
					},
				},
				"oneStarCount": bson.M{
					"$sum": bson.M{
						"$cond": bson.A{bson.M{"$eq": bson.A{"$rating", 1}}, 1, 0},
					},
				},
			},
		},
	}

	cursor, err := common.ListingReviewCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return models.Rating{}, err
	}
	defer cursor.Close(ctx)

	var result models.Rating
	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return models.Rating{}, err
		}
	}

	averageRating := float64(int(result.AverageRating*100)) / 100
	result.AverageRating = averageRating

	return result, nil
}

func (rs *ReviewServiceImpl) CalculateShopRating(ctx context.Context, shopID primitive.ObjectID) (models.Rating, error) {
	twelveMonthsAgo := time.Now().AddDate(0, -12, 0)

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"shop_id":    shopID,
				"created_at": bson.M{"$gte": twelveMonthsAgo},
			},
		},
		{
			"$group": bson.M{
				"_id":           nil,
				"averageRating": bson.M{"$avg": "$rating"},
				"reviewCount":   bson.M{"$sum": 1},
				"fiveStarCount": bson.M{
					"$sum": bson.M{
						"$cond": bson.A{bson.M{"$eq": bson.A{"$rating", 5}}, 1, 0},
					},
				},
				"fourStarCount": bson.M{
					"$sum": bson.M{
						"$cond": bson.A{bson.M{"$eq": bson.A{"$rating", 4}}, 1, 0},
					},
				},
				"threeStarCount": bson.M{
					"$sum": bson.M{
						"$cond": bson.A{bson.M{"$eq": bson.A{"$rating", 3}}, 1, 0},
					},
				},
				"twoStarCount": bson.M{
					"$sum": bson.M{
						"$cond": bson.A{bson.M{"$eq": bson.A{"$rating", 2}}, 1, 0},
					},
				},
				"oneStarCount": bson.M{
					"$sum": bson.M{
						"$cond": bson.A{bson.M{"$eq": bson.A{"$rating", 1}}, 1, 0},
					},
				},
			},
		},
	}

	cursor, err := common.ListingReviewCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return models.Rating{}, err
	}
	defer cursor.Close(ctx)

	var result models.Rating
	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return models.Rating{}, err
		}
	}

	averageRating := float64(int(result.AverageRating*100)) / 100
	result.AverageRating = averageRating

	return result, nil
}

func (rs *ReviewServiceImpl) CreateListingReview(ctx context.Context, userID, listingID primitive.ObjectID, req models.ReviewRequest) (primitive.ObjectID, error) {
	now := time.Now()

	reviewID := primitive.NewObjectID()
	callback := func(sessionContext mongo.SessionContext) (any, error) {
		var userProfile models.User
		err := common.UserCollection.FindOne(sessionContext, bson.M{"_id": userID}).Decode(&userProfile)
		if err != nil {
			return nil, err
		}

		listingReviewData := models.ListingReview{
			Id:           reviewID,
			UserId:       userID,
			ListingId:    listingID,
			ShopId:       req.ShopId,
			Review:       req.Review,
			ReviewAuthor: strings.Join([]string{userProfile.FirstName, userProfile.LastName}, " "),
			Thumbnail:    userProfile.Thumbnail,
			Rating:       req.Rating,
			CreatedAt:    now,
			Status:       models.ReviewStatusApproved,
		}
		_, err = common.ListingReviewCollection.InsertOne(sessionContext, listingReviewData)
		if err != nil {
			log.Println("Failed to insert review:", err)
			return nil, err
		}

		listingRating, err := rs.CalculateListingRating(sessionContext, listingID)
		if err != nil {
			log.Println("Failed to calculate listing rating:", err)
			return nil, err
		}

		updateResult, err := common.ListingCollection.UpdateOne(
			sessionContext,
			bson.M{"_id": listingID},
			bson.M{"$set": bson.M{"rating": listingRating, "date.modified_at": now}},
		)
		if err != nil {
			log.Println("Failed to update listing rating:", err)
			return nil, err
		}

		shopRating, err := rs.CalculateShopRating(sessionContext, req.ShopId)
		if err != nil {
			log.Println("Failed to calculate shop rating:", err)
			return nil, err
		}

		_, err = common.ShopCollection.UpdateOne(
			sessionContext,
			bson.M{"_id": req.ShopId},
			bson.M{"$set": bson.M{"rating": shopRating, "modified_at": now}},
		)
		if err != nil {
			log.Println("Failed to update shop rating:", err)
			return nil, err
		}

		return updateResult, nil
	}

	_, err := ExecuteTransaction(ctx, callback)
	return reviewID, err
}

func (rs *ReviewServiceImpl) GetListingReviews(ctx context.Context, listingID primitive.ObjectID, pagination util.PaginationArgs) ([]models.ListingReview, int64, error) {
	filter := bson.M{"listing_id": listingID}
	find := options.Find().SetLimit(int64(pagination.Limit)).SetSkip(int64(pagination.Skip))

	result, err := common.ListingReviewCollection.Find(ctx, filter, find)
	if err != nil {
		return nil, 0, err
	}

	var reviews []models.ListingReview
	if err = result.All(ctx, &reviews); err != nil {
		return nil, 0, err
	}

	count, err := common.ListingReviewCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return reviews, count, nil
}

func (rs *ReviewServiceImpl) GetShopReviews(ctx context.Context, shopID primitive.ObjectID, pagination util.PaginationArgs) ([]interface{}, int64, error) {
	pipeline := []bson.M{
		{"$match": bson.M{"shop_id": shopID, "status": models.ReviewStatusApproved}},
		{
			"$lookup": bson.M{
				"from":         "Listing",
				"localField":   "listing_id",
				"foreignField": "_id",
				"as":           "listing",
			},
		},
		{
			"$unwind": bson.M{
				"path":                       "$listing",
				"preserveNullAndEmptyArrays": true,
			},
		},
		{
			"$project": bson.M{
				"_id":           1,
				"user_id":       1,
				"listing_id":    1,
				"shop_id":       1,
				"review":        1,
				"review_author": 1,
				"thumbnail":     1,
				"rating":        1,
				"created_at":    1,
				"status":        1,
				"listing": bson.M{
					"_id":        "$listing._id",
					"title":      "$listing.details.title",
					"slug":       "$listing.slug",
					"main_image": "$listing.main_image",
					"images": bson.M{
						"$arrayElemAt": bson.A{"$listing.images", 0},
					},
				},
			},
		},
		{"$sort": bson.M{"created_at": -1}},
		{"$skip": int64(pagination.Skip)},
		{"$limit": int64(pagination.Limit)},
	}

	cursor, err := common.ListingReviewCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var reviews []bson.M
	if err = cursor.All(ctx, &reviews); err != nil {
		return nil, 0, err
	}

	// Get total count
	countPipeline := []bson.M{
		{"$match": bson.M{"shop_id": shopID, "status": models.ReviewStatusApproved}},
		{"$count": "total"},
	}

	countCursor, err := common.ListingReviewCollection.Aggregate(ctx, countPipeline)
	if err != nil {
		return nil, 0, err
	}
	defer countCursor.Close(ctx)

	var countResult []bson.M
	if err = countCursor.All(ctx, &countResult); err != nil {
		return nil, 0, err
	}

	var totalCount int64 = 0
	if len(countResult) > 0 {
		if count, ok := countResult[0]["total"].(int32); ok {
			totalCount = int64(count)
		}
	}

	result := make([]any, len(reviews))
	for i, v := range reviews {
		result[i] = v
	}

	return result, totalCount, nil
}

// DeleteMyListingReview deletes the current user's review for a listing
func (rs *ReviewServiceImpl) DeleteMyListingReview(ctx context.Context, userID, listingID primitive.ObjectID) error {

	var shopID primitive.ObjectID

	callback := func(ctx mongo.SessionContext) (any, error) {
		filter := bson.M{"listing_id": listingID, "user_id": userID}
		var reviewToDelete models.ListingReview
		err := common.ListingReviewCollection.FindOne(ctx, filter).Decode(&reviewToDelete)
		if err != nil {
			return nil, err
		}
		shopID = reviewToDelete.ShopId

		_, err = common.ListingReviewCollection.DeleteOne(ctx, filter)
		if err != nil {
			return nil, err
		}

		_, err = common.UserCollection.UpdateOne(
			ctx,
			bson.M{"_id": userID},
			bson.M{"$inc": bson.M{"review_count": -1}},
		)
		if err != nil {
			log.Println("Failed to update review count:", err)
			return nil, err
		}

		listingRating, err := rs.CalculateListingRating(ctx, listingID)
		if err != nil {
			log.Println("Failed to calculate listing rating:", err)
			return nil, err
		}

		_, err = common.ListingCollection.UpdateOne(
			ctx,
			bson.M{"_id": listingID},
			bson.M{"$set": bson.M{"rating": listingRating}},
		)
		if err != nil {
			log.Println("Failed to update listing rating:", err)
			return nil, err
		}

		shopRating, err := rs.CalculateShopRating(ctx, shopID)
		if err != nil {
			log.Println("Failed to calculate shop rating:", err)
			return nil, err
		}

		_, err = common.ShopCollection.UpdateOne(
			ctx,
			bson.M{"_id": shopID},
			bson.M{"$set": bson.M{"rating": shopRating}},
		)
		if err != nil {
			log.Println("Failed to update shop rating:", err)
			return nil, err
		}

		return nil, nil
	}

	_, err := ExecuteTransaction(ctx, callback)
	return err
}

// DeleteOtherListingReview deletes another user's review (owner operation)
func (rs *ReviewServiceImpl) DeleteOtherListingReview(ctx context.Context, ownerID, listingID, userToRemoveID primitive.ObjectID) error {
	err := rs.verifyListingOwnership(ctx, ownerID, listingID)
	if err != nil {
		return err
	}

	var shopID primitive.ObjectID

	callback := func(ctx mongo.SessionContext) (any, error) {
		// Get the review first to obtain the shop ID
		filter := bson.M{"listing_id": listingID, "user_id": userToRemoveID}
		var reviewToDelete models.ListingReview
		err = common.ListingReviewCollection.FindOne(ctx, filter).Decode(&reviewToDelete)
		if err != nil {
			return nil, err
		}
		shopID = reviewToDelete.ShopId

		// Delete review
		_, err = common.ListingReviewCollection.DeleteOne(ctx, filter)
		if err != nil {
			return nil, err
		}

		// Recalculate and update listing rating
		listingRating, err := rs.CalculateListingRating(ctx, listingID)
		if err != nil {
			log.Println("Failed to calculate listing rating:", err)
			return nil, err
		}

		_, err = common.ListingCollection.UpdateOne(
			ctx,
			bson.M{"_id": listingID},
			bson.M{"$set": bson.M{"rating": listingRating}},
		)
		if err != nil {
			log.Println("Failed to update listing rating:", err)
			return nil, err
		}

		shopRating, err := rs.CalculateShopRating(ctx, shopID)
		if err != nil {
			log.Println("Failed to calculate shop rating:", err)
			return nil, err
		}

		_, err = common.ShopCollection.UpdateOne(
			ctx,
			bson.M{"_id": shopID},
			bson.M{"$set": bson.M{"rating": shopRating}},
		)
		if err != nil {
			log.Println("Failed to update shop rating:", err)
			return nil, err
		}

		return nil, nil
	}

	_, err = ExecuteTransaction(ctx, callback)
	return err
}
