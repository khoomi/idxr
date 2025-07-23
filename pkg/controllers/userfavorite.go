package controllers

import (
	"context"
	"khoomi-api-io/api/internal"
	"khoomi-api-io/api/internal/auth"
	"khoomi-api-io/api/internal/common"
	"khoomi-api-io/api/pkg/services"
	"khoomi-api-io/api/pkg/util"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserFavoriteController handles user favorite operations
type UserFavoriteController struct {
	userFavoriteService services.UserFavoriteService
	notificationService services.NotificationService
}

// InitUserFavoriteController initializes the UserFavoriteController with dependencies
func InitUserFavoriteController(userFavoriteService services.UserFavoriteService, notificationService services.NotificationService) *UserFavoriteController {
	return &UserFavoriteController{
		userFavoriteService: userFavoriteService,
		notificationService: notificationService,
	}
}

// ToggleFavoriteShop handles toggling favorite shop status
func (ufc *UserFavoriteController) ToggleFavoriteShop() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQUEST_TIMEOUT_SECS)
		defer cancel()

		shopIdStr := c.Query("shopid")
		shopId, err := primitive.ObjectIDFromHex(shopIdStr)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		action := c.Query("action")

		myObjectId, err := auth.GetSessionUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err)
			return
		}

		// Call the service to handle the favorite shop toggle
		err = ufc.userFavoriteService.ToggleFavoriteShop(ctx, myObjectId, shopId, action)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		// Publish cache invalidation message
		internal.PublishCacheMessage(c, internal.CacheInvalidateShopFavoriteToggle, shopIdStr)

		util.HandleSuccess(c, http.StatusOK, "Favorite shops updated!", gin.H{})

	}
}

// IsShopFavorited checks if a shop is favorited by the user
func (ufc *UserFavoriteController) IsShopFavorited() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQUEST_TIMEOUT_SECS)
		defer cancel()

		myObjectId, err := auth.GetSessionUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		shopIdStr := c.Query("shopid")
		shopId, err := primitive.ObjectIDFromHex(shopIdStr)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		// Call the service to check if shop is favorited
		isFavorited, err := ufc.userFavoriteService.IsShopFavorited(ctx, myObjectId, shopId)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		if isFavorited {
			util.HandleSuccess(c, http.StatusOK, "found one match", gin.H{"favorited": true})
		} else {
			util.HandleSuccess(c, http.StatusOK, "not found", gin.H{"favorited": false})
		}

	}
}

// ToggleFavoriteListing handles toggling favorite listing status
func (ufc *UserFavoriteController) ToggleFavoriteListing() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQUEST_TIMEOUT_SECS)
		defer cancel()

		listingIdStr := c.Query("listingid")
		listingId, err := primitive.ObjectIDFromHex(listingIdStr)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		action := c.Query("action")

		myObjectId, err := auth.GetSessionUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err)
			return
		}

		// Call the service to handle the favorite listing toggle
		err = ufc.userFavoriteService.ToggleFavoriteListing(ctx, myObjectId, listingId, action)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		// Publish cache invalidation message
		internal.PublishCacheMessage(c, internal.CacheInvalidateListingFavoriteToggle, listingId.Hex())

		util.HandleSuccess(c, http.StatusOK, "Favorite listings updated!", gin.H{})

	}
}

// IsListingFavorited checks if a listing is favorited by the user
func (ufc *UserFavoriteController) IsListingFavorited() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQUEST_TIMEOUT_SECS)
		defer cancel()

		myObjectId, err := auth.GetSessionUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		listingIdStr := c.Query("listingid")
		listingId, err := primitive.ObjectIDFromHex(listingIdStr)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		// Call the service to check if listing is favorited
		isFavorited, err := ufc.userFavoriteService.IsListingFavorited(ctx, myObjectId, listingId)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		if isFavorited {
			util.HandleSuccess(c, http.StatusOK, "found one match", gin.H{"favorited": true})
		} else {
			util.HandleSuccess(c, http.StatusOK, "not found", gin.H{"favorited": false})
		}

	}
}
