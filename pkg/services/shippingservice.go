package services

import (
	"context"
	"time"

	"khoomi-api-io/api/internal"
	"khoomi-api-io/api/internal/common"
	"khoomi-api-io/api/pkg/models"
	"khoomi-api-io/api/pkg/util"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type shippingService struct{}

func NewShippingService() ShippingService {
	return &shippingService{}
}

func (s *shippingService) CreateShopShippingProfile(ctx context.Context, userID, shopID primitive.ObjectID, req models.ShopShippingProfileRequest) (primitive.ObjectID, error) {
	now := time.Now()
	if err := common.Validate.Struct(&req); err != nil {
		return primitive.NilObjectID, err
	}

	shippingID := primitive.NewObjectID()
	callback := func(ctx mongo.SessionContext) (any, error) {

		shippingProfile := models.ShopShippingProfile{
			ID:                 shippingID,
			ShopID:             shopID,
			Title:              req.Title,
			HandlingFee:        req.HandlingFee,
			OriginState:        req.OriginState,
			OriginPostalCode:   req.OriginPostalCode,
			MinDeliveryDays:    req.MinDeliveryDays,
			MaxDeliveryDays:    req.MaxDeliveryDays,
			PrimaryPrice:       req.PrimaryPrice,
			DestinationBy:      req.DestinationBy,
			Destinations:       req.Destinations,
			SecondaryPrice:     req.SecondaryPrice,
			OffersFreeShipping: req.OffersFreeShipping,
			CreatedAt:          primitive.NewDateTimeFromTime(now),
			ModifiedAt:         primitive.NewDateTimeFromTime(now),
			IsDefault:          req.IsDefault,
			Processing:         req.Processing,
			AcceptReturns:      req.AcceptReturns,
			AcceptExchange:     req.AcceptExchange,
			ReturnPeriod:       req.ReturnPeriod,
			ReturnUnit:         req.ReturnUnit,
			Conditions:         req.Conditions,
		}

		res, err := common.ShippingProfileCollection.InsertOne(ctx, shippingProfile)
		if err != nil {
			return nil, err
		}

		if req.IsOnboarding {
			filter := bson.M{"_id": userID}
			update := bson.M{"$set": bson.M{"modified_at": now, "seller_onboarding_level": models.OnboardingLevelShipping}}
			_, err = common.UserCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				return nil, err
			}
		}

		return res, err
	}

	_, err := ExecuteTransaction(ctx, callback)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return shippingID, nil
}

func (s *shippingService) GetShopShippingProfile(ctx context.Context, profileID primitive.ObjectID) (*models.ShopShippingProfile, error) {
	var shippingProfile models.ShopShippingProfile
	err := common.ShippingProfileCollection.FindOne(ctx, bson.M{"_id": profileID}).Decode(&shippingProfile)
	if err != nil {
		return nil, err
	}

	return &shippingProfile, nil
}

func (s *shippingService) GetShopShippingProfiles(ctx context.Context, shopID primitive.ObjectID, pagination util.PaginationArgs) ([]models.ShopShippingProfile, int64, error) {
	filter := bson.M{"shop_id": shopID}
	findOptions := options.Find().
		SetLimit(int64(pagination.Limit)).
		SetSkip(int64(pagination.Skip)).
		SetSort(bson.D{{Key: "date", Value: -1}})

	cursor, err := common.ShippingProfileCollection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var shippingProfiles []models.ShopShippingProfile
	if err = cursor.All(ctx, &shippingProfiles); err != nil {
		return nil, 0, err
	}

	count, err := common.ShippingProfileCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return shippingProfiles, count, nil
}

func (s *shippingService) UpdateShippingProfile(ctx context.Context, shopId primitive.ObjectID, shippingId primitive.ObjectID, req models.UpdateShopShippingProfileRequest) (any, error) {

	now := time.Now()
	updateDoc := bson.M{
		"title":                req.Title,
		"destination_by":       req.DestinationBy,
		"origin_state":         req.OriginState,
		"methods":              req.ShippingMethod,
		"destinations":         req.Destinations,
		"processing":           req.Processing,
		"secondary_price":      req.SecondaryPrice,
		"primary_price":        req.PrimaryPrice,
		"handling_fee":         req.HandlingFee,
		"origin_postal_code":   req.OriginPostalCode,
		"max_delivery_days":    req.MaxDeliveryDays,
		"min_delivery_days":    req.MinDeliveryDays,
		"is_default":           req.IsDefault,
		"offers_free_shipping": req.OffersFreeShipping,
		"return_unit":          req.ReturnUnit,
		"conditons":            req.Conditions,
		"return_period":        req.ReturnPeriod,
		"accept_returns":       req.AcceptReturns,
		"accept_exchange":      req.AcceptExchange,
		"modified_at":          primitive.NewDateTimeFromTime(now),
	}

	res, err := common.ShippingProfileCollection.UpdateOne(ctx, bson.M{"_id": shippingId, "shop_id": shopId}, bson.M{"$set": updateDoc})
	if err != nil {
		return nil, err
	}

	internal.PublishCacheMessage(ctx, internal.CacheInvalidateShopShipping, shopId.Hex())

	return res.UpsertedID, nil
}

func (s *shippingService) DeleteShippingProfile(ctx context.Context, shopId primitive.ObjectID, shippingId primitive.ObjectID) (int64, error) {

	res, err := common.ShippingProfileCollection.DeleteOne(ctx, bson.M{"_id": shippingId, "shop_id": shopId})
	if err != nil {
		return 0, err
	}

	internal.PublishCacheMessage(ctx, internal.CacheInvalidateShopShipping, shopId.Hex())

	return res.DeletedCount, nil
}

func (s *shippingService) ChangeDefaultShippingProfile(ctx context.Context, shopId primitive.ObjectID, shippingId primitive.ObjectID) error {
	callback := func(ctx mongo.SessionContext) (any, error) {
		// Set all other profile to non-default
		err := SetOtherRecordsToFalse(ctx, common.ShippingProfileCollection, "shop_id", shopId, shippingId, "is_default")
		if err != nil {
			return nil, err
		}

		// Set the specified profile as default
		filter := bson.M{"shop_id": shopId, "_id": shippingId}
		result, err := common.ShippingProfileCollection.UpdateOne(
			ctx,
			filter,
			bson.M{"$set": bson.M{"is_default": true}},
		)
		if err != nil {
			return nil, err
		}

		if result.ModifiedCount == 0 {
			return nil, errors.New("shipping profile not found")
		}

		return result, nil
	}

	_, err := ExecuteTransaction(ctx, callback)
	if err == nil {
		internal.PublishCacheMessage(ctx, internal.CacheInvalidateShopShipping, shopId.Hex())
	}

	return err
}
