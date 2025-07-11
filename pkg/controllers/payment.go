package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"khoomi-api-io/api/internal"
	auth "khoomi-api-io/api/internal/auth"
	"khoomi-api-io/api/internal/common"
	"khoomi-api-io/api/pkg/models"
	"khoomi-api-io/api/pkg/util"

	"github.com/durango/go-credit-card"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

// SetOtherAddressesToFalse sets IsDefaultShippingAddress to false for other addresses belonging to the user
func setOtherPaymentsToFalse(ctx context.Context, collection *mongo.Collection, userId primitive.ObjectID, paymentId primitive.ObjectID) error {
	filter := bson.M{
		"user_id":    userId,
		"_id":        bson.M{"$ne": paymentId},
		"is_default": true,
	}

	update := bson.M{
		"$set": bson.M{"is_default": false},
	}

	_, err := collection.UpdateMany(ctx, filter, update)
	return err
}

// setOtherUserPaymentCardsToFalse sets is_default to false for other user payment cards
func setOtherUserPaymentCardsToFalse(ctx context.Context, collection *mongo.Collection, userId primitive.ObjectID, paymentId primitive.ObjectID) error {
	filter := bson.M{
		"userId":     userId,
		"_id":        bson.M{"$ne": paymentId},
		"is_default": true,
	}

	update := bson.M{
		"$set": bson.M{"is_default": false},
	}

	_, err := collection.UpdateMany(ctx, filter, update)
	return err
}

// CreateSellerPaymentInformation -> POST /shop/:shopId/payment-information/
func CreateSellerPaymentInformation() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		userId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		res, err := common.IsSeller(c, userId)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		if !res {
			util.HandleError(c, http.StatusUnauthorized, errors.New("only sellers can perform this action"))
			return
		}

		var paymentInfo models.SellerPaymentInformationRequest
		if err := c.BindJSON(&paymentInfo); err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		if len(paymentInfo.AccountNumber) != 10 {
			util.HandleError(c, http.StatusBadRequest, errors.New("account number must be 10 digits"))
			return
		}

		if len(paymentInfo.BankName) < 3 {
			util.HandleError(c, http.StatusBadRequest, errors.New("invalid bank name"))
			return
		}

		paymentInfoToUpload := models.SellerPaymentInformation{
			ID:            primitive.NewObjectID(),
			UserID:        userId,
			BankName:      paymentInfo.BankName,
			AccountName:   paymentInfo.AccountName,
			AccountNumber: paymentInfo.AccountNumber,
			IsDefault:     paymentInfo.IsDefault,
		}

		count, err := common.SellerPaymentInformationCollection.CountDocuments(ctx, bson.M{"user_id": userId})
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		if count >= 5 {
			util.HandleError(c, http.StatusInternalServerError, errors.New("Max allowed payment information reached. Please delete other payment information to accommodate a new one."))
			return
		}

		wc := writeconcern.New(writeconcern.WMajority())
		txnOptions := options.Transaction().SetWriteConcern(wc)
		session, err := util.DB.StartSession()
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
		}

		defer session.EndSession(ctx)

		callback := func(ctx mongo.SessionContext) (any, error) {
			log.Println("Start mongo transaction for new payament information creation")
			if paymentInfoToUpload.IsDefault {
				// Set IsDefaultShippingAddress to false for other addresses belonging to the user
				err = setOtherPaymentsToFalse(ctx, common.SellerPaymentInformationCollection, userId, paymentInfoToUpload.ID)
				if err != nil {
					return nil, err
				}

			}

			insertRes, insertErr := common.SellerPaymentInformationCollection.InsertOne(ctx, paymentInfoToUpload)
			if insertErr != nil {
				return nil, insertErr
			}

			return insertRes, nil
		}

		_, err = session.WithTransaction(ctx, callback, txnOptions)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		if err := session.CommitTransaction(ctx); err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		log.Printf("User %v added their payment account information", userId)

		internal.PublishCacheMessage(c, internal.CacheRevalidatePayment, userId.Hex())

		util.HandleSuccess(c, http.StatusOK, "Payment account information created successfully", paymentInfoToUpload.ID.Hex())
	}
}

// GetSellerPaymentInformations -> GET /shop/:shopId/payment-information/
func GetSellerPaymentInformations() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		userId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}
		res, err := common.IsSeller(c, userId)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}
		if !res {
			util.HandleError(c, http.StatusUnauthorized, errors.New("Only sellers can perform this action"))
			return
		}

		filter := bson.M{"user_id": userId}
		paginationArgs := common.GetPaginationArgs(c)
		findOptions := options.Find().
			SetLimit(int64(paginationArgs.Limit)).
			SetSkip(int64(paginationArgs.Skip)).
			SetSort(bson.D{{Key: "date", Value: -1}})
		cursor, err := common.SellerPaymentInformationCollection.Find(ctx, filter, findOptions)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}
		defer cursor.Close(ctx)

		var paymentInfos []models.SellerPaymentInformation
		if err := cursor.All(ctx, &paymentInfos); err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}

		count, err := common.SellerPaymentInformationCollection.CountDocuments(ctx, filter)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		util.HandleSuccessMeta(c, http.StatusOK, "success", paymentInfos, gin.H{
			"pagination": util.Pagination{
				Limit: paginationArgs.Limit,
				Skip:  paginationArgs.Skip,
				Count: count,
			},
		})
	}
}

// ChangeDefaultSellerPaymentInformation -> PUT /shop/:shopId/payment-information/:paymentInfoId
func ChangeDefaultSellerPaymentInformation() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		userId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err)
			return
		}
		res, err := common.IsSeller(c, userId)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}
		if !res {
			util.HandleError(c, http.StatusUnauthorized, errors.New("Only sellers can perform this action"))
			return
		}

		paymentInfoID := c.Param("paymentInfoId")
		if paymentInfoID == "" {
			util.HandleError(c, http.StatusBadRequest, errors.New("No payment id was provided!"))
			return
		}

		paymentObjectID, err := primitive.ObjectIDFromHex(paymentInfoID)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, errors.New("bad payment id"))
			return
		}

		// Set all other payment information records to is_default=false
		_, err = common.SellerPaymentInformationCollection.UpdateMany(ctx, bson.M{"user_id": userId, "_id": bson.M{"$ne": paymentObjectID}}, bson.M{"$set": bson.M{"is_default": false}})
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		filter := bson.M{"user_id": userId, "_id": paymentObjectID}
		insertRes, insertErr := common.SellerPaymentInformationCollection.UpdateOne(ctx, filter, bson.M{"$set": bson.M{"is_default": true}})
		if insertErr != nil {
			util.HandleError(c, http.StatusNotModified, err)
			return
		}

		internal.PublishCacheMessage(c, internal.CacheRevalidatePayment, userId.Hex())

		util.HandleSuccess(c, http.StatusOK, "Default payment has been succesfuly changed.", insertRes.ModifiedCount)
	}
}

// DeleteSellerPaymentInformation -> DELETE /shop/:shopId/payment-information/:paymentInfoId
func DeleteSellerPaymentInformation() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		paymentInfoID := c.Param("paymentInfoId")
		paymentObjectID, err := primitive.ObjectIDFromHex(paymentInfoID)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, errors.New("bad payment id"))
			return
		}
		userId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		res, err := common.IsSeller(c, userId)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}
		if !res {
			util.HandleError(c, http.StatusUnauthorized, errors.New("Only sellers can perform this action"))
			return
		}
		filter := bson.M{"_id": paymentObjectID, "user_id": userId}
		result, err := common.SellerPaymentInformationCollection.DeleteOne(ctx, filter)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}

		if result.DeletedCount == 0 {
			util.HandleError(c, http.StatusNotFound, errors.New("No records deleted. Make sure you're using the correct _id"))
			return
		}

		internal.PublishCacheMessage(c, internal.CacheRevalidatePayment, userId.Hex())

		util.HandleSuccess(c, http.StatusOK, "Payment information deleted successfully", result.DeletedCount)
	}
}

func CompletedPaymentOnboarding() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		userId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}
		res, err := common.IsSeller(c, userId)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}
		if !res {
			util.HandleError(c, http.StatusUnauthorized, errors.New("Only sellers can perform this action"))
			return
		}
		filter := bson.M{"user_id": userId}
		findOptions := options.Find().SetLimit(1)
		cursor, err := common.SellerPaymentInformationCollection.Find(ctx, filter, findOptions)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err)
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

// / CreatePaymentInformation -> POST /:userId/payment/cards
func CreatePaymentCard() gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		userId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		var cardInfo models.PaymentCardInformationRequest
		if err := c.BindJSON(&cardInfo); err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		if validationErr := common.Validate.Struct(cardInfo); validationErr != nil {
			util.HandleError(c, http.StatusBadRequest, validationErr)
			return
		}

		card := creditcard.Card{
			Number:  cardInfo.CardNumber,
			Cvv:     cardInfo.CVV,
			Month:   string(cardInfo.ExpiryMonth),
			Year:    string(cardInfo.ExpiryYear),
			Company: creditcard.Company{},
		}
		err = card.Validate(true)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}
		lastFour, err := card.LastFour()
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		cardToUpload := models.PaymentCardInformation{
			ID:             primitive.NewObjectID(),
			UserID:         userId,
			IsDefault:      cardInfo.IsDefault,
			CardHolderName: cardInfo.CardHolderName,
			CardNumber:     card.Number,
			ExpiryMonth:    card.Month,
			ExpiryYear:     card.Year,
			CVV:            card.Cvv,
			LastFourDigits: lastFour,
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		count, err := common.UserPaymentCardsTable.CountDocuments(ctx, bson.M{"userId": userId})
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		if count >= 5 {
			errorMsg := "Max allowed payment cards reached. Please delete other cards to accommodate a new one."
			util.HandleError(c, http.StatusInternalServerError, errors.New(errorMsg))
			return
		}

		wc := writeconcern.New(writeconcern.WMajority())
		txnOptions := options.Transaction().SetWriteConcern(wc)
		session, err := util.DB.StartSession()
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
		}

		defer session.EndSession(ctx)

		callback := func(ctx mongo.SessionContext) (any, error) {
			log.Println("Starting mongo transaction for new payment card creation")
			if cardToUpload.IsDefault {
				// Set IsDefaultShippingAddress to false for other addresses belonging to the user
				err = setOtherUserPaymentCardsToFalse(ctx, common.UserPaymentCardsTable, userId, cardToUpload.ID)
				if err != nil {
					return nil, err
				}

			}

			insertRes, insertErr := common.UserPaymentCardsTable.InsertOne(ctx, cardToUpload)
			if insertErr != nil {
				return nil, insertErr
			}

			return insertRes, nil
		}

		_, err = session.WithTransaction(ctx, callback, txnOptions)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		if err := session.CommitTransaction(ctx); err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		log.Printf("User %v added a card", userId)

		internal.PublishCacheMessage(c, internal.CacheRevalidatePayment, userId.Hex())
		util.HandleSuccess(c, http.StatusOK, "new Card created successfully", cardToUpload.ID.Hex())
	}
}

// / GetPaymentCards-> GET /:userId/payment/cards
func GetPaymentCards() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		userId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		filter := bson.M{"userId": userId}
		paginationArgs := common.GetPaginationArgs(c)
		findOptions := options.Find().
			SetLimit(int64(paginationArgs.Limit)).
			SetSkip(int64(paginationArgs.Skip)).
			SetSort(bson.D{{Key: "createdAt", Value: -1}})
		cursor, err := common.UserPaymentCardsTable.Find(ctx, filter, findOptions)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}
		defer cursor.Close(ctx)

		var paymentInfos []models.PaymentCardInformation
		if err := cursor.All(ctx, &paymentInfos); err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}

		count, err := common.UserPaymentCardsTable.CountDocuments(ctx, filter)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		util.HandleSuccessMeta(c, http.StatusOK, "success", paymentInfos, gin.H{
			"pagination": util.Pagination{
				Limit: paginationArgs.Limit,
				Skip:  paginationArgs.Skip,
				Count: count,
			},
		})
	}
}

// / ChangeDefaulterPaymentCard-> PUT /:userId/payment/cards/:id
func ChangeDefaultPaymentCard() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		userId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err)
			return
		}

		paymentInfoID := c.Param("id")
		if paymentInfoID == "" {
			util.HandleError(c, http.StatusBadRequest, errors.New("No card id was provided!"))
			return
		}

		paymentObjectID, err := primitive.ObjectIDFromHex(paymentInfoID)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, errors.New("bad card id"))
			return
		}

		// Set all other payment information records to is_default=false
		_, err = common.UserPaymentCardsTable.UpdateMany(ctx, bson.M{"userId": userId, "_id": bson.M{"$ne": paymentObjectID}}, bson.M{"$set": bson.M{"is_default": false}})
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		filter := bson.M{"userId": userId, "_id": paymentObjectID}
		insertRes, insertErr := common.UserPaymentCardsTable.UpdateOne(ctx, filter, bson.M{"$set": bson.M{"is_default": true}})
		if insertErr != nil {
			util.HandleError(c, http.StatusNotModified, err)
			return
		}

		internal.PublishCacheMessage(c, internal.CacheRevalidatePayment, userId.Hex())
		util.HandleSuccess(c, http.StatusOK, "Default card has been succesfuly changed.", insertRes.ModifiedCount)
	}
}

// / DeletePaymentCard-> DELETE /user/:userId/payment/card/:id
func DeletePaymentCard() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		paymentInfoID := c.Param("id")
		paymentObjectID, err := primitive.ObjectIDFromHex(paymentInfoID)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, errors.New("bad card id"))
			return
		}
		userId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		filter := bson.M{"_id": paymentObjectID, "userId": userId}
		result, err := common.UserPaymentCardsTable.DeleteOne(ctx, filter)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}

		if result.DeletedCount == 0 {
			util.HandleError(c, http.StatusNotFound, errors.New("No records deleted. Make sure you're using the correct _id"))
			return
		}

		internal.PublishCacheMessage(c, internal.CacheRevalidatePayment, userId.Hex())

		util.HandleSuccess(c, http.StatusOK, "card deleted successfully", result.DeletedCount)
	}
}
