package services

import (
	"context"
	"errors"
	"time"

	"khoomi-api-io/api/internal/common"
	"khoomi-api-io/api/pkg/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// VerificationServiceImpl implements the VerificationService interface
type VerificationServiceImpl struct{}

// NewVerificationService creates a new instance of VerificationService
func NewVerificationService() VerificationService {
	return &VerificationServiceImpl{}
}

// CreateSellerVerificationProfile creates a new seller verification profile
func (vs *VerificationServiceImpl) CreateSellerVerificationProfile(ctx context.Context, userID, shopID primitive.ObjectID, req models.CreateSellerVerificationRequest) (primitive.ObjectID, error) {
	// Verify shop ownership
	err := vs.verifyShopOwnership(ctx, userID, shopID)
	if err != nil {
		return primitive.NilObjectID, err
	}

	now := time.Now()
	verificationID := primitive.NewObjectID()

	verification := models.SellerVerification{
		ID:                 verificationID,
		ShopId:             shopID,
		FirstName:          req.FirstName,
		LastName:           req.LastName,
		Card:               req.Card,
		CardNumber:         req.CardNumber,
		IsVerified:         false,
		VerifiedAt:         now,
		DOB:                req.DOB,
		CountryOfResidence: req.CountryOfResidence,
		Nationality:        req.Nationality,
		CreatedAt:          now,
		ModifiedAt:         now,
	}

	_, err = common.SellerVerificationCollection.InsertOne(ctx, verification)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return verificationID, nil
}

// GetSellerVerificationProfile retrieves a seller verification profile
func (vs *VerificationServiceImpl) GetSellerVerificationProfile(ctx context.Context, userID, shopID primitive.ObjectID) (*models.SellerVerification, error) {
	// Verify shop ownership
	err := vs.verifyShopOwnership(ctx, userID, shopID)
	if err != nil {
		return nil, err
	}

	var verificationProfile models.SellerVerification
	err = common.SellerVerificationCollection.FindOne(ctx, bson.M{"shop_id": shopID}).Decode(&verificationProfile)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, err
		}
		return nil, err
	}

	return &verificationProfile, nil
}

// verifyShopOwnership is a temporary helper method to verify shop ownership
// TODO: This should be refactored to use ShopService dependency injection
func (vs *VerificationServiceImpl) verifyShopOwnership(ctx context.Context, userID, shopID primitive.ObjectID) error {
	var shop struct {
		UserID primitive.ObjectID `bson:"user_id"`
	}

	err := common.ShopCollection.FindOne(ctx, bson.M{"_id": shopID}).Decode(&shop)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("shop not found")
		}
		return err
	}

	if shop.UserID != userID {
		return errors.New("unauthorized: you don't own this shop")
	}

	return nil
}
