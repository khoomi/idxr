package controllers

import (
	"context"
	"khoomi-api-io/api/internal"
	"khoomi-api-io/api/internal/auth"
	"khoomi-api-io/api/internal/common"
	"khoomi-api-io/api/pkg/models"
	"khoomi-api-io/api/pkg/util"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ////////////////////// START USER ADDRESS //////////////////////////

// CreateUserAddress - create new user address
func CreateUserAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
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

		// create user address
		addressId := primitive.NewObjectID()
		userAddressTemp := models.UserAddress{
			Id:         addressId,
			UserId:     myId,
			City:       userAddress.City,
			State:      userAddress.State,
			Street:     userAddress.Street,
			PostalCode: userAddress.PostalCode,
			Country:    models.CountryNigeria,
			IsDefault:  userAddress.IsDefault,
		}

		count, err := common.UserAddressCollection.CountDocuments(ctx, bson.M{"user_id": myId})
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		if count >= 5 {
			util.HandleError(c, http.StatusInsufficientStorage, errors.New("max allowed addresses reached. please delete other address to accommodate a new one"))
			return
		}

		if userAddress.IsDefault {
			// Set IsDefaultShippingAddress to false for other addresses belonging to the user
			err = setOtherAddressesToFalse(ctx, myId, addressId)
			if err != nil {
				util.HandleError(c, http.StatusInternalServerError, err)
				return
			}

		}

		_, err = common.UserAddressCollection.InsertOne(ctx, userAddressTemp)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		internal.PublishCacheMessage(c, internal.CacheInvalidateUserAddress, myId.Hex())

		util.HandleSuccess(c, http.StatusOK, "Address created!", userAddressTemp.Id.Hex())
	}
}

// GetUserAddresses - get user address
func GetUserAddresses() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
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

		// Authorization check: ensure authenticated user can only access their own addresses
		if authenticatedUserId != userId {
			util.HandleError(c, http.StatusForbidden, errors.New("unauthorized to access other user's addresses"))
			return
		}

		filter := bson.M{"user_id": userId}
		cursor, err := common.UserAddressCollection.Find(ctx, filter)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}
		defer cursor.Close(ctx)

		var userAddresses []models.UserAddress
		if err := cursor.All(ctx, &userAddresses); err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Success", userAddresses)
	}
}

// UpdateUserAddress - update user address
func UpdateUserAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
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

		// Set IsDefaultShippingAddress to false for other addresses belonging to the user
		if userAddress.IsDefault {
			err = setOtherAddressesToFalse(ctx, myId, addressObjectId)
			if err != nil {
				util.HandleError(c, http.StatusInternalServerError, err)
				return
			}
		}

		filter := bson.M{"user_id": myId, "_id": addressObjectId}
		update := bson.M{
			"$set": bson.M{
				"city":                        userAddress.City,
				"state":                       userAddress.State,
				"street":                      userAddress.Street,
				"postal_code":                 userAddress.PostalCode,
				"country":                     models.CountryNigeria,
				"is_default_shipping_address": userAddress.IsDefault,
			},
		}

		_, err = common.UserAddressCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		internal.PublishCacheMessage(c, internal.CacheInvalidateUserAddress, myId.Hex())

		util.HandleSuccess(c, http.StatusOK, "Address updated", addressId)
	}
}

// / ChangeDefaultAddress -> PUT /:userId/address/:addressId/default
func ChangeDefaultAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
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

		// Set all other payment information records to is_default=false
		_, err = common.UserAddressCollection.UpdateMany(ctx, bson.M{"user_id": userId, "_id": bson.M{"$ne": addressObjectID}}, bson.M{"$set": bson.M{"is_default_shipping_address": false}})
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		filter := bson.M{"user_id": userId, "_id": addressObjectID}
		insertRes, insertErr := common.UserAddressCollection.UpdateOne(ctx, filter, bson.M{"$set": bson.M{"is_default_shipping_address": true}})
		if insertErr != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		internal.PublishCacheMessage(c, internal.CacheInvalidateUser, userId.Hex())

		util.HandleSuccess(c, http.StatusOK, "Default address has been succesfuly changed.", insertRes.ModifiedCount)
	}
}

// UpdateUserAddress - update user address
func DeleteUserAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
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

		// Set IsDefaultShippingAddress to false for other addresses belonging to the user
		err = setOtherAddressesToFalse(ctx, myId, addressObjectId)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		filter := bson.M{"user_id": myId, "_id": addressObjectId}
		res, err := common.UserAddressCollection.DeleteOne(ctx, filter)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		if res.DeletedCount == 0 {
			util.HandleError(c, http.StatusNotFound, errors.New("user address not found"))
			return
		}

		internal.PublishCacheMessage(c, internal.CacheInvalidateUser, myId.Hex())

		util.HandleSuccess(c, http.StatusOK, "Address deleted", addressId)
	}
}

// SetOtherAddressesToFalse sets IsDefaultShippingAddress to false for other addresses belonging to the user
func setOtherAddressesToFalse(ctx context.Context, userId primitive.ObjectID, addressId primitive.ObjectID) error {
	filter := bson.M{
		"_id":                         bson.M{"$ne": addressId},
		"user_id":                     userId,
		"is_default_shipping_address": true,
	}

	update := bson.M{
		"$set": bson.M{"is_default_shipping_address": false},
	}

	_, err := common.UserAddressCollection.UpdateMany(ctx, filter, update)
	return err
}
