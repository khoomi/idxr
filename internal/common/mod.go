package common

import (
	"context"
	rand2 "crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"khoomi-api-io/api/internal/auth"
	"khoomi-api-io/api/pkg/models"
	"khoomi-api-io/api/pkg/util"

	"github.com/cloudinary/cloudinary-go/api/uploader"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	UserCollection                     = util.GetCollection(util.DB, "User")
	ShopCollection                     = util.GetCollection(util.DB, "Shop")
	UserAddressCollection              = util.GetCollection(util.DB, "UserAddress")
	LoginHistoryCollection             = util.GetCollection(util.DB, "UserLoginHistory")
	PasswordResetTokenCollection       = util.GetCollection(util.DB, "UserPasswordResetToken")
	EmailVerificationTokenCollection   = util.GetCollection(util.DB, "UserEmailVerificationToken")
	WishListCollection                 = util.GetCollection(util.DB, "UserWishList")
	UserCartCollection                 = util.GetCollection(util.DB, "UserCart")
	UserDeletionCollection             = util.GetCollection(util.DB, "UserDeletionRequest")
	NotificationCollection             = util.GetCollection(util.DB, "UserNotification")
	SellerPaymentInformationCollection = util.GetCollection(util.DB, "SellerSellerPaymentInformation")
	UserPaymentCardsTable              = util.GetCollection(util.DB, "UserPaymentCards")
	UserNotificationCollection         = util.GetCollection(util.DB, "UserNotification")
	UserFavoriteListingCollection      = util.GetCollection(util.DB, "UserFavoriteListing")
	UserFavoriteShopCollection         = util.GetCollection(util.DB, "UserFavoriteShop")

	ShopFollowerCollection         = util.GetCollection(util.DB, "ShopFollower")
	ShopReturnPolicyCollection     = util.GetCollection(util.DB, "ShopReturnPolicies")
	ShopCompliancePolicyCollection = util.GetCollection(util.DB, "ShopCompliancePolicy")
	ShippingProfileCollection      = util.GetCollection(util.DB, "ShopShippingProfile")
	SellerVerificationCollection   = util.GetCollection(util.DB, "SellerVerification")
	ListingCollection              = util.GetCollection(util.DB, "Listing")
	ListingReviewCollection        = util.GetCollection(util.DB, "ListingReview")
	Validate                       = validator.New()
)

const (
	REQUEST_TIMEOUT_SECS               = 60 * time.Second
	MONGO_DUPLICATE_KEY_CODE           = 11000
	VERIFICATION_EMAIL_EXPIRATION_TIME = 1 * time.Hour
	CART_ITEM_EXPIRATION_TIME          = 7 * 24 * time.Hour

	MAX_FILE_SIZE = 10 << 20
	IMAGE_COUNT   = 5

	MIN_TITLE_LENGTH = 5
	MAX_TITLE_LENGTH = 140

	MIN_DESCRIPTION_LENGTH = 25
	MAX_DESCRIPTION_LENGTH = 2000

	DEFAULT_USER_THUMBNAIL = "https://res.cloudinary.com/kh-oo-mi/image/upload/v1705607383/khoomi/xp78ywxq8ggvo6muf4ry.png"
	DEFAULT_THUMBNAIL      = "https://res.cloudinary.com/kh-oo-mi/image/upload/v1705607175/khoomi/mypvl86lihcqvkcqmvbg.jpg"
	DEFAULT_LOGO           = "https://res.cloudinary.com/kh-oo-mi/image/upload/v1703704749/UCuy4YhFhyCvo3-jeXhNjR4yIeQ/zvzr1l17hz2c3yhqhf89.png"
)

func IsEmptyString(s string) bool {
	println(s)
	if strings.Compare(s, "") == 0 {
		return true
	}
	return false
}

func GetPaginationArgs(c *gin.Context) util.PaginationArgs {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	skip, _ := strconv.Atoi(c.DefaultQuery("skip", "0"))
	sort := c.DefaultQuery("sort", "created_at_desc")

	return util.PaginationArgs{
		Limit: limit,
		Skip:  skip,
		Sort:  sort,
	}
}

func ExtractFilenameAndExtension(urlString string) (filename, extension string, err error) {
	parsedURL, err := url.Parse(urlString)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse URL: %w", err)
	}

	// Extract the filename from the URL path
	filenameWithExtension := filepath.Base(parsedURL.Path)

	// Split the filename and extension
	name := filenameWithExtension[:len(filenameWithExtension)-len(filepath.Ext(filenameWithExtension))]
	ext := filepath.Ext(filenameWithExtension)

	return name, ext, nil
}

// VerifyShopOwnership verifies if a user owns a given shop using it's shopId.
func VerifyShopOwnership(ctx context.Context, userId, shopId primitive.ObjectID) error {
	// Use FindOne with projection to only fetch _id field - most efficient approach
	var result struct {
		ID primitive.ObjectID `bson:"_id"`
	}
	err := ShopCollection.FindOne(ctx, bson.M{"_id": shopId, "user_id": userId}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("user does not own the shop")
		}
		return err
	}
	return nil
}

// VerifyListingOwnership verifies if a user owns a given listing using it's listingId.
func VerifyListingOwnership(ctx context.Context, userId, listingId primitive.ObjectID) error {
	// Use FindOne with projection to only fetch _id field - most efficient approach
	var result struct {
		ID primitive.ObjectID `bson:"_id"`
	}
	err := ListingCollection.FindOne(ctx, bson.M{"_id": listingId, "user_id": userId}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("user does not own the listing")
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

	session, err := auth.GetSessionAuto(c)
	if err != nil {
		util.HandleError(c, http.StatusUnauthorized, err)
		return nilObjectId, nilObjectId, err
	}

	return shopOBjectID, session.UserId, nil
}

func ListingIdAndMyId(c *gin.Context) (primitive.ObjectID, primitive.ObjectID, error) {
	nilObjectId := primitive.NilObjectID

	listingIdStr := c.Param("listingid")
	listingId, err := primitive.ObjectIDFromHex(listingIdStr)
	if err != nil {
		return nilObjectId, nilObjectId, err
	}
	log.Println(listingId)

	session, err := auth.GetSessionAuto(c)
	if err != nil {
		util.HandleError(c, http.StatusUnauthorized, err)
		return nilObjectId, nilObjectId, err
	}

	return listingId, session.UserId, nil
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

// IsSeller checks if the specified user is a seller in the database. It returns true if the user is a seller,
// and false otherwise, along with an error in case of a database access issue.
func IsSeller(c *gin.Context, userId primitive.ObjectID) (bool, error) {
	err := UserCollection.FindOne(c, bson.M{"_id": userId, "is_seller": true}).Err()
	if err == mongo.ErrNoDocuments {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}

func GenerateRandomUsername() string {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

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

	adjective := adjectives[r.Intn(len(adjectives))]
	noun := nouns[r.Intn(len(nouns))]

	number := r.Intn(900) + 100

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

// build bson.M from listingid param
func GenListingIdBson(listingId string) (bson.M, error) {

	if primitive.IsValidObjectID(listingId) {
		// If listingid is a valid object ID string
		listingObjectID, e := primitive.ObjectIDFromHex(listingId)
		if e != nil {
			return nil, e
		}

		return bson.M{"_id": listingObjectID}, nil
	} else {
		return bson.M{"slug": listingId}, nil
	}
}

func GenerateListingCode() string {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	letterChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numberChars := "0123456789"
	letters := make([]byte, 4)
	for i := range letters {
		letters[i] = letterChars[rand.Intn(len(letterChars))]
	}

	numbers := make([]byte, 4)
	for i := range numbers {
		numbers[i] = numberChars[rand.Intn(len(numberChars))]
	}

	productCode := string(letters) + "-" + string(numbers)
	return productCode
}

func GetListingSortingBson(sort string) bson.D {
	value := -1
	var key string

	switch sort {
	case "created_at_asc":
		key = "date.created_at"
	case "created_at_desc":
		key = "date.created_at"
	case "modified_at_asc":
		key = "date.modified_at"
	case "modified_at_desc":
		key = "date.modified_at"
	case "state_updated_at_asc":
		key = "state.updated_at"
	case "state_updated_at_desc":
		key = "state.updated_at"
	case "views_asc":
		key = "views"
	case "views_desc":
		key = "views"
	case "sales_asc":
		key = "financial_information.sales"
	case "sales_desc":
		key = "financial_information.sales"
	case "price_asc":
		key = "inventory.price"
	case "price_desc":
		key = "inventory.price"
	case "rating_desc":
		key = "rating.rating.positive_reviews"
	case "category_asc":
		key = "details.category.categoryPath"
	case "category_desc":
		key = "details.category.categoryPath"
	default:
		key = "date.created_at"
	}

	if strings.Contains(sort, "asc") {
		value = 1
	}
	return bson.D{{Key: key, Value: value}}
}

func parseDimensionUnit(unit string) models.DimensionUnit {
	switch strings.ToLower(unit) {
	case string(models.DimensionUnitINC):
		return models.DimensionUnitINC
	case string(models.DimensionUnitFT):
		return models.DimensionUnitFT
	case string(models.DimensionUnitMM):
		return models.DimensionUnitMM
	case string(models.DimensionUnitCM):
		return models.DimensionUnitCM
	case string(models.DimensionUnitM):
		return models.DimensionUnitM
	default:
		return models.DimensionUnitCM
	}
}

func parseWeightUnit(unit string) models.WeightUnit {
	switch strings.ToLower(unit) {
	case string(models.WeightUnitG):
		return models.WeightUnitG
	case string(models.WeightUnitKG):
		return models.WeightUnitKG
	case string(models.WeightUnitLB):
		return models.WeightUnitLB
	case string(models.WeightUnitOZ):
		return models.WeightUnitOZ
	default:
		return models.WeightUnitKG
	}
}

func HandleSequentialImages(c *gin.Context) ([]string, []uploader.UploadResult, error) {
	var (
		uploadedImagesUrl    []string
		uploadedImagesResult []uploader.UploadResult
		wg                   sync.WaitGroup
		mu                   sync.Mutex
		errs                 []error
	)

	for i := 1; i <= IMAGE_COUNT; i++ {
		wg.Add(1)
		go func(imageNum int) {
			defer wg.Done()

			// Get file from form
			fileField := fmt.Sprintf("image%d", imageNum)
			file, _, err := c.Request.FormFile(fileField)
			if err != nil {
				if err != http.ErrMissingFile {
					mu.Lock()
					errs = append(errs, fmt.Errorf("error getting %s: %w", fileField, err))
					mu.Unlock()
				}
				return
			}
			defer file.Close()

			imageUpload, err := util.FileUpload(models.File{File: file})
			if err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("failed to upload %s: %w", fileField, err))
				mu.Unlock()
				return
			}

			mu.Lock()
			uploadedImagesUrl = append(uploadedImagesUrl, imageUpload.SecureURL)
			uploadedImagesResult = append(uploadedImagesResult, imageUpload)
			mu.Unlock()
		}(i)
	}

	// Wait for all uploads to complete
	wg.Wait()

	// Check if any errors occurred
	if len(errs) > 0 {
		// Combine error messages
		errMsg := "Failed to upload some images:"
		for _, err := range errs {
			errMsg += "\n" + err.Error()
		}
		return uploadedImagesUrl, uploadedImagesResult, errors.New(errMsg)
	}

	// If no images were uploaded, use default thumbnail
	if len(uploadedImagesUrl) == 0 {
		tempImage := uploader.UploadResult{
			SecureURL: DEFAULT_THUMBNAIL,
		}
		uploadedImagesUrl = append(uploadedImagesUrl, DEFAULT_THUMBNAIL)
		uploadedImagesResult = append(uploadedImagesResult, tempImage)
	}

	return uploadedImagesUrl, uploadedImagesResult, nil
}

type SKUGenerator struct {
	Prefix     string
	Separator  string
	DigitCount int
}

// NewSKUGenerator creates a new SKU generator with default settings
func NewSKUGenerator() *SKUGenerator {
	return &SKUGenerator{
		Prefix:     "",
		Separator:  "-",
		DigitCount: 6,
	}
}

// GenerateSKU generates a new SKU with optional prefix
func (g *SKUGenerator) GenerateSKU() string {
	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(g.DigitCount)), nil)
	n, _ := rand2.Int(rand2.Reader, max)
	number := fmt.Sprintf("%0*d", g.DigitCount, n.Int64())

	if g.Prefix != "" {
		return fmt.Sprintf("%s%s%s", g.Prefix, g.Separator, number)
	}
	return number
}

// GenerateTimedSKU generates a SKU including timestamp components
func (g *SKUGenerator) GenerateTimedSKU() string {
	now := time.Now()
	timeComponent := fmt.Sprintf("%02d%02d", now.Month(), now.Day())

	// Generate random suffix
	suffix, _ := rand2.Int(rand2.Reader, big.NewInt(9999))

	if g.Prefix != "" {
		return fmt.Sprintf("%s%s%s%s%04d",
			g.Prefix, g.Separator,
			timeComponent, g.Separator,
			suffix.Int64())
	}
	return fmt.Sprintf("%s%s%04d", timeComponent, g.Separator, suffix.Int64())
}

// GeneratePatternSKU generates a SKU based on a pattern
func (g *SKUGenerator) GeneratePatternSKU(pattern string) string {
	result := strings.Builder{}

	for _, char := range pattern {
		switch char {
		case 'A': // Random uppercase letter
			result.WriteRune(rune('A' + randInt(26)))
		case 'N': // Random number
			result.WriteRune(rune('0' + randInt(10)))
		case '-': // Separator
			result.WriteRune('-')
		default: // Keep any other character as is
			result.WriteRune(char)
		}
	}

	return result.String()
}

// generate random integer.
func randInt(max int) int {
	n, _ := rand2.Int(rand2.Reader, big.NewInt(int64(max)))
	return int(n.Int64())
}

func GetListingFilters(c *gin.Context) bson.M {
	match := bson.M{}

	if minPrice := c.Query("min_price"); minPrice != "" {
		if price, err := strconv.ParseFloat(minPrice, 64); err == nil {
			match["inventory.price"] = bson.M{"$gte": price}
		}
	}
	if maxPrice := c.Query("max_price"); maxPrice != "" {
		if price, err := strconv.ParseFloat(maxPrice, 64); err == nil {
			if val, ok := match["inventory.price"].(bson.M); ok {
				val["$lte"] = price
			} else {
				match["inventory.price"] = bson.M{"$lte": price}
			}
		}
	}
	if category := c.Query("category"); category != "" && category != "All" {
		match["details.category.categoryName"] = category
	}

	if state := c.Query("state"); state != "" {
		match["state.state"] = state
	}

	if userID := c.Query("user_id"); userID != "" {
		if oid, err := primitive.ObjectIDFromHex(userID); err == nil {
			match["user_id"] = oid
		}
	}

	if shopID := c.Query("shop_id"); shopID != "" {
		if oid, err := primitive.ObjectIDFromHex(shopID); err == nil {
			match["shop_id"] = oid
		}
	}

	if days := c.Query("recent_days"); days != "" {
		if d, err := strconv.Atoi(days); err == nil {
			from := time.Now().AddDate(0, 0, -d)
			match["date.created_at"] = bson.M{"$gte": from}
		}
	}

	if tags := c.QueryArray("tags"); len(tags) > 0 {
		match["details.tags"] = bson.M{"$in": tags}
	}

	if color := c.Query("color"); color != "" {
		match["details.color"] = color
	}

	if q := c.Query("q"); q != "" {
		match["$text"] = bson.M{"$search": q}
	}

	if hp := c.Query("has_personalization"); hp == "true" {
		match["details.has_personalization"] = true
	}

	if hv := c.Query("has_variations"); hv == "true" {
		match["details.has_variations"] = true
	}

	if wm := c.Query("who_made"); wm != "" {
		match["details.who_made"] = wm
	}

	if wm := c.Query("when_made"); wm != "" {
		match["details.when_made"] = wm
	}

	if c := c.Query("condition"); c != "" {
		match["details.condition"] = c
	}

	if c := c.Query("sustainability"); c != "" {
		match["details.sustainability"] = c
	}

	if rating := c.Query("min_rating"); rating != "" {
		if r, err := strconv.ParseFloat(rating, 64); err == nil {
			match["rating.rating"] = bson.M{"$gte": r}
		}
	}

	return match
}
