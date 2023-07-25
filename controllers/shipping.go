package controllers

import (
	"context"
	"fmt"
	"khoomi-api-io/khoomi_api/auth"
	"khoomi-api-io/khoomi_api/configs"
	"khoomi-api-io/khoomi_api/models"
	"khoomi-api-io/khoomi_api/responses"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var shippingProfileCollection = configs.GetCollection(configs.DB, "ShopShippingProfile")

func verifyShopOwnership(ctx context.Context, shopID primitive.ObjectID, userID primitive.ObjectID) error {
	shop := models.Shop{}
	err := shopCollection.FindOne(ctx, bson.M{"_id": shopID, "owner_id": userID}).Decode(&shop)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("User does not own the shop")
		}
		return err
	}
	return nil
}

func CreateShopShippingProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		var shippingJson models.ShopShippingProfileRequest
		defer cancel()

		shopId := c.Param("shopId")
		shopIdObj, err := primitive.ObjectIDFromHex(shopId)
		if err != nil {
			c.JSON(http.StatusBadRequest,
				responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		// Check if the user owns the shop
		userID, err := auth.ExtractTokenID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, responses.UserResponse{Status: http.StatusUnauthorized, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		err = verifyShopOwnership(ctx, shopIdObj, userID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, responses.UserResponse{Status: http.StatusUnauthorized, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		err = c.BindJSON(&shippingJson)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		// Validate request body
		if validationErr := validate.Struct(&shippingJson); validationErr != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": validationErr.Error()}})
			return
		}

		shippingId := primitive.NewObjectID()
		ShippingProfile := models.ShopShippingProfile{
			ID:                  shippingId,
			ShopID:              shopIdObj,
			Title:               shippingJson.Title,
			MinProcessingTime:   shippingJson.MinProcessingTime,
			MaxProcessingTime:   shippingJson.MaxProcessingTime,
			ProcessingTimeUnit:  shippingJson.ProcessingTimeUnit,
			DomesticHandlingFee: shippingJson.DomesticHandlingFee,
			OriginState:         shippingJson.OriginState,
			OriginPostalCode:    shippingJson.OriginPostalCode,
			MinDeliveryDays:     shippingJson.MinDeliveryDays,
			MaxDeliveryDays:     shippingJson.MaxDeliveryDays,
			PrimaryCost:         shippingJson.PrimaryCost,
		}

		res, err := shippingProfileCollection.InsertOne(ctx, ShippingProfile)
		if err != nil {
			c.JSON(http.StatusNotModified, responses.UserResponse{Status: http.StatusNotModified, Message: "error", Data: map[string]interface{}{"error": err}})
			return
		}

		result := fmt.Sprintf("New Shop shipping profile has been added successfully, %v\n", res.InsertedID)
		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": result}})
	}
}

func GetShopShippingProfileInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		var shippingJson models.ShopShippingProfile
		defer cancel()

		profileIdString := c.Param("infoId")
		profileId, err := primitive.ObjectIDFromHex(profileIdString)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		err = shippingProfileCollection.FindOne(ctx, bson.M{"_id": profileId}).Decode(&shippingJson)
		if err != nil {
			c.JSON(http.StatusNotModified, responses.UserResponse{Status: http.StatusNotModified, Message: "error", Data: map[string]interface{}{"error": err}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": shippingJson}})
	}
}

func UpdateShopShippingProfileInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		var shippingJson models.ShopShippingProfileRequest
		defer cancel()

		shopId := c.Param("shopId")
		shopIdObj, err := primitive.ObjectIDFromHex(shopId)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		// Check if the user owns the shop
		userID, err := auth.ExtractTokenID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, responses.UserResponse{Status: http.StatusUnauthorized, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		err = verifyShopOwnership(ctx, shopIdObj, userID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, responses.UserResponse{Status: http.StatusUnauthorized, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		err = c.BindJSON(&shippingJson)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		// Validate request body
		if validationErr := validate.Struct(&shippingJson); validationErr != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": validationErr.Error()}})
			return
		}

		profileIdString := c.Param("infoId")
		profileId, err := primitive.ObjectIDFromHex(profileIdString)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		ShippingProfile := models.ShopShippingProfile{
			ShopID:              shopIdObj,
			Title:               shippingJson.Title,
			MinProcessingTime:   shippingJson.MinProcessingTime,
			MaxProcessingTime:   shippingJson.MaxProcessingTime,
			ProcessingTimeUnit:  shippingJson.ProcessingTimeUnit,
			DomesticHandlingFee: shippingJson.DomesticHandlingFee,
			OriginState:         shippingJson.OriginState,
			OriginPostalCode:    shippingJson.OriginPostalCode,
			MinDeliveryDays:     shippingJson.MinDeliveryDays,
			MaxDeliveryDays:     shippingJson.MaxDeliveryDays,
			PrimaryCost:         shippingJson.PrimaryCost,
		}

		res, err := shippingProfileCollection.UpdateOne(ctx, bson.M{"_id": profileId}, bson.M{"$set": ShippingProfile})
		if err != nil {
			c.JSON(http.StatusNotModified, responses.UserResponse{Status: http.StatusNotModified, Message: "error", Data: map[string]interface{}{"error": err}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": res}})
	}
}
