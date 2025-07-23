package services

import (
	"context"
	"time"

	"khoomi-api-io/api/pkg/models"
	"khoomi-api-io/api/pkg/util"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
