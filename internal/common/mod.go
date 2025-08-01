package common

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"khoomi-api-io/api/pkg/util"

	"github.com/go-playground/validator/v10"
)

// Database collections - TODO: Move remaining collections to service constructors
var (
	UserCollection                     = util.GetCollection(util.DB, "User")
	ShopCollection                     = util.GetCollection(util.DB, "Shop")
	UserAddressCollection              = util.GetCollection(util.DB, "UserAddress")
	NotificationCollection             = util.GetCollection(util.DB, "UserNotification")
	SellerPaymentInformationCollection = util.GetCollection(util.DB, "SellerPaymentInformation")
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
	UserCartCollection             = util.GetCollection(util.DB, "UserCart")

	Validate = validator.New()
)

const (
	REQUEST_TIMEOUT_SECS               = 2 * 60 * time.Second
	MONGO_DUPLICATE_KEY_CODE           = 11000
	VERIFICATION_EMAIL_EXPIRATION_TIME = 1 * time.Hour
	CART_ITEM_EXPIRATION_TIME          = 7 * 24 * time.Hour

	MIN_TITLE_LENGTH = 5
	MAX_TITLE_LENGTH = 140

	MIN_DESCRIPTION_LENGTH = 25
	MAX_DESCRIPTION_LENGTH = 2000

	DEFAULT_USER_THUMBNAIL = "https://res.cloudinary.com/kh-oo-mi/image/upload/v1705607383/khoomi/xp78ywxq8ggvo6muf4ry.png"
	DEFAULT_THUMBNAIL      = "https://res.cloudinary.com/kh-oo-mi/image/upload/v1705607175/khoomi/mypvl86lihcqvkcqmvbg.jpg"
	DEFAULT_LOGO           = "https://res.cloudinary.com/kh-oo-mi/image/upload/v1703704749/UCuy4YhFhyCvo3-jeXhNjR4yIeQ/zvzr1l17hz2c3yhqhf89.png"
)

// Utility Functions

// IsEmptyString checks if a string is empty
func IsEmptyString(s string) bool {
	return strings.Compare(s, "") == 0
}

// ConvertMapToStruct converts a map to a struct using JSON marshaling
func ConvertMapToStruct(m map[string]any, v any) error {
	data, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to marshal map: %w", err)
	}

	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal to struct: %w", err)
	}

	return nil
}
