package user_controllers

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"khoomi-api-io/khoomi_api/auth"
	"khoomi-api-io/khoomi_api/configs"
	"khoomi-api-io/khoomi_api/models/user_models"
	"khoomi-api-io/khoomi_api/responses"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var userCollection = configs.GetCollection(configs.DB, "User")
var validate = validator.New()

func CreateUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var jsonUser user_models.UserRegistrationBody
		defer cancel()

		//validate the request body
		if err := c.BindJSON(&jsonUser); err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"data": err.Error()}})
			return
		}

		if validationErr := validate.Struct(&jsonUser); validationErr != nil {
			c.JSON(http.StatusUnprocessableEntity, responses.UserResponse{Status: http.StatusUnprocessableEntity, Message: "error", Data: map[string]interface{}{"data": validationErr.Error()}})
			return
		}

		// validate and hash user_models password
		err := configs.ValidatePassword(jsonUser.Password)
		if err != nil {
			c.JSON(http.StatusExpectationFailed, responses.UserResponse{Status: http.StatusExpectationFailed, Message: "error", Data: map[string]interface{}{"data": err.Error()}})
			return
		}

		hashedPassword, errHashPassword := configs.HashPassword(jsonUser.Password)
		if errHashPassword != nil {
			c.JSON(http.StatusExpectationFailed, responses.UserResponse{Status: http.StatusExpectationFailed, Message: "error", Data: map[string]interface{}{"data": errHashPassword.Error()}})
			return
		}

		// validate login name
		errLoginName := configs.ValidateLoginName(jsonUser.LoginName)
		if errLoginName != nil {
			c.JSON(http.StatusExpectationFailed, responses.UserResponse{Status: http.StatusExpectationFailed, Message: "error", Data: map[string]interface{}{"data": errLoginName.Error()}})
			return
		}

		// Check if username already in database
		var tempUser user_models.User
		errLoginNameExist := userCollection.FindOne(ctx, bson.M{"login_name": jsonUser.LoginName}).Decode(tempUser)
		if errLoginNameExist != nil {
			if errLoginNameExist != mongo.ErrNoDocuments {
				c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"data": errLoginNameExist.Error()}})
				return
			}
		}

		now := time.Now()
		newUser := user_models.User{
			Id:           primitive.NewObjectID(),
			LoginName:    jsonUser.LoginName,
			PrimaryEmail: jsonUser.Email,
			FirstName:    "",
			LastName:     "",
			Auth: user_models.UserAuthData{
				EmailVerified:  false,
				ModifiedAt:     time.Now(),
				PasswordDigest: hashedPassword,
			},
			Thumbnail:      "",
			ProfileUid:     primitive.NilObjectID,
			LoginCounts:    0,
			LastLogin:      now,
			CreatedAt:      now,
			ModifiedAt:     now,
			ReferredByUser: "",
			Role:           user_models.Regular,
			Status:         user_models.Inactive,
			Shops:          []string{},
			FavoriteShops:  []string{},
		}

		result, err := userCollection.InsertOne(ctx, newUser)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"data": err.Error()}})
			return
		}

		c.JSON(http.StatusCreated, responses.UserResponse{Status: http.StatusCreated, Message: "success", Data: map[string]interface{}{"data": result}})
	}
}

func AuthenticateUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
		var jsonUser user_models.UserLoginBody
		defer cancel()

		if err := c.BindJSON(&jsonUser); err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"data": err.Error()}})
			c.Abort()
			return
		}

		var validUser user_models.User
		if err := userCollection.FindOne(ctx, bson.M{"primary_email": jsonUser.Email}).Decode(&validUser); err != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"data": err.Error()}})
			c.Abort()
			return
		}

		if errPasswordCheck := configs.CheckPassword(validUser.Auth.PasswordDigest, jsonUser.Password); errPasswordCheck != nil {
			c.JSON(http.StatusUnauthorized, responses.UserResponse{Status: http.StatusUnauthorized, Message: "error", Data: map[string]interface{}{"data": errPasswordCheck.Error()}})
			c.Abort()
			return
		}

		tokenString, err := auth.GenerateJWT(validUser.PrimaryEmail, validUser.LoginName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"data": err.Error()}})
			c.Abort()
			return
		}

		c.JSON(http.StatusCreated, responses.UserResponse{Status: http.StatusCreated, Message: "success", Data: map[string]interface{}{"token": tokenString}})
	}
}

func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
		userId := c.Param("userId")
		var user user_models.User
		defer cancel()

		Id, _ := primitive.ObjectIDFromHex(userId)
		err := userCollection.FindOne(ctx, bson.M{"_id": Id}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"data": err.Error()}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": user}})
	}
}
