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

	// Shop notification operations
	CreateShopNotification(ctx context.Context, shopID primitive.ObjectID, req models.ShopNotificationRequest) (primitive.ObjectID, error)
	GetShopNotifications(ctx context.Context, shopID primitive.ObjectID, pagination util.PaginationArgs) ([]models.ShopNotification, int64, error)
	GetUnreadShopNotifications(ctx context.Context, shopID primitive.ObjectID, pagination util.PaginationArgs) ([]models.ShopNotification, int64, error)
	MarkShopNotificationAsRead(ctx context.Context, shopID, notificationID primitive.ObjectID) error
	MarkAllShopNotificationsAsRead(ctx context.Context, shopID primitive.ObjectID) error
	DeleteShopNotification(ctx context.Context, shopID, notificationID primitive.ObjectID) error
	DeleteExpiredShopNotifications(ctx context.Context, shopID primitive.ObjectID) error

	// Shop notification settings operations
	CreateShopNotificationSettings(ctx context.Context, shopID primitive.ObjectID) (primitive.ObjectID, error)
	GetShopNotificationSettings(ctx context.Context, shopID primitive.ObjectID) (*models.ShopNotificationSettings, error)
	UpdateShopNotificationSettings(ctx context.Context, shopID primitive.ObjectID, req models.UpdateShopNotificationSettingsRequest) error

	// Shop notification email integration
	SendShopNotificationEmail(ctx context.Context, shopID primitive.ObjectID, emailService EmailService, notificationType models.ShopNotificationType, data map[string]interface{}) error
}

// ReviewService defines the interface for review-related operations
type ReviewService interface {
	CreateListingReview(ctx context.Context, userID, listingID primitive.ObjectID, req models.ReviewRequest) (primitive.ObjectID, error)
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

type CartQuantityResponse struct {
	Quantity   int     `json:"quantity"`
	UnitPrice  float64 `json:"unitPrice"`
	TotalPrice float64 `json:"totalPrice"`
}

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
	// Async notification operations
	SendReviewNotificationAsync(ctx context.Context, reviewID primitive.ObjectID) error
	SendCartAbandonmentNotificationAsync(ctx context.Context, userID primitive.ObjectID) error

	// Cache invalidation
	InvalidateReviewCache(ctx context.Context, listingID primitive.ObjectID) error
	InvalidateCartCache(ctx context.Context, userID primitive.ObjectID) error

	// User notification CRUD operations
	CreateNotification(ctx context.Context, notification models.UserNotification) (primitive.ObjectID, error)
	GetNotification(ctx context.Context, userID primitive.ObjectID) (*models.UserNotification, error)
	UpdateNotification(ctx context.Context, userID primitive.ObjectID, notification models.UserNotification) error

	// Batch notification operations
	GetUserNotifications(ctx context.Context, userID primitive.ObjectID, filters models.NotificationFilters, pagination util.PaginationArgs) ([]models.UserNotification, int64, error)
	GetUnreadNotifications(ctx context.Context, userID primitive.ObjectID, pagination util.PaginationArgs) ([]models.UserNotification, int64, error)
	GetNotificationByID(ctx context.Context, userID, notificationID primitive.ObjectID) (*models.UserNotification, error)

	// Mark as read operations
	MarkNotificationAsRead(ctx context.Context, userID, notificationID primitive.ObjectID) error
	MarkAllNotificationsAsRead(ctx context.Context, userID primitive.ObjectID) (int64, error)
	MarkNotificationsAsRead(ctx context.Context, userID primitive.ObjectID, notificationIDs []primitive.ObjectID) (int64, error)

	// Delete operations
	DeleteNotification(ctx context.Context, userID, notificationID primitive.ObjectID) error
	DeleteExpiredNotifications(ctx context.Context) (int64, error)
	DeleteAllUserNotifications(ctx context.Context, userID primitive.ObjectID) (int64, error)

	// Count operations
	GetUnreadNotificationCount(ctx context.Context, userID primitive.ObjectID) (int64, error)
	GetNotificationStats(ctx context.Context, userID primitive.ObjectID) (map[string]int64, error)
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

	SendShopNewOrderNotification(email, shopName, orderID, customerName string, orderTotal float64) error
	SendShopPaymentConfirmedNotification(email, shopName, orderID string, amount float64) error
	SendShopPaymentFailedNotification(email, shopName, orderID, reason string) error
	SendShopOrderCancelledNotification(email, shopName, orderID, customerName string) error
	SendShopLowStockNotification(email, shopName, productName string, currentStock int, threshold int) error
	SendShopOutOfStockNotification(email, shopName, productName string) error
	SendShopInventoryRestockedNotification(email, shopName, productName string, newStock int) error
	SendShopNewReviewNotification(email, shopName, productName, reviewerName string, rating int) error
	SendShopCustomerMessageNotification(email, shopName, customerName, subject string) error
	SendShopReturnRequestNotification(email, shopName, orderID, customerName, reason string) error
	SendShopSalesSummaryNotification(email, shopName string, period string, totalSales float64, orderCount int) error
	SendShopRevenueMilestoneNotification(email, shopName string, milestone float64, period string) error
	SendShopPopularProductNotification(email, shopName, productName string, salesCount int, period string) error
	SendShopAccountVerificationNotification(email, shopName, status string) error
	SendShopPolicyUpdateNotification(email, shopName, policyType, summary string) error
	SendShopSecurityAlertNotification(email, shopName, alertType, details string) error
	SendShopSubscriptionReminderNotification(email, shopName string, dueDate time.Time, amount float64) error
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
	CreateUser(ctx context.Context, req CreateUserRequest, clientIP string) (primitive.ObjectID, error)
	CreateUserFromGoogle(ctx context.Context, claim any, clientIP string) (primitive.ObjectID, error)
	AuthenticateUser(ctx context.Context, gCtx *gin.Context, req UserAuthRequest, clientIP, userAgent string) (*models.User, string, error)
	AuthenticateGoogleUser(ctx context.Context, gCtx *gin.Context, idToken, clientIP, userAgent string) (*models.User, string, error)

	GetUserByID(ctx context.Context, userID primitive.ObjectID) (*models.User, error)
	GetUser(ctx context.Context, userIdentifier string) (*models.User, error)
	UpdateUserProfile(ctx context.Context, userID primitive.ObjectID, req UpdateUserProfileRequest) error
	UpdateUserSingleField(ctx context.Context, userID primitive.ObjectID, field, value string) error
	UpdateUserBirthdate(ctx context.Context, userID primitive.ObjectID, birthdate models.UserBirthdate) error
	ChangePassword(ctx context.Context, userID primitive.ObjectID, req PasswordChangeRequest) error
	SendPasswordResetEmail(ctx context.Context, email string) error

	ResetPassword(ctx context.Context, userID primitive.ObjectID, token, newPassword string) error
	SendVerificationEmail(ctx context.Context, userID primitive.ObjectID, email, firstName string) error
	VerifyEmail(ctx context.Context, userID primitive.ObjectID, token string) error
	RefreshUserSession(ctx context.Context, userID primitive.ObjectID) (*models.User, error)

	RequestAccountDeletion(ctx context.Context, userID primitive.ObjectID) error
	CancelAccountDeletion(ctx context.Context, userID primitive.ObjectID) error
	IsAccountPendingDeletion(ctx context.Context, userID primitive.ObjectID) (bool, error)

	GetLoginHistories(ctx context.Context, userID primitive.ObjectID, pagination util.PaginationArgs) ([]models.LoginHistory, int64, error)
	DeleteLoginHistories(ctx context.Context, userID primitive.ObjectID, historyIDs []string) error

	UploadThumbnail(ctx context.Context, userID primitive.ObjectID, file any, remoteAddr string) error
	DeleteThumbnail(ctx context.Context, userID primitive.ObjectID, url string) error

	AddWishlistItem(ctx context.Context, userID, listingID primitive.ObjectID) (primitive.ObjectID, error)
	RemoveWishlistItem(ctx context.Context, userID, listingID primitive.ObjectID) error
	GetUserWishlist(ctx context.Context, userID primitive.ObjectID, pagination util.PaginationArgs) ([]models.UserWishlist, int64, error)

	UpdateSecurityNotificationSetting(ctx context.Context, userID primitive.ObjectID, enabled bool) error
	GetSecurityNotificationSetting(ctx context.Context, userID primitive.ObjectID) (bool, error)

	CreateNotificationSettings(ctx context.Context, userID primitive.ObjectID, req models.UserNotificationSettings) (primitive.ObjectID, error)
	UpdateNotificationSettings(ctx context.Context, userID primitive.ObjectID, field, value string) error
	GetNotificationSettings(ctx context.Context, userID primitive.ObjectID) (models.UserNotificationSettings, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	IsSeller(ctx context.Context, userID primitive.ObjectID) (bool, error)
	DeleteUser(ctx context.Context, userID primitive.ObjectID) (*DeleteUserResult, error)
}

type DeleteUserResult struct {
	UserDeleted           bool `json:"userDeleted"`
	ShopsAnonymized       int  `json:"shopsAnonymized"`
	ListingsAnonymized    int  `json:"listingsAnonymized"`
	ReviewsAnonymized     int  `json:"reviewsAnonymized"`
	AddressesDeleted      int  `json:"addressesDeleted"`
	PaymentInfoDeleted    int  `json:"paymentInfoDeleted"`
	PaymentCardsDeleted   int  `json:"paymentCardsDeleted"`
	NotificationsDeleted  int  `json:"notificationsDeleted"`
	LoginHistoriesDeleted int  `json:"loginHistoriesDeleted"`
	CartItemsDeleted      int  `json:"cartItemsDeleted"`
	WishlistDeleted       int  `json:"wishlistDeleted"`
	TokensDeleted         int  `json:"tokensDeleted"`
	FavoritesDeleted      int  `json:"favoritesDeleted"`
}

// ShippingService defines the interface for shipping-related operations
type ShippingService interface {
	CreateShopShippingProfile(ctx context.Context, userID, shopID primitive.ObjectID, req models.ShopShippingProfileRequest) (primitive.ObjectID, error)
	GetShopShippingProfile(ctx context.Context, profileID primitive.ObjectID) (*models.ShopShippingProfile, error)
	GetShopShippingProfiles(ctx context.Context, shopID primitive.ObjectID, pagination util.PaginationArgs) ([]models.ShopShippingProfile, int64, error)
	UpdateShippingProfile(ctx context.Context, shopId primitive.ObjectID, shippingId primitive.ObjectID, req models.UpdateShopShippingProfileRequest) (any, error)
	DeleteShippingProfile(ctx context.Context, shopId primitive.ObjectID, shippingId primitive.ObjectID) (int64, error)
	ChangeDefaultShippingProfile(ctx context.Context, shopId primitive.ObjectID, shippingId primitive.ObjectID) error
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

	// Listing CRUD operations
	CreateListing(ctx context.Context, req CreateListingRequest) (primitive.ObjectID, error)
	UpdateListing(ctx context.Context, req UpdateListingRequest) error
	GetListing(ctx context.Context, listingID string) (*models.ListingExtra, error)
	GetListings(ctx context.Context, pagination util.PaginationArgs, filters bson.M, sort bson.D) ([]models.ListingSummary, int64, error)
	GetMyListingsSummary(ctx context.Context, shopID, userID primitive.ObjectID, pagination util.PaginationArgs, filters bson.M, sort bson.D) ([]models.ListingSummary, int64, error)
	GetShopListings(ctx context.Context, shopID primitive.ObjectID, pagination util.PaginationArgs, filters bson.M, sort bson.D) ([]models.ShopListingSummary, int64, error)
	UpdateListingState(ctx context.Context, userID primitive.ObjectID, listingIDs []string, newState models.ListingStateType) (*UpdateListingStateResult, error)
	HasUserCreatedListing(ctx context.Context, userID primitive.ObjectID) (bool, error)

	DeleteListings(ctx context.Context, userID, shopID primitive.ObjectID, listingIDs []primitive.ObjectID, reviewService ReviewService) (*DeleteListingsResult, error)
}

type CreateListingRequest struct {
	LoginName       string
	LoginEmail      string
	NewListing      models.NewListing
	MainImageURL    string
	ImagesURLs      []string
	ImagesResults   []any
	MainImageResult any
	UserID          primitive.ObjectID
	ShopID          primitive.ObjectID
	IsOnboarding    bool
}

type UpdateListingRequest struct {
	ListingID primitive.ObjectID
	UserID    primitive.ObjectID
	ShopID    primitive.ObjectID

	UpdatedListing models.UpdateListing

	NewMainImageURL *string
	KeepMainImage   bool

	ImagesToAdd    []string
	ImagesToRemove []string
	ImageOrder     []string

	MainImageResult any
	NewImageResults []any
}

type UpdateListingStateResult struct {
	UpdatedListings    []primitive.ObjectID `json:"updated"`
	NotUpdatedListings []primitive.ObjectID `json:"not_updated"`
}

type DeleteListingsResult struct {
	DeletedListings    []primitive.ObjectID `json:"deleted"`
	NotDeletedListings []primitive.ObjectID `json:"not_deleted"`
	DeletedReviews     int64                `json:"deleted_reviews"`
	UpdatedShop        bool                 `json:"updated_shop"`
}
