package services

import (
	"context"
	"time"

	"khoomi-api-io/api/internal"
	"khoomi-api-io/api/internal/common"
	"khoomi-api-io/api/pkg/models"
	"khoomi-api-io/api/pkg/util"

	creditcard "github.com/durango/go-credit-card"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type paymentService struct{}

func NewPaymentService() PaymentService {
	return &paymentService{}
}

func (p *paymentService) CreateSellerPaymentInformation(ctx context.Context, userID primitive.ObjectID, req models.SellerPaymentInformationRequest) (primitive.ObjectID, error) {
	now := time.Now()
	if len(req.AccountNumber) != 10 {
		return primitive.NilObjectID, errors.New("account number must be 10 digits")
	}

	if len(req.BankName) < 3 {
		return primitive.NilObjectID, errors.New("invalid bank name")
	}

	paymentInfoToUpload := models.SellerPaymentInformation{
		ID:            primitive.NewObjectID(),
		UserID:        userID,
		BankName:      req.BankName,
		AccountName:   req.AccountName,
		AccountNumber: req.AccountNumber,
		IsDefault:     req.IsDefault,
	}

	err := CheckRecordLimit(ctx, common.SellerPaymentInformationCollection, "user_id", userID, 10, "maximum payment information limit reached")
	if err != nil {
		return primitive.NilObjectID, err
	}

	callback := func(ctx mongo.SessionContext) (any, error) {
		if paymentInfoToUpload.IsDefault {
			err = SetOtherRecordsToFalse(ctx, common.SellerPaymentInformationCollection, "user_id", userID, paymentInfoToUpload.ID, "is_default")
			if err != nil {
				return nil, err
			}
		}

		insertRes, insertErr := common.SellerPaymentInformationCollection.InsertOne(ctx, paymentInfoToUpload)
		if insertErr != nil {
			return nil, insertErr
		}

		if req.IsOnboarding {
			filter := bson.M{"_id": userID}
			update := bson.M{"$set": bson.M{"modified_at": now, "seller_onboarding_level": models.OnboardingLevelPayment}}
			_, err = common.UserCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				return nil, err
			}
		}

		return insertRes, nil
	}

	_, err = ExecuteTransaction(ctx, callback)
	if err != nil {
		return primitive.NilObjectID, err
	}

	internal.PublishCacheMessage(ctx, internal.CacheInvalidatePayment, userID.Hex())

	return paymentInfoToUpload.ID, nil
}

func (p *paymentService) GetSellerPaymentInformations(ctx context.Context, userID primitive.ObjectID, pagination util.PaginationArgs) ([]models.SellerPaymentInformation, int64, error) {
	filter := bson.M{"user_id": userID}
	findOptions := options.Find().
		SetLimit(int64(pagination.Limit)).
		SetSkip(int64(pagination.Skip)).
		SetSort(bson.D{{Key: "date", Value: -1}})

	cursor, err := common.SellerPaymentInformationCollection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var paymentInfos []models.SellerPaymentInformation
	if err := cursor.All(ctx, &paymentInfos); err != nil {
		return nil, 0, err
	}

	count, err := common.SellerPaymentInformationCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return paymentInfos, count, nil
}

func (p *paymentService) ChangeDefaultSellerPaymentInformation(ctx context.Context, userID, paymentInfoID primitive.ObjectID) error {
	_, err := common.SellerPaymentInformationCollection.UpdateMany(ctx, bson.M{"user_id": userID, "_id": bson.M{"$ne": paymentInfoID}}, bson.M{"$set": bson.M{"is_default": false}})
	if err != nil {
		return err
	}

	filter := bson.M{"user_id": userID, "_id": paymentInfoID}
	_, err = common.SellerPaymentInformationCollection.UpdateOne(ctx, filter, bson.M{"$set": bson.M{"is_default": true}})
	if err != nil {
		return err
	}

	internal.PublishCacheMessage(ctx, internal.CacheInvalidatePayment, userID.Hex())

	return nil
}

func (p *paymentService) DeleteSellerPaymentInformation(ctx context.Context, userID, paymentInfoID primitive.ObjectID) error {
	filter := bson.M{"_id": paymentInfoID, "user_id": userID}
	result, err := common.SellerPaymentInformationCollection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("No records deleted. Make sure you're using the correct _id")
	}

	internal.PublishCacheMessage(ctx, internal.CacheInvalidatePayment, userID.Hex())

	return nil
}

func (p *paymentService) HasSellerPaymentInformation(ctx context.Context, userID primitive.ObjectID) (bool, error) {
	filter := bson.M{"user_id": userID}
	findOptions := options.Find().SetLimit(1)
	cursor, err := common.SellerPaymentInformationCollection.Find(ctx, filter, findOptions)
	if err != nil {
		return false, err
	}
	defer cursor.Close(ctx)

	return cursor.Next(ctx), nil
}

func (p *paymentService) CreatePaymentCard(ctx context.Context, userID primitive.ObjectID, req models.PaymentCardInformationRequest) (primitive.ObjectID, error) {
	now := time.Now()

	if err := common.Validate.Struct(req); err != nil {
		return primitive.NilObjectID, err
	}

	card := creditcard.Card{
		Number:  req.CardNumber,
		Cvv:     req.CVV,
		Month:   string(req.ExpiryMonth),
		Year:    string(req.ExpiryYear),
		Company: creditcard.Company{},
	}

	err := card.Validate(true)
	if err != nil {
		return primitive.NilObjectID, err
	}

	lastFour, err := card.LastFour()
	if err != nil {
		return primitive.NilObjectID, err
	}

	cardToUpload := models.PaymentCardInformation{
		ID:             primitive.NewObjectID(),
		UserID:         userID,
		IsDefault:      req.IsDefault,
		CardHolderName: req.CardHolderName,
		CardNumber:     card.Number,
		ExpiryMonth:    card.Month,
		ExpiryYear:     card.Year,
		CVV:            card.Cvv,
		LastFourDigits: lastFour,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err = CheckRecordLimit(ctx, common.UserPaymentCardsTable, "userId", userID, 10, "maximum payment cards limit reached")
	if err != nil {
		return primitive.NilObjectID, err
	}

	callback := func(ctx mongo.SessionContext) (any, error) {
		if cardToUpload.IsDefault {
			err = SetOtherRecordsToFalse(ctx, common.UserPaymentCardsTable, "userId", userID, cardToUpload.ID, "is_default")
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

	_, err = ExecuteTransaction(ctx, callback)
	if err != nil {
		return primitive.NilObjectID, err
	}

	internal.PublishCacheMessage(ctx, internal.CacheInvalidatePayment, userID.Hex())

	return cardToUpload.ID, nil
}

func (p *paymentService) GetPaymentCards(ctx context.Context, userID primitive.ObjectID, pagination util.PaginationArgs) ([]models.PaymentCardInformation, int64, error) {
	filter := bson.M{"userId": userID}
	findOptions := options.Find().
		SetLimit(int64(pagination.Limit)).
		SetSkip(int64(pagination.Skip)).
		SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := common.UserPaymentCardsTable.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var paymentInfos []models.PaymentCardInformation
	if err := cursor.All(ctx, &paymentInfos); err != nil {
		return nil, 0, err
	}

	count, err := common.UserPaymentCardsTable.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return paymentInfos, count, nil
}

func (p *paymentService) ChangeDefaultPaymentCard(ctx context.Context, userID, cardID primitive.ObjectID) error {
	_, err := common.UserPaymentCardsTable.UpdateMany(ctx, bson.M{"userId": userID, "_id": bson.M{"$ne": cardID}}, bson.M{"$set": bson.M{"is_default": false}})
	if err != nil {
		return err
	}

	filter := bson.M{"userId": userID, "_id": cardID}
	_, err = common.UserPaymentCardsTable.UpdateOne(ctx, filter, bson.M{"$set": bson.M{"is_default": true}})
	if err != nil {
		return err
	}

	internal.PublishCacheMessage(ctx, internal.CacheInvalidatePayment, userID.Hex())

	return nil
}

func (p *paymentService) DeletePaymentCard(ctx context.Context, userID, cardID primitive.ObjectID) error {
	filter := bson.M{"_id": cardID, "userId": userID}
	result, err := common.UserPaymentCardsTable.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("No records deleted. Make sure you're using the correct _id")
	}

	internal.PublishCacheMessage(ctx, internal.CacheInvalidatePayment, userID.Hex())

	return nil
}
