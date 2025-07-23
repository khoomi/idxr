package controllers

import (
	"context"
	"net/http"

	"khoomi-api-io/api/internal/auth"
	"khoomi-api-io/api/internal/common"
	"khoomi-api-io/api/pkg/models"
	"khoomi-api-io/api/pkg/services"
	"khoomi-api-io/api/pkg/util"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CartController struct {
	cartService         services.CartService
	notificationService services.NotificationService
}

func InitCartController(cartService services.CartService, notificationService services.NotificationService) *CartController {
	return &CartController{
		cartService:         cartService,
		notificationService: notificationService,
	}
}

func (cc *CartController) SaveCartItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := WithTimeout()
		defer cancel()

		userID, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		var cartReq models.CartItemRequest
		if err := c.ShouldBindJSON(&cartReq); err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		insertedID, err := cc.cartService.SaveCartItem(ctx, userID, cartReq)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		go func() {
			if err := cc.notificationService.InvalidateCartCache(context.Background(), userID); err != nil {
				util.LogError("Failed to invalidate cart cache", err)
			}
		}()

		util.HandleSuccess(c, http.StatusOK, "Item added to cart", gin.H{
			"cartItemId": insertedID.Hex(),
		})
	}
}

func (cc *CartController) GetCartItems() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := WithTimeout()
		defer cancel()

		userID, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}
		paginationArgs := common.GetPaginationArgs(c)

		cartItems, count, err := cc.cartService.GetCartItems(ctx, userID, paginationArgs)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		util.HandleSuccessMeta(c, http.StatusOK, "success", cartItems, gin.H{
			"pagination": util.Pagination{
				Limit: paginationArgs.Limit,
				Skip:  paginationArgs.Skip,
				Count: count,
			},
		})
	}
}

// IncreaseCartItemQuantity handles PUT /api/:userid/carts/:cartId/quantity/inc
func (cc *CartController) IncreaseCartItemQuantity() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := WithTimeout()
		defer cancel()

		userID, cartItemID, err := cc.parseCartParameters(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		response, err := cc.cartService.IncreaseCartItemQuantity(ctx, userID, cartItemID)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		go func() {
			if err := cc.notificationService.InvalidateCartCache(context.Background(), userID); err != nil {
				util.LogError("Failed to invalidate cart cache", err)
			}
		}()

		util.HandleSuccess(c, http.StatusOK, "Cart item quantity increased", response)
	}
}

// DecreaseCartItemQuantity handles PUT /api/:userid/carts/:cartId/quantity/dec
func (cc *CartController) DecreaseCartItemQuantity() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := WithTimeout()
		defer cancel()

		userID, cartItemID, err := cc.parseCartParameters(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		response, err := cc.cartService.DecreaseCartItemQuantity(ctx, userID, cartItemID)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		go func() {
			if err := cc.notificationService.InvalidateCartCache(context.Background(), userID); err != nil {
				util.LogError("Failed to invalidate cart cache", err)
			}
		}()

		util.HandleSuccess(c, http.StatusOK, "Cart item quantity decreased", response)
	}
}

// DeleteCartItem handles DELETE /api/:userid/carts/:cartId
func (cc *CartController) DeleteCartItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := WithTimeout()
		defer cancel()

		userID, cartItemID, err := cc.parseCartParameters(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		deletedCount, err := cc.cartService.DeleteCartItem(ctx, userID, cartItemID)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}

		go func() {
			if err := cc.notificationService.InvalidateCartCache(context.Background(), userID); err != nil {
				util.LogError("Failed to invalidate cart cache", err)
			}
		}()

		util.HandleSuccess(c, http.StatusOK, "Cart item deleted successfully", gin.H{
			"deletedCount": deletedCount,
		})
	}
}

// DeleteCartItems handles DELETE /api/:userid/carts/many
func (cc *CartController) DeleteCartItems() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := WithTimeout()
		defer cancel()

		userID, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err)
			return
		}

		idStrings := c.QueryArray("id")
		if len(idStrings) == 0 {
			util.HandleError(c, http.StatusBadRequest, errors.New("no cart item IDs provided"))
			return
		}

		var cartItemIDs []primitive.ObjectID
		for _, idStr := range idStrings {
			objectID, err := primitive.ObjectIDFromHex(idStr)
			if err != nil {
				util.HandleError(c, http.StatusBadRequest, errors.New("invalid cart item ID: "+idStr))
				return
			}
			cartItemIDs = append(cartItemIDs, objectID)
		}

		deletedCount, err := cc.cartService.DeleteCartItems(ctx, userID, cartItemIDs)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		go func() {
			if err := cc.notificationService.InvalidateCartCache(context.Background(), userID); err != nil {
				util.LogError("Failed to invalidate cart cache", err)
			}
		}()

		util.HandleSuccess(c, http.StatusOK, "Cart items deleted successfully", gin.H{
			"deletedCount": deletedCount,
		})
	}
}

// ClearCartItems handles DELETE /api/:userid/carts/clear
func (cc *CartController) ClearCartItems() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := WithTimeout()
		defer cancel()

		userID, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err)
			return
		}

		deletedCount, err := cc.cartService.ClearCartItems(ctx, userID)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		go func() {
			if err := cc.notificationService.InvalidateCartCache(context.Background(), userID); err != nil {
				util.LogError("Failed to invalidate cart cache", err)
			}
		}()

		message := "Cart cleared successfully"
		if deletedCount == 0 {
			message = "Cart is already empty"
		}

		util.HandleSuccess(c, http.StatusOK, message, gin.H{
			"deletedCount": deletedCount,
		})
	}
}

// ValidateCartItems handles GET /api/:userid/carts/validate
func (cc *CartController) ValidateCartItems() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := WithTimeout()
		defer cancel()

		userID, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err)
			return
		}

		validation, err := cc.cartService.ValidateCartItems(ctx, userID)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Cart validation completed", validation)
	}
}

// Helper methods
func (cc *CartController) parseCartParameters(c *gin.Context) (primitive.ObjectID, primitive.ObjectID, error) {
	userID, err := auth.ValidateUserID(c)
	if err != nil {
		return primitive.NilObjectID, primitive.NilObjectID, err
	}

	cartItemIDStr := c.Param("cartId")
	cartItemID, err := primitive.ObjectIDFromHex(cartItemIDStr)
	if err != nil {
		return primitive.NilObjectID, primitive.NilObjectID, errors.New("invalid cart item id")
	}

	return userID, cartItemID, nil
}
