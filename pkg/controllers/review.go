package controllers

import (
	"context"
	"fmt"
	"khoomi-api-io/api/internal"
	"khoomi-api-io/api/internal/common"
	"khoomi-api-io/api/pkg/models"
	"khoomi-api-io/api/pkg/util"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

// calculateListingRating recalculates the listing's average rating and star distribution
func calculateListingRating(ctx context.Context, listingId primitive.ObjectID) (models.Rating, error) {
	pipeline := []bson.M{
		{"$match": bson.M{"listing_id": listingId}},
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

// calculateShopRating recalculates the shop's average rating based on listing reviews from the past 12 months
func calculateShopRating(ctx context.Context, shopId primitive.ObjectID) (models.Rating, error) {
	// Calculate the date 12 months ago
	twelveMonthsAgo := time.Now().AddDate(0, -12, 0)

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"shop_id":    shopId,
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

// CreateListingReview - api/listing/:listingid/reviews
func CreateListingReview() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

		var listingReviewJson models.ReviewRequest
		now := time.Now()
		defer cancel()

		listingId, myId, err := common.ListingIdAndMyId(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		err = c.BindJSON(&listingReviewJson)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		if validationErr := common.Validate.Struct(&listingReviewJson); validationErr != nil {
			util.HandleError(c, http.StatusBadRequest, validationErr)
			return
		}

		wc := writeconcern.New(writeconcern.WMajority())
		txnOptions := options.Transaction().SetWriteConcern(wc)

		dbSession, err := util.DB.StartSession()
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}
		defer dbSession.EndSession(ctx)

		callback := func(sessionContext mongo.SessionContext) (interface{}, error) {
			var userProfile models.User
			err := common.UserCollection.FindOne(sessionContext, bson.M{"_id": myId}).Decode(&userProfile)
			if err != nil {
				return nil, err
			}
			reviewId := primitive.NewObjectID()

			// attempt to add review to listing review collection
			listingReviewData := models.ListingReview{
				Id:           reviewId,
				UserId:       myId,
				ListingId:    listingId,
				ShopId:       listingReviewJson.ShopId,
				Review:       listingReviewJson.Review,
				ReviewAuthor: strings.Join([]string{userProfile.FirstName, userProfile.LastName}, " "),
				Thumbnail:    userProfile.Thumbnail,
				Rating:       listingReviewJson.Rating,
				CreatedAt:    now,
				Status:       models.ReviewStatusApproved,
			}

			// Use upsert to create a new review if one doesn't exist, or update an existing one.
			_, err = common.ListingReviewCollection.InsertOne(sessionContext, listingReviewData)
			if err != nil {
				log.Println(err)
				return nil, err
			}

			// Calculate and update the listing's rating
			listingRating, err := calculateListingRating(sessionContext, listingId)
			if err != nil {
				log.Println("Failed to calculate listing rating:", err)
				return nil, err
			}

			updateResult, err := common.ListingCollection.UpdateOne(sessionContext, bson.M{"_id": listingId}, bson.M{"$set": bson.M{"rating": listingRating, "date.modified_at": now}})
			if err != nil {
				log.Println("Failed to update listing rating:", err)
				return nil, err
			}
			log.Printf("Listing update result: %+v\n", updateResult)

			// Calculate and update the shop's rating based on listing reviews from past 12 months
			shopRating, err := calculateShopRating(sessionContext, listingReviewData.ShopId)
			if err != nil {
				log.Println("Failed to calculate shop rating:", err)
				return nil, err
			}

			log.Println(shopRating)
			_, err = common.ShopCollection.UpdateOne(sessionContext, bson.M{"_id": listingReviewData.ShopId}, bson.M{"$set": bson.M{"rating": shopRating, "modified_at": now}})
			if err != nil {
				log.Println("Failed to update shop rating:", err)
				return nil, err
			}

			return updateResult, nil
		}

		result, err := dbSession.WithTransaction(ctx, callback, txnOptions)

		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		if err := dbSession.CommitTransaction(ctx); err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}
		dbSession.EndSession(ctx)

		internal.PublishCacheMessage(c, internal.CacheInvalidateListingReviews, listingId.Hex())
		util.HandleSuccess(c, http.StatusOK, "Review Added Successfully", result)
	}
}

// GetShopReviews - api/shops/:shopid/reviews?limit=50&skip=0
func GetShopReviews() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		shopIdStr := c.Param("shopid")
		shopId, err := primitive.ObjectIDFromHex(shopIdStr)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		paginationArgs := common.GetPaginationArgs(c)

		// Build aggregation pipeline to get listing reviews with listing info
		pipeline := []bson.M{
			{"$match": bson.M{"shop_id": shopId, "status": models.ReviewStatusApproved}},
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
			{"$skip": int64(paginationArgs.Skip)},
			{"$limit": int64(paginationArgs.Limit)},
		}

		cursor, err := common.ListingReviewCollection.Aggregate(ctx, pipeline)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}
		defer cursor.Close(ctx)

		var reviews []bson.M
		if err = cursor.All(ctx, &reviews); err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}

		// Get total count for pagination
		countPipeline := []bson.M{
			{"$match": bson.M{"shop_id": shopId, "status": models.ReviewStatusApproved}},
			{"$count": "total"},
		}

		countCursor, err := common.ListingReviewCollection.Aggregate(ctx, countPipeline)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}
		defer countCursor.Close(ctx)

		var countResult []bson.M
		if err = countCursor.All(ctx, &countResult); err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		var totalCount int64 = 0
		if len(countResult) > 0 {
			if count, ok := countResult[0]["total"].(int32); ok {
				totalCount = int64(count)
			}
		}

		util.HandleSuccessMeta(c, http.StatusOK, "success", reviews, gin.H{
			"pagination": util.Pagination{
				Limit: paginationArgs.Limit,
				Skip:  paginationArgs.Skip,
				Count: totalCount,
			},
		})
	}
}

// GetShopReviews - api/listing/:listingid/reviews?limit=50&skip=0
func GetListingReviews() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		listingIdStr := c.Param("listingid")
		listingId, err := primitive.ObjectIDFromHex(listingIdStr)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		filter := bson.M{"listing_id": listingId}
		paginationArgs := common.GetPaginationArgs(c)
		find := options.Find().SetLimit(int64(paginationArgs.Limit)).SetSkip(int64(paginationArgs.Skip))
		result, err := common.ListingReviewCollection.Find(ctx, filter, find)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}

		var reviews []models.ListingReview
		if err = result.All(ctx, &reviews); err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}

		count, err := common.ListingReviewCollection.CountDocuments(ctx, filter)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		util.HandleSuccessMeta(c, http.StatusOK, "success", reviews, gin.H{
			"pagination": util.Pagination{
				Limit: paginationArgs.Limit,
				Skip:  paginationArgs.Skip,
				Count: count,
			},
		})
	}
}

// DeleteMyReview - api/listing/:listingid/reviews
func DeleteMyListingReview() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		listingId, myId, err := common.ListingIdAndMyId(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		// Shop Member session
		wc := writeconcern.New(writeconcern.WMajority())
		txnOptions := options.Transaction().SetWriteConcern(wc)

		session, err := util.DB.StartSession()
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, fmt.Errorf("failed to start session: %v", err))
			return
		}
		defer session.EndSession(ctx)

		var deletedReviewId any
		var shopId primitive.ObjectID
		callback := func(ctx mongo.SessionContext) (any, error) {
			// Get the review first to obtain the shop ID
			filter := bson.M{"listing_id": listingId, "user_id": myId}
			var reviewToDelete models.ListingReview
			err := common.ListingReviewCollection.FindOne(ctx, filter).Decode(&reviewToDelete)
			if err != nil {
				return nil, err
			}
			shopId = reviewToDelete.ShopId

			// Attempt to remove review from review collection table
			_, err = common.ListingReviewCollection.DeleteOne(ctx, filter)
			if err != nil {
				return nil, err
			}

			// Attempt to remove member from embedded field in listing
			filter = bson.M{"_id": listingId}
			update := bson.M{"$pull": bson.M{"recent_reviews": bson.M{"user_id": myId}}, "$inc": bson.M{"rating.review_count": -1}}
			result2, err := common.ListingCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				return nil, err
			}

			if result2.ModifiedCount == 0 {
				return nil, errors.New("no matching documents found")
			}

			// attempt updating user reviewCount fields.
			_, err = common.UserCollection.UpdateOne(ctx, bson.M{"_id": myId}, bson.M{"$inc": bson.M{"review_count": -1}})
			if err != nil {
				log.Println("Failed to update review count:", err)
				return nil, err
			}

			// recalculate and update listing rating
			listingRating, err := calculateListingRating(ctx, listingId)
			if err != nil {
				log.Println("Failed to calculate listing rating:", err)
				return nil, err
			}

			// update listing with new rating
			_, err = common.ListingCollection.UpdateOne(ctx, bson.M{"_id": listingId}, bson.M{"$set": bson.M{"rating": listingRating}})
			if err != nil {
				log.Println("Failed to update listing rating:", err)
				return nil, err
			}

			// recalculate and update shop rating
			shopRating, err := calculateShopRating(ctx, shopId)
			if err != nil {
				log.Println("Failed to calculate shop rating:", err)
				return nil, err
			}

			_, err = common.ShopCollection.UpdateOne(ctx, bson.M{"_id": shopId}, bson.M{"$set": bson.M{"rating": shopRating}})
			if err != nil {
				log.Println("Failed to update shop rating:", err)
				return nil, err
			}

			deletedReviewId = result2.UpsertedID
			return result2, nil
		}

		_, err = session.WithTransaction(context.Background(), callback, txnOptions)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		internal.PublishCacheMessage(c, internal.CacheInvalidateListingReviews, listingId.Hex())

		util.HandleSuccess(c, http.StatusOK, "My review was deleted successfully", deletedReviewId)
	}
}

// DeleteOtherReview - api/listing/:listingid/reviews/other?userid={user_id to remove}
func DeleteOtherListingReview() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		userToBeRemoved := c.Query("userid")
		userToBeRemovedId, err := primitive.ObjectIDFromHex(userToBeRemoved)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		listingId, myId, err := common.ListingIdAndMyId(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}
		err = common.VerifyListingOwnership(c, myId, listingId)
		if err != nil {
			log.Printf("You don't have write access to this listing: %v", err.Error())
			util.HandleError(c, http.StatusUnauthorized, err)
			return
		}

		var deletedReviewId any
		var shopId primitive.ObjectID
		// Shop review session
		wc := writeconcern.New(writeconcern.WMajority())
		txnOptions := options.Transaction().SetWriteConcern(wc)
		session, err := util.DB.StartSession()
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, fmt.Errorf("failed to start session: %v", err))
			return
		}
		defer session.EndSession(ctx)

		callback := func(ctx mongo.SessionContext) (any, error) {
			// Get the review first to obtain the shop ID
			filter := bson.M{"listing_id": listingId, "user_id": userToBeRemovedId}
			var reviewToDelete models.ListingReview
			err = common.ListingReviewCollection.FindOne(ctx, filter).Decode(&reviewToDelete)
			if err != nil {
				return nil, err
			}
			shopId = reviewToDelete.ShopId

			// Attempt to remove review from review collection table
			_, err = common.ListingReviewCollection.DeleteOne(ctx, filter)
			if err != nil {
				return nil, err
			}

			// Attempt to remove review from recent review field in listing
			filter = bson.M{"_id": listingId}
			update := bson.M{"$pull": bson.M{"recent_reviews": bson.M{"user_id": userToBeRemovedId}}, "$inc": bson.M{"rating.review_count": -1}}
			result2, err := common.ListingCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				return nil, err
			}

			if result2.ModifiedCount == 0 {
				return nil, errors.New("no matching documents found")
			}

			// recalculate and update listing rating
			listingRating, err := calculateListingRating(ctx, listingId)
			if err != nil {
				log.Println("Failed to calculate listing rating:", err)
				return nil, err
			}

			// update listing with new rating
			_, err = common.ListingCollection.UpdateOne(ctx, bson.M{"_id": listingId}, bson.M{"$set": bson.M{"rating": listingRating}})
			if err != nil {
				log.Println("Failed to update listing rating:", err)
				return nil, err
			}

			// recalculate and update shop rating
			shopRating, err := calculateShopRating(ctx, shopId)
			if err != nil {
				log.Println("Failed to calculate shop rating:", err)
				return nil, err
			}

			_, err = common.ShopCollection.UpdateOne(ctx, bson.M{"_id": shopId}, bson.M{"$set": bson.M{"rating": shopRating}})
			if err != nil {
				log.Println("Failed to update shop rating:", err)
				return nil, err
			}

			deletedReviewId = result2.UpsertedID
			return result2, nil
		}

		_, err = session.WithTransaction(context.Background(), callback, txnOptions)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}

		internal.PublishCacheMessage(c, internal.CacheInvalidateListingReviews, listingId.Hex())

		util.HandleSuccess(c, http.StatusOK, "Other user review deleted successfully", deletedReviewId)
	}
}
