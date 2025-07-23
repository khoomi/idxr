package controllers

import (
	"log"
	"net/http"

	"khoomi-api-io/api/internal/common"
	"khoomi-api-io/api/pkg/models"
	"khoomi-api-io/api/pkg/services"
	"khoomi-api-io/api/pkg/util"

	"github.com/gin-gonic/gin"
)

type ShippingController struct {
	shippingService     services.ShippingService
	shopService         services.ShopService
	notificationService services.NotificationService
}

// InitShippingController initializes a new ShippingController with dependencies
func InitShippingController(shippingService services.ShippingService, shopService services.ShopService, notificationService services.NotificationService) *ShippingController {
	return &ShippingController{
		shippingService:     shippingService,
		shopService:         shopService,
		notificationService: notificationService,
	}
}

func (sc *ShippingController) CreateShopShippingProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := WithTimeout()
		defer cancel()

		shopIdObj, ok := ParseObjectIDParam(c, "shopid")
		if !ok {
			return
		}

		userID, ok := ValidateAndGetUserID(c)
		if !ok {
			return
		}

		err := sc.shopService.VerifyShopOwnership(c, userID, shopIdObj)
		if err != nil {
			log.Printf("Error you the shop owner: %s\n", err.Error())
			util.HandleError(c, http.StatusForbidden, err)
			return
		}

		var shippingJson models.ShopShippingProfileRequest
		if !BindJSONAndValidate(c, &shippingJson) {
			return
		}

		shippingID, err := sc.shippingService.CreateShopShippingProfile(ctx, userID, shopIdObj, shippingJson)
		if err != nil {
			log.Println(err)
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		util.HandleSuccess(c, http.StatusOK, "document inserted", shippingID)
	}
}

func (sc *ShippingController) GetShopShippingProfileInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := WithTimeout()
		defer cancel()

		profileId, ok := ParseObjectIDParam(c, "id")
		if !ok {
			return
		}

		shippingProfile, err := sc.shippingService.GetShopShippingProfile(ctx, profileId)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		util.HandleSuccess(c, http.StatusOK, "success", shippingProfile)
	}
}

func (sc *ShippingController) GetShopShippingProfileInfos() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := WithTimeout()
		defer cancel()

		shopIDObject, ok := ParseObjectIDParam(c, "shopid")
		if !ok {
			return
		}

		paginationArgs := common.GetPaginationArgs(c)
		shippingProfiles, count, err := sc.shippingService.GetShopShippingProfiles(ctx, shopIDObject, paginationArgs)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		HandlePaginationAndResponse(c, shippingProfiles, count, paginationArgs, "success")
	}
}

// func UpdateShopShippingProfileInfo() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		ctx, cancel := context.WithTimeout(context.Background(), REQUEST_TIMEOUT_SECS)
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
