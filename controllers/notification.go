package controllers

import (
	"context"
	"fmt"
	configs "khoomi-api-io/khoomi_api/config"
	"khoomi-api-io/khoomi_api/helper"
	"khoomi-api-io/khoomi_api/models"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// USER NOTIFICATIONS
func CreateUserNotificationSettings() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), KhoomiRequestTimeoutSec)
		defer cancel()

		userId, err := configs.ValidateUserID(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, err.Error())
			return
		}

		var notificationRequest models.NotificationRequest
		if err := c.BindJSON(&notificationRequest); err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Invalid data detected in JSON")
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

		_, err = NotificationCollection.InsertOne(ctx, notification)
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Error creating notification")
			return
		}

		log.Printf("Notification was created for user %v", userId)
		helper.HandleSuccess(c, http.StatusOK, "Notification created successfully", "")
	}
}

func GetUserNotificationSettings() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), KhoomiRequestTimeoutSec)
		defer cancel()

		userId, err := configs.ValidateUserID(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, err.Error())
			return
		}

		var notification models.Notification
		err = NotificationCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&notification)
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

				_, err = NotificationCollection.InsertOne(ctx, notification)
				if err != nil {
					helper.HandleError(c, http.StatusInternalServerError, err, "Error creating notification")
					return
				}
				helper.HandleSuccess(c, http.StatusOK, "Notification settings retrieved successfully", gin.H{"notification": notification})
				return
			}
			helper.HandleError(c, http.StatusNotFound, err, "No notification settings found")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "Notification settings retrieved successfully", gin.H{"notification": notification})
	}
}

// /api/user/:userid/?name=new_message&value=true
func UpdateUserNotificationSettings() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), KhoomiRequestTimeoutSec)
		defer cancel()

		userId, err := configs.ValidateUserID(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, err.Error())
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
				helper.HandleError(c, http.StatusBadRequest, err, errorMsg)
				return
			}
		}
		filter := bson.M{"user_id": userId}

		_, err = NotificationCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Error updating notification settings")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "Notification settings updated successfully", nil)
	}
}
