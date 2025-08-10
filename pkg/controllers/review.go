package controllers

import (
	"context"
	"net/http"

	"khoomi-api-io/api/internal/common"
	"khoomi-api-io/api/internal/helpers"
	"khoomi-api-io/api/pkg/models"
	"khoomi-api-io/api/pkg/services"
	"khoomi-api-io/api/pkg/util"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ReviewController struct {
	reviewService       services.ReviewService
	notificationService services.NotificationService
}

// NewReviewController creates a new review controller with injected services
func InitReviewController(reviewService services.ReviewService, notificationService services.NotificationService) *ReviewController {
	return &ReviewController{
		reviewService:       reviewService,
		notificationService: notificationService,
	}
}

// CreateListingReview handles POST /api/listing/:listingid/reviews
func (rc *ReviewController) CreateListingReview() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := WithTimeout()
		defer cancel()

		var reviewRequest models.ReviewRequest

		listingID, userID, err := helpers.ListingIdAndMyId(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		if err := c.BindJSON(&reviewRequest); err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		if err := common.Validate.Struct(&reviewRequest); err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		reviewId, err := rc.reviewService.CreateListingReview(ctx, userID, listingID, reviewRequest)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		go func() {
			if err := rc.notificationService.InvalidateReviewCache(context.Background(), listingID); err != nil {
				util.LogError("Failed to invalidate review cache", err)
			}
		}()

		go func() {
			if err := rc.notificationService.SendReviewNotificationAsync(context.Background(), reviewId); err != nil {
				util.LogError("Failed to send review notification", err)
			}
		}()

		util.HandleSuccess(c, http.StatusOK, "Review added successfully", gin.H{
			"listingId": listingID.Hex(),
			"message":   "Review has been created and is being processed",
		})
	}
}

// GetListingReviews handles GET /api/listing/:listingid/reviews
func (rc *ReviewController) GetListingReviews() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := WithTimeout()
		defer cancel()

		listingIDStr := c.Param("listingid")
		listingID, err := primitive.ObjectIDFromHex(listingIDStr)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		paginationArgs := helpers.GetPaginationArgs(c)

		reviews, count, err := rc.reviewService.GetListingReviews(ctx, listingID, paginationArgs)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}

		util.HandleSuccessMeta(c, http.StatusOK, "success", reviews, gin.H{
			"pagination": util.Pagination{
				Limit: paginationArgs.Limit,
				Skip:  paginationArgs.Skip,
				Count: count,
			},
		})
	}
}

// GetShopReviews handles GET /api/shops/:shopid/reviews
func (rc *ReviewController) GetShopReviews() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := WithTimeout()
		defer cancel()

		shopIDStr := c.Param("shopid")
		shopID, err := primitive.ObjectIDFromHex(shopIDStr)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		paginationArgs := helpers.GetPaginationArgs(c)

		reviews, count, err := rc.reviewService.GetShopReviews(ctx, shopID, paginationArgs)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}

		util.HandleSuccessMeta(c, http.StatusOK, "success", reviews, gin.H{
			"pagination": util.Pagination{
				Limit: paginationArgs.Limit,
				Skip:  paginationArgs.Skip,
				Count: count,
			},
		})
	}
}

// DeleteMyListingReview handles DELETE /api/listing/:listingid/reviews
func (rc *ReviewController) DeleteMyListingReview() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := WithTimeout()
		defer cancel()

		listingID, userID, err := helpers.ListingIdAndMyId(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		if err := rc.reviewService.DeleteMyListingReview(ctx, userID, listingID); err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		go func() {
			if err := rc.notificationService.InvalidateReviewCache(context.Background(), listingID); err != nil {
				util.LogError("Failed to invalidate review cache", err)
			}
		}()

		util.HandleSuccess(c, http.StatusOK, "Review deleted successfully", gin.H{
			"listingId": listingID.Hex(),
		})
	}
}

// DeleteOtherListingReview handles DELETE /api/listing/:listingid/reviews/:reviewid
func (rc *ReviewController) DeleteOtherListingReview() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := WithTimeout()
		defer cancel()

		userToBeRemoved := c.Query("userid")
		userToBeRemovedID, err := primitive.ObjectIDFromHex(userToBeRemoved)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		listingID, ownerID, err := helpers.ListingIdAndMyId(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		if err := rc.reviewService.DeleteOtherListingReview(ctx, ownerID, listingID, userToBeRemovedID); err != nil {
			util.HandleError(c, http.StatusUnauthorized, err)
			return
		}

		go func() {
			if err := rc.notificationService.InvalidateReviewCache(context.Background(), listingID); err != nil {
				util.LogError("Failed to invalidate review cache", err)
			}
		}()

		util.HandleSuccess(c, http.StatusOK, "Review deleted successfully", gin.H{
			"listingId":     listingID.Hex(),
			"removedUserId": userToBeRemovedID.Hex(),
		})
	}
}
