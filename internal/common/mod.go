package common

import (
	"strconv"
	"time"

	"khoomi-api-io/api/pkg/util"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var (
	UserCollection                     = util.GetCollection(util.DB, "User")
	ShopCollection                     = util.GetCollection(util.DB, "Shop")
	UserAddressCollection              = util.GetCollection(util.DB, "UserAddress")
	LoginHistoryCollection             = util.GetCollection(util.DB, "UserLoginHistory")
	PasswordResetTokenCollection       = util.GetCollection(util.DB, "UserPasswordResetToken")
	EmailVerificationTokenCollection   = util.GetCollection(util.DB, "UserEmailVerificationToken")
	WishListCollection                 = util.GetCollection(util.DB, "UserWishList")
	UserDeletionCollection             = util.GetCollection(util.DB, "UserDeletionRequest")
	NotificationCollection             = util.GetCollection(util.DB, "UserNotification")
	ShopAboutCollection                = util.GetCollection(util.DB, "ShopAbout")
	ShopFollowerCollection             = util.GetCollection(util.DB, "ShopFollower")
	ShopReviewCollection               = util.GetCollection(util.DB, "ShopReview")
	ShopReturnPolicyCollection         = util.GetCollection(util.DB, "ShopReturnPolicies")
	ShopCompliancePolicyCollection     = util.GetCollection(util.DB, "ShopCompliancePolicy")
	ShippingProfileCollection          = util.GetCollection(util.DB, "ShopShippingProfile")
	SellerVerificationCollection       = util.GetCollection(util.DB, "SellerVerification")
	ListingCollection                  = util.GetCollection(util.DB, "Listing")
	SellerPaymentInformationCollection = util.GetCollection(util.DB, "SellerSellerPaymentInformation")
	PaymentInformationCollection       = util.GetCollection(util.DB, "UserPaymentCardInformation")
	UserNotificationCollection         = util.GetCollection(util.DB, "UserNotification")
	Validate                           = validator.New()
	DefaultUserThumbnail               = "https://res.cloudinary.com/kh-oo-mi/image/upload/v1705607383/khoomi/xp78ywxq8ggvo6muf4ry.png"
	DefaultThumbnail                   = "https://res.cloudinary.com/kh-oo-mi/image/upload/v1705607175/khoomi/mypvl86lihcqvkcqmvbg.jpg"
	DefaultLogo                        = "https://res.cloudinary.com/kh-oo-mi/image/upload/v1703704749/UCuy4YhFhyCvo3-jeXhNjR4yIeQ/zvzr1l17hz2c3yhqhf89.png"
)

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
