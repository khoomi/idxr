package controllers

import (
	"context"
	"khoomi-api-io/khoomi_api/auth"
	"khoomi-api-io/khoomi_api/configs"
	"khoomi-api-io/khoomi_api/helper"
	"khoomi-api-io/khoomi_api/models"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var paymentInformationCollection = configs.GetCollection(configs.DB, "PaymentInformation")

// / CreatePaymentInformation -> POST /:userId/payment-information/
func CreatePaymentInformation() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		defer cancel()

		userId, err := auth.ExtractTokenID(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Unauthorized")
			return
		}

		isSeller, err := auth.IsSeller(c) // Check if the user is a seller
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Unauthorized")
			return
		}
		if !isSeller {
			helper.HandleError(c, http.StatusUnauthorized, errors.New("Only sellers can perform this action"), "Unauthorized")
			return
		}

		var paymentInfo models.PaymentInformationRequest
		if err := c.BindJSON(&paymentInfo); err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Invalid data detected in JSON")
			return
		}

		if len(paymentInfo.AccountNumber) != 10 {
			helper.HandleError(c, http.StatusBadRequest, errors.New("Account number must be 10 digits"), "Invalid account number")
			return
		}

		paymentInfoToUpload := models.PaymentInformation{
			ID:            primitive.NewObjectID(),
			UserID:        userId,
			BankName:      paymentInfo.BankName,
			AccountName:   paymentInfo.AccountName,
			AccountNumber: paymentInfo.AccountNumber,
			IsDefault:     paymentInfo.IsDefault,
		}

		count, err := paymentInformationCollection.CountDocuments(ctx, bson.M{"user_id": userId})
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Error counting current payment information")
			return
		}

		if count >= 3 {
			helper.HandleError(c, http.StatusInternalServerError, errors.New("Max allowed payment information reached. Please delete other payment information to accommodate a new one."), "Max allowed payment information reached")
			return
		}

		res, err := paymentInformationCollection.InsertOne(ctx, paymentInfoToUpload)
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Error creating payment information")
			return
		}

		log.Printf("User %v added their payment account information", userId)
		helper.HandleSuccess(c, http.StatusOK, "Payment account information created successfully", gin.H{"inserted_id": res.InsertedID})
	}
}

// / GetPaymentInformations -> GET /:userId/payment-information/
func GetPaymentInformations() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		defer cancel()

		userId, err := auth.ExtractTokenID(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Unauthorized")
			return
		}
		isSeller, err := auth.IsSeller(c) // Check if the user is a seller
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Unauthorized")
			return
		}
		if !isSeller {
			helper.HandleError(c, http.StatusUnauthorized, errors.New("Only sellers can perform this action"), "Unauthorized")
			return
		}

		filter := bson.M{"user_id": userId}
		findOptions := options.Find().SetLimit(3)
		cursor, err := paymentInformationCollection.Find(ctx, filter, findOptions)
		if err != nil {
			log.Printf("Error fetching payment informations: %v", err)
			helper.HandleError(c, http.StatusNotFound, err, "Error fetching payment informations")
			return
		}
		defer cursor.Close(ctx)

		var paymentInfos []models.PaymentInformation
		if err := cursor.All(ctx, &paymentInfos); err != nil {
			helper.HandleError(c, http.StatusNotFound, err, "Error retrieving payment informations")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "Success", gin.H{"accounts": paymentInfos})
	}
}

// / ChangeDefaultPaymentInformation -> PUT /:userId/payment-information/:paymentInfoId
func ChangeDefaultPaymentInformation() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		defer cancel()

		userId, err := auth.ExtractTokenID(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Unauthorized")
			return
		}
		isSeller, err := auth.IsSeller(c) // Check if the user is a seller
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Unauthorized")
			return
		}
	 	if !isSeller {
			helper.HandleError(c, http.StatusUnauthorized, errors.New("Only sellers can perform this action!"), "Unauthorized")
			return
		}

		paymentInfoID := c.Param("paymentInfoId")
		if paymentInfoID == "" {
			helper.HandleError(c, http.StatusBadRequest, errors.New("No payment id was provided!"), "bad request")
			return
		}

		paymentObjectID, err := primitive.ObjectIDFromHex(paymentInfoID)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, errors.New("bad payment id"), "bad request")
			return
		}

		// Set all other payment information records to is_default=false
		_, err = paymentInformationCollection.UpdateMany(ctx, bson.M{"user_id": userId, "_id": bson.M{"$ne": paymentObjectID}}, bson.M{"$set": bson.M{"is_default": false}})
		if err != nil {
			log.Printf("Error updating payment informations: %v", err)
			helper.HandleError(c, http.StatusInternalServerError, err, "error modifying payment information")
			return
		}

		filter := bson.M{"user_id": userId, "_id": paymentObjectID}
		res, err := paymentInformationCollection.UpdateOne(ctx, filter, bson.M{"$set": bson.M{"is_default": true}})
		if err != nil {
			log.Printf("Error updating default payment informations: %v", err)
			helper.HandleError(c, http.StatusNotFound, err, "error modifying payment information")
			return
		}

		if res.ModifiedCount < 1 {
			log.Printf("Error updating default payment informations: %v", err)
			helper.HandleError(c, http.StatusNotFound, err, "payment information not modified")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "Default payment has been succesfuly changed.", gin.H{"modified": res.ModifiedCount})
	}
}

// / DeletePaymentInformation -> DELETE /:userId/payment-information/:paymentInfoId
func DeletePaymentInformation() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		defer cancel()

		paymentInfoID := c.Param("paymentInfoId")
		paymentObjectID, err := primitive.ObjectIDFromHex(paymentInfoID)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, errors.New("bad payment id"), "bad request")
			return
		}
		userID, err := auth.ExtractTokenID(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Unauthorized")
			return
		}

		isSeller, err := auth.IsSeller(c) // Check if the user is a seller
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Unauthorized")
			return
		}
		if !isSeller {
			helper.HandleError(c, http.StatusUnauthorized, errors.New("Only sellers can perform this action"), "Unauthorized")
			return
		}

		filter := bson.M{"_id": paymentObjectID, "user_id": userID}
		result, err := paymentInformationCollection.DeleteOne(ctx, filter)
		if err != nil {
			log.Printf("Error deleting payment information: %v", err)
			helper.HandleError(c, http.StatusNotFound, err, "Error deleting payment information")
			return
		}

		if result.DeletedCount == 0 {
			helper.HandleError(c, http.StatusNotFound, errors.New("No records deleted. Make sure you're using the correct _id"), "Data not found")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "Payment information deleted successfully",
			gin.H{"_id": paymentInfoID, "count": result.DeletedCount})
	}
}
