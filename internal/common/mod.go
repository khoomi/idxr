package common

import (
	"khoomi-api-io/api/pkg/util"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
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
