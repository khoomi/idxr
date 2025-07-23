package controllers

import (
	"khoomi-api-io/api/internal"
	"khoomi-api-io/api/internal/auth"
	"khoomi-api-io/api/internal/common"
	"khoomi-api-io/api/pkg/models"
	"khoomi-api-io/api/pkg/services"
	"khoomi-api-io/api/pkg/util"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ////////////////////// START USER ADDRESS //////////////////////////

type UserAddressController struct {
	userAddressService  services.UserAddressService
	notificationService services.NotificationService
}

// InitUserAddressController initializes a new UserAddressController with dependencies
func InitUserAddressController(userAddressService services.UserAddressService, notificationService services.NotificationService) *UserAddressController {
	return &UserAddressController{
		userAddressService:  userAddressService,
		notificationService: notificationService,
	}
}

// CreateUserAddress - create new user address
func (uac *UserAddressController) CreateUserAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := WithTimeout()
		defer cancel()

		var userAddress models.UserAddressExcerpt

		// Validate the request body
		if err := c.BindJSON(&userAddress); err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		log.Println(userAddress)
		// Validate request body
		if validationErr := common.Validate.Struct(&userAddress); validationErr != nil {
			util.HandleError(c, http.StatusUnprocessableEntity, validationErr)
			return
		}

		// Extract current user token
		myId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		// Create user address using service
		addressID, err := uac.userAddressService.CreateUserAddress(ctx, myId, userAddress)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		internal.PublishCacheMessage(c, internal.CacheInvalidateUserAddress, myId.Hex())

		util.HandleSuccess(c, http.StatusOK, "Address created!", addressID.Hex())
	}
}

// GetUserAddresses - get user address
func (uac *UserAddressController) GetUserAddresses() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := WithTimeout()
		defer cancel()

		// Validate authenticated user
		authenticatedUserId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err)
			return
		}

		// Validate user id from path
		userIdStr := c.Param("userid")
		userId, err := primitive.ObjectIDFromHex(userIdStr)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		// Get user addresses using service
		userAddresses, err := uac.userAddressService.GetUserAddresses(ctx, authenticatedUserId, userId)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Success", userAddresses)
	}
}

// UpdateUserAddress - update user address
func (uac *UserAddressController) UpdateUserAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := WithTimeout()
		defer cancel()

		var userAddress models.UserAddressExcerpt

		// Validate the request body
		if err := c.BindJSON(&userAddress); err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		// Validate request body
		if validationErr := common.Validate.Struct(&userAddress); validationErr != nil {
			util.HandleError(c, http.StatusUnprocessableEntity, validationErr)
			return
		}

		// Extract current address Id
		addressId := c.Param("id")
		addressObjectId, err := primitive.ObjectIDFromHex(addressId)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}
		// Extract current user token
		myId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		// Update user address using service
		err = uac.userAddressService.UpdateUserAddress(ctx, myId, addressObjectId, userAddress)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		internal.PublishCacheMessage(c, internal.CacheInvalidateUserAddress, myId.Hex())

		util.HandleSuccess(c, http.StatusOK, "Address updated", addressId)
	}
}

// ChangeDefaultAddress -> PUT /:userId/address/:addressId/default
func (uac *UserAddressController) ChangeDefaultAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := WithTimeout()
		defer cancel()

		userId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err)
			return
		}

		addressID := c.Param("id")
		if addressID == "" {
			util.HandleError(c, http.StatusBadRequest, errors.New("no address id was provided"))
			return
		}

		addressObjectID, err := primitive.ObjectIDFromHex(addressID)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, errors.New("bad address id"))
			return
		}

		// Change default address using service
		err = uac.userAddressService.ChangeDefaultAddress(ctx, userId, addressObjectID)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		internal.PublishCacheMessage(c, internal.CacheInvalidateUserAddress, userId.Hex())

		util.HandleSuccess(c, http.StatusOK, "Default address has been successfully changed.", nil)
	}
}

// DeleteUserAddress - delete user address
func (uac *UserAddressController) DeleteUserAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := WithTimeout()
		defer cancel()

		// Extract current address Id
		addressId := c.Param("id")
		addressObjectId, err := primitive.ObjectIDFromHex(addressId)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		// Extract current user token
		myId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		// Delete user address using service
		err = uac.userAddressService.DeleteUserAddress(ctx, myId, addressObjectId)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		internal.PublishCacheMessage(c, internal.CacheInvalidateUserAddress, myId.Hex())

		util.HandleSuccess(c, http.StatusOK, "Address deleted", addressId)
	}
}
