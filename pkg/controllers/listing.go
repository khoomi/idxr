package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"khoomi-api-io/api/internal"
	auth "khoomi-api-io/api/internal/auth"
	"khoomi-api-io/api/internal/common"
	"khoomi-api-io/api/pkg/models"
	"khoomi-api-io/api/pkg/util"
	"khoomi-api-io/api/pkg/services"

	"github.com/cloudinary/cloudinary-go/api/uploader"
	"github.com/gin-gonic/gin"
	slug2 "github.com/gosimple/slug"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ListingController struct {
	listingService      services.ListingService
	shopService         services.ShopService
	notificationService services.NotificationService
}

// InitListingController initializes a new ListingController with dependencies
func InitListingController(listingService services.ListingService, shopService services.ShopService, notificationService services.NotificationService) *ListingController {
	return &ListingController{
		listingService:      listingService,
		shopService:         shopService,
		notificationService: notificationService,
	}
}

func CreateListing() gin.HandlerFunc {
	return CreateListingWithEmailService(nil)
}

func CreateListingWithEmailService(emailService services.EmailService) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQUEST_TIMEOUT_SECS)
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

		// Verify shop ownership before allowing listing creation
		// Note: This function needs to be refactored to use dependency injection
		// For now, we'll create a temporary service instance
		tempShopService := services.NewShopService()
		err = tempShopService.VerifyShopOwnership(ctx, myId, shopId)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, errors.New("only listing owners can create listings"))
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
			mainImageUploadUrl.SecureURL = common.DEFAULT_THUMBNAIL
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
			err := common.ShippingProfileCollection.FindOne(ctx, bson.M{"shop_id": shopId, "is_default_profile": true}).Decode(&shipping)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					// No default shipping profile found, set to nil
					shippingId = primitive.NilObjectID
				} else {
					// Database error, return early
					util.HandleError(c, http.StatusInternalServerError, fmt.Errorf("failed to fetch default shipping profile: %v", err))
					return
				}
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
			Dynamic:            newListing.Details.Dynamic,
			DynamicType:        newListing.Details.DynamicType,
			WhoMade:            newListing.Details.WhoMade,
			Keywords:           newListing.Details.Keywords,
			WhenMade:           newListing.Details.WhenMade,
			Category:           newListing.Details.Category,
			Condition:          newListing.Details.Condition,
			Description:        newListing.Details.Description,
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

		// Create temporary listing service instance for code generation
		tempListingService := services.NewListingService()
		
		listing := models.Listing{
			ID:                   primitive.NewObjectID(),
			Code:                 tempListingService.GenerateListingCode(),
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
			Rating:               listingRating,
			Measurements:         newListing.Measurements,
			FinancialInformation: listingFinancialInformation,
			Views:                0,
			FavorersCount:        0,
		}

		res, err := common.ListingCollection.InsertOne(ctx, listing)
		if err != nil {
			// Cleanup uploaded images on listing creation failure
			if mainImageUploadUrl.PublicID != "" {
				if _, destroyErr := util.DestroyMedia(mainImageUploadUrl.PublicID); destroyErr != nil {
					log.Printf("Failed to cleanup main image %s: %v", mainImageUploadUrl.PublicID, destroyErr)
				}
			}
			for _, file := range uploadedImagesResult {
				if _, destroyErr := util.DestroyMedia(file.PublicID); destroyErr != nil {
					log.Printf("Failed to cleanup image %s: %v", file.PublicID, destroyErr)
				}
			}

			// Return specific error based on MongoDB error type
			if mongo.IsDuplicateKeyError(err) {
				util.HandleError(c, http.StatusConflict, errors.New("listing with similar data already exists"))
			} else {
				util.HandleError(c, http.StatusInternalServerError, fmt.Errorf("database error while creating listing: %v", err))
			}
			return
		}

		// update shop listing active count
		filter := bson.M{"_id": shopId}
		update := bson.M{"$inc": bson.M{"listing_active_count": 1}}
		updateResult, err := common.ShopCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			log.Printf("Failed to update shop listing count for shop %s: %v", shopId.Hex(), err)
		} else if updateResult.MatchedCount == 0 {
			log.Printf("Warning: Shop %s not found when updating listing count", shopId.Hex())
		}

		if emailService == nil {
			emailService = services.NewEmailService()
		}
		
		// send new listing email notification to user
		emailService.SendNewListingEmail(loginEmail, loginName, newListing.Details.Title)

		internal.PublishCacheMessage(c, internal.CacheInvalidateShopListings, shopId.Hex())

		util.HandleSuccess(c, http.StatusOK, "Listing was created successfully", res.InsertedID)
	}
}

func GetListing() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQUEST_TIMEOUT_SECS)
		defer cancel()

		listingId := c.Param("listingid")
		tempListingService := services.NewListingService()
		listingIdentifier, e := tempListingService.GenerateListingBson(listingId)
		if e != nil {
			util.HandleError(c, http.StatusBadRequest, e)
			return
		}
		log.Println(listingIdentifier)
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
		ctx, cancel := context.WithTimeout(context.Background(), common.REQUEST_TIMEOUT_SECS)
		defer cancel()

		paginationArgs := common.GetPaginationArgs(c)
		tempListingService := services.NewListingService()
		match := tempListingService.GetListingFilters(c)
		sort := tempListingService.GetListingSortingBson(paginationArgs.Sort)

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
			{"$skip": int64(paginationArgs.Skip)},
			{"$limit": int64(paginationArgs.Limit)},
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
		ctx, cancel := context.WithTimeout(context.Background(), common.REQUEST_TIMEOUT_SECS)
		defer cancel()

		shopId, myId, err := common.MyShopIdAndMyId(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		paginationArgs := common.GetPaginationArgs(c)
		tempListingService := services.NewListingService()
		findOptions := options.Find().
			SetLimit(int64(paginationArgs.Limit)).
			SetSkip(int64(paginationArgs.Skip)).
			SetSort(tempListingService.GetListingSortingBson(paginationArgs.Sort))
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
		ctx, cancel := context.WithTimeout(context.Background(), common.REQUEST_TIMEOUT_SECS)
		defer cancel()

		shopId := c.Param("shopid")
		shopObjectId, err := primitive.ObjectIDFromHex(shopId)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		tempListingService := services.NewListingService()
		match := tempListingService.GetListingFilters(c)
		match["shop_id"] = shopObjectId
		paginationArgs := common.GetPaginationArgs(c)
		findOptions := options.Find().
			SetLimit(int64(paginationArgs.Limit)).
			SetSkip(int64(paginationArgs.Skip)).
			SetSort(tempListingService.GetListingSortingBson(paginationArgs.Sort))
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
		ctx, cancel := context.WithTimeout(context.Background(), common.REQUEST_TIMEOUT_SECS)
		defer cancel()

		_, userId, err := common.MyShopIdAndMyId(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
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
		ctx, cancel := context.WithTimeout(context.Background(), common.REQUEST_TIMEOUT_SECS)
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

		internal.PublishCacheMessage(c, internal.CacheInvalidateListings, "")
		util.HandleSuccess(c, http.StatusOK, "Listing(s) deleted", gin.H{"deleted": deletedObjectIDs, "not_deleted": notDeletedObjectIDs})
	}
}

func DeactivateListings() gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), common.REQUEST_TIMEOUT_SECS)
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

		internal.PublishCacheMessage(c, internal.CacheInvalidateListings, "")

		util.HandleSuccess(c, http.StatusOK, "Listing(s) deleted", gin.H{"deactivated": deletedObjectIDs, "not_deactivated": notDeletedObjectIDs})
	}
}
