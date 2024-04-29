package controllers

import (
	"context"
	"khoomi-api-io/khoomi_api/config"
	configs "khoomi-api-io/khoomi_api/config"
	"khoomi-api-io/khoomi_api/helper"
	"khoomi-api-io/khoomi_api/models"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var UserCollection = configs.GetCollection(configs.DB, "User")
var ShopCollection = configs.GetCollection(configs.DB, "Shop")
var UserAddressCollection = configs.GetCollection(configs.DB, "UserAddress")
var LoginHistoryCollection = configs.GetCollection(configs.DB, "UserLoginHistory")
var PasswordResetTokenCollection = configs.GetCollection(configs.DB, "UserPasswordResetToken")
var EmailVerificationTokenCollection = configs.GetCollection(configs.DB, "UserEmailVerificationToken")
var WishListCollection = configs.GetCollection(configs.DB, "UserWishList")
var UserDeletionCollection = configs.GetCollection(configs.DB, "UserDeletionRequest")
var NotificationCollection = configs.GetCollection(configs.DB, "UserNotification")
var ShopAboutCollection = configs.GetCollection(configs.DB, "ShopAbout")
var ShopFollowerCollection = configs.GetCollection(configs.DB, "ShopFollower")
var ShopReviewCollection = configs.GetCollection(configs.DB, "ShopReview")
var ShopReturnPolicyCollection = configs.GetCollection(configs.DB, "ShopReturnPolicies")
var ShopCompliancePolicyCollection = configs.GetCollection(configs.DB, "ShopCompliancePolicy")
var ShippingProfileCollection = configs.GetCollection(configs.DB, "ShopShippingProfile")
var SellerVerificationCollection = configs.GetCollection(configs.DB, "SellerVerification")
var ListingCollection = configs.GetCollection(configs.DB, "Listing")
var PaymentInformationCollection = configs.GetCollection(configs.DB, "SellerPaymentInformation")
var UserNotificationCollection = configs.GetCollection(configs.DB, "UserNotification")
var Validate = validator.New()
var DefaultUserThumbnail = "https://res.cloudinary.com/kh-oo-mi/image/upload/v1705607383/khoomi/xp78ywxq8ggvo6muf4ry.png"
var DefaultThumbnail = "https://res.cloudinary.com/kh-oo-mi/image/upload/v1705607175/khoomi/mypvl86lihcqvkcqmvbg.jpg"
var DefaultLogo = "https://res.cloudinary.com/kh-oo-mi/image/upload/v1703704749/UCuy4YhFhyCvo3-jeXhNjR4yIeQ/zvzr1l17hz2c3yhqhf89.png"

const (
	REQ_TIMEOUT_SECS                   = 50 * time.Second
	MongoDuplicateKeyCode              = 11000
	VERIFICATION_EMAIL_EXPIRATION_TIME = 1 * time.Hour
)

// CurrentUser get current user using userId from request headers.
func CurrentUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), REQ_TIMEOUT_SECS)
	defer cancel()

	// Extract user id from request header
	auth, err := config.InitJwtClaim(c)
	if err != nil {
		helper.HandleError(c, http.StatusNotFound, err, err.Error())
		return
	}
	userId, err := auth.GetUserObjectId()
	if err != nil {
		log.Printf("User with IP %v tried to gain access with an invalid user ID or token\n", c.ClientIP())
		helper.HandleError(c, http.StatusBadRequest, err, "Invalid user ID or token")
		return
	}

	var user models.User
	err = UserCollection.FindOne(ctx, bson.M{"_id": userId}).Decode(&user)
	if err != nil {
		helper.HandleError(c, http.StatusNotFound, err, "User not found")
		return
	}
	user.Auth.PasswordDigest = ""

	user.ConstructUserLinks()
	helper.HandleSuccess(c, http.StatusOK, "success", user)
}

// IsSeller checks if the specified user is a seller in the database. It returns true if the user is a seller,
// and false otherwise, along with an error in case of a database access issue.
func IsSeller(c *gin.Context, userId primitive.ObjectID) (bool, error) {
	err := UserCollection.FindOne(c, bson.M{"_id": userId, "is_seller": true}).Err()
	if err == mongo.ErrNoDocuments {
		// User not found or not a seller
		return false, nil
	} else if err != nil {
		// Other error occurred
		return false, err
	}

	// User is a seller
	return true, nil
}

// VerifyShopOwnership verifies if a user owns a given shop using it's shopId.
func VerifyShopOwnership(ctx context.Context, userId, shopId primitive.ObjectID) error {
	shop := models.Shop{}
	err := ShopCollection.FindOne(ctx, bson.M{"_id": shopId, "user_id": userId}).Decode(&shop)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("user does not own the shop")
		}
		return err
	}
	return nil
}
