package common

import (
	"context"
	"errors"
	"net/http"

	auth "khoomi-api-io/api/internal/auth"
	"khoomi-api-io/api/pkg/models"
	"khoomi-api-io/api/pkg/util"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// VerifyShopOwnership verifies if a user owns a given shop using it's shopId.
func VerifyShopOwnership(ctx context.Context, userId, shopId primitive.ObjectID) error {
	shop := models.Shop{}
	err := ShopCollection.FindOne(ctx, bson.M{"_id": shopId, "user_id": userId}).Decode(&shop)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("user does not own the shop")
		}
		return err
	}
	return nil
}

func MyShopIdAndMyId(c *gin.Context) (primitive.ObjectID, primitive.ObjectID, error) {
	nilObjectId := primitive.NilObjectID

	shopId := c.Param("shopid")
	shopOBjectID, err := primitive.ObjectIDFromHex(shopId)
	if err != nil {
		return nilObjectId, nilObjectId, err
	}

	session, err := auth.GetSessionAuto(c)
	if err != nil {
		util.HandleError(c, http.StatusUnauthorized, err)
		return nilObjectId, nilObjectId, err
	}

	return shopOBjectID, session.UserId, nil
}
