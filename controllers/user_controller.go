package controllers

import (
	"context"
	"fmt"
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
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/api/uploader"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"go.mongodb.org/mongo-driver/x/bsonx"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var userCollection = configs.GetCollection(configs.DB, "User")
var userAddressCollection = configs.GetCollection(configs.DB, "UserAddress")
var loginHistoryCollection = configs.GetCollection(configs.DB, "UserLoginHistory")
var passwordResetTokenCollection = configs.GetCollection(configs.DB, "UserPasswordResetToken")
var emailVerificationTokenCollection = configs.GetCollection(configs.DB, "UserEmailVerificationToken")
var wishListCollection = configs.GetCollection(configs.DB, "UserWishList")

var validate = validator.New()

const (
	UserRequestTimeout = 20
)

func CreateUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		var jsonUser models.UserRegistrationBody
		now := time.Now()
		defer cancel()

		// bind the request body
		if err := c.BindJSON(&jsonUser); err != nil {
			log.Printf("Error binding request body: %s\n", err.Error())
			helper.HandleError(c, http.StatusBadRequest, err, "invalid or missing data in request body")
			return
		}

		// validate request body
		if err := validate.Struct(&jsonUser); err != nil {
			log.Printf("Error validating request body: %s\n", err.Error())
			helper.HandleError(c, http.StatusBadRequest, err, "invalid or missing data in request body")
			return
		}

		// Verify current user email
		err := configs.ValidateEmailAddress(jsonUser.Email)
		if err != nil {
			log.Printf("Invalid email address from user %s with IP %s at %s: %s\n", jsonUser.FirstName, c.ClientIP(), time.Now().Format(time.RFC3339), err.Error())
			helper.HandleError(c, http.StatusBadRequest, err, "invalid email address")
			return
		}

		// validate user password
		err := configs.ValidatePassword(jsonUser.Password)
		if err != nil {
			log.Printf("Error validating password: %s\n", err.Error())
			helper.HandleError(c, http.StatusBadRequest, err, err.Error())
			return
		}

		// hash user password
		hashedPassword, errHashPassword := configs.HashPassword(jsonUser.Password)
		if errHashPassword != nil {
			log.Printf("Error hashing password: %s\n", errHashPassword.Error())
			helper.HandleError(c, http.StatusBadRequest, err, "Internal error")
			return
		}

		// validate login name
		//errLoginName := configs.ValidateLoginName(jsonUser.LoginName)
		//if errLoginName != nil {
		//	log.Printf("Error validating login name: %s\n", errLoginName.Error())
		//	c.JSON(http.StatusExpectationFailed, responses.UserResponse{Status: http.StatusExpectationFailed, Message: "error", Data: map[string]interface{}{"error": errLoginName.Error()}})
		//	return
		//}

		userAuth := models.UserAuthData{
			EmailVerified:  false,
			ModifiedAt:     now,
			PasswordDigest: hashedPassword,
		}
		jsonUser.Email = strings.ToLower(jsonUser.Email)
		newUser := bson.M{
			"_id":                    primitive.NewObjectID(),
			"login_name":             bsonx.Null(),
			"primary_email":          strings.ToLower(jsonUser.Email),
			"first_name":             jsonUser.FirstName,
			"last_name":              bsonx.Null(),
			"auth":                   userAuth,
			"thumbnail":              bsonx.Null(),
			"bio":                    bsonx.Null(),
			"phone":                  bsonx.Null(),
			"birthdate":              models.UserBirthdate{},
			"is_seller":              false,
			"transaction_buy_count":  0,
			"transaction_sold_count": 0,
			"referred_by_user":       bsonx.Null(),
			"role":                   models.Regular,
			"status":                 models.Inactive,
			"shops":                  []string{},
			"favorite_shops":         []string{},
			"created_at":             now,
			"modified_at":            now,
			"last_login":             now,
			"login_counts":           0,
			"last_login_ip":          c.ClientIP(),
		}

		// insert user data to db
		result, err := userCollection.InsertOne(ctx, newUser)
		if err != nil {
			log.Printf("Mongo Error: Request could not be completed %s\n", err.Error())
			helper.HandleError(c, http.StatusInternalServerError, err, "internal server error while attempting to create user, plese try again later")
			return
		}

		// Send welcome email.
		email.SendWelcomeEmail(jsonUser.Email, jsonUser.FirstName)

		helper.HandleSuccess(c, http.StatusOK, "signup successful", result.InsertedID)
	}
}

// HandleUserAuthentication - Authenticate user into the server
func HandleUserAuthentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		var jsonUser models.UserLoginBody
		clientIp := c.ClientIP()
		now := time.Now()
		defer cancel()

		if err := c.BindJSON(&jsonUser); err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}

		// Validate request body
		if validationErr := validate.Struct(&jsonUser); validationErr != nil {
			c.JSON(http.StatusUnprocessableEntity, responses.UserResponse{Status: http.StatusUnprocessableEntity, Message: validationErr.Error(), Data: map[string]interface{}{}})
			return
		}

		// check and return valid user with email from db
		var validUser models.User
		if err := userCollection.FindOne(ctx, bson.M{"primary_email": jsonUser.Email}).Decode(&validUser); err != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}

		// check and validate user password digest
		if errPasswordCheck := configs.CheckPassword(validUser.Auth.PasswordDigest, jsonUser.Password); errPasswordCheck != nil {
			c.JSON(http.StatusUnauthorized, responses.UserResponse{Status: http.StatusUnauthorized, Message: errPasswordCheck.Error(), Data: map[string]interface{}{}})
			return
		}

		// start auth session
		wc := writeconcern.New(writeconcern.WMajority())
		txnOptions := options.Transaction().SetWriteConcern(wc)
		session, err := configs.DB.StartSession()
		if err != nil {
			defer func() {
				if r := recover(); r != nil {
					log.Println("Recovered from panic:", r)
				}
				c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "failed to start database session", Data: map[string]interface{}{}})
				return
			}()
		}
		defer session.EndSession(ctx)

		callback := func(ctx mongo.SessionContext) (interface{}, error) {
			// update user login counts
			filter := bson.M{"primary_email": validUser.PrimaryEmail}
			update := bson.M{"$set": bson.M{"last_login": now, "login_counts": validUser.LoginCounts + 1, "last_login_ip": clientIp}}
			errUpdateLoginCounts := userCollection.FindOneAndUpdate(ctx, filter, update).Decode(&validUser)
			if errUpdateLoginCounts != nil {
				log.Println(errUpdateLoginCounts)
				return nil, errUpdateLoginCounts
			}

			// create and insert login history to db
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
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: err.Error(), Data: map[string]interface{}{}})
		}

		if err := session.CommitTransaction(context.Background()); err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}
		session.EndSession(context.Background())

		tokenString, err := auth.GenerateJWT(validUser.Id.Hex(), validUser.PrimaryEmail, validUser.LoginName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}

		// Send new login IP notification on condition
		email.SendNewIpLoginNotification(validUser.PrimaryEmail, validUser.LoginName, validUser.LastLoginIp, validUser.LastLogin)

		c.JSON(http.StatusCreated, responses.UserResponse{Status: http.StatusCreated, Message: "success", Data: map[string]interface{}{"token": tokenString, "role": validUser.Role, "email": validUser.PrimaryEmail, "name": validUser.FirstName, "thumbnail": validUser.Thumbnail, "email_verified": validUser.Auth.EmailVerified}})
	}
}

// Logout - Log user out and invalidate session key
func Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		token := auth.ExtractToken(c)
		defer cancel()

		log.Printf("Logging user with ip %v out\n", c.ClientIP())
		_ = helper.InvalidateToken(c, configs.REDIS, token)
	}
}

func CurrentUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
	defer cancel()
	// Extract user id from request header
	userId, err := auth.ExtractTokenID(c)
	if err != nil {
		log.Printf("user with ip %v tried to gain access with invalid userid or token\n", c.ClientIP())
		c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: err.Error(), Data: map[string]interface{}{}})
		c.Abort()
		return
	}

	user, err := services.GetUserById(ctx, userId)
	if err != nil {
		log.Printf("Logging user with ip %v out\n", c.ClientIP())
		c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: err.Error(), Data: map[string]interface{}{}})
		c.Abort()
		return
	}

	user.Auth.PasswordDigest = ""
	c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": user}})
}

// GetUserByIDOrEmail - Get user by id or email endpoint
func GetUserByIDOrEmail() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		defer cancel()

		userID := c.Query("id")
		emailId := c.Query("email")

		// Check if either the user ID or email is missing
		if userID == "" && emailId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing user ID or email"})
			return
		}

		// Prepare the filter based on the available query parameter
		filter := bson.M{}
		if userID != "" {
			filter["_id"] = userID
		} else {
			filter["email"] = emailId
		}

		// Query the database to find the user based on the specified field and value
		var user models.User
		err := userCollection.FindOne(ctx, filter).Decode(&user)
		if err != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusBadRequest, Message: "User not found"})
			return
		}

		// Return the user data in the response
		c.JSON(http.StatusOK, gin.H{"data": user})
	}
}

// SendVerifyEmail - api/send-verify-email?email=...&name=user_login_name
func SendVerifyEmail() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		emailCurrent := c.Query("email")
		firstName := c.Query("name")
		now := time.Now()
		defer cancel()

		// Verify current user
		userId, err := auth.ExtractTokenID(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: err.Error()})
			return
		}

		// Verify current user email
		errEmail := configs.ValidateEmailAddress(emailCurrent)
		if errEmail != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "invalid email address", Data: map[string]interface{}{}})
			return
		}

		// generate secure and unique token
		token := middleware.GenerateSecureToken(8)

		expirationTime := now.Add(1 * time.Hour)
		verifyEmail := models.UserVerifyEmailToken{
			UserId:      userId,
			TokenDigest: token,
			CreatedAt:   primitive.NewDateTimeFromTime(now),
			ExpiresAt:   primitive.NewDateTimeFromTime(expirationTime),
		}
		opts := options.Replace().SetUpsert(true)
		filter := bson.M{"user_uid": userId}
		_, err = emailVerificationTokenCollection.ReplaceOne(ctx, filter, verifyEmail, opts)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}

		link := fmt.Sprintf("https://khoomi.com/verify-email?token=%v&id=%v", token, userId)
		// Send welcome email.
		email.SendVerifyEmailNotification(emailCurrent, firstName, link)

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": "verification email successfully sent"}})
	}
}

// VerifyEmail - api/send-verify-email?email=...&name=user_login_name
func VerifyEmail() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		currentId := c.Query("id")
		currentToken := c.Query("token")
		var emailVerificationData models.UserVerifyEmailToken
		var user models.User
		defer cancel()

		userId, err := primitive.ObjectIDFromHex(currentId)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "invalid given user id", Data: map[string]interface{}{}})
			return
		}

		// get and delete email verification
		err = emailVerificationTokenCollection.FindOneAndDelete(ctx, bson.M{"user_uid": userId}).Decode(&emailVerificationData)
		if err != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}

		// Check if reset token has expired
		now := primitive.NewDateTimeFromTime(time.Now())
		if now.Time().Unix() > emailVerificationData.ExpiresAt.Time().Unix() {
			if err != nil {
				c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "password reset token has expired.", Data: map[string]interface{}{}})
				return
			}
		}

		// Check if reset token is correct.
		if currentToken != emailVerificationData.TokenDigest {
			if err != nil {
				c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "email verify token is incorrect or expired. ", Data: map[string]interface{}{}})
				return
			}
		}

		// Change user email verify status.
		filter := bson.M{"_id": emailVerificationData.UserId}
		update := bson.M{"$set": bson.M{"status": "Active", "modified_at": now, "auth.modified_at": now, "auth.email_verified": true}}
		err = userCollection.FindOneAndUpdate(ctx, filter, update).Decode(&user)
		if err != nil {
			c.JSON(http.StatusNotModified, responses.UserResponse{Status: http.StatusNotModified, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": "your email been verified successfully."}})

	}
}

// UpdateFirstLastName -> Update first and last name for current user
func UpdateFirstLastName() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		var firstLastName models.FirstLastName
		defer cancel()

		myId, err := auth.ExtractTokenID(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}

		err = c.BindJSON(&firstLastName)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}

		// Validate the request body
		if err := c.BindJSON(&firstLastName); err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}

		done, err := regexp.MatchString("([A-Z][a-zA-Z]*)", firstLastName.FirstName)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}
		if !done {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "first name should follow naming rule", Data: map[string]interface{}{}})
			return
		}

		done, err = regexp.MatchString("([A-Z][a-zA-Z]*)", firstLastName.LastName)
		if !done {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "last name should follow naming rule", Data: map[string]interface{}{}})
			return
		}
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}

		filter := bson.M{"_id": myId}
		update := bson.M{"$set": bson.M{"first_name": firstLastName.FirstName, "last_name": firstLastName.LastName}}
		result, err := userCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}

		c.JSON(http.StatusCreated, responses.UserResponse{Status: http.StatusCreated, Message: "success", Data: map[string]interface{}{"data": result}})
	}
}

// ////////////////////// START USER LOGIN HISTORY //////////////////////////

// GetLoginHistories - Get user login histories (/api/users/login-history?limit=50&skip=0)
func GetLoginHistories() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		defer cancel()
		userId, err := auth.ExtractTokenID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, responses.UserResponse{Status: http.StatusUnauthorized, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}

		paginationArgs := services.GetPaginationArgs(c)
		filter := bson.M{"user_uid": userId}
		find := options.Find().SetLimit(int64(paginationArgs.Limit)).SetSkip(int64(paginationArgs.Skip))
		result, err := loginHistoryCollection.Find(ctx, filter, find)
		if err != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}

		count, err := loginHistoryCollection.CountDocuments(ctx, bson.M{"user_uid": userId})
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error counting shops", Data: map[string]interface{}{}})
			return
		}

		var loginHistory []models.LoginHistory
		if err = result.All(ctx, &loginHistory); err != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponsePagination{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": loginHistory}, Pagination: responses.Pagination{
			Limit: paginationArgs.Limit,
			Skip:  paginationArgs.Skip,
			Count: count,
		}})
	}
}

func DeleteLoginHistories() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		userId, err := auth.ExtractTokenID(c)
		var historyIDs models.LoginHistoryIds
		defer cancel()

		// validate the request body
		if err := c.BindJSON(&historyIDs); err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}

		// start delete login history session
		wc := writeconcern.New(writeconcern.WMajority())
		txnOptions := options.Transaction().SetWriteConcern(wc)
		session, err := configs.DB.StartSession()
		if err != nil {
			defer func() {
				if r := recover(); r != nil {
					log.Println("Recovered from panic:", r)
				}
				c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "failed to start database session", Data: map[string]interface{}{}})
			}()
			log.Println("Failed to start database session: " + err.Error())
		}
		defer session.EndSession(context.TODO())

		var IdsToDelete []primitive.ObjectID
		for _, id := range historyIDs.IDs {
			objId, _ := primitive.ObjectIDFromHex(id)
			IdsToDelete = append(IdsToDelete, objId)
		}
		callback := func(ctx mongo.SessionContext) (interface{}, error) {
			// update user login counts
			filter := bson.M{"_id": bson.M{"$in": IdsToDelete}, "user_uid": userId}
			result, err := loginHistoryCollection.DeleteMany(ctx, filter)
			if err != nil {
				return nil, err
			}

			return result, nil
		}

		_, err = session.WithTransaction(context.Background(), callback, txnOptions)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		if err := session.CommitTransaction(context.Background()); err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}
		session.EndSession(context.Background())
		// end delete login history session

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": "Login histories deleted successfully"}})
	}

}

// ////////////////////// START USER PASSWORD RESET //////////////////////////

// PasswordResetEmail - api/send-password-reset?email=borngracedd@gmail.com
func PasswordResetEmail() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		currentEmail := strings.ToLower(c.Query("email"))
		var user models.User
		defer cancel()

		err := userCollection.FindOne(ctx, bson.M{"primary_email": currentEmail}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "user with email now found", Data: map[string]interface{}{}})
			return
		}

		token := middleware.GenerateSecureToken(8)
		now := time.Now()
		expirationTime := now.Add(1 * time.Hour)
		passwordReset := models.UserPasswordResetToken{
			UserId:      user.Id,
			TokenDigest: token,
			CreatedAt:   primitive.NewDateTimeFromTime(now),
			ExpiresAt:   primitive.NewDateTimeFromTime(expirationTime),
		}

		opts := options.Replace().SetUpsert(true)
		filter := bson.M{"user_uid": user.Id}
		_, err = passwordResetTokenCollection.ReplaceOne(ctx, filter, passwordReset, opts)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}

		// send password reset email
		link := fmt.Sprintf("https://khoomi.com/password-reset/?id=%v&token=%v", user.Id.Hex(), token)
		email.SendPasswordResetEmail(user.PrimaryEmail, user.FirstName, link)

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": "Password reset email send successfully"}})
	}
}

// PasswordReset - api/password-reset/userid?token=..&newpassword=..&id=user_uid
func PasswordReset() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		currentId := c.Query("id")
		currentToken := c.Query("token")
		newPassword := c.Query("newpassword")
		var passwordResetData models.UserPasswordResetToken
		var user models.User
		defer cancel()

		userId, err := primitive.ObjectIDFromHex(currentId)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "invalid userid", Data: map[string]interface{}{}})
			return
		}

		err = passwordResetTokenCollection.FindOneAndDelete(ctx, bson.M{"user_uid": userId}).Decode(&passwordResetData)
		if err != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}

		// Check if reset token has expired
		now := primitive.NewDateTimeFromTime(time.Now())
		if now.Time().Unix() > passwordResetData.ExpiresAt.Time().Unix() {
			if err != nil {
				c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "password reset token has expired. Please restart the reset process", Data: map[string]interface{}{}})
				return
			}
		}

		// Check if reset token is correct.
		if currentToken != passwordResetData.TokenDigest {
			if err != nil {
				c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "password reset token is incorrect or expired. Please restart the reset process or use a valid token", Data: map[string]interface{}{}})
				return
			}
		}

		// Validate and hash new given password.
		err = configs.ValidatePassword(newPassword)
		if err != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}
		hashedPassword, err := configs.HashPassword(newPassword)
		if err != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}

		// Change user password.
		filter := bson.M{"_id": passwordResetData.UserId}
		update := bson.M{"$set": bson.M{"auth.password_digest": hashedPassword, "auth.modified_at": now, "auth.email_verified": true}}
		err = userCollection.FindOneAndUpdate(ctx, filter, update).Decode(&user)
		if err != nil {
			c.JSON(http.StatusNotModified, responses.UserResponse{Status: http.StatusNotModified, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}

		// Send password reset successfully email to user.
		email.SendPasswordResetSuccessfulEmail(user.PrimaryEmail, user.FirstName)

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "your password been changed successfully.", Data: map[string]interface{}{}})

	}
}

// ////////////////////// START USER THUMBNAIL //////////////////////////

// UploadThumbnail - Upload user profile picture/thumbnail
// api/user/thumbnail?kind=(remote | file)&url=..
func UploadThumbnail() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		kind := c.Query("kind")
		defer cancel()

		currentId, err := auth.ExtractTokenID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, responses.UserResponse{Status: http.StatusUnauthorized, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		now := time.Now()
		filter := bson.M{"_id": currentId}

		// if user wants remote upload we proceed here
		if kind == "remote" {
			url := c.Query("url")
			uploadUrl, err := services.NewMediaUpload().RemoteUpload(models.Url{Url: url})
			if err != nil {
				c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: err.Error(), Data: map[string]interface{}{}})
				return
			}

			update := bson.M{"$set": bson.M{"thumbnail": uploadUrl, "modified_at": now}}
			_, err = userCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				c.JSON(http.StatusExpectationFailed, responses.UserResponse{Status: http.StatusExpectationFailed, Message: err.Error(), Data: map[string]interface{}{}})
				return
			}
		}

		formFile, _, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}

		uploadUrl, err := services.NewMediaUpload().FileUpload(models.File{File: formFile})
		if err != nil {
			log.Printf("Thumbnail Image upload failed - %v", err.Error())
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}

		update := bson.M{"$set": bson.M{"thumbnail": uploadUrl, "modified_at": now}}
		_, err = userCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusExpectationFailed, responses.UserResponse{Status: http.StatusExpectationFailed, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "your thumbnail has been changed successfully.", Data: map[string]interface{}{}})
	}
}

// DeleteThumbnail - Upload user profile picture/thumbnail
// api/user/thumbnail?name=thumbnail_name&type=jpg
func DeleteThumbnail() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		name := c.Query("name")
		kind := c.Query("type")
		defer cancel()

		myId, err := auth.ExtractTokenID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, responses.UserResponse{Status: http.StatusUnauthorized, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}

		_, errOnDelete := helper.ImageDeletionHelper(uploader.DestroyParams{
			PublicID:     name,
			Type:         "upload",
			ResourceType: kind,
			Invalidate:   false,
		})
		if errOnDelete != nil {
			c.JSON(http.StatusExpectationFailed, responses.UserResponse{Status: http.StatusExpectationFailed, Message: errOnDelete.Error(), Data: map[string]interface{}{}})
			return
		}

		now := time.Now()
		filter := bson.M{"_id": myId}
		update := bson.M{"$set": bson.M{"thumbnail": nil, "modified_at": now}}
		_, err = userCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			log.Printf("Thumbnail deletion failed %v", err)
			c.JSON(http.StatusExpectationFailed, responses.UserResponse{Status: http.StatusExpectationFailed, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "your thumbnail has been deleted successfully.", Data: map[string]interface{}{}})
	}
}

// ////////////////////// START USER ADDRESS //////////////////////////

// CreateUserAddress - create new user address
func CreateUserAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		var userAddress models.UserAddress
		defer cancel()

		// Validate the request body
		if err := c.BindJSON(&userAddress); err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}

		// Validate request body
		if validationErr := validate.Struct(&userAddress); validationErr != nil {
			c.JSON(http.StatusUnprocessableEntity, responses.UserResponse{Status: http.StatusUnprocessableEntity, Message: validationErr.Error(), Data: map[string]interface{}{}})
			return
		}

		// Extract current user token
		myId, err := auth.ExtractTokenID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, responses.UserResponse{Status: http.StatusUnauthorized, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}

		// create user address
		userAddress.UserId = myId
		userAddress.Id = primitive.NewObjectID()
		_, err = userAddressCollection.InsertOne(ctx, userAddress)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "address created!", Data: map[string]interface{}{}})

	}
}

// GetUserAddress - update user address
func GetUserAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		defer cancel()

		// Validate user id
		userIdStr := c.Param("userid")
		userId, err := primitive.ObjectIDFromHex(userIdStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, responses.UserResponse{Status: http.StatusUnauthorized, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}

		filter := bson.M{"user_id": userId}
		res, err := userAddressCollection.Find(ctx, filter)
		if err != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: err.Error(), Data: map[string]interface{}{}})
		}

		var userAddresses []models.UserAddress
		if err = res.All(ctx, &userAddresses); err != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": userAddresses}})

	}
}

// UpdateUserAddress - update user address
func UpdateUserAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		var userAddress models.UserAddress
		defer cancel()

		// Validate the request body
		if err := c.BindJSON(&userAddress); err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}

		// Validate request body
		if validationErr := validate.Struct(&userAddress); validationErr != nil {
			c.JSON(http.StatusUnprocessableEntity, responses.UserResponse{Status: http.StatusUnprocessableEntity, Message: validationErr.Error(), Data: map[string]interface{}{}})
			return
		}

		// Extract current user token
		myId, err := auth.ExtractTokenID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, responses.UserResponse{Status: http.StatusUnauthorized, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}

		filter := bson.M{"user_id": myId}
		update := bson.M{"$set": bson.M{"city": userAddress.City, "state": userAddress.State, "street": userAddress.Street, "postal_code": userAddress.PostalCode, "country": models.CountryNigeria}}
		_, err = userAddressCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: err.Error(), Data: map[string]interface{}{}})
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "address updated", Data: map[string]interface{}{}})

	}
}

// ////////////////////// START USER BIRTHDATE //////////////////////////

// UpdateUserBirthdate - update user birthdate
func UpdateUserBirthdate() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		var birthDate models.UserBirthdate
		defer cancel()

		errBind := c.BindJSON(&birthDate)
		if errBind != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: errBind.Error(), Data: map[string]interface{}{}})
			return
		}

		// Validate request body
		if validationErr := validate.Struct(&birthDate); validationErr != nil {
			c.JSON(http.StatusUnprocessableEntity, responses.UserResponse{Status: http.StatusUnprocessableEntity, Message: validationErr.Error(), Data: map[string]interface{}{}})
			return
		}

		myId, err := auth.ExtractTokenID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, responses.UserResponse{Status: http.StatusUnauthorized, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}
		filter := bson.M{"_id": myId}
		update := bson.M{"$set": bson.M{"birthdate.day": birthDate.Day, "birthdate.month": birthDate.Month, "birthdate.year": birthDate.Year}}
		result, errUpdateName := userCollection.UpdateOne(ctx, filter, update)
		if errUpdateName != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: errUpdateName.Error(), Data: map[string]interface{}{}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": result}})
	}
}

// UpdateUserSingleField - update user single field like Phone, Bio
// api/user/update?field=phone&value=8084051523
func UpdateUserSingleField() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		field := c.Query("field")
		value := c.Query("value")
		defer cancel()

		myId, err := auth.ExtractTokenID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, responses.UserResponse{Status: http.StatusUnauthorized, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}

		if strings.Contains(field, ".") {
			c.JSON(http.StatusUnauthorized, responses.UserResponse{Status: http.StatusUnauthorized, Message: fmt.Sprintf("No way!, %v can't contain a .", field), Data: map[string]interface{}{}})
			return
		}

		notAllowedFields := []string{"role", "login_counts", "modified_at", "created_at", "favorite_shops", "shops", "status", "referred_by_user", "address_id", "transaction_sold_count", "transaction_buy_count", "birthdate", "thumbnail", "auth", "primary_email", "login_name", "_id"}

		for _, n := range notAllowedFields {
			if strings.ToLower(field) == n {
				log.Printf("User (%v) is trying to change their %v", myId.Hex(), n)
				c.JSON(http.StatusUnauthorized, responses.UserResponse{Status: http.StatusUnauthorized, Message: fmt.Sprintf("no way!, you can't change your %v", n), Data: map[string]interface{}{}})
				return
			}
		}

		filter := bson.M{"_id": myId}
		update := bson.M{"$set": bson.M{field: value}}
		result, errUpdateName := userCollection.UpdateOne(ctx, filter, update)
		if errUpdateName != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: errUpdateName.Error(), Data: map[string]interface{}{}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": result}})
	}
}

// AddRemoveFavoriteShop - update user single field like Phone, Bio
// api/user/update?shopid=phone&value=8084051523
func AddRemoveFavoriteShop() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		shop := c.Query("shopid")
		action := c.Query("action")
		defer cancel()

		myObjectId, err := auth.ExtractTokenID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, responses.UserResponse{Status: http.StatusUnauthorized, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}

		filter := bson.M{"_id": myObjectId}
		if action != "add" {
			update := bson.M{"$push": bson.M{"favorite_shops": shop}}
			res, err := userCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: err.Error(), Data: map[string]interface{}{}})
				return
			}

			c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": res}})
			return
		}

		if action != "remove" {
			update := bson.M{"pull": bson.M{"favorite_shops": shop}}
			res, err := userCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: err.Error(), Data: map[string]interface{}{}})
				return
			}

			log.Println("Only add or remove keywords are recognize for the endpoint")
			c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": res}})
			return
		}

		c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "action not recognized", Data: map[string]interface{}{}})
	}
}

// AddWishListItem - Add to user wish list
// api/user/wishlist?listing_id=8084051523
func AddWishListItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		defer cancel()

		listingId := c.Query("listing_id")
		listingObjectId, err := primitive.ObjectIDFromHex(listingId)
		if err != nil {
			c.JSON(http.StatusUnauthorized, responses.UserResponse{Status: http.StatusUnauthorized, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}

		MyId, err := auth.ExtractTokenID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, responses.UserResponse{Status: http.StatusUnauthorized, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}

		now := time.Now()
		data := models.UserWishlist{
			ID:        primitive.NewObjectID(),
			UserID:    MyId,
			ListingId: listingObjectId,
			CreatedAt: now,
		}
		res, err := wishListCollection.InsertOne(ctx, data)
		if err != nil {
			c.JSON(http.StatusNotModified, responses.UserResponse{Status: http.StatusNotModified, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}

		result := fmt.Sprintf("New Wishlist item added with ID %v\n", res.InsertedID)
		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": result}})

	}
}

// RemoveWishListItem - Add to user wish list
// api/user/wishlist?listing_id=8084051523
func RemoveWishListItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		defer cancel()

		listingId := c.Query("listing_id")
		listingObjectId, err := primitive.ObjectIDFromHex(listingId)
		if err != nil {
			c.JSON(http.StatusUnauthorized, responses.UserResponse{Status: http.StatusUnauthorized, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}

		MyId, err := auth.ExtractTokenID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, responses.UserResponse{Status: http.StatusUnauthorized, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}

		filter := bson.M{"user_id": MyId, "listing_id": listingObjectId}
		res, err := wishListCollection.DeleteOne(ctx, filter)
		if err != nil {
			c.JSON(http.StatusNotModified, responses.UserResponse{Status: http.StatusNotModified, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}

		result := fmt.Sprintf("removed %v\n item from my Wishlist", res.DeletedCount)
		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": result}})

	}
}

// GetWishListItems - Get all wishlist items
// api/user/wishlist?limit=10&skip=0
func GetWishListItems() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		defer cancel()

		MyId, err := auth.ExtractTokenID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, responses.UserResponse{Status: http.StatusUnauthorized, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}

		paginationArgs := services.GetPaginationArgs(c)
		filter := bson.M{"user_id": MyId}
		find := options.Find().SetLimit(int64(paginationArgs.Limit)).SetSkip(int64(paginationArgs.Skip))
		result, err := wishListCollection.Find(ctx, filter, find)
		if err != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}

		var myWishLists []models.UserWishlist
		if err = result.All(ctx, &myWishLists); err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: err.Error(), Data: map[string]interface{}{}})
			return
		}

		count, err := wishListCollection.CountDocuments(ctx,
			bson.M{
				"user_id": MyId,
			})
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error counting wishlist", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponsePagination{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": myWishLists}, Pagination: responses.Pagination{
			Limit: paginationArgs.Limit,
			Skip:  paginationArgs.Skip,
			Count: count,
		}})
	}
}
