package controllers

import (
	"context"
	auth "khoomi-api-io/api/internal/auth"
	"khoomi-api-io/api/internal/common"
	"khoomi-api-io/api/pkg/models"
	"khoomi-api-io/api/pkg/util"
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
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		shopId := c.Param("shopid")
		shopIdObj, err := primitive.ObjectIDFromHex(shopId)
		if err != nil {
			log.Println(err)
			util.HandleError(c, http.StatusUnprocessableEntity, err, "invalid shopid")
			return
		}

		// Check if the user owns the shop
		auth_, err := auth.InitJwtClaim(c)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err, "Invalid user token, id or access")
			return
		}
		userID, err := auth_.GetUserObjectId()
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err, "Invalid user token, id or access")
			return
		}

		err = common.VerifyShopOwnership(c, userID, shopIdObj)
		if err != nil {
			log.Printf("Error you the shop owner: %s\n", err.Error())
			util.HandleError(c, http.StatusForbidden, err, "shop ownership validation error")
			return
		}

		var shippingJson models.ShopShippingProfileRequest
		err = c.BindJSON(&shippingJson)
		if err != nil {
			log.Println(err)
			util.HandleError(c, http.StatusUnprocessableEntity, err, "invalid request body")
			return
		}

		// Validate request body
		if err := common.Validate.Struct(&shippingJson); err != nil {
			util.HandleError(c, http.StatusUnprocessableEntity, err, "invalid request body")
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
		res, err := common.ShippingProfileCollection.InsertOne(ctx, ShippingProfile)
		if err != nil {
			log.Println(err)
			util.HandleError(c, http.StatusInternalServerError, err, "document insert error")
			return
		}

		util.HandleSuccess(c, http.StatusOK, "document inserted", res.InsertedID)
	}
}

func GetShopShippingProfileInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		profileIdString := c.Param("id")
		profileId, err := primitive.ObjectIDFromHex(profileIdString)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err, "invalid shipping profile id")
			return
		}

		var shippingProfile models.ShopShippingProfile
		err = common.ShippingProfileCollection.FindOne(ctx, bson.M{"_id": profileId}).Decode(&shippingProfile)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err, "invalid shipping profile id")
			return
		}

		util.HandleSuccess(c, http.StatusOK, "success", shippingProfile)
	}
}

func GetShopShippingProfileInfos() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		shopIDStr := c.Param("shopid")
		shopIDObject, err := primitive.ObjectIDFromHex(shopIDStr)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err, "Invalid user token, id or access")
			return
		}

		paginationArgs := common.GetPaginationArgs(c)
		filter := bson.M{"shop_id": shopIDObject}
		findOptions := options.Find().
			SetLimit(int64(paginationArgs.Limit)).
			SetSkip(int64(paginationArgs.Skip)).
			SetSort(bson.D{{Key: "date", Value: -1}}) // Sort by date field in descending order (-1)

		result, err := common.ShippingProfileCollection.Find(ctx, filter, findOptions)
		if err != nil {
			log.Println(err)
			util.HandleError(c, http.StatusInternalServerError, err, "Failed to find shipping profiles")
			return
		}

		count, err := common.ShippingProfileCollection.CountDocuments(ctx, filter)
		if err != nil {
			log.Println(err)
			util.HandleError(c, http.StatusInternalServerError, err, "Failed to count shipping profiles")
			return
		}

		var shippingProfiles []models.ShopShippingProfile
		if err = result.All(ctx, &shippingProfiles); err != nil {
			log.Println(err)
			util.HandleError(c, http.StatusInternalServerError, err, "Failed to decode shipping profiles")
			return
		}

		util.HandleSuccessMeta(c, http.StatusOK, "success", shippingProfiles, gin.H{
			"pagination": util.Pagination{
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
// 			c.JSON(http.StatusBadRequest, util.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
// 			return
// 		}

// 		// Check if the user owns the shop
// 		userID, err := util.ExtractTokenID(c)
// 		if err != nil {
// 			c.JSON(http.StatusUnauthorized, util.UserResponse{Status: http.StatusUnauthorized, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
// 			return
// 		}

// 		err = verifyShopOwnership(ctx, userID, shopIdObj)
// 		if err != nil {
// 			c.JSON(http.StatusUnauthorized, util.UserResponse{Status: http.StatusUnauthorized, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
// 			return
// 		}

// 		err = c.BindJSON(&shippingJson)
// 		if err != nil {
// 			c.JSON(http.StatusBadRequest, util.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
// 			return
// 		}

// 		// Validate request body
// 		if validationErr := validate.Struct(&shippingJson); validationErr != nil {
// 			c.JSON(http.StatusBadRequest, util.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": validationErr.Error()}})
// 			return
// 		}

// 		profileIdString := c.Param("infoId")
// 		profileId, err := primitive.ObjectIDFromHex(profileIdString)
// 		if err != nil {
// 			c.JSON(http.StatusBadRequest, util.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
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
// 			c.JSON(http.StatusNotModified, util.UserResponse{Status: http.StatusNotModified, Message: "error", Data: map[string]interface{}{"error": err}})
// 			return
// 		}

// 		c.JSON(http.StatusOK, util.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": res}})
// 	}
// }
