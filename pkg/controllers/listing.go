package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	auth "khoomi-api-io/api/internal/auth"
	"khoomi-api-io/api/internal/common"
	"khoomi-api-io/api/pkg/models"
	"khoomi-api-io/api/pkg/util"
	email "khoomi-api-io/api/web/email"

	"github.com/cloudinary/cloudinary-go/api/uploader"
	"github.com/gin-gonic/gin"
	slug2 "github.com/gosimple/slug"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateListing() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		shopId, myId, err := common.MyShopIdAndMyId(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}
		session, err := auth.GetSessionAuto(c)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err)
			return
		}

		loginName, loginEmail := session.LoginName, session.Email
		listingJson := c.PostForm("listing")
		if listingJson == "" {
			util.HandleError(c, http.StatusBadRequest, errors.New("missing listing payload"))
			return
		}

		var newListing models.NewListing
		if err := json.Unmarshal([]byte(listingJson), &newListing); err != nil {
			util.HandleError(c, http.StatusBadRequest, fmt.Errorf("invalid JSON: %v", err))
			return
		}

		if validationErr := common.Validate.Struct(newListing); validationErr != nil {
			util.HandleError(c, http.StatusBadRequest, validationErr)
			return
		}

		if err := newListing.Details.SetDynamicToTypedField(); err != nil {
			util.HandleError(c, http.StatusBadRequest, fmt.Errorf("invalid dynamic data: %v", err))
			return
		}

		// main_image file handling
		mainImage, _, err := c.Request.FormFile("mainImage")
		var mainImageUploadUrl uploader.UploadResult
		if err == nil {
			mainImageUploadUrl, err = util.FileUpload(models.File{File: mainImage})
			if err != nil {
				errMsg := fmt.Sprintf("Logo failed to upload - %v", err.Error())
				util.HandleError(c, http.StatusInternalServerError, errors.New(errMsg))
				return
			}
		} else {
			mainImageUploadUrl = uploader.UploadResult{}
			mainImageUploadUrl.SecureURL = common.DefaultThumbnail
		}

		if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		uploadedImagesUrl, uploadedImagesResult, err := common.HandleSequentialImages(c)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		// Verify shipping id, if null, find default shipping profile from db and use for listing.
		var shippingId primitive.ObjectID
		shippingObj, err := primitive.ObjectIDFromHex(newListing.Details.ShippingProfileId)
		if err != nil {
			var shipping models.ShopShippingProfile
			err := common.ShippingProfileCollection.FindOne(ctx, bson.M{"shop_id": shopId, "is_default_profile": true}).Decode(shipping)
			if err != nil {
				shippingId = primitive.NilObjectID
			} else {
				shippingId = shipping.ID
			}
		} else {
			shippingId = shippingObj
		}

		now := time.Now()
		listingDetails := models.Details{
			Type:               newListing.Details.Type,
			Tags:               newListing.Details.Tags,
			Title:              newListing.Details.Title,
			Color:              newListing.Details.Color,
			Dynamic:            newListing.Details.Dynamic,
			DynamicType:        newListing.Details.DynamicType,
			WhoMade:            newListing.Details.WhoMade,
			Keywords:           newListing.Details.Keywords,
			WhenMade:           newListing.Details.WhenMade,
			Category:           newListing.Details.Category,
			Condition:          newListing.Details.Condition,
			Description:        newListing.Details.Description,
			HasVariations:      newListing.Details.HasVariations,
			Sustainability:     newListing.Details.Sustainability,
			HasPersonalization: newListing.Details.HasPersonalization,
			Personalization:    newListing.Details.Personalization,

			ClothingData:              newListing.Details.ClothingData,
			FurnitureData:             newListing.Details.FurnitureData,
			GiftsAndOccasionsData:     newListing.Details.GiftsAndOccasionsData,
			ArtAndCollectiblesData:    newListing.Details.ArtAndCollectiblesData,
			AceessoriesAndJewelryData: newListing.Details.AceessoriesAndJewelryData,
			HomeAndLivingData:         newListing.Details.HomeAndLivingData,
		}

		listingInventory := models.Inventory{
			DomesticPricing: newListing.Inventory.DomesticPricing,
			DomesticPrice:   newListing.Inventory.DomesticPrice,
			Price:           newListing.Inventory.Price,
			InitialQuantity: newListing.Inventory.Quantity,
			Quantity:        newListing.Inventory.Quantity,
			SKU:             newListing.Inventory.SKU,
			CurrencyCode:    "NGN",
			ModifiedAt:      now,
		}

		listingDate := models.ListingDateMeta{
			CreatedAt:  now,
			EndingAt:   now,
			ModifiedAt: now,
		}

		listingRating := models.ListingRating{
			Rating:          0,
			ReviewCount:     0,
			PositiveReviews: 0,
			NegativeReviews: 0,
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
			ID:                   primitive.NewObjectID(),
			Code:                 common.GenerateListingCode(),
			UserId:               myId,
			ShopId:               shopId,
			MainImage:            mainImageUploadUrl.SecureURL,
			Images:               uploadedImagesUrl,
			Details:              listingDetails,
			Slug:                 slug2.Make(newListing.Details.Title),
			Date:                 listingDate,
			State:                models.ListingState{State: models.ListingStateActive, StateUpdatedAt: now},
			ShippingProfileId:    shippingId,
			NonTaxable:           true,
			ShouldAutoRenew:      false,
			Variations:           newListing.Variations,
			Inventory:            listingInventory,
			RecentReviews:        nil,
			Rating:               listingRating,
			Measurements:         newListing.Measurements,
			FinancialInformation: listingFinancialInformation,
			Views:                0,
			FavorersCount:        0,
		}

		res, err := common.ListingCollection.InsertOne(ctx, listing)
		if err != nil {
			// delete images
			_, err := util.DestroyMedia(mainImageUploadUrl.PublicID)
			for _, file := range uploadedImagesResult {
				_, err := util.DestroyMedia(file.PublicID)
				if err != nil {
					log.Println("Failed to destroy media:", err)
				}
			}
			// return error
			util.HandleError(c, http.StatusInternalServerError, fmt.Errorf("failed to create listing — %v", err))
			return
		}

		// send new listing email notification to user
		email.SendNewListingEmail(loginEmail, loginName, newListing.Details.Title)

		util.HandleSuccess(c, http.StatusOK, "Listing was created successfully", res.InsertedID)
	}
}

func GetListing() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		var listingIdentifier bson.M

		listingId := c.Param("listingid")
		if primitive.IsValidObjectID(listingId) {
			// If listingid is a valid object ID string
			listingObjectID, e := primitive.ObjectIDFromHex(listingId)
			if e != nil {
				util.HandleError(c, http.StatusBadRequest, e)
				return
			}

			listingIdentifier = bson.M{"_id": listingObjectID}
		} else {
			listingIdentifier = bson.M{"slug": listingId}
		}

		pipeline := []bson.M{
			{"$match": listingIdentifier},
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
						"is_live":       "$shop.is_live",
					},
					"shipping": bson.M{
						"title":                "$shipping.title",
						"min_processing_time":  "$shipping.min_processing_time",
						"max_processing_time":  "$shipping.max_processing_time",
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
						"service":              "$shop.service",
						"policy":               "$shipping.policy",
					},
				},
			},
		}

		cursor, err := common.ListingCollection.Aggregate(ctx, pipeline)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		var listing models.ListingExtra

		if cursor.Next(ctx) {
			if err := cursor.Decode(&listing); err != nil {
				util.HandleError(c, http.StatusInternalServerError, err)
				return
			}
		} else {
			util.HandleError(c, http.StatusNotFound, errors.New("no listing found"))
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Success", listing)
	}
}

func GetListings() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		paginationArgs := common.GetPaginationArgs(c)
		match := common.GetListingFilters(c)
		sort := common.GetListingSortingBson(paginationArgs.Sort)

		pipeline := []bson.M{
			{"$match": match},
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
			{"$skip": int64(paginationArgs.Skip)},
			{"$limit": int64(paginationArgs.Limit)},
			{"$sort": sort},
		}

		cursor, err := common.ListingCollection.Aggregate(ctx, pipeline)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		var listings []models.ListingExtra
		if err := cursor.All(ctx, &listings); err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		countPipeline := []bson.M{
			{"$match": bson.M{}},
			{"$count": "total"},
		}
		countCursor, err := common.ListingCollection.Aggregate(ctx, countPipeline)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}
		var countResult struct {
			Total int64 `bson:"total"`
		}
		if countCursor.Next(ctx) {
			if err := countCursor.Decode(&countResult); err != nil {
				util.HandleError(c, http.StatusInternalServerError, err)
				return
			}
		}

		util.HandleSuccessMeta(c, http.StatusOK, "success", listings, gin.H{
			"pagination": util.Pagination{
				Limit: paginationArgs.Limit,
				Skip:  paginationArgs.Skip,
				Count: countResult.Total,
			},
		})
	}
}

func GetMyListingsSummary() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		shopId, myId, err := common.MyShopIdAndMyId(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		paginationArgs := common.GetPaginationArgs(c)
		findOptions := options.Find().
			SetLimit(int64(paginationArgs.Limit)).
			SetSkip(int64(paginationArgs.Skip)).
			SetSort(common.GetListingSortingBson(paginationArgs.Sort))
		filter := bson.M{"shop_id": shopId, "user_id": myId}
		cursor, err := common.ListingCollection.Find(ctx, filter, findOptions)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}
		defer cursor.Close(ctx)
		var listings []models.ListingsSummary
		if err := cursor.All(ctx, &listings); err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}
		count, err := common.ListingCollection.CountDocuments(ctx, filter)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		util.HandleSuccessMeta(c, http.StatusOK, "success", listings, gin.H{
			"pagination": util.Pagination{
				Limit: paginationArgs.Limit,
				Skip:  paginationArgs.Skip,
				Count: count,
			},
		})
	}
}

// GetShopListings - Get single shop listings.
func GetShopListings() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		shopId := c.Param("shopid")
		shopObjectId, err := primitive.ObjectIDFromHex(shopId)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		match := common.GetListingFilters(c)
		match["shop_id"] = shopObjectId
		paginationArgs := common.GetPaginationArgs(c)
		findOptions := options.Find().
			SetLimit(int64(paginationArgs.Limit)).
			SetSkip(int64(paginationArgs.Skip)).
			SetSort(common.GetListingSortingBson(paginationArgs.Sort))
		cursor, err := common.ListingCollection.Find(ctx, match, findOptions)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}
		defer cursor.Close(ctx)

		count, err := common.ListingCollection.CountDocuments(ctx, match)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		var listings []models.Listing
		if err = cursor.All(ctx, &listings); err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		util.HandleSuccessMeta(c, http.StatusOK, "success", listings, gin.H{
			"pagination": util.Pagination{
				Limit: paginationArgs.Limit,
				Skip:  paginationArgs.Skip,
				Count: count,
			},
		})
	}
}

func HasUserCreatedListingOnboarding() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		shopId, userId, err := common.MyShopIdAndMyId(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		err = common.VerifyShopOwnership(c, userId, shopId)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, errors.New("Only sellers can perform this action"))
			return
		}

		filter := bson.M{"user_id": userId}
		findOptions := options.Find().SetLimit(1)
		cursor, err := common.ListingCollection.Find(ctx, filter, findOptions)
		if err != nil {
			log.Printf("error retrieving user listing: %v", err)
			util.HandleError(c, http.StatusNotFound, err)
			return
		}
		defer cursor.Close(ctx)

		var listing []models.Listing
		if err := cursor.All(ctx, &listing); err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}

		if len(listing) == 0 {
			util.HandleError(c, http.StatusNotFound, errors.New("User has no listings"))
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Success", len(listing) > 0)
	}
}

func DeleteListings() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		myId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err)
			return
		}

		listingIDs := c.PostFormArray("ids")
		if len(listingIDs) < 1 {
			util.HandleError(c, http.StatusBadRequest, errors.New("no listing IDs provided"))
			return
		}

		var deletedObjectIDs []primitive.ObjectID
		var notDeletedObjectIDs []primitive.ObjectID

		for _, id := range listingIDs {
			idObjectID, err := primitive.ObjectIDFromHex(id)
			if err != nil {
				notDeletedObjectIDs = append(notDeletedObjectIDs, idObjectID)
				continue
			}

			_, err = common.ListingCollection.DeleteOne(ctx, bson.M{"_id": idObjectID, "user_id": myId})
			if err != nil {
				notDeletedObjectIDs = append(notDeletedObjectIDs, idObjectID)
				continue
			}

			deletedObjectIDs = append(deletedObjectIDs, idObjectID)
		}

		util.HandleSuccess(c, http.StatusOK, "Listing(s) deleted", gin.H{"deleted": deletedObjectIDs, "not_deleted": notDeletedObjectIDs})
	}
}

func DeactivateListings() gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		session, err := auth.GetSessionAuto(c)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err)
			return
		}
		listingIDs := c.PostFormArray("id")
		if len(listingIDs) < 1 {
			util.HandleError(c, http.StatusBadRequest, errors.New("no listing IDs provided"))
			return
		}

		var deletedObjectIDs []primitive.ObjectID
		var notDeletedObjectIDs []primitive.ObjectID

		for _, id := range listingIDs {
			idObjectID, err := primitive.ObjectIDFromHex(id)
			if err != nil {
				notDeletedObjectIDs = append(notDeletedObjectIDs, idObjectID)
				continue
			}

			_, err = common.ListingCollection.UpdateOne(ctx, bson.M{"_id": idObjectID, "user_id": session.UserId}, bson.M{"$set": bson.M{"state.state": models.ListingStateDeactivated, "state.state_updated_at": now, "date.modified_at": now}})
			if err != nil {
				notDeletedObjectIDs = append(notDeletedObjectIDs, idObjectID)
				continue
			}

			deletedObjectIDs = append(deletedObjectIDs, idObjectID)
		}

		util.HandleSuccess(c, http.StatusOK, "Listing(s) deleted", gin.H{"deactivated": deletedObjectIDs, "not_deactivated": notDeletedObjectIDs})
	}
}
