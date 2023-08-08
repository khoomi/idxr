package controllers

import (
	"context"
	"khoomi-api-io/khoomi_api/configs"
	"khoomi-api-io/khoomi_api/helper"
	"khoomi-api-io/khoomi_api/models"
	"khoomi-api-io/khoomi_api/responses"
	"khoomi-api-io/khoomi_api/services"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var shippingProfileCollection = configs.GetCollection(configs.DB, "ShopShippingProfile")

func CreateShopShippingProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		defer cancel()

		shopId := c.Param("shopid")
		shopIdObj, err := primitive.ObjectIDFromHex(shopId)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "invalid shopid")
			return
		}

		// Check if the user owns the shop
		userID, err := configs.ExtractTokenID(c)
		if err != nil {
			log.Println(err)
			helper.HandleError(c, http.StatusUnauthorized, err, "unauthorized")
			return
		}

		err = VerifyShopOwnership(c, userID, shopIdObj)
		if err != nil {
			log.Printf("Error you the shop owner: %s\n", err.Error())
			helper.HandleError(c, http.StatusUnauthorized, err, "shop ownership validation error")
			return
		}

		var shippingJson models.ShopShippingProfileRequest
		err = c.BindJSON(&shippingJson)
		if err != nil {
			log.Println(err)
			helper.HandleError(c, http.StatusBadRequest, err, "invalid request body")
			return
		}

		// Validate request body
		if validationErr := validate.Struct(&shippingJson); validationErr != nil {
			log.Println(validationErr)
			helper.HandleError(c, http.StatusBadRequest, validationErr, "invalid request body")
			return
		}
		shippingPolicy := models.ShippingPolicy{
			AcceptReturns:  shippingJson.Policy.AcceptReturns,
			AcceptExchange: shippingJson.Policy.AcceptExchange,
			ReturnPeriod:   shippingJson.Policy.ReturnPeriod,
			ReturnUnit:     shippingJson.Policy.ReturnUnit,
			Conditions:     shippingJson.Policy.Conditions,
		}

		shippingId := primitive.NewObjectID()
		now := time.Now()
		ShippingProfile := models.ShopShippingProfile{
			ID:                 shippingId,
			ShopID:             shopIdObj,
			Title:              shippingJson.Title,
			MinProcessingTime:  shippingJson.MinProcessingTime,
			MaxProcessingTime:  shippingJson.MaxProcessingTime,
			ProcessingTimeUnit: shippingJson.ProcessingTimeUnit,
			HandlingFee:        shippingJson.HandlingFee,
			OriginState:        shippingJson.OriginState,
			OriginPostalCode:   shippingJson.OriginPostalCode,
			MinDeliveryDays:    shippingJson.MinDeliveryDays,
			MaxDeliveryDays:    shippingJson.MaxDeliveryDays,
			PrimaryPrice:       shippingJson.PrimaryPrice,
			DestinationBy:      shippingJson.DestinationBy,
			Destinations:       shippingJson.Destinations,
			SecondaryPrice:     shippingJson.SecondaryPrice,
			ShippingService:    shippingJson.ShippingService,
			AutoCalculatePrice: shippingJson.AutoCalculatePrice,
			OffersFreeShipping: shippingJson.OffersFreeShipping,
			Policy:             shippingPolicy,
			CreatedAt:          primitive.NewDateTimeFromTime(now),
			ModifiedAt:         primitive.NewDateTimeFromTime(now),
		}
		res, err := shippingProfileCollection.InsertOne(ctx, ShippingProfile)
		if err != nil {
			log.Println(err)
			helper.HandleError(c, http.StatusBadRequest, err, "document insert error")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "document inserted", gin.H{"inserted_id": res.InsertedID})
	}
}

func GetShopShippingProfileInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		defer cancel()

		profileIdString := c.Param("shippingProfileId")
		profileId, err := primitive.ObjectIDFromHex(profileIdString)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "invalid shipping profile id")
			return
		}

		var shippingProfile models.ShopShippingProfile
		err = shippingProfileCollection.FindOne(ctx, bson.M{"_id": profileId}).Decode(&shippingProfile)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "invalid shipping profile id")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "success", gin.H{"shipping_profile": shippingProfile})
	}
}

func GetShopShippingProfileInfos() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		defer cancel()

		// Check if the user owns the shop
		shopIDStr := c.Param("shopid")
		shopIDObject, err := primitive.ObjectIDFromHex(shopIDStr)
		if err != nil {
			log.Println(err)
			helper.HandleError(c, http.StatusUnauthorized, err, "unauthorized")
			return
		}

		paginationArgs := services.GetPaginationArgs(c)
		filter := bson.M{"shop_id": shopIDObject}
		findOptions := options.Find().
			SetLimit(int64(paginationArgs.Limit)).
			SetSkip(int64(paginationArgs.Skip)).
			SetSort(bson.D{{Key: "date", Value: -1}}) // Sort by date field in descending order (-1)

		result, err := shippingProfileCollection.Find(ctx, filter, findOptions)
		if err != nil {
			helper.HandleError(c, http.StatusNotFound, err, "Failed to find shipping profiles")
			return
		}

		count, err := shippingProfileCollection.CountDocuments(ctx, filter)
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Failed to count shipping profiles")
			return
		}

		var shippingProfiles []models.ShopShippingProfile
		if err = result.All(ctx, &shippingProfiles); err != nil {
			helper.HandleError(c, http.StatusNotFound, err, "Failed to decode shipping profiles")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "success", gin.H{
			"shipping_profiles": shippingProfiles,
			"pagination": responses.Pagination{
				Limit: paginationArgs.Limit,
				Skip:  paginationArgs.Skip,
				Count: count,
			},
		})
	}
}

// func UpdateShopShippingProfileInfo() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
// 		var shippingJson models.ShopShippingProfileRequest
// 		defer cancel()

// 		shopId := c.Param("shopId")
// 		shopIdObj, err := primitive.ObjectIDFromHex(shopId)
// 		if err != nil {
// 			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
// 			return
// 		}

// 		// Check if the user owns the shop
// 		userID, err := configs.ExtractTokenID(c)
// 		if err != nil {
// 			c.JSON(http.StatusUnauthorized, responses.UserResponse{Status: http.StatusUnauthorized, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
// 			return
// 		}

// 		err = verifyShopOwnership(ctx, userID, shopIdObj)
// 		if err != nil {
// 			c.JSON(http.StatusUnauthorized, responses.UserResponse{Status: http.StatusUnauthorized, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
// 			return
// 		}

// 		err = c.BindJSON(&shippingJson)
// 		if err != nil {
// 			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
// 			return
// 		}

// 		// Validate request body
// 		if validationErr := validate.Struct(&shippingJson); validationErr != nil {
// 			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": validationErr.Error()}})
// 			return
// 		}

// 		profileIdString := c.Param("infoId")
// 		profileId, err := primitive.ObjectIDFromHex(profileIdString)
// 		if err != nil {
// 			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
// 			return
// 		}

// 		ShippingProfile := models.ShopShippingProfile{
// 			ShopID:              shopIdObj,
// 			Title:               shippingJson.Title,
// 			MinProcessingTime:   shippingJson.MinProcessingTime,
// 			MaxProcessingTime:   shippingJson.MaxProcessingTime,
// 			ProcessingTimeUnit:  shippingJson.ProcessingTimeUnit,
// 			DomesticHandlingFee: shippingJson.DomesticHandlingFee,
// 			OriginState:         shippingJson.OriginState,
// 			OriginPostalCode:    shippingJson.OriginPostalCode,
// 			MinDeliveryDays:     shippingJson.MinDeliveryDays,
// 			MaxDeliveryDays:     shippingJson.MaxDeliveryDays,
// 			PrimaryCost:         shippingJson.PrimaryCost,
// 		}

// 		res, err := shippingProfileCollection.UpdateOne(ctx, bson.M{"_id": profileId}, bson.M{"$set": ShippingProfile})
// 		if err != nil {
// 			c.JSON(http.StatusNotModified, responses.UserResponse{Status: http.StatusNotModified, Message: "error", Data: map[string]interface{}{"error": err}})
// 			return
// 		}

// 		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": res}})
// 	}
// }
