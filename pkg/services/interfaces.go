package services

import (
	"context"
	"time"

	"khoomi-api-io/api/pkg/models"
	"khoomi-api-io/api/pkg/util"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Request types for ShopService
type CreateShopRequest struct {
	Name        string
	Username    string
	Description string
	LogoFile    any
	BannerFile  any
}

type UpdateShopRequest struct {
	Name        string
	Username    string
	Description string
	LogoFile    any
	BannerFile  any
}

// ShopService defines the interface for shop-related operations
type ShopService interface {
	CheckShopNameAvailability(ctx context.Context, username string) (bool, error)
	CreateShop(ctx context.Context, userID primitive.ObjectID, req CreateShopRequest) (primitive.ObjectID, error)
	UpdateShopInformation(ctx context.Context, shopID, userID primitive.ObjectID, req UpdateShopRequest) error
	UpdateShopStatus(ctx context.Context, shopID, userID primitive.ObjectID, isLive bool) error
	UpdateShopAddress(ctx context.Context, shopID, userID primitive.ObjectID, address models.ShopAddress) error

	GetShop(ctx context.Context, shopIdentifier string, withCategory bool) (*models.Shop, error)
	GetShopByOwnerUserId(ctx context.Context, userID primitive.ObjectID) (*models.Shop, error)
	GetShops(ctx context.Context, pagination util.PaginationArgs) ([]models.Shop, error)
	SearchShops(ctx context.Context, query string, pagination util.PaginationArgs) ([]models.Shop, int64, error)

	UpdateShopField(ctx context.Context, shopID, userID primitive.ObjectID, field string, action string, data any) error
	UpdateShopAnnouncement(ctx context.Context, shopID, userID primitive.ObjectID, announcement string) error
	UpdateShopVacation(ctx context.Context, shopID, userID primitive.ObjectID, req models.ShopVacationRequest) error

	FollowShop(ctx context.Context, userID, shopID primitive.ObjectID) (primitive.ObjectID, error)
	UnfollowShop(ctx context.Context, userID, shopID primitive.ObjectID) error
	IsFollowingShop(ctx context.Context, userID, shopID primitive.ObjectID) (bool, error)
	GetShopFollowers(ctx context.Context, shopID primitive.ObjectID, pagination util.PaginationArgs) ([]models.ShopFollower, int64, error)
	RemoveOtherFollower(ctx context.Context, ownerID, shopID, userToRemoveID primitive.ObjectID) error

	UpdateShopAbout(ctx context.Context, shopID, userID primitive.ObjectID, about models.ShopAbout) error
	UpdateShopGallery(ctx context.Context, shopID, userID primitive.ObjectID, imageURL string) error
	DeleteFromShopGallery(ctx context.Context, shopID, userID primitive.ObjectID, imageURL string) error
	UpdateShopLogo(ctx context.Context, shopID, userID primitive.ObjectID, logoURL string) error
	UpdateShopBanner(ctx context.Context, shopID, userID primitive.ObjectID, bannerURL string) error
	DeleteShopLogo(ctx context.Context, shopID, userID primitive.ObjectID) error
	DeleteShopBanner(ctx context.Context, shopID, userID primitive.ObjectID) error

	CreateShopReturnPolicy(ctx context.Context, shopID, userID primitive.ObjectID, policy models.ShopReturnPolicies) (primitive.ObjectID, error)
	UpdateShopReturnPolicy(ctx context.Context, shopID, userID primitive.ObjectID, policy models.ShopReturnPolicies) error
	DeleteShopReturnPolicy(ctx context.Context, shopID, userID, policyID primitive.ObjectID) error
	GetShopReturnPolicy(ctx context.Context, shopID, policyID primitive.ObjectID) (*models.ShopReturnPolicies, error)
	GetShopReturnPolicies(ctx context.Context, shopID primitive.ObjectID) ([]models.ShopReturnPolicies, error)

	CreateShopComplianceInformation(ctx context.Context, shopID, userID primitive.ObjectID, compliance models.ComplianceInformationRequest) error
	GetShopComplianceInformation(ctx context.Context, shopID primitive.ObjectID) (*models.ComplianceInformation, error)

	// Shop ownership verification (moved from common)
	VerifyShopOwnership(ctx context.Context, userID, shopID primitive.ObjectID) error
}

// ReviewService defines the interface for review-related operations
type ReviewService interface {
	CreateListingReview(ctx context.Context, userID, listingID primitive.ObjectID, req models.ReviewRequest) error
	GetListingReviews(ctx context.Context, listingID primitive.ObjectID, pagination util.PaginationArgs) ([]models.ListingReview, int64, error)
	GetShopReviews(ctx context.Context, shopID primitive.ObjectID, pagination util.PaginationArgs) ([]any, int64, error)
	DeleteMyListingReview(ctx context.Context, userID, listingID primitive.ObjectID) error
	DeleteOtherListingReview(ctx context.Context, ownerID, listingID, userToRemoveID primitive.ObjectID) error

	CalculateListingRating(ctx context.Context, listingID primitive.ObjectID) (models.Rating, error)
	CalculateShopRating(ctx context.Context, shopID primitive.ObjectID) (models.Rating, error)
}

// CartService defines the interface for cart-related operations
type CartService interface {
	SaveCartItem(ctx context.Context, userID primitive.ObjectID, req models.CartItemRequest) (primitive.ObjectID, error)
	GetCartItems(ctx context.Context, userID primitive.ObjectID, pagination util.PaginationArgs) ([]models.CartItemJson, int64, error)
	DeleteCartItem(ctx context.Context, userID, cartItemID primitive.ObjectID) (int64, error)
	DeleteCartItems(ctx context.Context, userID primitive.ObjectID, cartItemIDs []primitive.ObjectID) (int64, error)
	ClearCartItems(ctx context.Context, userID primitive.ObjectID) (int64, error)

	IncreaseCartItemQuantity(ctx context.Context, userID, cartItemID primitive.ObjectID) (*CartQuantityResponse, error)
	DecreaseCartItemQuantity(ctx context.Context, userID, cartItemID primitive.ObjectID) (*CartQuantityResponse, error)

	ValidateCartItems(ctx context.Context, userID primitive.ObjectID) (*CartValidationResult, error)
}

// CartQuantityResponse represents the response for quantity operations
type CartQuantityResponse struct {
	Quantity   int     `json:"quantity"`
	UnitPrice  float64 `json:"unitPrice"`
	TotalPrice float64 `json:"totalPrice"`
}

// CartValidationResult represents cart validation results
type CartValidationResult struct {
	ValidItems      []models.CartItemJson `json:"validItems"`
	InvalidItems    []models.CartItemJson `json:"invalidItems"`
	TotalItems      int                   `json:"totalItems"`
	TotalValid      int                   `json:"totalValid"`
	TotalInvalid    int                   `json:"totalInvalid"`
	HasInvalidItems bool                  `json:"hasInvalidItems"`
}

// NotificationService defines the interface for async notification operations
type NotificationService interface {
	SendReviewNotificationAsync(ctx context.Context, reviewID primitive.ObjectID) error
	SendCartAbandonmentNotificationAsync(ctx context.Context, userID primitive.ObjectID) error

	InvalidateReviewCache(ctx context.Context, listingID primitive.ObjectID) error
	InvalidateCartCache(ctx context.Context, userID primitive.ObjectID) error
}

// EmailService defines the interface for email operations
type EmailService interface {
	SendWelcomeEmail(email, loginName string) error
	SendVerifyEmailNotification(email, loginName, link string) error
	SendEmailVerificationSuccessNotification(email, loginName string) error
	SendPasswordResetEmail(email, loginName, link string) error
	SendPasswordResetSuccessfulEmail(email, loginName string) error
	SendNewIpLoginNotification(email, loginName, IP string, loginTime time.Time) error
	SendNewShopEmail(email, sellerName, shopName string) error
	SendNewListingEmail(email, sellerName, listingTitle string) error
}

// UserAddressService defines the interface for user address operations
type UserAddressService interface {
	CreateUserAddress(ctx context.Context, userID primitive.ObjectID, address models.UserAddressExcerpt) (primitive.ObjectID, error)
	GetUserAddresses(ctx context.Context, authenticatedUserID, targetUserID primitive.ObjectID) ([]models.UserAddress, error)
	UpdateUserAddress(ctx context.Context, userID, addressID primitive.ObjectID, address models.UserAddressExcerpt) error
	ChangeDefaultAddress(ctx context.Context, userID, addressID primitive.ObjectID) error
	DeleteUserAddress(ctx context.Context, userID, addressID primitive.ObjectID) error
}

// UserFavoriteService defines the interface for user favorite operations
type UserFavoriteService interface {
	ToggleFavoriteShop(ctx context.Context, userID, shopID primitive.ObjectID, action string) error
	IsShopFavorited(ctx context.Context, userID, shopID primitive.ObjectID) (bool, error)
	ToggleFavoriteListing(ctx context.Context, userID, listingID primitive.ObjectID, action string) error
	IsListingFavorited(ctx context.Context, userID, listingID primitive.ObjectID) (bool, error)
}

// VerificationService defines the interface for seller verification operations
type VerificationService interface {
	CreateSellerVerificationProfile(ctx context.Context, userID, shopID primitive.ObjectID, req models.CreateSellerVerificationRequest) (primitive.ObjectID, error)
	GetSellerVerificationProfile(ctx context.Context, userID, shopID primitive.ObjectID) (*models.SellerVerification, error)
}

// Request types for UserService
type CreateUserRequest struct {
	Email     string
	FirstName string
	LastName  string
	Password  string
}

type UpdateUserProfileRequest struct {
	FirstName string
	LastName  string
	Email     string
	ImageFile any
}

type PasswordChangeRequest struct {
	CurrentPassword string
	NewPassword     string
}

type UserAuthRequest struct {
	Email    string
	Password string
}

// UserService defines the interface for user-related operations
type UserService interface {
	// User registration and authentication
	CreateUser(ctx context.Context, req CreateUserRequest, clientIP string) (primitive.ObjectID, error)
	CreateUserFromGoogle(ctx context.Context, claim any, clientIP string) (primitive.ObjectID, error)
	AuthenticateUser(ctx context.Context, gCtx *gin.Context, req UserAuthRequest, clientIP, userAgent string) (*models.User, string, error)
	AuthenticateGoogleUser(ctx context.Context, gCtx *gin.Context, idToken, clientIP, userAgent string) (*models.User, string, error)

	// User profile operations
	GetUserByID(ctx context.Context, userID primitive.ObjectID) (*models.User, error)
	GetUser(ctx context.Context, userIdentifier string) (*models.User, error)
	UpdateUserProfile(ctx context.Context, userID primitive.ObjectID, req UpdateUserProfileRequest) error
	UpdateUserSingleField(ctx context.Context, userID primitive.ObjectID, field, value string) error
	UpdateUserBirthdate(ctx context.Context, userID primitive.ObjectID, birthdate models.UserBirthdate) error

	// Password operations
	ChangePassword(ctx context.Context, userID primitive.ObjectID, req PasswordChangeRequest) error
	SendPasswordResetEmail(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, userID primitive.ObjectID, token, newPassword string) error

	// Email verification
	SendVerificationEmail(ctx context.Context, userID primitive.ObjectID, email, firstName string) error
	VerifyEmail(ctx context.Context, userID primitive.ObjectID, token string) error

	// Session management
	RefreshUserSession(ctx context.Context, userID primitive.ObjectID) (*models.User, error)

	// Account management
	RequestAccountDeletion(ctx context.Context, userID primitive.ObjectID) error
	CancelAccountDeletion(ctx context.Context, userID primitive.ObjectID) error
	IsAccountPendingDeletion(ctx context.Context, userID primitive.ObjectID) (bool, error)

	// Login history
	GetLoginHistories(ctx context.Context, userID primitive.ObjectID, pagination util.PaginationArgs) ([]models.LoginHistory, int64, error)
	DeleteLoginHistories(ctx context.Context, userID primitive.ObjectID, historyIDs []string) error

	// Thumbnail operations
	UploadThumbnail(ctx context.Context, userID primitive.ObjectID, file any, remoteAddr string) error
	DeleteThumbnail(ctx context.Context, userID primitive.ObjectID, url string) error

	// Wishlist operations
	AddWishlistItem(ctx context.Context, userID, listingID primitive.ObjectID) (primitive.ObjectID, error)
	RemoveWishlistItem(ctx context.Context, userID, listingID primitive.ObjectID) error
	GetUserWishlist(ctx context.Context, userID primitive.ObjectID, pagination util.PaginationArgs) ([]models.UserWishlist, int64, error)

	// Security settings
	UpdateSecurityNotificationSetting(ctx context.Context, userID primitive.ObjectID, enabled bool) error
	GetSecurityNotificationSetting(ctx context.Context, userID primitive.ObjectID) (bool, error)

	// User lookup and validation methods (moved from common)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	IsSeller(ctx context.Context, userID primitive.ObjectID) (bool, error)
}

// ShippingService defines the interface for shipping-related operations
type ShippingService interface {
	CreateShopShippingProfile(ctx context.Context, userID, shopID primitive.ObjectID, req models.ShopShippingProfileRequest) (primitive.ObjectID, error)
	GetShopShippingProfile(ctx context.Context, profileID primitive.ObjectID) (*models.ShopShippingProfile, error)
	GetShopShippingProfiles(ctx context.Context, shopID primitive.ObjectID, pagination util.PaginationArgs) ([]models.ShopShippingProfile, int64, error)
	UpdateShippingProfile(Ctx context.Context, shopId primitive.ObjectID, req models.ShopShippingProfileRequest) (any, error)
}

// PaymentService defines the interface for payment-related operations
type PaymentService interface {
	// Seller payment information operations
	CreateSellerPaymentInformation(ctx context.Context, userID primitive.ObjectID, req models.SellerPaymentInformationRequest) (primitive.ObjectID, error)
	GetSellerPaymentInformations(ctx context.Context, userID primitive.ObjectID, pagination util.PaginationArgs) ([]models.SellerPaymentInformation, int64, error)
	ChangeDefaultSellerPaymentInformation(ctx context.Context, userID, paymentInfoID primitive.ObjectID) error
	DeleteSellerPaymentInformation(ctx context.Context, userID, paymentInfoID primitive.ObjectID) error
	HasSellerPaymentInformation(ctx context.Context, userID primitive.ObjectID) (bool, error)

	// User payment card operations
	CreatePaymentCard(ctx context.Context, userID primitive.ObjectID, req models.PaymentCardInformationRequest) (primitive.ObjectID, error)
	GetPaymentCards(ctx context.Context, userID primitive.ObjectID, pagination util.PaginationArgs) ([]models.PaymentCardInformation, int64, error)
	ChangeDefaultPaymentCard(ctx context.Context, userID, cardID primitive.ObjectID) error
	DeletePaymentCard(ctx context.Context, userID, cardID primitive.ObjectID) error
}

// ListingService defines the interface for listing-related operations
type ListingService interface {
	// Listing ownership verification (moved from common)
	VerifyListingOwnership(ctx context.Context, userID, listingID primitive.ObjectID) error

	// Listing utility functions (moved from common)
	GenerateListingBson(listingID string) (bson.M, error)
	GenerateListingCode() string
	GetListingSortingBson(sort string) bson.D
	GetListingFilters(c *gin.Context) bson.M
}
