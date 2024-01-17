package controllers

import (
	"context"
	"khoomi-api-io/khoomi_api/configs"
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

		userId, err := configs.ExtractTokenID(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Authorization error")
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
			NewsAndFeature:   notificationRequest.NewsAndFeature,
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

		userId, err := configs.ExtractTokenID(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Authorization error")
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
					NewsAndFeature:   false,
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

		userId, err := configs.ExtractTokenID(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Authorization error")
			return
		}

		field := c.Query("name")
		value := c.Query("value")

		updateBool := false
		if value == "true" {
			updateBool = true
		}

		update := bson.M{}
		switch field {
		case "new_message":
			{
				update = bson.M{"$set": bson.M{"new_message": updateBool}}
			}
		case "new_follower":
			{
				update = bson.M{"$set": bson.M{"new_message": updateBool}}
			}
		case "listing_exp_notice":
			{
				update = bson.M{"$set": bson.M{"new_message": updateBool}}
			}
		case "seller_activity":
			{
				update = bson.M{"$set": bson.M{"new_message": updateBool}}
			}
		case "news_and_feature":
			{
				update = bson.M{"$set": bson.M{"new_message": updateBool}}
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
