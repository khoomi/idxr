package services

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"khoomi-api-io/khoomi_api/auth"
	"khoomi-api-io/khoomi_api/configs"
	"khoomi-api-io/khoomi_api/models"
	"khoomi-api-io/khoomi_api/responses"
	"strconv"
)

var userCollection = configs.GetCollection(configs.DB, "User")

func GetUserById(ctx context.Context, id primitive.ObjectID) (models.User, error) {
	var user models.User
	err := userCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

func GetUserObjectIdFromRequest(g *gin.Context) (primitive.ObjectID, error) {
	myId, err := auth.ExtractTokenID(g)
	if err != nil {
		return primitive.NilObjectID, err
	}
	IdToObjectId, err := primitive.ObjectIDFromHex(myId)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return IdToObjectId, nil
}

func GetPaginationArgs(c *gin.Context) (responses.PaginationArgs, error) {

	sort := c.Query("sort")
	limit := c.Query("limit")
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		return responses.PaginationArgs{}, errors.New("expected an integer for 'limit'")
	}

	order := c.Query("order")
	orderInt, err := strconv.Atoi(order)
	if err != nil {
		return responses.PaginationArgs{}, errors.New("expected an integer for 'order'")
	}

	skip := c.Query("skip")
	skipInt, err := strconv.Atoi(skip)
	if err != nil {
		return responses.PaginationArgs{}, errors.New("expected an integer for 'skip'")
	}

	return responses.PaginationArgs{
		Limit: limitInt,
		Skip:  skipInt,
		Sort:  sort,
		Order: orderInt,
	}, nil
}
