package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"khoomi-api-io/khoomi_api/configs"
	"khoomi-api-io/khoomi_api/email"
	"khoomi-api-io/khoomi_api/helper"
	"khoomi-api-io/khoomi_api/models"
	"khoomi-api-io/khoomi_api/responses"
	"khoomi-api-io/khoomi_api/services"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/api/uploader"
	"github.com/gin-gonic/gin"
	slug2 "github.com/gosimple/slug"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func generateListingCode() string {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	// Define characters for letters and numbers
	letterChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numberChars := "0123456789"
	// Generate 4 random letters
	letters := make([]byte, 4)
	for i := range letters {
		letters[i] = letterChars[rand.Intn(len(letterChars))]
	}
	// Generate 4 random numbers
	numbers := make([]byte, 4)
	for i := range numbers {
		numbers[i] = numberChars[rand.Intn(len(numberChars))]
	}
	// Combine letters and numbers with a hyphen
	productCode := string(letters) + "-" + string(numbers)
	return productCode
}

func getListingSortingBson(sort string) bson.D {
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
	case "code_asc":
		key = "code"
	case "code_desc":
		key = "code"
	case "state_updated_at_asc":
		key = "state.updated_at"
	case "state_updated_at_desc":
		key = "state.updated_at"
	case "views_asc":
		key = "views"
	case "views_desc":
		key = "views"
	case "processing_max_asc":
		key = "processing.max"
	case "processing_max_desc":
		key = "processing.max"
	case "processing_min_asc":
		key = "processing.min"
	case "processing_min_desc":
		key = "processing.min"
	case "favorers_count_asc":
		key = "favorers_count"
	case "favorers_count_desc":
		key = "favorers_count"
	case "total_orders_asc":
		key = "financial_information.total_orders"
	case "total_orders_desc":
		key = "financial_information.total_orders"
	case "sales_asc":
		key = "financial_information.sales"
	case "sales_desc":
		key = "financial_information.sales"
	default:
		key = "date.created_at"
	}

	if strings.Contains(sort, "asc") {
		value = 1
	}
	return bson.D{{Key: key, Value: value}}
}

func CreateListing() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
		defer cancel()

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Error getting shop ID and user ID")
			return
		}

		loginName, loginEmail, err := configs.ExtractTokenLoginNameEmail(c)
		if err != nil {
			helper.HandleError(c, http.StatusUnauthorized, err, "unathorized")
			return
		}

		var newListing models.NewListing

		jsonData := c.PostForm("data")
		// Unmarshal the JSON data to the NewListing struct
		if err := json.Unmarshal([]byte(jsonData), &newListing); err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Invalid JSON data")
			return
		}

		if validationErr := Validate.Struct(&newListing); validationErr != nil {
			log.Println(validationErr)
			helper.HandleError(c, http.StatusBadRequest, validationErr, "invalid or missing data in request body")
			return
		}

		// main_image file handling
		mainImage, _, err := c.Request.FormFile("main_image")
		var mainImageUploadUrl uploader.UploadResult
		if err == nil {
			mainImageUploadUrl, err = services.FileUpload(models.File{File: mainImage})
			if err != nil {
				errMsg := fmt.Sprintf("Logo failed to upload - %v", err.Error())
				log.Print(errMsg)
				helper.HandleError(c, http.StatusInternalServerError, err, errMsg)
				return
			}
		} else {
			mainImageUploadUrl = uploader.UploadResult{}
		}

		_, _, err = c.Request.FormFile("images")
		var uploadedImagesUrl []string
		var uploadedImagesResult []uploader.UploadResult

		if err == nil {
			// FormFile returns a single file, so you need to use MultipartForm to get multiple files
			err := c.Request.ParseMultipartForm(10 << 20)
			if err != nil {
				errMsg := fmt.Sprintf("Failed to parse multipart form - %v", err.Error())
				log.Print(errMsg)
				helper.HandleError(c, http.StatusInternalServerError, err, errMsg)
				return
			}

			files := c.Request.MultipartForm.File["images"]
			for _, fileHeader := range files {
				file, err := fileHeader.Open()
				if err != nil {
					errMsg := fmt.Sprintf("Failed to open file - %v", err.Error())
					log.Print(errMsg)
					helper.HandleError(c, http.StatusInternalServerError, err, errMsg)
					return
				}
				defer file.Close()

				// Here, you can upload each file to the desired location and get the URLs
				imageUpload, err := services.FileUpload(models.File{File: file})
				if err != nil {
					errMsg := fmt.Sprintf("File failed to upload - %v", err.Error())
					log.Print(errMsg)
					helper.HandleError(c, http.StatusInternalServerError, err, errMsg)
					return
				}

				// Append the URL to the logoUploadUrls slice
				uploadedImagesUrl = append(uploadedImagesUrl, imageUpload.SecureURL)
				uploadedImagesResult = append(uploadedImagesResult, imageUpload)
			}
		} else {
			uploadedImagesUrl = nil
		}

		now := time.Now()
		listingDetails := models.ListingDetails{
			Type:                        newListing.ListingDetails.Type,
			Tags:                        newListing.ListingDetails.Tags,
			Title:                       newListing.ListingDetails.Title,
			Color:                       newListing.ListingDetails.Color,
			Dynamic:                     newListing.ListingDetails.Dynamic,
			WhoMade:                     newListing.ListingDetails.WhoMade,
			Keywords:                    newListing.ListingDetails.Keywords,
			WhenMade:                    newListing.ListingDetails.WhenMade,
			Category:                    newListing.ListingDetails.Category,
			Condition:                   newListing.ListingDetails.Condition,
			Description:                 newListing.ListingDetails.Description,
			HasVariations:               newListing.ListingDetails.HasVariations,
			Sustainability:              newListing.ListingDetails.Sustainability,
			Personalization:             newListing.ListingDetails.Personalization,
			PersonalizationText:         newListing.ListingDetails.PersonalizationText,
			PersonalizationTextChars:    newListing.ListingDetails.PersonalizationTextChars,
			PersonalizationTextOptional: newListing.ListingDetails.PersonalizationTextOptional,
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

		listingProcessing := models.ListingProcessing{
			ProcessingMin:     newListing.Processing.ProcessingMin,
			ProcessingMinUnit: newListing.Processing.ProcessingMinUnit,
			ProcessingMax:     newListing.Processing.ProcessingMax,
			ProcessingMaxUnit: newListing.Processing.ProcessingMaxUnit,
		}

		listingRating := models.ListingRating{
			Rating:          0,
			ReviewCount:     0,
			PositiveReviews: 0,
			NegativeReviews: 0,
		}

		listingFinancialInformation := models.ListingFinancialInformation{
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
			Code:                 generateListingCode(),
			State:                models.ListingState{State: models.ListingStateActive, StateUpdatedAt: now},
			UserId:               myId,
			ShopId:               shopId,
			MainImage:            mainImageUploadUrl.SecureURL,
			Images:               uploadedImagesUrl,
			ListingDetails:       listingDetails,
			Date:                 listingDate,
			Slug:                 slug2.Make(newListing.ListingDetails.Title),
			Views:                0,
			FavorersCount:        0,
			ShippingProfileId:    primitive.NilObjectID,
			Processing:           listingProcessing,
			NonTaxable:           true,
			Variations:           newListing.Variations,
			ShouldAutoRenew:      false,
			Inventory:            listingInventory,
			RecentReviews:        nil,
			Rating:               listingRating,
			Measurements:         newListing.Measurements,
			FinancialInformation: listingFinancialInformation,
		}

		res, err := ListingCollection.InsertOne(ctx, listing)
		if err != nil {
			// delete images
			_, err := services.DestroyMedia(mainImageUploadUrl.PublicID)
			for _, file := range uploadedImagesResult {
				_, err := services.DestroyMedia(file.PublicID)
				log.Println(err)
			}
			// return error
			errMsg := fmt.Sprintf("Failed to create new listing — %v", err.Error())
			helper.HandleError(c, http.StatusInternalServerError, err, errMsg)
			return
		}

		// send new listing email notification to user
		email.SendNewListingEmail(loginEmail, loginName, newListing.ListingDetails.Title)

		helper.HandleSuccess(c, http.StatusOK, "Lisating was created successfully", gin.H{"_id": res.InsertedID})

	}
}

func GetListing() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
		defer cancel()

		var listingIdentifier bson.M

		listingId := c.Param("listingid")
		if primitive.IsValidObjectID(listingId) {
			// If listingid is a valid object ID string
			listingObjectID, e := primitive.ObjectIDFromHex(listingId)
			if e != nil {
				helper.HandleError(c, http.StatusBadRequest, e, "invalid listing id was provided")
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
				},
			},
		}

		cursor, err := ListingCollection.Aggregate(ctx, pipeline)
		if err != nil {
			log.Println(err)
			helper.HandleError(c, http.StatusInternalServerError, err, "error while retrieving listing")
			return
		}

		var listing models.ListingExtra

		if cursor.Next(ctx) {
			if err := cursor.Decode(&listing); err != nil {
				log.Println(err)
				helper.HandleError(c, http.StatusInternalServerError, err, "error while decoding listing")
				return
			}
		} else {
			helper.HandleError(c, http.StatusNotFound, errors.New("no listing found"), "no listing found")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "Success", gin.H{"listing": listing})
	}
}

func GetListings() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
		defer cancel()

		paginationArgs := services.GetPaginationArgs(c)
		sort := getListingSortingBson(paginationArgs.Sort)

		pipeline := []bson.M{
			{"$match": bson.M{}},
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
						"name":          "$shop.username",
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

		cursor, err := ListingCollection.Aggregate(ctx, pipeline)
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "error while retrieving listing")
			return
		}

		var listings []models.ListingExtra
		if err := cursor.All(ctx, &listings); err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "error while decoding listing")
			return
		}

		countPipeline := []bson.M{
			{"$match": bson.M{}},
			{"$count": "total"},
		}
		countCursor, err := ListingCollection.Aggregate(ctx, countPipeline)
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "error while counting listings")
			return
		}
		var countResult struct {
			Total int64 `bson:"total"`
		}
		if countCursor.Next(ctx) {
			if err := countCursor.Decode(&countResult); err != nil {
				helper.HandleError(c, http.StatusInternalServerError, err, "error while decoding count")
				return
			}
		}

		helper.HandleSuccess(c, http.StatusOK, "success", gin.H{
			"listings": listings,
			"pagination": responses.Pagination{
				Limit: paginationArgs.Limit,
				Skip:  paginationArgs.Skip,
				Count: countResult.Total,
			},
		})
	}
}

func GetMyListingsSummary() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
		defer cancel()

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Error getting shop ID and user ID")
			return
		}

		paginationArgs := services.GetPaginationArgs(c)
		findOptions := options.Find().
			SetLimit(int64(paginationArgs.Limit)).
			SetSkip(int64(paginationArgs.Skip)).
			SetSort(getListingSortingBson(paginationArgs.Sort))
		filter := bson.M{"shop_id": shopId, "user_id": myId}
		cursor, err := ListingCollection.Find(ctx, filter, findOptions)
		if err != nil {
			helper.HandleError(c, http.StatusNotFound, err, "no listing found")
			return
		}
		defer func() {
			if err := cursor.Close(ctx); err != nil {
				log.Println("Failed to close cursor:", err)
			}
		}()

		var listings []models.ListingsSummary
		if err := cursor.All(ctx, &listings); err != nil {
			helper.HandleError(c, http.StatusNotFound, err, "Failed to retrieve listings")
			return
		}
		count, err := ListingCollection.CountDocuments(ctx, filter)
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Error counting listings")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "success", gin.H{
			"listings": listings,
			"pagination": responses.Pagination{
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
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
		defer cancel()

		shopId := c.Param("shopid")
		shopObjectId, err := primitive.ObjectIDFromHex(shopId)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "invalid shop id was provided")
			return
		}

		paginationArgs := services.GetPaginationArgs(c)
		findOptions := options.Find().
			SetLimit(int64(paginationArgs.Limit)).
			SetSkip(int64(paginationArgs.Skip)).
			SetSort(getListingSortingBson(paginationArgs.Sort))

		result, err := ListingCollection.Find(ctx, bson.M{"shop_id": shopObjectId}, findOptions)
		if err != nil {
			helper.HandleError(c, http.StatusNotFound, err, "error retrieving listings")
			return
		}

		count, err := ListingCollection.CountDocuments(ctx, bson.M{"shop_id": shopObjectId})
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Failed to count shipping profiles")
			return
		}

		var listings []models.Listing
		if err = result.All(ctx, &listings); err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Failed to decode listings")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "success", gin.H{
			"listings": listings,
			"pagination": responses.Pagination{
				Limit: paginationArgs.Limit,
				Skip:  paginationArgs.Skip,
				Count: count,
			},
		})
	}
}

func HasUserCreatedListingOnboarding() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), KhoomiRequestTimeoutSec)
		defer cancel()

		shopId, userId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Error getting shop ID and user ID")
			return
		}

		err = VerifyShopOwnership(c, userId, shopId)
		if err != nil {
			helper.HandleError(c, http.StatusUnauthorized, errors.New("Only sellers can perform this action"), "unauthorized")
			return
		}

		filter := bson.M{"user_id": userId}
		findOptions := options.Find().SetLimit(1)
		cursor, err := ListingCollection.Find(ctx, filter, findOptions)
		if err != nil {
			log.Printf("error retrieving user listing: %v", err)
			helper.HandleError(c, http.StatusNotFound, err, "error retrieving user listing")
			return
		}
		defer cursor.Close(ctx)

		var listing []models.Listing
		if err := cursor.All(ctx, &listing); err != nil {
			helper.HandleError(c, http.StatusNotFound, err, "Error retrieving listing informations")
			return
		}

		if len(listing) == 0 {
			helper.HandleError(c, http.StatusNotFound, errors.New("User has no listings"), "User has no listings")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "Success", gin.H{"listing": listing[0]})
	}
}
