package controllers

import (
	"context"
	configs "khoomi-api-io/khoomi_api/config"
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
		ctx, cancel := context.WithTimeout(context.Background(), KhoomiRequestTimeoutSec)
		defer cancel()

		shopId := c.Param("shopid")
		shopIdObj, err := primitive.ObjectIDFromHex(shopId)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "invalid shopid")
		}

		// Check if the user owns the shop
		userID, err := configs.ExtractTokenID(c)
		if err != nil {
			helper.HandleError(c, http.StatusUnauthorized, err, "unauthorized")
			return
		}

		err = VerifyShopOwnership(c, userID, shopIdObj)
		if err != nil {
			log.Printf("Error you the shop owner: %s\n", err.Error())
			helper.HandleError(c, http.StatusUnauthorized, err, "shop ownership validation error")
			return
		}

		var verificationJson models.CreateSellerVerificationRequest
		err = c.BindJSON(&verificationJson)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "invalid request body")
		}

		// Validate request body
		if validationErr := Validate.Struct(&verificationJson); validationErr != nil {
			helper.HandleError(c, http.StatusBadRequest, validationErr, "invalid request body")
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
			helper.HandleError(c, http.StatusBadRequest, err, "document insert error")
		}

		helper.HandleSuccess(c, http.StatusOK, "successful", gin.H{"inserted_id": res.InsertedID})
	}
}

func GetSellerVerificationProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), KhoomiRequestTimeoutSec)
		defer cancel()

		shopId := c.Param("shopid")
		shopIdObj, err := primitive.ObjectIDFromHex(shopId)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "invalid shopid")
		}

		// Check if the user owns the shop
		userID, err := configs.ExtractTokenID(c)
		if err != nil {
			log.Println(err)
			helper.HandleError(c, http.StatusUnauthorized, err, "unauthorized")
		}
		err = VerifyShopOwnership(c, userID, shopIdObj)
		if err != nil {
			log.Printf("Error you the shop owner: %s\n", err.Error())
			helper.HandleError(c, http.StatusUnauthorized, err, "shop ownership validation error")
		}

		var verificationProfile models.SellerVerification
		err = SellerVerificationCollection.FindOne(ctx, bson.M{"shop_id": shopIdObj}).Decode(&verificationProfile)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				helper.HandleError(c, http.StatusNotFound, err, "Shop compliance information not found")
			}
			helper.HandleError(c, http.StatusInternalServerError, err, "error retrieving user information")
		}

		helper.HandleSuccess(c, http.StatusOK, "Seller verification profile retrieved successfully", gin.H{"verification_profile": verificationProfile})

	}
}
