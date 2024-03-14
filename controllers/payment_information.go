package controllers

import (
	"context"
	configs "khoomi-api-io/khoomi_api/config"
	"khoomi-api-io/khoomi_api/helper"
	"khoomi-api-io/khoomi_api/models"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

// SetOtherAddressesToFalse sets IsDefaultShippingAddress to false for other addresses belonging to the user
func setOtherPaymentsToFalse(ctx context.Context, userId primitive.ObjectID, paymentId primitive.ObjectID) error {
	filter := bson.M{
		"user_id":    userId,
		"_id":        bson.M{"$ne": paymentId},
		"is_default": true,
	}

	update := bson.M{
		"$set": bson.M{"is_default": false},
	}

	_, err := PaymentInformationCollection.UpdateMany(ctx, filter, update)
	return err
}

// / CreatePaymentInformation -> POST /:userId/payment-information/
func CreatePaymentInformation() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), KhoomiRequestTimeoutSec)
		defer cancel()

		userId, err := configs.ValidateUserID(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Unauthorized")
			return
		}

		res, err := IsSeller(c, userId)
		if err != nil {
			log.Println(err)
			helper.HandleError(c, http.StatusInternalServerError, err, "Error finding user")
		}
		if !res {
			helper.HandleError(c, http.StatusUnauthorized, errors.New("Only sellers can perform this action"), "Unauthorized")
		}

		var paymentInfo models.PaymentInformationRequest
		if err := c.BindJSON(&paymentInfo); err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Invalid data detected in JSON")
		}

		if len(paymentInfo.AccountNumber) != 10 {
			helper.HandleError(c, http.StatusBadRequest, errors.New("Account number must be 10 digits"), "Invalid account number")
		}

		if len(paymentInfo.BankName) < 3 {
			helper.HandleError(c, http.StatusBadRequest, errors.New("Invalid bank name"), "Invalid bank name")
		}

		paymentInfoToUpload := models.PaymentInformation{
			ID:            primitive.NewObjectID(),
			UserID:        userId,
			BankName:      paymentInfo.BankName,
			AccountName:   paymentInfo.AccountName,
			AccountNumber: paymentInfo.AccountNumber,
			IsDefault:     paymentInfo.IsDefault,
		}

		count, err := PaymentInformationCollection.CountDocuments(ctx, bson.M{"user_id": userId})
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Error counting current payment information")
		}

		if count >= 3 {
			helper.HandleError(c, http.StatusInternalServerError, errors.New("Max allowed payment information reached. Please delete other payment information to accommodate a new one."), "Max allowed payment information reached")
		}

		wc := writeconcern.New(writeconcern.WMajority())
		txnOptions := options.Transaction().SetWriteConcern(wc)
		session, err := configs.DB.StartSession()
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "failed to start mongodb session")
		}

		defer session.EndSession(ctx)

		callback := func(ctx mongo.SessionContext) (interface{}, error) {
			log.Println("Start mongo transaction for new payament information creation")
			if paymentInfoToUpload.IsDefault {
				// Set IsDefaultShippingAddress to false for other addresses belonging to the user
				err = setOtherPaymentsToFalse(ctx, userId, paymentInfoToUpload.ID)
				if err != nil {
					return nil, err
				}

			}

			insertRes, insertErr := PaymentInformationCollection.InsertOne(ctx, paymentInfoToUpload)
			if insertErr != nil {
				return nil, insertErr
			}

			return insertRes, nil
		}

		result, err := session.WithTransaction(ctx, callback, txnOptions)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Failed to execute transaction")
		}

		if err := session.CommitTransaction(ctx); err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Failed to commit transaction")
		}

		log.Printf("User %v added their payment account information", userId)
		helper.HandleSuccess(c, http.StatusOK, "Payment account information created successfully", result)
	}
}

// / GetPaymentInformations -> GET /:userId/payment-information/
func GetPaymentInformations() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), KhoomiRequestTimeoutSec)
		defer cancel()

		userId, err := configs.ValidateUserID(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Unauthorized")
		}
		res, err := IsSeller(c, userId)
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Error finding user")
		}
		if !res {
			helper.HandleError(c, http.StatusUnauthorized, errors.New("Only sellers can perform this action"), "Unauthorized")
		}

		filter := bson.M{"user_id": userId}
		findOptions := options.Find().SetLimit(3)
		cursor, err := PaymentInformationCollection.Find(ctx, filter, findOptions)
		if err != nil {
			log.Printf("Error fetching payment informations: %v", err)
			helper.HandleError(c, http.StatusNotFound, err, "Error fetching payment informations")
		}
		defer cursor.Close(ctx)

		var paymentInfos []models.PaymentInformation
		if err := cursor.All(ctx, &paymentInfos); err != nil {
			helper.HandleError(c, http.StatusNotFound, err, "Error retrieving payment informations")
		}

		helper.HandleSuccess(c, http.StatusOK, "Success", gin.H{"accounts": paymentInfos})
	}
}

// / ChangeDefaultPaymentInformation -> PUT /:userId/payment-information/:paymentInfoId
func ChangeDefaultPaymentInformation() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), KhoomiRequestTimeoutSec)
		defer cancel()

		userId, err := configs.ValidateUserID(c)
		if err != nil {
			helper.HandleError(c, http.StatusUnauthorized, err, "Unauthorized")
		}
		res, err := IsSeller(c, userId)
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Error finding user")
		}
		if !res {
			helper.HandleError(c, http.StatusUnauthorized, errors.New("Only sellers can perform this action"), "Unauthorized")
		}

		paymentInfoID := c.Param("paymentInfoId")
		if paymentInfoID == "" {
			helper.HandleError(c, http.StatusBadRequest, errors.New("No payment id was provided!"), "bad request")
		}

		paymentObjectID, err := primitive.ObjectIDFromHex(paymentInfoID)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, errors.New("bad payment id"), "bad request")
		}

		// Set all other payment information records to is_default=false
		_, err = PaymentInformationCollection.UpdateMany(ctx, bson.M{"user_id": userId, "_id": bson.M{"$ne": paymentObjectID}}, bson.M{"$set": bson.M{"is_default": false}})
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "error modifying payment information")
		}

		filter := bson.M{"user_id": userId, "_id": paymentObjectID}
		insertRes, insertErr := PaymentInformationCollection.UpdateOne(ctx, filter, bson.M{"$set": bson.M{"is_default": true}})
		if insertErr != nil {
			helper.HandleError(c, http.StatusNotFound, err, "error modifying payment information")
		}

		if insertRes.ModifiedCount < 1 {
			helper.HandleError(c, http.StatusNotFound, err, "payment information not modified")
		}

		helper.HandleSuccess(c, http.StatusOK, "Default payment has been succesfuly changed.", gin.H{"modified": insertRes.ModifiedCount})
	}
}

// / DeletePaymentInformation -> DELETE /:userId/payment-information/:paymentInfoId
func DeletePaymentInformation() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), KhoomiRequestTimeoutSec)
		defer cancel()

		paymentInfoID := c.Param("paymentInfoId")
		paymentObjectID, err := primitive.ObjectIDFromHex(paymentInfoID)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, errors.New("bad payment id"), "bad request")
		}
		userId, err := configs.ValidateUserID(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Unauthorized")
		}

		res, err := IsSeller(c, userId)
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Error finding user")
		}
		if res == false {
			helper.HandleError(c, http.StatusUnauthorized, errors.New("Only sellers can perform this action"), "Unauthorized")
		}
		filter := bson.M{"_id": paymentObjectID, "user_id": userId}
		result, err := PaymentInformationCollection.DeleteOne(ctx, filter)
		if err != nil {
			log.Printf("Error deleting payment information: %v", err)
			helper.HandleError(c, http.StatusNotFound, err, "Error deleting payment information")
		}

		if result.DeletedCount == 0 {
			helper.HandleError(c, http.StatusNotFound, errors.New("No records deleted. Make sure you're using the correct _id"), "Data not found")
		}

		helper.HandleSuccess(c, http.StatusOK, "Payment information deleted successfully",
			gin.H{"_id": paymentInfoID, "count": result.DeletedCount})
	}
}

func CompletedPaymentOnboarding() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), KhoomiRequestTimeoutSec)
		defer cancel()

		userId, err := configs.ValidateUserID(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Unauthorized")
		}
		res, err := IsSeller(c, userId)
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Error finding user")
		}
		if res == false {
			helper.HandleError(c, http.StatusUnauthorized, errors.New("Only sellers can perform this action"), "Unauthorized")
		}
		filter := bson.M{"user_id": userId}
		findOptions := options.Find().SetLimit(1)
		cursor, err := PaymentInformationCollection.Find(ctx, filter, findOptions)
		if err != nil {
			log.Printf("Error fetching payment information: %v", err)
			helper.HandleError(c, http.StatusNotFound, err, "Error fetching payment information")
		}
		defer cursor.Close(ctx)

		hasPaymentInfo := cursor.Next(ctx)

		if !hasPaymentInfo {
			hasPaymentInfo = false
		}

		helper.HandleSuccess(c, http.StatusOK, "Success", gin.H{"has_payment_information": hasPaymentInfo})
	}
}
