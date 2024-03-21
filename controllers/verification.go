package controllers

import (
	"context"
	"khoomi-api-io/khoomi_api/config"
	"khoomi-api-io/khoomi_api/helper"
	"khoomi-api-io/khoomi_api/models"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateSellerVerificationProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), REQ_TIMEOUT_SECS)
		defer cancel()

		shopId := c.Param("shopid")
		shopIdObj, err := primitive.ObjectIDFromHex(shopId)
		if err != nil {
			helper.HandleError(c, http.StatusUnprocessableEntity, err, "Invalid shop id")
			return
		}

		// Check if the user owns the shop
		auth, err := config.InitJwtClaim(c)
		if err != nil {
			helper.HandleError(c, http.StatusUnauthorized, err, "Invalid user token, id or access")
			return
		}
		userId, err := auth.GetUserObjectId()
		if err != nil {
			helper.HandleError(c, http.StatusUnauthorized, err, "Invalid user token, id or access")
			return
		}

		err = VerifyShopOwnership(c, userId, shopIdObj)
		if err != nil {
			log.Printf("Error you the shop owner: %s\n", err.Error())
			helper.HandleError(c, http.StatusForbidden, err, "shop ownership validation error")
			return
		}

		var verificationJson models.CreateSellerVerificationRequest
		err = c.BindJSON(&verificationJson)
		if err != nil {
			log.Println(err)
			helper.HandleError(c, http.StatusUnprocessableEntity, err, "invalid request body")
			return
		}

		// Validate request body
		if validationErr := Validate.Struct(&verificationJson); validationErr != nil {
			log.Println(validationErr)
			helper.HandleError(c, http.StatusUnprocessableEntity, err, "invalid request body")
			return
		}

		verificationId := primitive.NewObjectID()
		now := time.Now()
		ShippingProfile := models.SellerVerification{
			ID:                 verificationId,
			ShopId:             shopIdObj,
			FirstName:          verificationJson.FirstName,
			LastName:           verificationJson.LastName,
			Card:               verificationJson.Card,
			CardNumber:         verificationJson.CardNumber,
			IsVerified:         false,
			VerifiedAt:         now,
			DOB:                verificationJson.DOB,
			CountryOfResidence: verificationJson.CountryOfResidence,
			Nationality:        verificationJson.Nationality,
			CreatedAt:          now,
			ModifiedAt:         now,
		}
		res, err := SellerVerificationCollection.InsertOne(ctx, ShippingProfile)
		if err != nil {
			log.Println(err)
			helper.HandleError(c, http.StatusUnauthorized, err, "Internal server error while creating verification")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "successful", res.InsertedID)
	}
}

func GetSellerVerificationProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), REQ_TIMEOUT_SECS)
		defer cancel()

		shopId := c.Param("shopid")
		shopIdObj, err := primitive.ObjectIDFromHex(shopId)
		if err != nil {
			helper.HandleError(c, http.StatusUnprocessableEntity, err, "invalid shopid")
			return
		}

		// Check if the user owns the shop
		auth, err := config.InitJwtClaim(c)
		if err != nil {
			helper.HandleError(c, http.StatusUnauthorized, err, "Invalid user token, id or access")
			return
		}
		userId, err := auth.GetUserObjectId()
		if err != nil {
			helper.HandleError(c, http.StatusUnauthorized, err, "Invalid user token, id or access")
			return
		}
		err = VerifyShopOwnership(c, userId, shopIdObj)
		if err != nil {
			log.Printf("Error you the shop owner: %s\n", err.Error())
			helper.HandleError(c, http.StatusForbidden, err, "shop ownership validation error")
			return
		}

		var verificationProfile models.SellerVerification
		err = SellerVerificationCollection.FindOne(ctx, bson.M{"shop_id": shopIdObj}).Decode(&verificationProfile)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				helper.HandleError(c, http.StatusNotFound, err, "verification profile not found")
				return
			}
			helper.HandleError(c, http.StatusInternalServerError, err, "Internal server error while fetching verification profile")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "Seller verification profile retrieved successfully", verificationProfile)

	}
}
