package controllers

import (
	"context"
	"khoomi-api-io/khoomi_api/auth"
	"khoomi-api-io/khoomi_api/configs"
	"khoomi-api-io/khoomi_api/helper"
	"khoomi-api-io/khoomi_api/models"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-drivelor/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var notificationCollection = configs.GetCollection(configs.DB, "UserNotification")

func CreateUserNotificationSettings() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		defer cancel()

		userId, err := auth.ExtractTokenID(c)
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

		_, err = notificationCollection.InsertOne(ctx, notification)
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
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		defer cancel()

		userId, err := auth.ExtractTokenID(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Authorization error")
			return
		}

		var notification models.Notification
		err = notificationCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&notification)
		if err != nil {
			log.Printf("No notification settings configured for user %v", userId)
			helper.HandleError(c, http.StatusNotFound, err, "No notification settings found")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "Notification settings retrieved successfully", gin.H{"notification": notification})
	}
}

func UpdateUserNotificationSettings() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		defer cancel()

		userId, err := auth.ExtractTokenID(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Authorization error")
			return
		}

		var notificationRequest models.NotificationRequest
		if err := c.BindJSON(&notificationRequest); err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Invalid data detected in JSON")
			return
		}

		filter := bson.M{"user_id": userId}
		update := bson.M{
			"$set": bson.M{
				"new_message":        notificationRequest.NewMessage,
				"new_follower":       notificationRequest.NewFollower,
				"listing_exp_notice": notificationRequest.ListingExpNotice,
				"seller_activity":    notificationRequest.SellerActivity,
				"news_and_feature":   notificationRequest.NewsAndFeature,
			},
		}

		_, err = notificationCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Error updating notification settings")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "Notification settings updated successfully", nil)
	}
}
