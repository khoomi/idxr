package controllers

import (
	"context"
	"fmt"
	"github.com/cloudinary/cloudinary-go/api/uploader"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"khoomi-api-io/khoomi_api/auth"
	"khoomi-api-io/khoomi_api/configs"
	"khoomi-api-io/khoomi_api/email"
	"khoomi-api-io/khoomi_api/helper"
	"khoomi-api-io/khoomi_api/middleware"
	"khoomi-api-io/khoomi_api/models"
	"khoomi-api-io/khoomi_api/responses"
	"khoomi-api-io/khoomi_api/services"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var userCollection = configs.GetCollection(configs.DB, "User")
var userAddressCollection = configs.GetCollection(configs.DB, "User")
var loginHistoryCollection = configs.GetCollection(configs.DB, "UserLoginHistory")
var passwordResetTokenCollection = configs.GetCollection(configs.DB, "UserPasswordResetToken")
var userEmailVerificationTokenCollection = configs.GetCollection(configs.DB, "UserEmailVerificationToken")
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

		// Verify current user email
		errEmail := configs.ValidateEmailAddress(jsonUser.Email)
		if errEmail != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": "Bad or invalid email address", "field": ""}})
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

		now := time.Now()
		userAuth := models.UserAuthData{
			EmailVerified:  false,
			ModifiedAt:     time.Now(),
			PasswordDigest: hashedPassword,
		}
		newUser := models.User{
			Id:                   primitive.NewObjectID(),
			LoginName:            jsonUser.LoginName,
			PrimaryEmail:         jsonUser.Email,
			FirstName:            "",
			LastName:             "",
			Auth:                 userAuth,
			Thumbnail:            "",
			Bio:                  "",
			Phone:                "0000000000",
			Birthdate:            models.UserBirthdate{},
			IsSeller:             false,
			TransactionBuyCount:  0,
			TransactionSoldCount: 0,
			AddressId:            primitive.NilObjectID,
			ReferredByUser:       "",
			Role:                 models.Regular,
			Status:               models.Inactive,
			Shops:                []string{},
			FavoriteShops:        []string{},
			CreatedAt:            now,
			ModifiedAt:           now,
			LastLogin:            now,
			LoginCounts:          0,
			LastLoginIp:          c.ClientIP(),
		}

		result, err := userCollection.InsertOne(ctx, newUser)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		// Send welcome email.
		email.SendWelcomeEmail(newUser.PrimaryEmail, newUser.LoginName)

		c.JSON(http.StatusCreated, responses.UserResponse{Status: http.StatusCreated, Message: "success", Data: map[string]interface{}{"data": result}})
	}
}

// AuthenticateUser - AAuthenticate user into the server
func AuthenticateUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var jsonUser models.UserLoginBody
		clientIp := c.ClientIP()
		now := time.Now()
		defer cancel()

		if err := c.BindJSON(&jsonUser); err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		var validUser models.User
		if err := userCollection.FindOne(ctx, bson.M{"primary_email": jsonUser.Email}).Decode(&validUser); err != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": "primary_email"}})
			return
		}

		if errPasswordCheck := configs.CheckPassword(validUser.Auth.PasswordDigest, jsonUser.Password); errPasswordCheck != nil {
			c.JSON(http.StatusUnauthorized, responses.UserResponse{Status: http.StatusUnauthorized, Message: "error", Data: map[string]interface{}{"error": errPasswordCheck.Error(), "field": "password"}})
			return
		}

		wc := writeconcern.New(writeconcern.WMajority())
		txnOptions := options.Transaction().SetWriteConcern(wc)

		session, err := configs.DB.StartSession()
		if err != nil {
			panic(err)
		}
		defer session.EndSession(ctx)
		callback := func(ctx mongo.SessionContext) (interface{}, error) {
			// update user login counts
			filter := bson.M{"primary_email": validUser.PrimaryEmail}
			update := bson.M{"$set": bson.M{"last_login": now, "login_counts": validUser.LoginCounts + 1, "last_login_ip": clientIp}}
			errUpdateLoginCounts := userCollection.FindOneAndUpdate(ctx, filter, update).Decode(&validUser)
			if errUpdateLoginCounts != nil {
				log.Fatal(errUpdateLoginCounts)
				return nil, errUpdateLoginCounts
			}

			// create login history
			doc := models.LoginHistory{
				Id:        primitive.NewObjectID(),
				UserUid:   validUser.Id,
				Date:      now,
				UserAgent: c.Request.UserAgent(),
				IpAddr:    clientIp,
			}
			result, errLoginHistory := loginHistoryCollection.InsertOne(ctx, doc)
			if errLoginHistory != nil {
				return result, errLoginHistory
			}

			return result, nil
		}

		_, err = session.WithTransaction(context.Background(), callback, txnOptions)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
		}

		tokenString, err := auth.GenerateJWT(validUser.Id.Hex(), validUser.PrimaryEmail, validUser.LoginName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": "auth_token"}})
			return
		}

		// Send new login IP notification on condition
		if validUser.LastLoginIp != clientIp {
			email.SendNewIpLoginNotification(validUser.PrimaryEmail, validUser.LoginName, clientIp, time.Now().String())
		}

		c.JSON(http.StatusCreated, responses.UserResponse{Status: http.StatusCreated, Message: "success", Data: map[string]interface{}{"token": tokenString}})
	}
}

// Logout - Log user out and invalidate session key
func Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		token := auth.ExtractToken(c)
		defer cancel()

		_ = helper.InvalidateToken(c, configs.REDIS, token)
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
	user, errMongo := services.GetUserById(ctx, Id)
	if errMongo != nil {
		c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"error": errMongo.Error(), "field": "user id"}})
		c.Abort()
		return
	}

	user.Auth.PasswordDigest = ""
	c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": user}})
}

// GetUser => Get user by id endpoint
func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		userId := c.Param("userId")
		defer cancel()

		Id, _ := primitive.ObjectIDFromHex(userId)
		user, errMongo := services.GetUserById(ctx, Id)
		if errMongo != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"data": errMongo.Error(), "field": "user id"}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": user}})
	}
}

// ////////////////////// START USER EMAIL VERIFICATION //////////////////////////

// SendVerifyEmail -> api/send-verify-email?email=...&name=user_login_name
func SendVerifyEmail() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		emailCurrent := c.Query("email")
		loginName := c.Query("name")
		defer cancel()

		// Verify current user
		userId, err := auth.ExtractTokenID(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		// Verify current user email
		errEmail := configs.ValidateEmailAddress(emailCurrent)
		if errEmail != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": "Bad or invalid email address", "field": ""}})
			return
		}

		userObjId, _ := primitive.ObjectIDFromHex(userId)
		token := middleware.GenerateSecureToken(8)

		now := time.Now()
		expirationTime := now.Add(1 * time.Hour)
		verifyEmail := models.UserVerifyEmailToken{
			UserId:      userObjId,
			TokenDigest: token,
			CreatedAt:   primitive.NewDateTimeFromTime(now),
			ExpiresAt:   primitive.NewDateTimeFromTime(expirationTime),
		}
		opts := options.Replace().SetUpsert(true)
		filter := bson.M{"user_uid": userObjId}
		_, err = userEmailVerificationTokenCollection.ReplaceOne(ctx, filter, verifyEmail, opts)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": "email"}})
			return
		}

		link := fmt.Sprintf("https://khoomi.com/verify-email?token=%v&id=%v", token, userId)
		email.SendVerifyEmailNotification(emailCurrent, loginName, link)

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": "Password reset email send successfully"}})
	}
}

// VerifyEmail -> api/send-verify-email?email=...&name=user_login_name
func VerifyEmail() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		currentId := c.Query("id")
		currentToken := c.Query("token")
		var emailVerificationData models.UserVerifyEmailToken
		var user models.User
		defer cancel()

		userId, err := primitive.ObjectIDFromHex(currentId)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": "Error decoding userid", "field": "userid"}})
			return
		}

		err = userEmailVerificationTokenCollection.FindOneAndDelete(ctx, bson.M{"user_uid": userId}).Decode(&emailVerificationData)
		if err != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": "userid"}})
			return
		}

		// Check if reset token has expired
		now := primitive.NewDateTimeFromTime(time.Now())
		if now.Time().Unix() > emailVerificationData.ExpiresAt.Time().Unix() {
			if err != nil {
				c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"error": "Password reset token has expired. Please restart the reset process", "field": "expired"}})
				return
			}
		}

		// Check if reset token is correct.
		if currentToken != emailVerificationData.TokenDigest {
			if err != nil {
				c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"error": "Email verify token is incorrect or expired. Please restart the verification process or use a valid token", "field": "expired"}})
				return
			}
		}

		// Change user email verify status.
		filter := bson.M{"_id": emailVerificationData.UserId}
		update := bson.M{"$set": bson.M{"status": "Active", "modified_at": now, "auth.modified_at": now, "auth.email_verified": true}}
		err = userCollection.FindOneAndUpdate(ctx, filter, update).Decode(&user)
		if err != nil {
			c.JSON(http.StatusNotModified, responses.UserResponse{Status: http.StatusNotModified, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": "Your email been verified successfully."}})

	}
}

// UpdateFirstLastName -> Update first and last name for current user
func UpdateFirstLastName() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var firstLastName models.FirstLastName
		defer cancel()

		myId, err := auth.ExtractTokenID(c)
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

// ////////////////////// START USER LOGIN HISTORY //////////////////////////

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
			return
		}

		skip := c.Query("skip")
		skipInt, err := strconv.Atoi(skip)
		if err != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"error": "Expected an integer for 'skip'", "field": "skip"}})
			return
		}

		userObj, _ := primitive.ObjectIDFromHex(userId)
		filter := bson.M{"user_uid": userObj}
		find := options.Find().SetLimit(int64(limitInt)).SetSkip(int64(skipInt)).SetSort(bson.M{sort: 1})
		result, err := loginHistoryCollection.Find(ctx, filter, find)
		if err != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		var loginHistory []models.LoginHistory
		if err = result.All(ctx, &loginHistory); err != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
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
		var IdsToDelete []primitive.ObjectID
		for _, id := range historyIDs.IDs {
			objId, _ := primitive.ObjectIDFromHex(id)
			IdsToDelete = append(IdsToDelete, objId)
		}
		log.Println(IdsToDelete)
		callback := func(ctx mongo.SessionContext) (interface{}, error) {
			// update user login counts
			filter := bson.M{"_id": bson.M{"$in": IdsToDelete}, "user_uid": userObj}
			result, err := loginHistoryCollection.DeleteMany(ctx, filter)
			if err != nil {
				return nil, err
			}

			return result, nil
		}

		_, err = session.WithTransaction(context.Background(), callback, txnOptions)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": "Login histories deleted successfully"}})
	}

}

// ////////////////////// START USER PASSWORD RESET //////////////////////////

// PasswordResetEmail - api/send-password-reset?email=borngracedd@gmail.com
func PasswordResetEmail() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		currentEmail := c.Query("email")
		var user models.User
		defer cancel()

		err := userCollection.FindOne(ctx, bson.M{"primary_email": currentEmail}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": "user with email now found", "field": "email"}})
			return
		}

		token := middleware.GenerateSecureToken(8)
		now := time.Now()
		expirationTime := now.Add(1 * time.Hour)
		passwordReset := models.UserPasswordResetToken{
			UserId:      primitive.ObjectID{},
			TokenDigest: token,
			CreatedAt:   primitive.NewDateTimeFromTime(now),
			ExpiresAt:   primitive.NewDateTimeFromTime(expirationTime),
		}

		opts := options.Replace().SetUpsert(true)
		filter := bson.M{"user_uid": user.Id}
		_, err = passwordResetTokenCollection.ReplaceOne(ctx, filter, passwordReset, opts)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": "email"}})
			return
		}

		link := fmt.Sprintf("https://khoomi.com/%v/password-reset/?token=%v", user.Id.Hex(), token)
		email.SendPasswordResetEmail(user.PrimaryEmail, user.LoginName, link)

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": "Password reset email send successfully"}})
	}
}

// PasswordReset - api/password-reset/userid?token=..&newpassword=..&id=user_uid
func PasswordReset() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		currentId := c.Query("id")
		currentToken := c.Query("token")
		newPassword := c.Query("newpassword")
		var passwordResetData models.UserPasswordResetToken
		var user models.User
		defer cancel()

		userId, err := primitive.ObjectIDFromHex(currentId)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": "Error decoding userid", "field": "userid"}})
			return
		}

		err = passwordResetTokenCollection.FindOneAndDelete(ctx, bson.M{"user_uid": userId}).Decode(&passwordResetData)
		if err != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": "userid"}})
			return
		}

		// Check if reset token has expired
		now := primitive.NewDateTimeFromTime(time.Now())
		if now.Time().Unix() > passwordResetData.ExpiresAt.Time().Unix() {
			if err != nil {
				c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"error": "Password reset token has expired. Please restart the reset process", "field": "expired"}})
				return
			}
		}

		// Check if reset token is correct.
		if currentToken != passwordResetData.TokenDigest {
			if err != nil {
				c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"error": "Password reset token is incorrect or expired. Please restart the reset process or use a valid token", "field": "expired"}})
				return
			}
		}

		// Validate and hash new given password.
		err = configs.ValidatePassword(newPassword)
		if err != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": "userid"}})
			return
		}
		hashedPassword, err := configs.HashPassword(newPassword)
		if err != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": "userid"}})
			return
		}

		// Change user password.
		filter := bson.M{"_id": passwordResetData.UserId}
		update := bson.M{"$set": bson.M{"auth.password_digest": hashedPassword, "auth.modified_at": now, "auth.email_verified": true}}
		err = userCollection.FindOneAndUpdate(ctx, filter, update).Decode(&user)
		if err != nil {
			c.JSON(http.StatusNotModified, responses.UserResponse{Status: http.StatusNotModified, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		// Send password reset successfully email to user.
		email.SendPasswordResetSuccessfulEmail(user.PrimaryEmail, user.LoginName)

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": "Your password been changed successfully."}})

	}
}

// ////////////////////// START USER THUMBNAIL //////////////////////////

// UploadThumbnail - Upload user profile picture/thumbnail
// api/user/thumbnail?kind=(remote | file)&url=..
func UploadThumbnail() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		kind := c.Query("kind")
		defer cancel()

		currentIdStr, err := auth.ExtractTokenID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, responses.UserResponse{Status: http.StatusUnauthorized, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": "file"}})
			return
		}
		currentId, _ := primitive.ObjectIDFromHex(currentIdStr)

		now := time.Now()
		filter := bson.M{"_id": currentId}

		// if user wants remote upload we proceed here
		if kind == "remote" {
			url := c.Query("url")
			uploadUrl, err := services.NewMediaUpload().RemoteUpload(models.Url{Url: url})
			if err != nil {
				c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": "file"}})
				return
			}

			update := bson.M{"$set": bson.M{"thumbnail": uploadUrl, "modified_at": now}}
			_, err = userCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				c.JSON(http.StatusExpectationFailed, responses.UserResponse{Status: http.StatusExpectationFailed, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": "file"}})
				return
			}
		}

		formFile, _, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": "file"}})
			return
		}

		uploadUrl, err := services.NewMediaUpload().FileUpload(models.File{File: formFile})
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": "file"}})
			return
		}

		update := bson.M{"$set": bson.M{"thumbnail": uploadUrl, "modified_at": now}}
		_, err = userCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusExpectationFailed, responses.UserResponse{Status: http.StatusExpectationFailed, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": "file"}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": "Your thumbnail has been changed successfully."}})
	}
}

// DeleteThumbnail - Upload user profile picture/thumbnail
// api/user/thumbnail?name=thumbnail_name&type=jpg
func DeleteThumbnail() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		name := c.Query("name")
		kind := c.Query("type")
		defer cancel()

		currentIdStr, err := auth.ExtractTokenID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, responses.UserResponse{Status: http.StatusUnauthorized, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": "file"}})
			return
		}
		currentId, _ := primitive.ObjectIDFromHex(currentIdStr)

		_, errOnDelete := helper.ImageDeletionHelper(uploader.DestroyParams{
			PublicID:     name,
			Type:         "upload",
			ResourceType: kind,
			Invalidate:   false,
		})
		if errOnDelete != nil {
			c.JSON(http.StatusExpectationFailed, responses.UserResponse{Status: http.StatusExpectationFailed, Message: "error", Data: map[string]interface{}{"error": errOnDelete.Error(), "field": "file"}})
			return
		}

		now := time.Now()
		filter := bson.M{"_id": currentId}
		update := bson.M{"$set": bson.M{"thumbnail": nil, "modified_at": now}}
		_, err = userCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusExpectationFailed, responses.UserResponse{Status: http.StatusExpectationFailed, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": "file"}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": "Your thumbnail has been deleted successfully."}})
	}
}

// ////////////////////// START USER ADDRESS //////////////////////////

// CreateUserAddress - create new user address
func CreateUserAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var userAddress models.UserAddress
		defer cancel()

		// Extract current user token
		userIdStr, err := auth.ExtractTokenID(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		userId, err := primitive.ObjectIDFromHex(userIdStr)

		// Validate the request body
		if err := c.BindJSON(&userAddress); err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		wc := writeconcern.New(writeconcern.WMajority())
		txnOptions := options.Transaction().SetWriteConcern(wc)
		session, err := configs.DB.StartSession()
		if err != nil {
			panic(err)
		}
		defer session.EndSession(ctx)
		callback := func(ctx mongo.SessionContext) (interface{}, error) {
			// create user address
			userAddress.UserId = userId
			userAddress.Id = primitive.NewObjectID()
			_, err = userAddressCollection.InsertOne(ctx, userAddress)
			if err != nil {
				return nil, err
			}

			// update user profile address_id
			filter := bson.M{"_id": userId}
			update := bson.M{"$set": bson.M{"address_id": userAddress.Id, "modified_at": time.Now()}}
			res, err := userCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				return nil, err
			}

			return res, nil
		}

		_, err = session.WithTransaction(context.Background(), callback, txnOptions)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": "Address created."}})

	}
}

// UpdateUserAddress - update user address
func UpdateUserAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var userAddress models.UserAddress
		defer cancel()

		// Extract current user token
		userIdStr, err := auth.ExtractTokenID(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		userId, err := primitive.ObjectIDFromHex(userIdStr)

		// Validate the request body
		if err := c.BindJSON(&userAddress); err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		filter := bson.M{"user_id": userId}
		update := bson.M{"$set": bson.M{"city": userAddress.City, "state": userAddress.State, "street": userAddress.Street, "postal_code": userAddress.PostalCode, "country": models.CountryNigeria}}
		_, err = userAddressCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": "Address updated."}})

	}
}

// ////////////////////// START USER BIRTHDATE //////////////////////////

// UpdateUserBirthdate - update user birthdate
func UpdateUserBirthdate() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var birthDate models.UserBirthdate
		defer cancel()

		myId, err := auth.ExtractTokenID(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": "Authorization"}})
			return
		}

		errBind := c.BindJSON(&birthDate)
		if errBind != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": errBind.Error(), "field": ""}})
			return
		}

		IdToObjectId, errId := primitive.ObjectIDFromHex(myId)
		if errId != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": errId.Error(), "field": "user id"}})
			return
		}
		filter := bson.M{"_id": IdToObjectId}
		update := bson.M{"$set": bson.M{"birthdate.day": birthDate.Day, "birthdate.month": birthDate.Month, "birthdate.year": birthDate.Year}}
		result, errUpdateName := userCollection.UpdateOne(ctx, filter, update)
		if errUpdateName != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": errUpdateName.Error(), "field": "user id"}})
			return
		}

		log.Println("Okay")

		c.JSON(http.StatusCreated, responses.UserResponse{Status: http.StatusCreated, Message: "success", Data: map[string]interface{}{"data": result}})
	}
}

// UpdateUserSingleField - update user single field like Phone, Bio
// api/user/update?field=phone&value=8084051523
func UpdateUserSingleField() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		field := c.Query("field")
		value := c.Query("value")
		defer cancel()

		notAllowedFields := []string{"role", "login_counts", "modified_at", "created_at", "favorite_shops", "shops", "status", "referred_by_user", "address_id", "transaction_sold_count", "transaction_buy_count", "birthdate", "thumbnail", "auth", "primary_email", "login_name", "_id"}

		for _, n := range notAllowedFields {
			if strings.ToLower(field) == n {
				c.JSON(http.StatusUnauthorized, responses.UserResponse{Status: http.StatusUnauthorized, Message: "error", Data: map[string]interface{}{"error": fmt.Sprintf("No way!, you can't change your %v", n), "field": "Authorization"}})
				return
			}
		}

		if strings.ToLower(field) == "login_counts" {
			c.JSON(http.StatusUnauthorized, responses.UserResponse{Status: http.StatusUnauthorized, Message: "error", Data: map[string]interface{}{"error": "No way!, you can't change your login_counts", "field": "Authorization"}})
			return
		}

		if strings.ToLower(field) == "login_counts" {
			c.JSON(http.StatusUnauthorized, responses.UserResponse{Status: http.StatusUnauthorized, Message: "error", Data: map[string]interface{}{"error": "No way!, you can't change your login_counts", "field": "Authorization"}})
			return
		}

		myId, err := auth.ExtractTokenID(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": "Authorization"}})
			return
		}
		IdToObjectId, errId := primitive.ObjectIDFromHex(myId)
		if errId != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": errId.Error(), "field": "user id"}})
			return
		}

		filter := bson.M{"_id": IdToObjectId}
		update := bson.M{"$set": bson.M{field: value}}
		result, errUpdateName := userCollection.UpdateOne(ctx, filter, update)
		if errUpdateName != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": errUpdateName.Error(), "field": field}})
			return
		}

		c.JSON(http.StatusCreated, responses.UserResponse{Status: http.StatusCreated, Message: "success", Data: map[string]interface{}{"data": result}})
	}
}
