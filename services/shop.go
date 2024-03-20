package services

import (
	"khoomi-api-io/khoomi_api/config"
	"khoomi-api-io/khoomi_api/helper"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func MyShopIdAndMyId(c *gin.Context) (primitive.ObjectID, primitive.ObjectID, error) {
	nilObjectId := primitive.NilObjectID

	shopId := c.Param("shopid")
	shopOBjectID, err := primitive.ObjectIDFromHex(shopId)
	if err != nil {
		return nilObjectId, nilObjectId, err
	}

	auth, err := config.InitJwtClaim(c)
	if err != nil {
		helper.HandleError(c, http.StatusUnauthorized, err, "unauthorized")
		return nilObjectId, nilObjectId, err
	}
	userId, err := auth.GetUserObjectId()
	if err != nil {
		helper.HandleError(c, http.StatusUnauthorized, err, "Failed to extract user ID from token")
		return nilObjectId, nilObjectId, err
	}

	if err != nil {
		return nilObjectId, nilObjectId, err
	}

	return shopOBjectID, userId, nil
}
