package controllers

import (
	"context"
	"fmt"
	"khoomi-api-io/api/internal"
	"khoomi-api-io/api/internal/auth"
	"khoomi-api-io/api/internal/common"
	"khoomi-api-io/api/pkg/util"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

func ToggleFavoriteShop() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		shopIdStr := c.Query("shopid")
		shopId, err := primitive.ObjectIDFromHex(shopIdStr)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		action := c.Query("action")

		myObjectId, err := auth.GetSessionUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err)
			return
		}

		// Favorite shop session.
		wc := writeconcern.New(writeconcern.WMajority())
		txnOptions := options.Transaction().SetWriteConcern(wc)
		session, err := util.DB.StartSession()
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, fmt.Errorf("failed to start session: %v", err))
			return
		}
		callback := func(ctx mongo.SessionContext) (any, error) {
			filter := bson.M{"_id": myObjectId}
			// update user favorite shops field
			switch action {
			case "add":
				update := bson.M{"$push": bson.M{"favorite_shops": shopId}}
				_, err := common.UserCollection.UpdateOne(ctx, filter, update)
				if err != nil {
					return nil, err
				}

				result, err := common.UserFavoriteShopCollection.InsertOne(ctx, bson.M{"shopId": shopId, "userId": myObjectId})
				if err != nil {
					return nil, err
				}
				return result, nil
			case "remove":
				update := bson.M{"$pull": bson.M{"favorite_shops": shopId}}
				_, err := common.UserCollection.UpdateOne(ctx, filter, update)
				if err != nil {
					return nil, err
				}

				result, err := common.UserFavoriteShopCollection.DeleteOne(ctx, bson.M{"shopId": shopId, "userId": myObjectId})
				if err != nil {
					return nil, err
				}
				return result, nil

			default:
				return nil, errors.New("action query is missing from url")
			}

		}

		_, err = session.WithTransaction(ctx, callback, txnOptions)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}
		if err := session.CommitTransaction(context.Background()); err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}
		session.EndSession(context.Background())

		internal.PublishCacheMessage(c, internal.CacheInvalidateShopFavoriteToggle, shopIdStr)

		util.HandleSuccess(c, http.StatusOK, "Favorite shops updated!", gin.H{})

	}
}

// IsShopFavorited
func IsShopFavorited() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		myObjectId, err := auth.GetSessionUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		shopIdStr := c.Query("shopid")
		shopId, err := primitive.ObjectIDFromHex(shopIdStr)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		result := common.UserFavoriteListingCollection.FindOne(ctx, bson.M{"shopId": shopId, "userId": myObjectId})
		if result.Err() != nil {
			util.HandleSuccess(c, http.StatusOK, "not found", gin.H{"favorited": false})
			return
		}

		util.HandleSuccess(c, http.StatusOK, "found one match", gin.H{"favorited": true})

	}
}

// ToggleFavoriteListing
func ToggleFavoriteListing() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		listingIdStr := c.Query("listingid")
		fmt.Println(listingIdStr)
		listingId, err := primitive.ObjectIDFromHex(listingIdStr)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		action := c.Query("action")

		myObjectId, err := auth.GetSessionUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err)
			return
		}

		// Favorite listing session.
		wc := writeconcern.New(writeconcern.WMajority())
		txnOptions := options.Transaction().SetWriteConcern(wc)
		session, err := util.DB.StartSession()
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, fmt.Errorf("failed to start session: %v", err))
			return
		}
		callback := func(ctx mongo.SessionContext) (any, error) {
			filter := bson.M{"_id": myObjectId}
			// update user favorite listings field
			switch action {
			case "add":
				update := bson.M{"$push": bson.M{"favorite_listings": listingIdStr}}
				_, err := common.UserCollection.UpdateOne(ctx, filter, update)
				if err != nil {
					return nil, err
				}

				result, err := common.UserFavoriteListingCollection.InsertOne(ctx, bson.M{"listingId": listingId, "userId": myObjectId})
				if err != nil {
					return nil, err
				}
				return result, nil
			case "remove":
				update := bson.M{"$pull": bson.M{"favorite_listings": listingIdStr}}
				_, err := common.UserCollection.UpdateOne(ctx, filter, update)
				if err != nil {
					return nil, err
				}

				result, err := common.UserFavoriteListingCollection.DeleteOne(ctx, bson.M{"listingId": listingId, "userId": myObjectId})
				if err != nil {
					return nil, err
				}
				return result, nil

			default:
				return nil, errors.New("action query is missing from url")
			}

		}

		_, err = session.WithTransaction(ctx, callback, txnOptions)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}
		if err := session.CommitTransaction(context.Background()); err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}
		session.EndSession(context.Background())

		internal.PublishCacheMessage(c, internal.CacheInvalidateListingFavoriteToggle, listingId.Hex())

		util.HandleSuccess(c, http.StatusOK, "Favorite listings updated!", gin.H{})

	}
}

// IsListingFavorited
func IsListingFavorited() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		myObjectId, err := auth.GetSessionUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		listingIdStr := c.Query("listingid")
		listingId, err := primitive.ObjectIDFromHex(listingIdStr)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		result := common.UserFavoriteListingCollection.FindOne(ctx, bson.M{"listingId": listingId, "userId": myObjectId})
		if result.Err() != nil {
			util.HandleSuccess(c, http.StatusOK, "not found", gin.H{"favorited": false})
			return
		}

		util.HandleSuccess(c, http.StatusOK, "found one match", gin.H{"favorited": true})

	}
}