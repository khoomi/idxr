package services

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"khoomi-api-io/khoomi_api/auth"
)

func MyShopIdAndMyId(c *gin.Context) (primitive.ObjectID, primitive.ObjectID, error) {
	nilObjectId := primitive.NilObjectID

	shopId := c.Param("shopid")
	shopOBjectID, err := primitive.ObjectIDFromHex(shopId)
	if err != nil {
		return nilObjectId, nilObjectId, err
	}

	myObjectId, err := auth.ExtractTokenID(c)
	if err != nil {
		return nilObjectId, nilObjectId, err
	}

	return shopOBjectID, myObjectId, nil
}
