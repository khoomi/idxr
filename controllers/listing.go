package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"khoomi-api-io/khoomi_api/configs"
	"khoomi-api-io/khoomi_api/helper"
	"khoomi-api-io/khoomi_api/models"
	"khoomi-api-io/khoomi_api/services"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	slug2 "github.com/gosimple/slug"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var listingCollection = configs.GetCollection(configs.DB, "ListingCollection")

func CreateListing() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
		defer cancel()

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			log.Println(err)
			helper.HandleError(c, http.StatusBadRequest, err, "Error getting shop ID and user ID")
			return
		}

		// loginName, loginEmail, err := auth.ExtractTokenLoginNameEmail(c)
		// if err != nil {
		// 	log.Println(err)
		// 	helper.HandleError(c, http.StatusUnauthorized, err, "unathorized")
		// 	return
		// }

		var newListing models.NewListing

		jsonData := c.PostForm("data")
		// Unmarshal the JSON data to the NewListing struct
		if err := json.Unmarshal([]byte(jsonData), &newListing); err != nil {
			log.Println(err)
			helper.HandleError(c, http.StatusBadRequest, err, "Invalid JSON data")
			return
		}

		if validationErr := validate.Struct(&newListing); validationErr != nil {
			log.Println(validationErr)
			helper.HandleError(c, http.StatusBadRequest, validationErr, "invalid or missing data in request body")
			return
		}
		// main_image file handling
		mainImage, _, err := c.Request.FormFile("main_image")
		var mainImageUploadUrl string
		if err == nil {
			mainImageUploadUrl, err = services.NewMediaUpload().FileUpload(models.File{File: mainImage})
			if err != nil {
				errMsg := fmt.Sprintf("Logo failed to upload - %v", err.Error())
				log.Print(errMsg)
				helper.HandleError(c, http.StatusInternalServerError, err, errMsg)
				return
			}
		} else {
			mainImageUploadUrl = ""
		}

		_, _, err = c.Request.FormFile("images")
		var imagesUploadUrls []string

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
				imageUploadUrl, err := services.NewMediaUpload().FileUpload(models.File{File: file})
				if err != nil {
					errMsg := fmt.Sprintf("File failed to upload - %v", err.Error())
					log.Print(errMsg)
					helper.HandleError(c, http.StatusInternalServerError, err, errMsg)
					return
				}

				// Append the URL to the logoUploadUrls slice
				imagesUploadUrls = append(imagesUploadUrls, imageUploadUrl)
			}
		} else {
			imagesUploadUrls = nil
		}

		now := time.Now()
		listingDetails := models.ListingDetails{
			Type:                        newListing.ListingDetails.Type,
			Tags:                        newListing.ListingDetails.Tags,
			Title:                       newListing.ListingDetails.Title,
			Dynamic:                     newListing.ListingDetails.Dynamic,
			WhoMade:                     newListing.ListingDetails.WhoMade,
			Keywords:                    newListing.ListingDetails.Keywords,
			WhenMade:                    newListing.ListingDetails.WhenMade,
			Category:                    newListing.ListingDetails.Category,
			Condition:                   newListing.ListingDetails.Condition,
			Description:                 newListing.ListingDetails.Description,
			HasVariations:               newListing.ListingDetails.HasVariations,
			Personalization:             newListing.ListingDetails.Personalization,
			PersonalizationText:         newListing.ListingDetails.PersonalizationText,
			PersonalizationTextChars:    newListing.ListingDetails.PersonalizationTextChars,
			PersonalizationTextOptional: newListing.ListingDetails.PersonalizationTextOptional,
		}

		listingInventory := models.Inventory{
			DomesticPricing: newListing.Inventory.DomesticPricing,
			DomesticPrice:   newListing.Inventory.DomesticPrice,
			Price:           newListing.Inventory.Price,
			Quantity:        newListing.Inventory.Quantity,
			SKU:             newListing.Inventory.SKU,
			CurrencyCode:    "NGN",
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
			Rating:      0,
			ReviewCount: 0,
		}

		listing := models.Listing{
			ID:                primitive.NewObjectID(),
			State:             models.ListingState{State: models.ListingStateActive, StateUpdatedAt: now},
			UserId:            myId,
			ShopId:            shopId,
			MainImage:         mainImageUploadUrl,
			Images:            imagesUploadUrls,
			ListingDetails:    listingDetails,
			Date:              listingDate,
			Slug:              slug2.Make(newListing.ListingDetails.Title),
			Views:             0,
			FavorersCount:     0,
			ShippingProfileId: primitive.NilObjectID,
			Processing:        listingProcessing,
			NonTaxable:        true,
			Variations:        newListing.Variations,
			ShouldAutoRenew:   false,
			Inventory:         listingInventory,
			RecentReviews:     nil,
			Rating:            listingRating,
		}

		res, err := listingCollection.InsertOne(ctx, listing)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to create new listing — %v", err.Error())
			log.Print(errMsg)
			helper.HandleError(c, http.StatusInternalServerError, err, errMsg)
			return
		}

		// send new listing email notification to user
		// email.SendNewListingEmail(loginEmail, loginName, newListing.ListingDetails.Title)

		helper.HandleSuccess(c, http.StatusOK, "Lisating was created successfully", gin.H{"_id": res.InsertedID})

	}

}
