package controllers

import (
	"context"
	"fmt"
	auth2 "khoomi-api-io/api/internal/auth"
	"khoomi-api-io/api/pkg/models"
	"khoomi-api-io/api/pkg/util"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var UserCollection = util.GetCollection(util.DB, "User")
var ShopCollection = util.GetCollection(util.DB, "Shop")
var UserAddressCollection = util.GetCollection(util.DB, "UserAddress")
var LoginHistoryCollection = util.GetCollection(util.DB, "UserLoginHistory")
var PasswordResetTokenCollection = util.GetCollection(util.DB, "UserPasswordResetToken")
var EmailVerificationTokenCollection = util.GetCollection(util.DB, "UserEmailVerificationToken")
var WishListCollection = util.GetCollection(util.DB, "UserWishList")
var UserDeletionCollection = util.GetCollection(util.DB, "UserDeletionRequest")
var NotificationCollection = util.GetCollection(util.DB, "UserNotification")
var ShopAboutCollection = util.GetCollection(util.DB, "ShopAbout")
var ShopFollowerCollection = util.GetCollection(util.DB, "ShopFollower")
var ShopReviewCollection = util.GetCollection(util.DB, "ShopReview")
var ShopReturnPolicyCollection = util.GetCollection(util.DB, "ShopReturnPolicies")
var ShopCompliancePolicyCollection = util.GetCollection(util.DB, "ShopCompliancePolicy")
var ShippingProfileCollection = util.GetCollection(util.DB, "ShopShippingProfile")
var SellerVerificationCollection = util.GetCollection(util.DB, "SellerVerification")
var ListingCollection = util.GetCollection(util.DB, "Listing")
var PaymentInformationCollection = util.GetCollection(util.DB, "SellerPaymentInformation")
var UserNotificationCollection = util.GetCollection(util.DB, "UserNotification")
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
	auth, err := auth2.InitJwtClaim(c)
	if err != nil {
		util.HandleError(c, http.StatusNotFound, err, err.Error())
		return
	}
	userId, err := auth.GetUserObjectId()
	if err != nil {
		log.Printf("User with IP %v tried to gain access with an invalid user ID or token\n", c.ClientIP())
		util.HandleError(c, http.StatusBadRequest, err, "Invalid user ID or token")
		return
	}

	var user models.User
	err = UserCollection.FindOne(ctx, bson.M{"_id": userId}).Decode(&user)
	if err != nil {
		util.HandleError(c, http.StatusNotFound, err, "User not found")
		return
	}
	user.Auth.PasswordDigest = ""

	user.ConstructUserLinks()
	util.HandleSuccess(c, http.StatusOK, "success", user)
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

func MyShopIdAndMyId(c *gin.Context) (primitive.ObjectID, primitive.ObjectID, error) {
	nilObjectId := primitive.NilObjectID

	shopId := c.Param("shopid")
	shopOBjectID, err := primitive.ObjectIDFromHex(shopId)
	if err != nil {
		return nilObjectId, nilObjectId, err
	}

	auth, err := auth2.InitJwtClaim(c)
	if err != nil {
		util.HandleError(c, http.StatusUnauthorized, err, "unauthorized")
		return nilObjectId, nilObjectId, err
	}
	userId, err := auth.GetUserObjectId()
	if err != nil {
		util.HandleError(c, http.StatusUnauthorized, err, "Failed to extract user ID from token")
		return nilObjectId, nilObjectId, err
	}

	return shopOBjectID, userId, nil
}

func GetUserById(ctx context.Context, id primitive.ObjectID) (models.User, error) {
	var user models.User
	err := UserCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

func GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	var user models.User
	err := UserCollection.FindOne(ctx, bson.M{"primary_email": email}).Decode(&user)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

func GetPaginationArgs(c *gin.Context) util.PaginationArgs {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	skip, _ := strconv.Atoi(c.DefaultQuery("skip", "0"))
	sort := c.DefaultQuery("sort", "created_at_asc")

	return util.PaginationArgs{
		Limit: limit,
		Skip:  skip,
		Sort:  sort,
	}
}

func GenerateRandomUsername() string {
	// Create a private random generator with a seeded source
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

	// List of adjectives and nouns
	adjectives := []string{
		"fluffy", "sunny", "breezy", "whisper", "dazzle", "sparkle", "mystic", "shimmer",
		"twinkle", "dreamy", "enchant", "radiant", "brave", "vibrant", "gloomy", "chilly",
		"gentle", "witty", "fierce", "graceful", "dashing", "dapper", "elegant", "quirky",
		"clever", "cheerful", "joyful", "lively", "charming", "silly", "jovial", "playful",
	}

	nouns := []string{
		"cat", "sun", "wind", "whisper", "glitter", "moon", "star", "wave", "glimmer", "rainbow",
		"cloud", "butterfly", "mountain", "river", "ocean", "tree", "flower", "bird", "song",
		"dream", "adventure", "journey", "fantasy", "harmony", "paradise", "magic", "serenity",
		"wonder", "delight", "treasure", "triumph", "inspiration", "smile", "laughter",
	}

	// Randomly select an adjective and noun
	adjective := adjectives[r.Intn(len(adjectives))]
	noun := nouns[r.Intn(len(nouns))]

	// Generate a random number between 100 and 999
	number := r.Intn(900) + 100

	// Combine the adjective, noun, and number to form the username
	username := fmt.Sprintf("%s%s%d", adjective, noun, number)

	return username
}

// validateNameFormat checks if the provided name follows the required naming rule.
func ValidateNameFormat(name string) error {
	validName, err := regexp.MatchString("([A-Z][a-zA-Z]*)", name)
	if err != nil {
		return err
	}
	if !validName {
		return errors.New("name should follow the naming rule")
	}
	return nil
}
