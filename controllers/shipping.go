package controllers

import (
	"context"
	"khoomi-api-io/khoomi_api/config"
	"khoomi-api-io/khoomi_api/helper"
	"khoomi-api-io/khoomi_api/models"
	"khoomi-api-io/khoomi_api/services"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateShopShippingProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), REQ_TIMEOUT_SECS)
		defer cancel()

		shopId := c.Param("shopid")
		shopIdObj, err := primitive.ObjectIDFromHex(shopId)
		if err != nil {
			log.Println(err)
			helper.HandleError(c, http.StatusUnprocessableEntity, err, "invalid shopid")
			return
		}

		// Check if the user owns the shop
		auth, err := config.InitJwtClaim(c)
		if err != nil {
			helper.HandleError(c, http.StatusUnauthorized, err, "Invalid user token, id or access")
			return
		}
		userID, err := auth.GetUserObjectId()
		if err != nil {
			helper.HandleError(c, http.StatusUnauthorized, err, "Invalid user token, id or access")
			return
		}

		err = VerifyShopOwnership(c, userID, shopIdObj)
		if err != nil {
			log.Printf("Error you the shop owner: %s\n", err.Error())
			helper.HandleError(c, http.StatusForbidden, err, "shop ownership validation error")
			return
		}

		var shippingJson models.ShopShippingProfileRequest
		err = c.BindJSON(&shippingJson)
		if err != nil {
			log.Println(err)
			helper.HandleError(c, http.StatusUnprocessableEntity, err, "invalid request body")
			return
		}

		// Validate request body
		if err := Validate.Struct(&shippingJson); err != nil {
			helper.HandleError(c, http.StatusUnprocessableEntity, err, "invalid request body")
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
			ID:                       shippingId,
			ShopID:                   shopIdObj,
			Title:                    shippingJson.Title,
			MinProcessingTime:        shippingJson.MinProcessingTime,
			MaxProcessingTime:        shippingJson.MaxProcessingTime,
			ProcessingTimeUnit:       shippingJson.ProcessingTimeUnit,
			HandlingFee:              shippingJson.HandlingFee,
			OriginState:              shippingJson.OriginState,
			OriginPostalCode:         shippingJson.OriginPostalCode,
			MinDeliveryDays:          shippingJson.MinDeliveryDays,
			MaxDeliveryDays:          shippingJson.MaxDeliveryDays,
			PrimaryPrice:             shippingJson.PrimaryPrice,
			DestinationBy:            shippingJson.DestinationBy,
			Destinations:             shippingJson.Destinations,
			SecondaryPrice:           shippingJson.SecondaryPrice,
			ShippingService:          shippingJson.ShippingService,
			AutoCalculatePrice:       shippingJson.AutoCalculatePrice,
			OffersFreeShipping:       shippingJson.OffersFreeShipping,
			Policy:                   shippingPolicy,
			CreatedAt:                primitive.NewDateTimeFromTime(now),
			ModifiedAt:               primitive.NewDateTimeFromTime(now),
			IsDefaultShippingProfile: shippingJson.IsDefaultShippingProfile,
		}
		res, err := ShippingProfileCollection.InsertOne(ctx, ShippingProfile)
		if err != nil {
			log.Println(err)
			helper.HandleError(c, http.StatusInternalServerError, err, "document insert error")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "document inserted", res.InsertedID)
	}
}

func GetShopShippingProfileInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), REQ_TIMEOUT_SECS)
		defer cancel()

		profileIdString := c.Param("id")
		profileId, err := primitive.ObjectIDFromHex(profileIdString)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "invalid shipping profile id")
			return
		}

		var shippingProfile models.ShopShippingProfile
		err = ShippingProfileCollection.FindOne(ctx, bson.M{"_id": profileId}).Decode(&shippingProfile)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "invalid shipping profile id")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "success", shippingProfile)
	}
}

func GetShopShippingProfileInfos() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), REQ_TIMEOUT_SECS)
		defer cancel()

		shopIDStr := c.Param("shopid")
		shopIDObject, err := primitive.ObjectIDFromHex(shopIDStr)
		if err != nil {
			helper.HandleError(c, http.StatusUnauthorized, err, "Invalid user token, id or access")
			return
		}

		paginationArgs := services.GetPaginationArgs(c)
		filter := bson.M{"shop_id": shopIDObject}
		findOptions := options.Find().
			SetLimit(int64(paginationArgs.Limit)).
			SetSkip(int64(paginationArgs.Skip)).
			SetSort(bson.D{{Key: "date", Value: -1}}) // Sort by date field in descending order (-1)

		result, err := ShippingProfileCollection.Find(ctx, filter, findOptions)
		if err != nil {
			log.Println(err)
			helper.HandleError(c, http.StatusInternalServerError, err, "Failed to find shipping profiles")
			return
		}

		count, err := ShippingProfileCollection.CountDocuments(ctx, filter)
		if err != nil {
			log.Println(err)
			helper.HandleError(c, http.StatusInternalServerError, err, "Failed to count shipping profiles")
			return
		}

		var shippingProfiles []models.ShopShippingProfile
		if err = result.All(ctx, &shippingProfiles); err != nil {
			log.Println(err)
			helper.HandleError(c, http.StatusInternalServerError, err, "Failed to decode shipping profiles")
			return
		}

		helper.HandleSuccessMeta(c, http.StatusOK, "success", shippingProfiles, gin.H{
			"pagination": helper.Pagination{
				Limit: paginationArgs.Limit,
				Skip:  paginationArgs.Skip,
				Count: count,
			},
		})
	}
}

// func UpdateShopShippingProfileInfo() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		ctx, cancel := context.WithTimeout(context.Background(), REQ_TIMEOUT_SECS)
// 		var shippingJson models.ShopShippingProfileRequest
// 		defer cancel()

// 		shopId := c.Param("shopId")
// 		shopIdObj, err := primitive.ObjectIDFromHex(shopId)
// 		if err != nil {
// 			c.JSON(http.StatusBadRequest, helper.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
// 			return
// 		}

// 		// Check if the user owns the shop
// 		userID, err := configs.ExtractTokenID(c)
// 		if err != nil {
// 			c.JSON(http.StatusUnauthorized, helper.UserResponse{Status: http.StatusUnauthorized, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
// 			return
// 		}

// 		err = verifyShopOwnership(ctx, userID, shopIdObj)
// 		if err != nil {
// 			c.JSON(http.StatusUnauthorized, helper.UserResponse{Status: http.StatusUnauthorized, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
// 			return
// 		}

// 		err = c.BindJSON(&shippingJson)
// 		if err != nil {
// 			c.JSON(http.StatusBadRequest, helper.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
// 			return
// 		}

// 		// Validate request body
// 		if validationErr := validate.Struct(&shippingJson); validationErr != nil {
// 			c.JSON(http.StatusBadRequest, helper.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": validationErr.Error()}})
// 			return
// 		}

// 		profileIdString := c.Param("infoId")
// 		profileId, err := primitive.ObjectIDFromHex(profileIdString)
// 		if err != nil {
// 			c.JSON(http.StatusBadRequest, helper.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
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

// 		res, err := ShippingProfileCollection.UpdateOne(ctx, bson.M{"_id": profileId}, bson.M{"$set": ShippingProfile})
// 		if err != nil {
// 			c.JSON(http.StatusNotModified, helper.UserResponse{Status: http.StatusNotModified, Message: "error", Data: map[string]interface{}{"error": err}})
// 			return
// 		}

// 		c.JSON(http.StatusOK, helper.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": res}})
// 	}
// }
