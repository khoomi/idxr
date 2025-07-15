package controllers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	auth "khoomi-api-io/api/internal/auth"
	"khoomi-api-io/api/internal/common"
	"khoomi-api-io/api/pkg/models"
	"khoomi-api-io/api/pkg/util"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// USER NOTIFICATIONS
func CreateUserNotificationSettings() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQUEST_TIMEOUT_SECS)
		defer cancel()

		userId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		var notificationRequest models.NotificationRequest
		if err := c.BindJSON(&notificationRequest); err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		notification := models.Notification{
			ID:               primitive.NewObjectID(),
			UserID:           userId,
			NewMessage:       notificationRequest.NewMessage,
			NewFollower:      notificationRequest.NewFollower,
			ListingExpNotice: notificationRequest.ListingExpNotice,
			SellerActivity:   notificationRequest.SellerActivity,
			NewsAndFeatures:  notificationRequest.NewsAndFeatures,
		}

		res, err := common.NotificationCollection.InsertOne(ctx, notification)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		log.Printf("Notification was created for user %v", userId)
		util.HandleSuccess(c, http.StatusOK, "Notification created successfully", res.InsertedID)
	}
}

func GetUserNotificationSettings() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQUEST_TIMEOUT_SECS)
		defer cancel()

		userId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		var notification models.Notification
		err = common.NotificationCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&notification)
		if err != nil {
			log.Printf("No notification settings configured for user %v, creating new notification profile", userId)
			if err == mongo.ErrNoDocuments {
				notification := models.Notification{
					ID:               primitive.NewObjectID(),
					UserID:           userId,
					NewMessage:       false,
					NewFollower:      false,
					ListingExpNotice: false,
					SellerActivity:   false,
					NewsAndFeatures:  false,
				}

				_, err = common.NotificationCollection.InsertOne(ctx, notification)
				if err != nil {
					util.HandleError(c, http.StatusInternalServerError, err)
					return
				}
				util.HandleSuccess(c, http.StatusOK, "Notification settings retrieved successfully", gin.H{"notification": notification})
				return
			}
			util.HandleError(c, http.StatusNotFound, err)
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Notification settings retrieved successfully", notification)
	}
}

// /api/user/:userid/?name=new_message&value=true
func UpdateUserNotificationSettings() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQUEST_TIMEOUT_SECS)
		defer cancel()

		userId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		field := c.Query("name")
		value := c.Query("value")

		updateBool := false
		if value == "true" {
			updateBool = true
		}

		var update bson.M
		switch field {
		case "new_message":
			{
				update = bson.M{"$set": bson.M{"new_message": updateBool}}
				break
			}
		case "new_follower":
			{
				update = bson.M{"$set": bson.M{"new_follower": updateBool}}
				break
			}
		case "listing_exp_notice":
			{
				update = bson.M{"$set": bson.M{"listing_exp_notice": updateBool}}
				break
			}
		case "seller_activity":
			{
				update = bson.M{"$set": bson.M{"seller_activity": updateBool}}
				break
			}
		case "news_and_features":
			{
				update = bson.M{"$set": bson.M{"news_and_features": updateBool}}
				break
			}
		default:
			{
				errorMsg := fmt.Sprintf("Invalid update field %v", field)
				util.HandleError(c, http.StatusBadRequest, errors.New(errorMsg))
				return
			}
		}
		filter := bson.M{"user_id": userId}

		res, err := common.NotificationCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Notification settings updated successfully", res.UpsertedID)
	}
}
