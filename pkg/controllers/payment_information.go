package controllers

import (
	"context"
	auth "khoomi-api-io/api/internal/auth"
	"khoomi-api-io/api/internal/common"
	"khoomi-api-io/api/pkg/models"
	"khoomi-api-io/api/pkg/util"
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

	_, err := common.PaymentInformationCollection.UpdateMany(ctx, filter, update)
	return err
}

// / CreatePaymentInformation -> POST /:userId/payment-information/
func CreatePaymentInformation() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		userId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err, "Unauthorized")
			return
		}

		res, err := common.IsSeller(c, userId)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err, "Error finding user")
			return
		}
		if !res {
			util.HandleError(c, http.StatusUnauthorized, errors.New("Only sellers can perform this action"), "Unauthorized")
			return
		}

		var paymentInfo models.PaymentInformationRequest
		if err := c.BindJSON(&paymentInfo); err != nil {
			util.HandleError(c, http.StatusBadRequest, err, "Invalid data detected in JSON")
			return
		}

		if len(paymentInfo.AccountNumber) != 10 {
			util.HandleError(c, http.StatusBadRequest, errors.New("Account number must be 10 digits"), "Invalid account number")
			return
		}

		if len(paymentInfo.BankName) < 3 {
			util.HandleError(c, http.StatusBadRequest, errors.New("Invalid bank name"), "Invalid bank name")
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

		count, err := common.PaymentInformationCollection.CountDocuments(ctx, bson.M{"user_id": userId})
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err, "Error counting current payment information")
			return
		}

		if count >= 3 {
			util.HandleError(c, http.StatusInternalServerError, errors.New("Max allowed payment information reached. Please delete other payment information to accommodate a new one."), "Max allowed payment information reached")
			return
		}

		wc := writeconcern.New(writeconcern.WMajority())
		txnOptions := options.Transaction().SetWriteConcern(wc)
		session, err := util.DB.StartSession()
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err, "failed to start mongodb session")
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

			insertRes, insertErr := common.PaymentInformationCollection.InsertOne(ctx, paymentInfoToUpload)
			if insertErr != nil {
				return nil, insertErr
			}

			return insertRes, nil
		}

		_, err = session.WithTransaction(ctx, callback, txnOptions)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err, "Failed to execute transaction")
			return
		}

		if err := session.CommitTransaction(ctx); err != nil {
			util.HandleError(c, http.StatusBadRequest, err, "Failed to commit transaction")
			return
		}

		log.Printf("User %v added their payment account information", userId)
		util.HandleSuccess(c, http.StatusOK, "Payment account information created successfully", paymentInfoToUpload.ID.Hex())
	}
}

// / GetPaymentInformations -> GET /:userId/payment-information/
func GetPaymentInformations() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		userId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err, "Unauthorized")
			return
		}
		res, err := common.IsSeller(c, userId)
		if err != nil {
			log.Println(err)
			util.HandleError(c, http.StatusInternalServerError, err, "Error finding user")
			return
		}
		if !res {
			util.HandleError(c, http.StatusUnauthorized, errors.New("Only sellers can perform this action"), "Unauthorized")
			return
		}

		filter := bson.M{"user_id": userId}
		findOptions := options.Find().SetLimit(3)
		cursor, err := common.PaymentInformationCollection.Find(ctx, filter, findOptions)
		if err != nil {
			log.Printf("Error fetching payment informations: %v", err)
			util.HandleError(c, http.StatusNotFound, err, "Error fetching payment informations")
			return
		}
		defer cursor.Close(ctx)

		var paymentInfos []models.PaymentInformation
		if err := cursor.All(ctx, &paymentInfos); err != nil {
			util.HandleError(c, http.StatusNotFound, err, "Error retrieving payment informations")
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Success", paymentInfos)
	}
}

// / ChangeDefaultPaymentInformation -> PUT /:userId/payment-information/:paymentInfoId
func ChangeDefaultPaymentInformation() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		userId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err, "Unauthorized")
			return
		}
		res, err := common.IsSeller(c, userId)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err, "Error finding user")
			return
		}
		if !res {
			util.HandleError(c, http.StatusUnauthorized, errors.New("Only sellers can perform this action"), "Unauthorized")
			return
		}

		paymentInfoID := c.Param("paymentInfoId")
		if paymentInfoID == "" {
			util.HandleError(c, http.StatusBadRequest, errors.New("No payment id was provided!"), "bad request")
			return
		}

		paymentObjectID, err := primitive.ObjectIDFromHex(paymentInfoID)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, errors.New("bad payment id"), "bad request")
			return
		}

		// Set all other payment information records to is_default=false
		_, err = common.PaymentInformationCollection.UpdateMany(ctx, bson.M{"user_id": userId, "_id": bson.M{"$ne": paymentObjectID}}, bson.M{"$set": bson.M{"is_default": false}})
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err, "error modifying payment information")
			return
		}

		filter := bson.M{"user_id": userId, "_id": paymentObjectID}
		insertRes, insertErr := common.PaymentInformationCollection.UpdateOne(ctx, filter, bson.M{"$set": bson.M{"is_default": true}})
		if insertErr != nil {
			util.HandleError(c, http.StatusNotModified, err, "error modifying payment information")
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Default payment has been succesfuly changed.", insertRes.ModifiedCount)
	}
}

// / DeletePaymentInformation -> DELETE /:userId/payment-information/:paymentInfoId
func DeletePaymentInformation() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		paymentInfoID := c.Param("paymentInfoId")
		paymentObjectID, err := primitive.ObjectIDFromHex(paymentInfoID)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, errors.New("bad payment id"), "bad request")
			return
		}
		userId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err, "Unauthorized")
			return
		}

		res, err := common.IsSeller(c, userId)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err, "Error finding user")
			return
		}
		if res == false {
			util.HandleError(c, http.StatusUnauthorized, errors.New("Only sellers can perform this action"), "Unauthorized")
			return
		}
		filter := bson.M{"_id": paymentObjectID, "user_id": userId}
		result, err := common.PaymentInformationCollection.DeleteOne(ctx, filter)
		if err != nil {
			log.Printf("Error deleting payment information: %v", err)
			util.HandleError(c, http.StatusNotFound, err, "Error deleting payment information")
			return
		}

		if result.DeletedCount == 0 {
			util.HandleError(c, http.StatusNotFound, errors.New("No records deleted. Make sure you're using the correct _id"), "Data not found")
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Payment information deleted successfully", result.DeletedCount)
	}
}

func CompletedPaymentOnboarding() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		userId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err, "Unauthorized")
			return
		}
		res, err := common.IsSeller(c, userId)
		if err != nil {
			log.Println(err)
			util.HandleError(c, http.StatusInternalServerError, err, "Error finding user")
			return
		}
		if res == false {
			util.HandleError(c, http.StatusUnauthorized, errors.New("Only sellers can perform this action"), "Unauthorized")
			return
		}
		filter := bson.M{"user_id": userId}
		findOptions := options.Find().SetLimit(1)
		cursor, err := common.PaymentInformationCollection.Find(ctx, filter, findOptions)
		if err != nil {
			log.Printf("Error fetching payment information: %v", err)
			util.HandleError(c, http.StatusNotFound, err, "Error fetching payment information")
			return
		}
		defer cursor.Close(ctx)

		hasPaymentInfo := cursor.Next(ctx)

		if !hasPaymentInfo {
			hasPaymentInfo = false
		}

		util.HandleSuccess(c, http.StatusOK, "Success", hasPaymentInfo)
	}
}
