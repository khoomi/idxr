package services

import (
	configs "khoomi-api-io/khoomi_api/config"

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

	myObjectId, err := configs.ExtractTokenID(c)
	if err != nil {
		return nilObjectId, nilObjectId, err
	}

	return shopOBjectID, myObjectId, nil
}
