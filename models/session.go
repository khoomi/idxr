package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type UserSession struct {
	ID           primitive.ObjectID `bson:"_id" json:"_id"`
	UserID       primitive.ObjectID `bson:"user_id" json:"user_id"`
	RefreshToken string             `bson:"refreshToken" json:"refreshToken"`
	UserAgent    string             `bson:"useragent" json:"useragent"`
	UserIP       string             `bson:"userip" json:"userip"`
	IsBlocked    bool               `bson:"is_blocked" json:"is_blocked"`
	ExpiresAt    primitive.DateTime `bson:"expires_at" json:"expires_at"`
	CreatedAt    bool               `bson:"created_at" json:"created_at"`
}
