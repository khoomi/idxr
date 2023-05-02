package controllers

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"khoomi-api-io/khoomi_api/auth"
	"khoomi-api-io/khoomi_api/configs"
	"khoomi-api-io/khoomi_api/models"
	"khoomi-api-io/khoomi_api/responses"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var userCollection = configs.GetCollection(configs.DB, "User")
var loginHistoryCollection = configs.GetCollection(configs.DB, "UserLoginHistory")
var validate = validator.New()

func CreateUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var jsonUser models.UserRegistrationBody
		defer cancel()

		//validate the request body
		if err := c.BindJSON(&jsonUser); err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		if validationErr := validate.Struct(&jsonUser); validationErr != nil {
			c.JSON(http.StatusUnprocessableEntity, responses.UserResponse{Status: http.StatusUnprocessableEntity, Message: "error", Data: map[string]interface{}{"error": validationErr.Error(), "field": ""}})
			return
		}

		// validate and hash user_models password
		err := configs.ValidatePassword(jsonUser.Password)
		if err != nil {
			c.JSON(http.StatusExpectationFailed, responses.UserResponse{Status: http.StatusExpectationFailed, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		hashedPassword, errHashPassword := configs.HashPassword(jsonUser.Password)
		if errHashPassword != nil {
			c.JSON(http.StatusExpectationFailed, responses.UserResponse{Status: http.StatusExpectationFailed, Message: "error", Data: map[string]interface{}{"error": errHashPassword.Error(), "field": ""}})
			return
		}

		// validate login name
		errLoginName := configs.ValidateLoginName(jsonUser.LoginName)
		if errLoginName != nil {
			c.JSON(http.StatusExpectationFailed, responses.UserResponse{Status: http.StatusExpectationFailed, Message: "error", Data: map[string]interface{}{"error": errLoginName.Error(), "field": "login_name"}})
			return
		}

		// Check if username already in database
		var tempUser models.User
		errLoginNameExist := userCollection.FindOne(ctx, bson.M{"login_name": jsonUser.LoginName, "primary_email": jsonUser.Email}).Decode(tempUser)
		if errLoginNameExist != nil {
			if errLoginNameExist != mongo.ErrNoDocuments {
				c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": errLoginNameExist.Error(), "field": "login_name"}})
				return
			}
		}

		// Check if email already in database
		errEmailExist := userCollection.FindOne(ctx, bson.M{"login_name": jsonUser.LoginName, "primary_email": jsonUser.Email}).Decode(tempUser)
		if errEmailExist != nil {
			if errEmailExist != mongo.ErrNoDocuments {
				c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": errEmailExist.Error(), "field": "email"}})
				return
			}
		}

		now := time.Now()
		newUser := models.User{
			Id:           primitive.NewObjectID(),
			LoginName:    jsonUser.LoginName,
			PrimaryEmail: jsonUser.Email,
			Auth: models.UserAuthData{
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
			Role:           models.Regular,
			Status:         models.Inactive,
			Shops:          []string{},
			FavoriteShops:  []string{},
		}

		result, err := userCollection.InsertOne(ctx, newUser)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		c.JSON(http.StatusCreated, responses.UserResponse{Status: http.StatusCreated, Message: "success", Data: map[string]interface{}{"data": result}})
	}
}

func AuthenticateUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var jsonUser models.UserLoginBody
		defer cancel()

		if err := c.BindJSON(&jsonUser); err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			c.Abort()
			return
		}

		var validUser models.User
		if err := userCollection.FindOne(ctx, bson.M{"primary_email": jsonUser.Email}).Decode(&validUser); err != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": "primary_email"}})
			c.Abort()
			return
		}

		if errPasswordCheck := configs.CheckPassword(validUser.Auth.PasswordDigest, jsonUser.Password); errPasswordCheck != nil {
			c.JSON(http.StatusUnauthorized, responses.UserResponse{Status: http.StatusUnauthorized, Message: "error", Data: map[string]interface{}{"error": errPasswordCheck.Error(), "field": "password"}})
			c.Abort()
			return
		}

		wc := writeconcern.New(writeconcern.WMajority())
		txnOptions := options.Transaction().SetWriteConcern(wc)

		session, err := configs.DB.StartSession()
		if err != nil {
			panic(err)
		}
		defer session.EndSession(context.TODO())
		callback := func(ctx mongo.SessionContext) (interface{}, error) {
			// update user login counts
			filter := bson.M{"primary_email": validUser.PrimaryEmail}
			update := bson.M{"$set": bson.M{"last_login": time.Now(), "login_counts": validUser.LoginCounts + 1}}
			errUpdateLoginCounts := userCollection.FindOneAndUpdate(ctx, filter, update).Decode(&validUser)
			if errUpdateLoginCounts != nil {
				log.Fatal(errUpdateLoginCounts)
				return nil, errUpdateLoginCounts
			}

			// create login history
			doc := models.LoginHistory{
				Id:        primitive.NewObjectID(),
				UserUid:   validUser.Id,
				Date:      time.Now(),
				UserAgent: c.Request.UserAgent(),
				IpAddr:    c.ClientIP(),
			}
			log.Println(doc)
			result, errLoginHistory := loginHistoryCollection.InsertOne(ctx, doc)
			if errLoginHistory != nil {
				log.Println(errLoginHistory)
				return result, errLoginHistory
			}

			return result, nil
		}

		_, err = session.WithTransaction(context.Background(), callback, txnOptions)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			c.Abort()
		}

		tokenString, err := auth.GenerateJWT(validUser.Id.Hex(), validUser.PrimaryEmail, validUser.LoginName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": "auth_token"}})
			c.Abort()
			return
		}

		c.JSON(http.StatusCreated, responses.UserResponse{Status: http.StatusCreated, Message: "success", Data: map[string]interface{}{"token": tokenString}})
	}
}

func CurrentUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// Extract user id from request header
	userId, err := auth.ExtractTokenID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
		c.Abort()
		return
	}

	Id, _ := primitive.ObjectIDFromHex(userId)
	user, errMongo := GetUserById(ctx, Id)
	if errMongo != nil {
		c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"error": errMongo.Error(), "field": "user id"}})
		c.Abort()
		return
	}

	user.Auth.PasswordDigest = ""
	c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": user}})
}

func GetUserById(ctx context.Context, id primitive.ObjectID) (models.User, error) {
	var user models.User
	err := userCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		userId := c.Param("userId")
		defer cancel()

		Id, _ := primitive.ObjectIDFromHex(userId)
		user, errMongo := GetUserById(ctx, Id)
		if errMongo != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"data": errMongo.Error(), "field": "user id"}})
			c.Abort()
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": user}})
	}
}

// UpdateFirstLastName -> Update first and last name for current user
func UpdateFirstLastName() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		myId, err := auth.ExtractTokenID(c)
		var firstLastName models.FirstLastName
		defer cancel()

		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": "Authorization"}})
			return
		}

		errBind := c.BindJSON(&firstLastName)
		if errBind != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": errBind.Error(), "field": ""}})
			return
		}

		done, errFirstName := regexp.MatchString("([A-Z][a-zA-Z]*)", firstLastName.FirstName)
		if errFirstName != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": errFirstName.Error(), "field": "first_name"}})
			return
		}
		if !done {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": "First name should follow naming rule", "field": "first_name"}})
			return
		}

		done, errLastName := regexp.MatchString("([A-Z][a-zA-Z]*)", firstLastName.LastName)
		if !done {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": "Last name should follow naming rule", "field": "last_name"}})
			return
		}
		if errLastName != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": errLastName.Error(), "field": "last_name"}})
			return
		}

		IdToObjectId, errId := primitive.ObjectIDFromHex(myId)
		if errId != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": errId.Error(), "field": "user id"}})
			return
		}
		filter := bson.M{"_id": IdToObjectId}
		update := bson.M{"$set": bson.M{"first_name": firstLastName.FirstName, "last_name": firstLastName.LastName}}
		result, errUpdateName := userCollection.UpdateOne(ctx, filter, update)
		if errUpdateName != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": errUpdateName.Error(), "field": "user id"}})
			return
		}

		log.Println("Okay")

		c.JSON(http.StatusCreated, responses.UserResponse{Status: http.StatusCreated, Message: "success", Data: map[string]interface{}{"data": result}})
	}
}

// GetLoginHistories -> Get user login histories (/api/users/63ae3eb4b3cd579527549d97/login-history?limit=50&skip=0&sort=date)
func GetLoginHistories() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		userId := c.Param("userId")
		sort := c.Query("sort")

		limit := c.Query("limit")
		limitInt, err := strconv.Atoi(limit)
		if err != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"error": "Expected an integer for 'limit'", "field": "limit"}})
			c.Abort()
			return
		}

		skip := c.Query("skip")
		skipInt, err := strconv.Atoi(skip)
		if err != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"error": "Expected an integer for 'skip'", "field": "skip"}})
			c.Abort()
			return
		}

		userObj, _ := primitive.ObjectIDFromHex(userId)
		filter := bson.M{"user_uid": userObj}
		find := options.Find().SetLimit(int64(limitInt)).SetSkip(int64(skipInt)).SetSort(bson.M{sort: 1})
		result, err := loginHistoryCollection.Find(ctx, filter, find)
		if err != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			c.Abort()
			return
		}

		var loginHistory []models.LoginHistory
		if err = result.All(ctx, &loginHistory); err != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			c.Abort()
			return
		}

		c.JSON(http.StatusOK, responses.UserResponsePagination{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": loginHistory}, Pagination: responses.Pagination{
			Limit: limitInt,
			Skip:  skipInt,
			Sort:  sort,
			Total: len(loginHistory),
		}})
	}
}

func DeleteLoginHistories() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		userId := c.Param("userId")
		var historyIDs models.LoginHistoryIds
		defer cancel()

		// validate the request body
		if err := c.BindJSON(&historyIDs); err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		wc := writeconcern.New(writeconcern.WMajority())
		txnOptions := options.Transaction().SetWriteConcern(wc)

		session, err := configs.DB.StartSession()
		if err != nil {
			panic(err)
		}
		defer session.EndSession(context.TODO())

		userObj, _ := primitive.ObjectIDFromHex(userId)
		var DeleteingIds []primitive.ObjectID
		for _, id := range historyIDs.IDs {
			objId, _ := primitive.ObjectIDFromHex(id)
			DeleteingIds = append(DeleteingIds, objId)
		}
		log.Println(DeleteingIds)
		callback := func(ctx mongo.SessionContext) (interface{}, error) {
			// update user login counts
			filter := bson.M{"_id": bson.M{"$in": DeleteingIds}, "user_uid": userObj}
			result, err := loginHistoryCollection.DeleteMany(ctx, filter)
			if err != nil {
				return nil, err
			}

			return result, nil
		}

		_, err = session.WithTransaction(context.Background(), callback, txnOptions)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			c.Abort()
			return
		}

		c.JSON(http.StatusCreated, responses.UserResponse{Status: http.StatusCreated, Message: "success", Data: map[string]interface{}{"data": "Login histories deleted successfully"}})
	}

}
