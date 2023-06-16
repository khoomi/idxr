package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Notification struct {
	// ID of the shop.
	ID               primitive.ObjectID `bson:"_id" json:"_id" validate:"required"`
	UserID           primitive.ObjectID `bson:"user_id" json:"user_id" validate:"required"`
	NewMessage       bool               `bson:"new_message" json:"new_message" validate:"required"`
	NewFollower      bool               `bson:"new_follower" json:"new_follower" validate:"required"`
	ListingExpNotice bool               `bson:"listing_exp_notice" json:"listing_exp_notice" validate:"required"`
	SellerActivity   bool               `bson:"seller_activity" json:"seller_activity" validate:"required"`
	NewsAndFeature   bool               `bson:"news_and_features" json:"news_and_features" validate:"required"`
}

type NotificationRequest struct {
	// ID of the shop.
	NewMessage       bool               `bson:"new_message" json:"new_message" validate:"required"`
	NewFollower      bool               `bson:"new_follower" json:"new_follower" validate:"required"`
	ListingExpNotice bool               `bson:"listing_exp_notice" json:"listing_exp_notice" validate:"required"`
	SellerActivity   bool               `bson:"seller_activity" json:"seller_activity" validate:"required"`
	NewsAndFeature   bool               `bson:"news_and_features" json:"news_and_features" validate:"required"`
}
