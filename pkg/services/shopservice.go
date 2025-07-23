package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"khoomi-api-io/api/internal/common"
	"khoomi-api-io/api/pkg/models"
	"khoomi-api-io/api/pkg/util"

	slug2 "github.com/gosimple/slug"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ShopServiceImpl implements the ShopService interface
type ShopServiceImpl struct{}

// NewShopService creates a new instance of ShopService
func NewShopService() ShopService {
	return &ShopServiceImpl{}
}

// CheckShopNameAvailability checks if a shop username is available
func (ss *ShopServiceImpl) CheckShopNameAvailability(ctx context.Context, username string) (bool, error) {
	var shop models.Shop
	filter := bson.M{"username": username}
	err := common.ShopCollection.FindOne(ctx, filter).Decode(&shop)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return true, nil
		}
		return false, err
	}
	return false, nil
}

// CreateShop creates a new shop
func (ss *ShopServiceImpl) CreateShop(ctx context.Context, userID primitive.ObjectID, req CreateShopRequest) (primitive.ObjectID, error) {
	now := time.Now()
	shopID := primitive.NewObjectID()

	slug := slug2.Make(req.Username)
	policy := models.ShopPolicy{
		PaymentPolicy:  "",
		ShippingPolicy: "",
		RefundPolicy:   "",
		AdditionalInfo: "",
	}

	shopRating := models.Rating{
		AverageRating:  0.0,
		ReviewCount:    0,
		FiveStarCount:  0,
		FourStarCount:  0,
		ThreeStarCount: 0,
		TwoStarCount:   0,
		OneStarCount:   0,
	}

	shopAboutData := models.ShopAbout{
		Headline:  fmt.Sprintf("Welcome to %v!", req.Username),
		Story:     fmt.Sprintf("Thank you for visiting our online artisan shop. We are passionate about craftsmanship and dedicated to providing unique, handcrafted items that reflect the creativity and skill of our artisans. Explore our collection and discover the beauty of handmade products that carry a story of craftsmanship and tradition.\n\nAt %v, we believe in the art of creating something special. Each piece in our collection is carefully crafted with attention to detail and a commitment to quality. We aim to connect artisans with appreciative buyers, creating a community that values and supports the artistry behind every creation.\n\nJoin us on this journey of celebrating craftsmanship and supporting talented artisans from around the world. Your purchase not only adds a unique piece to your life but also contributes to the livelihood of skilled individuals who pour their heart and soul into their work.\n\nThank you for being a part of our community. Happy shopping!", req.Username),
		Instagram: fmt.Sprintf("@%v", req.Username),
		Facebook:  fmt.Sprintf("@%v", req.Username),
		X:         fmt.Sprintf("@%v", req.Username),
	}

	// Handle logo URL
	logoURL := common.DEFAULT_LOGO
	if req.LogoFile != nil {
		if logoStr, ok := req.LogoFile.(string); ok && logoStr != "" {
			logoURL = logoStr
		}
	}

	// Handle banner URL
	bannerURL := common.DEFAULT_THUMBNAIL
	if req.BannerFile != nil {
		if bannerStr, ok := req.BannerFile.(string); ok && bannerStr != "" {
			bannerURL = bannerStr
		}
	}

	shop := models.Shop{
		ID:                 shopID,
		Name:               req.Name,
		Description:        req.Description,
		Username:           req.Username,
		UserID:             userID,
		ListingActiveCount: 0,
		Announcement:       "",
		IsVacation:         false,
		VacationMessage:    "",
		Slug:               slug,
		LogoURL:            logoURL,
		BannerURL:          bannerURL,
		Gallery:            []string{},
		FollowerCount:      0,
		Followers:          []models.ShopFollower{},
		Status:             models.ShopStatusActive,
		IsLive:             true,
		CreatedAt:          now,
		ModifiedAt:         now,
		Policy:             policy,
		ReviewsCount:       0,
		Rating:             shopRating,
		About:              shopAboutData,
	}

	_, err := common.ShopCollection.InsertOne(ctx, shop)
	if err != nil {
		return primitive.NilObjectID, err
	}

	// Update user profile shop
	filter := bson.M{"_id": userID}
	update := bson.M{"$set": bson.M{"shop_id": shopID, "is_seller": true, "modified_at": now}}
	_, err = common.UserCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return shopID, nil
}

// UpdateShopInformation updates shop basic information
func (ss *ShopServiceImpl) UpdateShopInformation(ctx context.Context, shopID, userID primitive.ObjectID, req UpdateShopRequest) error {
	updateData := bson.M{}

	if req.Name != "" {
		if err := util.ValidateShopName(req.Name); err != nil {
			return err
		}
		updateData["name"] = req.Name
	}

	if req.Username != "" {
		if err := util.ValidateShopUserName(req.Username); err != nil {
			return err
		}
		updateData["username"] = req.Username
	}

	if req.Description != "" {
		updateData["description"] = req.Description
	}

	// Handle logo file
	if req.LogoFile != nil {
		if logoStr, ok := req.LogoFile.(string); ok && logoStr != "" {
			updateData["logo_url"] = logoStr
		}
	}

	// Handle banner file
	if req.BannerFile != nil {
		if bannerStr, ok := req.BannerFile.(string); ok && bannerStr != "" {
			updateData["banner_url"] = bannerStr
		}
	}

	if len(updateData) == 0 {
		return errors.New("no update data provided")
	}

	updateData["modified_at"] = time.Now()

	filter := bson.M{"_id": shopID, "user_id": userID}
	update := bson.M{"$set": updateData}

	_, err := common.ShopCollection.UpdateOne(ctx, filter, update)
	return err
}

// UpdateShopStatus updates shop live status
func (ss *ShopServiceImpl) UpdateShopStatus(ctx context.Context, shopID, userID primitive.ObjectID, isLive bool) error {
	filter := bson.M{"_id": shopID, "user_id": userID}
	update := bson.M{"$set": bson.M{"is_live": isLive, "modified_at": time.Now()}}
	res, err := common.ShopCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("no matching documents found")
	}
	return nil
}

// UpdateShopAddress updates shop address
func (ss *ShopServiceImpl) UpdateShopAddress(ctx context.Context, shopID, userID primitive.ObjectID, address models.ShopAddress) error {
	now := time.Now()
	address.ModifiedAt = now
	filter := bson.M{"_id": shopID, "user_id": userID}
	update := bson.M{"$set": bson.M{"address": address, "modified_at": now}}
	res, err := common.ShopCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("unknown error while trying to update shop")
	}
	return nil
}

// GetShop retrieves a shop by ID or slug
func (ss *ShopServiceImpl) GetShop(ctx context.Context, shopIdentifier string, withCategory bool) (*models.Shop, error) {
	var shopFilter bson.M

	if primitive.IsValidObjectID(shopIdentifier) {
		shopObjectID, err := primitive.ObjectIDFromHex(shopIdentifier)
		if err != nil {
			return nil, err
		}
		shopFilter = bson.M{"_id": shopObjectID}
	} else {
		shopFilter = bson.M{"slug": shopIdentifier}
	}

	shopPipeline := []bson.M{
		{"$match": shopFilter},
		{
			"$lookup": bson.M{
				"from":         "User",
				"localField":   "user_id",
				"foreignField": "_id",
				"as":           "user",
			},
		},
		{
			"$unwind": bson.M{
				"path":                       "$user",
				"preserveNullAndEmptyArrays": true,
			},
		},
		{
			"$project": bson.M{
				"_id":                      1,
				"name":                     1,
				"description":              1,
				"user_id":                  1,
				"username":                 1,
				"user_address_id":          1,
				"listing_active_count":     1,
				"announcement":             1,
				"announcement_modified_at": 1,
				"is_vacation":              1,
				"vacation_message":         1,
				"slug":                     1,
				"logo_url":                 1,
				"banner_url":               1,
				"gallery":                  1,
				"follower_count":           1,
				"followers":                1,
				"status":                   1,
				"is_live":                  1,
				"created_at":               1,
				"modified_at":              1,
				"policy":                   1,
				"recent_reviews":           1,
				"reviews_count":            1,
				"sales_message":            1,
				"rating":                   1,
				"address":                  1,
				"about":                    1,
				"user": bson.M{
					"login_name":             "$user.login_name",
					"first_name":             "$user.first_name",
					"last_name":              "$user.last_name",
					"thumbnail":              "$user.thumbnail",
					"transaction_buy_count":  "$user.transaction_buy_count",
					"transaction_sold_count": "$user.transaction_sold_count",
				},
			},
		},
	}

	cursor, err := common.ShopCollection.Aggregate(ctx, shopPipeline)
	if err != nil {
		return nil, err
	}

	var shop models.Shop
	if cursor.Next(ctx) {
		if err := cursor.Decode(&shop); err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("no shop found")
	}

	shop.ConstructShopLinks()

	if withCategory {
		listingPipeline := []bson.M{
			{"$match": bson.M{"shop_id": shop.ID}},
			{"$group": bson.M{"_id": "$details.category.category_name", "count": bson.M{"$sum": 1}}},
			{"$project": bson.M{"name": "$_id", "count": 1, "_id": 0, "path": "$details.category.category_path"}},
		}

		categoryCursor, err := common.ListingCollection.Aggregate(ctx, listingPipeline)
		if err != nil && err != mongo.ErrNoDocuments {
			return nil, err
		}

		var shopCategories []models.ShopCategory
		if categoryCursor.Next(ctx) {
			var shopCategory models.ShopCategory
			if err := categoryCursor.Decode(&shopCategory); err != nil {
				return nil, err
			}
			shopCategories = append(shopCategories, shopCategory)
		}
		shop.Categories = shopCategories
	}

	return &shop, nil
}

// GetShopByOwnerUserId retrieves a shop by owner user ID
func (ss *ShopServiceImpl) GetShopByOwnerUserId(ctx context.Context, userID primitive.ObjectID) (*models.Shop, error) {
	var shop models.Shop
	err := common.ShopCollection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&shop)
	if err != nil {
		return nil, err
	}
	return &shop, nil
}

// GetShops retrieves all active shops with pagination
func (ss *ShopServiceImpl) GetShops(ctx context.Context, pagination util.PaginationArgs) ([]models.Shop, error) {
	filter := bson.D{{Key: "status", Value: models.ShopStatusActive}}
	find := options.Find().SetLimit(int64(pagination.Limit)).SetSkip(int64(pagination.Skip))
	result, err := common.ShopCollection.Find(ctx, filter, find)
	if err != nil {
		return nil, err
	}

	var shops []models.Shop
	if err = result.All(ctx, &shops); err != nil {
		return nil, err
	}

	return shops, nil
}

// SearchShops searches for shops by name or description
func (ss *ShopServiceImpl) SearchShops(ctx context.Context, query string, pagination util.PaginationArgs) ([]models.Shop, int64, error) {
	searchFilter := bson.M{
		"$or": []bson.M{
			{"shop_name": bson.M{"$regex": primitive.Regex{Pattern: query, Options: "i"}}},
			{"description": bson.M{"$regex": primitive.Regex{Pattern: query, Options: "i"}}},
		},
	}

	shops, err := common.ShopCollection.Find(ctx, searchFilter,
		options.Find().SetSkip(int64(pagination.Skip)).SetLimit(int64(pagination.Limit)))
	if err != nil {
		return nil, 0, err
	}

	count, err := common.ShopCollection.CountDocuments(ctx, searchFilter)
	if err != nil {
		return nil, 0, err
	}

	var serializedShops []models.Shop
	for shops.Next(ctx) {
		var shop models.Shop
		if err := shops.Decode(&shop); err != nil {
			return nil, 0, err
		}
		serializedShops = append(serializedShops, shop)
	}

	return serializedShops, count, nil
}

// UpdateShopField updates a specific shop field
func (ss *ShopServiceImpl) UpdateShopField(ctx context.Context, shopID, userID primitive.ObjectID, field string, action string, data interface{}) error {
	now := time.Now()
	filter := bson.M{"_id": shopID, "user_id": userID}

	switch field {
	case "vacation":
		vacation, ok := data.(models.ShopVacationRequest)
		if !ok {
			return errors.New("invalid vacation data")
		}
		update := bson.M{"$set": bson.M{"vacation_message": vacation.Message, "is_vacation": vacation.IsVacation, "modified_at": now}}
		res, err := common.ShopCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			return err
		}
		if res.ModifiedCount == 0 {
			return errors.New("no matching documents found")
		}
	case "basic":
		basic, ok := data.(models.ShopBasicInformationRequest)
		if !ok {
			return errors.New("invalid basic data")
		}
		err := util.ValidateShopName(basic.Name)
		if err != nil {
			return err
		}
		err = util.ValidateShopDescription(basic.Description)
		if err != nil {
			return err
		}
		update := bson.M{"$set": bson.M{"name": basic.Name, "is_live": basic.IsLive, "description": basic.Description, "sales_message": basic.SalesMessage, "announcement": basic.Announcement, "modified_at": now}}
		res, err := common.ShopCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			return err
		}
		if res.ModifiedCount == 0 {
			return errors.New("no matching documents found")
		}
	case "policy":
		payload, ok := data.(models.ShopPolicy)
		if !ok {
			return errors.New("invalid policy data")
		}
		update := bson.M{"$set": bson.M{"policy": payload, "modified_at": now}}
		res, err := common.ShopCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			return err
		}
		if res.ModifiedCount == 0 {
			return errors.New("unknown error while trying to update shop")
		}
	default:
		return errors.New("unsupported field")
	}

	return nil
}

// UpdateShopAnnouncement updates shop announcement
func (ss *ShopServiceImpl) UpdateShopAnnouncement(ctx context.Context, shopID, userID primitive.ObjectID, announcement string) error {
	if announcement == "" {
		return errors.New("announcement cannot be empty")
	}
	if len(announcement) > 100 {
		return errors.New("announcement is too long")
	}

	now := time.Now()
	filter := bson.M{"_id": shopID, "user_id": userID}
	update := bson.M{"$set": bson.M{"announcement": announcement, "announcement_modified_at": now, "modified_at": now}}
	res, err := common.ShopCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("no matching documents found")
	}
	return nil
}

// UpdateShopVacation updates shop vacation status
func (ss *ShopServiceImpl) UpdateShopVacation(ctx context.Context, shopID, userID primitive.ObjectID, req models.ShopVacationRequest) error {
	now := time.Now()
	filter := bson.M{"_id": shopID, "user_id": userID}
	update := bson.M{"$set": bson.M{"vacation_message": req.Message, "is_vacation": req.IsVacation, "modified_at": now}}
	res, err := common.ShopCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("no matching documents found")
	}
	return nil
}

// FollowShop allows a user to follow a shop
func (ss *ShopServiceImpl) FollowShop(ctx context.Context, userID, shopID primitive.ObjectID) (primitive.ObjectID, error) {
	now := time.Now()
	followerId := primitive.NewObjectID()

	var user models.User
	err := common.UserCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return primitive.NilObjectID, err
	}

	var currentShop models.Shop
	err = common.ShopCollection.FindOne(ctx, bson.M{"_id": shopID}).Decode(&currentShop)
	if err != nil {
		return primitive.NilObjectID, err
	}

	shopMemberData := models.ShopFollower{
		Id:        followerId,
		UserId:    userID,
		ShopId:    shopID,
		LoginName: user.LoginName,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Thumbnail: user.Thumbnail,
		IsOwner:   currentShop.UserID == userID,
		JoinedAt:  time.Now(),
	}
	_, err = common.ShopFollowerCollection.InsertOne(ctx, shopMemberData)
	if err != nil {
		return primitive.NilObjectID, err
	}

	inner := models.ShopFollowerExcerpt{
		Id:        followerId,
		UserId:    userID,
		LoginName: user.LoginName,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Thumbnail: user.Thumbnail,
		IsOwner:   currentShop.UserID == userID,
	}
	filter := bson.M{"_id": shopID, "followers": bson.M{"$not": bson.M{"$elemMatch": bson.M{"user_id": &user.Id}}}}
	update := bson.M{
		"$push": bson.M{
			"followers": bson.M{
				"$each":  bson.A{inner},
				"$sort":  -1,
				"$slice": -5,
			},
		},
		"$set": bson.M{"modified_at": now},
		"$inc": bson.M{"follower_count": 1},
	}
	result, err := common.ShopCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return primitive.NilObjectID, err
	}

	if result.ModifiedCount == 0 {
		return primitive.NilObjectID, errors.New("no matching documents found")
	}

	return followerId, nil
}

// UnfollowShop allows a user to unfollow a shop
func (ss *ShopServiceImpl) UnfollowShop(ctx context.Context, userID, shopID primitive.ObjectID) error {
	filter := bson.M{"shop_id": shopID, "user_id": userID}
	_, err := common.ShopFollowerCollection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	filter = bson.M{"_id": shopID}
	update := bson.M{"$pull": bson.M{"followers": bson.M{"user_id": userID}}, "$inc": bson.M{"follower_count": -1}}
	result, err := common.ShopCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.ModifiedCount == 0 {
		return errors.New("no matching documents found")
	}

	return nil
}

// IsFollowingShop checks if a user is following a shop
func (ss *ShopServiceImpl) IsFollowingShop(ctx context.Context, userID, shopID primitive.ObjectID) (bool, error) {
	filter := bson.M{"user_id": userID, "shop_id": shopID}
	var follower models.ShopFollower
	err := common.ShopFollowerCollection.FindOne(ctx, filter).Decode(&follower)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetShopFollowers retrieves shop followers with pagination
func (ss *ShopServiceImpl) GetShopFollowers(ctx context.Context, shopID primitive.ObjectID, pagination util.PaginationArgs) ([]models.ShopFollower, int64, error) {
	filter := bson.M{"shop_id": shopID}
	find := options.Find().SetLimit(int64(pagination.Limit)).SetSkip(int64(pagination.Skip))
	result, err := common.ShopFollowerCollection.Find(ctx, filter, find)
	if err != nil {
		return nil, 0, err
	}

	count, err := common.ShopFollowerCollection.CountDocuments(ctx, bson.M{"shop_id": shopID})
	if err != nil {
		return nil, 0, err
	}

	var shopFollowers []models.ShopFollower
	if err = result.All(ctx, &shopFollowers); err != nil {
		return nil, 0, err
	}

	return shopFollowers, count, nil
}

// RemoveOtherFollower removes another user from shop followers
func (ss *ShopServiceImpl) RemoveOtherFollower(ctx context.Context, ownerID, shopID, userToRemoveID primitive.ObjectID) error {
	if ownerID == userToRemoveID {
		return errors.New("cannot remove yourself")
	}

	err := ss.VerifyShopOwnership(ctx, ownerID, shopID)
	if err != nil {
		return err
	}

	filter := bson.M{"shop_id": shopID, "user_id": userToRemoveID}
	_, err = common.ShopFollowerCollection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	filter = bson.M{"_id": shopID}
	update := bson.M{"$pull": bson.M{"followers": bson.M{"user_id": userToRemoveID}}, "$inc": bson.M{"follower_count": -1}}
	_, err = common.ShopCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

// UpdateShopAbout updates shop about information
func (ss *ShopServiceImpl) UpdateShopAbout(ctx context.Context, shopID, userID primitive.ObjectID, about models.ShopAbout) error {
	err := ss.VerifyShopOwnership(ctx, userID, shopID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": shopID}
	update := bson.M{"$set": bson.M{"about": about}}
	res, err := common.ShopCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("no matching documents found")
	}
	return nil
}

// UpdateShopGallery adds an image to shop gallery
func (ss *ShopServiceImpl) UpdateShopGallery(ctx context.Context, shopID, userID primitive.ObjectID, imageURL string) error {
	now := time.Now()
	filter := bson.M{"_id": shopID, "user_id": userID}
	update := bson.M{"$push": bson.M{"gallery": imageURL}, "modified_at": now}
	res, err := common.ShopCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("failed to update gallery")
	}
	return nil
}

// DeleteFromShopGallery removes an image from shop gallery
func (ss *ShopServiceImpl) DeleteFromShopGallery(ctx context.Context, shopID, userID primitive.ObjectID, imageURL string) error {
	now := time.Now()
	filter := bson.M{"_id": shopID, "user_id": userID}
	update := bson.M{"$pull": bson.M{"gallery": imageURL}, "modified_at": now}
	res, err := common.ShopCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("no matching documents found")
	}
	return nil
}

// CreateShopReturnPolicy creates a new return policy
func (ss *ShopServiceImpl) CreateShopReturnPolicy(ctx context.Context, shopID, userID primitive.ObjectID, policy models.ShopReturnPolicies) (primitive.ObjectID, error) {
	err := ss.VerifyShopOwnership(ctx, userID, shopID)
	if err != nil {
		return primitive.NilObjectID, err
	}

	policy.ID = primitive.NewObjectID()
	policy.ShopId = shopID

	_, err = common.ShopReturnPolicyCollection.InsertOne(ctx, policy)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return policy.ID, nil
}

// UpdateShopReturnPolicy updates an existing return policy
func (ss *ShopServiceImpl) UpdateShopReturnPolicy(ctx context.Context, shopID, userID primitive.ObjectID, policy models.ShopReturnPolicies) error {
	err := ss.VerifyShopOwnership(ctx, userID, shopID)
	if err != nil {
		return err
	}

	filter := bson.M{"shop_id": shopID}
	update := bson.M{"$set": bson.M{"accepts_return": policy.AcceptsReturn, "accepts_echanges": policy.AcceptsExchanges, "deadline": policy.Deadline}}
	res, err := common.ShopReturnPolicyCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if res.ModifiedCount == 0 {
		return errors.New("no matching documents found")
	}

	return nil
}

// DeleteShopReturnPolicy deletes a return policy
func (ss *ShopServiceImpl) DeleteShopReturnPolicy(ctx context.Context, shopID, userID, policyID primitive.ObjectID) error {
	err := ss.VerifyShopOwnership(ctx, userID, shopID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": policyID, "shop_id": shopID}
	_, err = common.ShopReturnPolicyCollection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	return nil
}

// GetShopReturnPolicy retrieves a specific return policy
func (ss *ShopServiceImpl) GetShopReturnPolicy(ctx context.Context, shopID, policyID primitive.ObjectID) (*models.ShopReturnPolicies, error) {
	var currentPolicy models.ShopReturnPolicies
	filter := bson.M{"_id": policyID, "shop_id": shopID}
	err := common.ShopReturnPolicyCollection.FindOne(ctx, filter).Decode(&currentPolicy)
	if err != nil {
		return nil, err
	}
	return &currentPolicy, nil
}

// GetShopReturnPolicies retrieves all return policies for a shop
func (ss *ShopServiceImpl) GetShopReturnPolicies(ctx context.Context, shopID primitive.ObjectID) ([]models.ShopReturnPolicies, error) {
	cursor, err := common.ShopReturnPolicyCollection.Find(ctx, bson.M{"shop_id": shopID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var policies []models.ShopReturnPolicies
	for cursor.Next(ctx) {
		var policy models.ShopReturnPolicies
		if err := cursor.Decode(&policy); err != nil {
			return nil, err
		}
		policies = append(policies, policy)
	}

	return policies, nil
}

// CreateShopComplianceInformation creates compliance information
func (ss *ShopServiceImpl) CreateShopComplianceInformation(ctx context.Context, shopID, userID primitive.ObjectID, compliance models.ComplianceInformationRequest) error {
	err := ss.VerifyShopOwnership(ctx, userID, shopID)
	if err != nil {
		return err
	}

	complianceInformation := models.ComplianceInformation{
		ID:                   primitive.NewObjectID(),
		ShopID:               shopID,
		TermsOfUse:           compliance.TermsOfUse,
		IntellectualProperty: compliance.IntellectualProperty,
		SellerPolicie:        compliance.SellerPolicie,
	}

	_, err = common.ShopCompliancePolicyCollection.InsertOne(ctx, complianceInformation)
	if err != nil {
		return err
	}

	return nil
}

// GetShopComplianceInformation retrieves compliance information
func (ss *ShopServiceImpl) GetShopComplianceInformation(ctx context.Context, shopID primitive.ObjectID) (*models.ComplianceInformation, error) {
	var complianceInformation models.ComplianceInformation
	err := common.ShopCompliancePolicyCollection.FindOne(ctx, bson.M{"shop_id": shopID}).Decode(&complianceInformation)
	if err != nil {
		return nil, err
	}
	return &complianceInformation, nil
}

// UpdateShopLogo updates shop logo
func (ss *ShopServiceImpl) UpdateShopLogo(ctx context.Context, shopID, userID primitive.ObjectID, logoURL string) error {
	now := time.Now()
	filter := bson.M{"_id": shopID, "user_id": userID}
	update := bson.M{"$set": bson.M{"logo_url": logoURL, "modified_at": now}}
	res, err := common.ShopCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("failed to update logo")
	}
	return nil
}

// UpdateShopBanner updates shop banner
func (ss *ShopServiceImpl) UpdateShopBanner(ctx context.Context, shopID, userID primitive.ObjectID, bannerURL string) error {
	now := time.Now()
	filter := bson.M{"_id": shopID, "user_id": userID}
	update := bson.M{"$set": bson.M{"banner_url": bannerURL, "modified_at": now}}
	res, err := common.ShopCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("failed to update banner")
	}
	return nil
}

// DeleteShopLogo resets shop logo to default
func (ss *ShopServiceImpl) DeleteShopLogo(ctx context.Context, shopID, userID primitive.ObjectID) error {
	now := time.Now()
	filter := bson.M{"_id": shopID, "user_id": userID}
	update := bson.M{"$set": bson.M{"logo_url": common.DEFAULT_LOGO, "modified_at": now}}
	res, err := common.ShopCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("failed to reset logo")
	}
	return nil
}

// DeleteShopBanner resets shop banner to default
func (ss *ShopServiceImpl) DeleteShopBanner(ctx context.Context, shopID, userID primitive.ObjectID) error {
	now := time.Now()
	filter := bson.M{"_id": shopID, "user_id": userID}
	update := bson.M{"$set": bson.M{"banner_url": common.DEFAULT_THUMBNAIL, "modified_at": now}}
	res, err := common.ShopCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("failed to reset banner")
	}
	return nil
}

// VerifyShopOwnership verifies if a user owns a given shop using its shopId
func (ss *ShopServiceImpl) VerifyShopOwnership(ctx context.Context, userID, shopID primitive.ObjectID) error {
	// Use FindOne with projection to only fetch _id field - most efficient approach
	var result struct {
		ID primitive.ObjectID `bson:"_id"`
	}
	err := common.ShopCollection.FindOne(ctx, bson.M{"_id": shopID, "user_id": userID}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("user does not own the shop")
		}
		return err
	}
	return nil
}
