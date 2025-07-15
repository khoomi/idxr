package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"khoomi-api-io/api/internal/auth"
	"khoomi-api-io/api/internal/common"
	"khoomi-api-io/api/pkg/models"
	"khoomi-api-io/api/pkg/util"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateSellerVerificationProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), common.REQUEST_TIMEOUT_SECS)
		defer cancel()

		shopId := c.Param("shopid")
		shopIdObj, err := primitive.ObjectIDFromHex(shopId)
		if err != nil {
			util.HandleError(c, http.StatusUnprocessableEntity, err)
			return
		}

		// Check if the user owns the shop
		session_, err := auth.GetSessionAuto(c)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err)
			return
		}

		err = common.VerifyShopOwnership(c, session_.UserId, shopIdObj)
		if err != nil {
			log.Printf("Error you the shop owner: %s\n", err.Error())
			util.HandleError(c, http.StatusForbidden, err)
			return
		}

		var verificationJson models.CreateSellerVerificationRequest
		err = c.BindJSON(&verificationJson)
		if err != nil {
			util.HandleError(c, http.StatusUnprocessableEntity, err)
			return
		}

		// Validate request body
		if validationErr := common.Validate.Struct(&verificationJson); validationErr != nil {
			util.HandleError(c, http.StatusUnprocessableEntity, validationErr)
			return
		}

		verificationId := primitive.NewObjectID()
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
		res, err := common.SellerVerificationCollection.InsertOne(ctx, ShippingProfile)
		if err != nil {
			util.HandleError(c, http.StatusFound, err)
			return
		}

		util.HandleSuccess(c, http.StatusOK, "successful", res.InsertedID)
	}
}

func GetSellerVerificationProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQUEST_TIMEOUT_SECS)
		defer cancel()

		shopId := c.Param("shopid")
		shopIdObj, err := primitive.ObjectIDFromHex(shopId)
		if err != nil {
			util.HandleError(c, http.StatusUnprocessableEntity, err)
			return
		}

		// Check if the user owns the shop
		session_, err := auth.GetSessionAuto(c)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err)
			return
		}
		err = common.VerifyShopOwnership(c, session_.UserId, shopIdObj)
		if err != nil {
			log.Printf("Error you the shop owner: %s\n", err.Error())
			util.HandleError(c, http.StatusForbidden, err)
			return
		}

		var verificationProfile models.SellerVerification
		err = common.SellerVerificationCollection.FindOne(ctx, bson.M{"shop_id": shopIdObj}).Decode(&verificationProfile)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				util.HandleError(c, http.StatusNotFound, err)
				return
			}
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Seller verification profile retrieved successfully", verificationProfile)
	}
}
