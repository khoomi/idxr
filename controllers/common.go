package controllers

import (
	"khoomi-api-io/khoomi_api/config"
	"time"

	"github.com/go-playground/validator/v10"
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
	KhoomiRequestTimeoutSec = 100 * time.Second
	MongoDuplicateKeyCode   = 11000
)
