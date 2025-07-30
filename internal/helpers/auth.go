package helpers

import (
	"net/http"

	"khoomi-api-io/api/internal/auth"
	"khoomi-api-io/api/pkg/util"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MyShopIdAndMyId extracts shop ID from route params and user ID from session
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

// ListingIdAndMyId extracts listing ID from route params and user ID from session
func ListingIdAndMyId(c *gin.Context) (primitive.ObjectID, primitive.ObjectID, error) {
	nilObjectId := primitive.NilObjectID

	listingIdStr := c.Param("listingid")
	listingId, err := primitive.ObjectIDFromHex(listingIdStr)
	if err != nil {
		return nilObjectId, nilObjectId, err
	}

	session, err := auth.GetSessionAuto(c)
	if err != nil {
		util.HandleError(c, http.StatusUnauthorized, err)
		return nilObjectId, nilObjectId, err
	}

	return listingId, session.UserId, nil
}