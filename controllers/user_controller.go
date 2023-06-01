package controllers

import (
	"context"
	"errors"
	"fmt"
	"github.com/cloudinary/cloudinary-go/api/uploader"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"go.mongodb.org/mongo-driver/x/bsonx"
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
		defer cancel()

		var jsonUser models.UserRegistrationBody
		if err := c.BindJSON(&jsonUser); err != nil {
			log.Printf("Error binding request body: %s\n", err.Error())
			helper.HandleError(c, http.StatusBadRequest, err, "Invalid or missing data in request body")
			return
		}

		if err := validate.Struct(&jsonUser); err != nil {
			log.Printf("Error validating request body: %s\n", err.Error())
			helper.HandleError(c, http.StatusBadRequest, err, "Invalid or missing data in request body")
			return
		}

		errEmail := configs.ValidateEmailAddress(jsonUser.Email)
		if errEmail != nil {
			log.Printf("Invalid email address from user %s with IP %s at %s: %s\n", jsonUser.FirstName, c.ClientIP(), time.Now().Format(time.RFC3339), errEmail.Error())
			helper.HandleError(c, http.StatusExpectationFailed, errEmail, "Invalid email address")
			return
		}

		err := configs.ValidatePassword(jsonUser.Password)
		if err != nil {
			log.Printf("Error validating password: %s\n", err.Error())
			helper.HandleError(c, http.StatusExpectationFailed, err, err.Error())
			return
		}

		hashedPassword, errHashPassword := configs.HashPassword(jsonUser.Password)
		if errHashPassword != nil {
			log.Printf("Error hashing password: %s\n", errHashPassword.Error())
			helper.HandleError(c, http.StatusExpectationFailed, errHashPassword, errHashPassword.Error())
			return
		}

		userAuth := models.UserAuthData{
			EmailVerified:  false,
			ModifiedAt:     time.Now(),
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
			"created_at":             time.Now(),
			"modified_at":            time.Now(),
			"last_login":             time.Now(),
			"login_counts":           0,
			"last_login_ip":          c.ClientIP(),
		}

		result, err := userCollection.InsertOne(ctx, newUser)
		if err != nil {
			log.Printf("Mongo Error: Request could not be completed %s\n", err.Error())
			helper.HandleError(c, http.StatusInternalServerError, err, "Failed to create user")
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
		defer cancel()

		var jsonUser models.UserLoginBody
		clientIP := c.ClientIP()
		now := time.Now()

		if err := c.BindJSON(&jsonUser); err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Failed to bind request body")
			return
		}

		if validationErr := validate.Struct(&jsonUser); validationErr != nil {
			helper.HandleError(c, http.StatusUnprocessableEntity, validationErr, "Invalid or missing data in request body")
			return
		}

		var validUser models.User
		if err := userCollection.FindOne(ctx, bson.M{"primary_email": jsonUser.Email}).Decode(&validUser); err != nil {
			helper.HandleError(c, http.StatusNotFound, err, "User not found")
			return
		}

		if errPasswordCheck := configs.CheckPassword(validUser.Auth.PasswordDigest, jsonUser.Password); errPasswordCheck != nil {
			helper.HandleError(c, http.StatusUnauthorized, errPasswordCheck, "Invalid password")
			return
		}

		wc := writeconcern.New(writeconcern.WMajority())
		txnOptions := options.Transaction().SetWriteConcern(wc)
		session, err := configs.DB.StartSession()
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Failed to start database session")
			return
		}
		defer session.EndSession(ctx)

		callback := func(ctx mongo.SessionContext) (interface{}, error) {
			filter := bson.M{"primary_email": validUser.PrimaryEmail}
			update := bson.M{
				"$set": bson.M{
					"last_login":    now,
					"login_counts":  validUser.LoginCounts + 1,
					"last_login_ip": clientIP,
				},
			}
			errUpdateLoginCounts := userCollection.FindOneAndUpdate(ctx, filter, update).Decode(&validUser)
			if errUpdateLoginCounts != nil {
				log.Println(errUpdateLoginCounts)
				return nil, errUpdateLoginCounts
			}

			doc := models.LoginHistory{
				Id:        primitive.NewObjectID(),
				UserUid:   validUser.Id,
				Date:      now,
				UserAgent: c.Request.UserAgent(),
				IpAddr:    clientIP,
			}
			result, errLoginHistory := loginHistoryCollection.InsertOne(ctx, doc)
			if errLoginHistory != nil {
				return result, errLoginHistory
			}

			return result, nil
		}

		_, err = session.WithTransaction(context.Background(), callback, txnOptions)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Failed to execute transaction")
			return
		}

		if err := session.CommitTransaction(context.Background()); err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Failed to commit transaction")
			return
		}
		session.EndSession(context.Background())

		tokenString, err := auth.GenerateJWT(validUser.Id.Hex(), validUser.PrimaryEmail, validUser.LoginName)
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Failed to generate JWT")
			return
		}

		email.SendNewIpLoginNotification(validUser.PrimaryEmail, validUser.LoginName, validUser.LastLoginIp, validUser.LastLogin)

		// Send new login IP notification on condition
		email.SendNewIpLoginNotification(validUser.PrimaryEmail, validUser.LoginName, validUser.LastLoginIp, validUser.LastLogin)

		helper.HandleSuccess(c, http.StatusCreated, "Authentication successful", gin.H{
			"token":          tokenString,
			"role":           validUser.Role,
			"email":          validUser.PrimaryEmail,
			"name":           validUser.FirstName,
			"thumbnail":      validUser.Thumbnail,
			"email_verified": validUser.Auth.EmailVerified,
		})
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

		helper.HandleSuccess(c, http.StatusOK, "logout successful", nil)
	}
}

func CurrentUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
	defer cancel()

	// Extract user id from request header
	userId, err := auth.ExtractTokenID(c)
	if err != nil {
		log.Printf("User with IP %v tried to gain access with an invalid user ID or token\n", c.ClientIP())
		helper.HandleError(c, http.StatusBadRequest, err, "Invalid user ID or token")
		return
	}

	user, err := services.GetUserById(ctx, userId)
	if err != nil {
		log.Printf("Logging out user with IP %v\n", c.ClientIP())
		helper.HandleError(c, http.StatusNotFound, err, "User not found")
		return
	}

	user.Auth.PasswordDigest = ""
	helper.HandleSuccess(c, http.StatusOK, "success", gin.H{"user": user})

}

// GetUserByIDOrEmail - Get user by id or email endpoint
func GetUserByIDOrEmail() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		defer cancel()

		userID := c.Query("id")
		userEmail := c.Query("email")

		// Check if either the user ID or email is missing
		if userID == "" && userEmail == "" {
			helper.HandleError(c, http.StatusBadRequest, errors.New("missing user ID or email"), "Missing user ID or email")
			return
		}

		// Prepare the filter based on the available query parameter
		filter := bson.M{}
		if userID != "" {
			filter["_id"] = userID
		} else {
			filter["email"] = userEmail
		}

		// Query the database to find the user based on the specified field and value
		var user models.User
		err := userCollection.FindOne(ctx, filter).Decode(&user)
		if err != nil {
			helper.HandleError(c, http.StatusNotFound, err, "User not found")
			return
		}

		// Return the user data in the response
		helper.HandleSuccess(c, http.StatusOK, "success", gin.H{"user": user})
	}
}

// SendVerifyEmail - api/send-verify-email?email=...&name=user_login_name
func SendVerifyEmail() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		defer cancel()

		emailCurrent := c.Query("email")
		firstName := c.Query("name")
		now := time.Now()

		// Verify current user
		userId, err := auth.ExtractTokenID(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, err.Error())
			return
		}

		// Verify current user email
		err = configs.ValidateEmailAddress(emailCurrent)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Invalid email address")
			return
		}

		// Generate secure and unique token
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
			helper.HandleError(c, http.StatusBadRequest, err, err.Error())
			return
		}

		link := fmt.Sprintf("https://khoomi.com/verify-email?token=%v&id=%v", token, userId)
		// Send welcome email
		email.SendVerifyEmailNotification(emailCurrent, firstName, link)

		helper.HandleSuccess(c, http.StatusOK, "Verification email successfully sent", nil)
	}
}

// VerifyEmail - api/send-verify-email?email=...&name=user_login_name
func VerifyEmail() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		defer cancel()

		currentId := c.Query("id")
		currentToken := c.Query("token")
		var emailVerificationData models.UserVerifyEmailToken
		var user models.User

		userId, err := primitive.ObjectIDFromHex(currentId)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Invalid user ID")
			return
		}

		// Get and delete email verification
		err = emailVerificationTokenCollection.FindOneAndDelete(ctx, bson.M{"user_uid": userId}).Decode(&emailVerificationData)
		if err != nil {
			helper.HandleError(c, http.StatusNotFound, err, "Email verification token not found")
			return
		}

		// Check if verification token has expired
		now := primitive.NewDateTimeFromTime(time.Now())
		if now.Time().Unix() > emailVerificationData.ExpiresAt.Time().Unix() {
			helper.HandleError(c, http.StatusNotFound, errors.New("email verification token has expired"), "Email verification token has expired")
			return
		}

		// Check if verification token is correct
		if currentToken != emailVerificationData.TokenDigest {
			helper.HandleError(c, http.StatusNotFound, errors.New("incorrect or expired email verification token"), "Incorrect or expired email verification token")
			return
		}

		// Change user email verification status
		filter := bson.M{"_id": emailVerificationData.UserId}
		update := bson.M{"$set": bson.M{"status": "Active", "modified_at": now, "auth.modified_at": now, "auth.email_verified": true}}
		err = userCollection.FindOneAndUpdate(ctx, filter, update).Decode(&user)
		if err != nil {
			helper.HandleError(c, http.StatusNotModified, err, "Failed to update user")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "Your email has been verified successfully.", user.PrimaryEmail)
	}
}

// UpdateFirstLastName -> Update first and last name for current user
func UpdateFirstLastName() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		defer cancel()

		var firstLastName models.FirstLastName

		myId, err := auth.ExtractTokenID(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Failed to extract user ID")
			return
		}

		// Bind JSON request body to firstLastName struct
		if err := c.BindJSON(&firstLastName); err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Invalid JSON format")
			return
		}

		// Validate first name
		validFirstName, err := regexp.MatchString("([A-Z][a-zA-Z]*)", firstLastName.FirstName)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Invalid first name format")
			return
		}
		if !validFirstName {
			helper.HandleError(c, http.StatusBadRequest, errors.New("first name should follow naming rule"), "Invalid first name format")
			return
		}

		// Validate last name
		validLastName, err := regexp.MatchString("([A-Z][a-zA-Z]*)", firstLastName.LastName)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Invalid last name format")
			return
		}
		if !validLastName {
			helper.HandleError(c, http.StatusBadRequest, errors.New("last name should follow naming rule"), "Invalid last name format")
			return
		}

		// Update user's first name and last name in the database
		filter := bson.M{"_id": myId}
		update := bson.M{"$set": bson.M{"first_name": firstLastName.FirstName, "last_name": firstLastName.LastName}}
		result, err := userCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Failed to update user's first name and last name")
			return
		}

		helper.HandleSuccess(c, http.StatusCreated, "success", result)
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
			helper.HandleError(c, http.StatusUnauthorized, err, "Failed to extract user ID")
			return
		}

		paginationArgs := services.GetPaginationArgs(c)
		filter := bson.M{"user_uid": userId}
		find := options.Find().SetLimit(int64(paginationArgs.Limit)).SetSkip(int64(paginationArgs.Skip))
		result, err := loginHistoryCollection.Find(ctx, filter, find)
		if err != nil {
			helper.HandleError(c, http.StatusNotFound, err, "Failed to find login histories")
			return
		}

		count, err := loginHistoryCollection.CountDocuments(ctx, bson.M{"user_uid": userId})
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Failed to count login histories")
			return
		}

		var loginHistory []models.LoginHistory
		if err = result.All(ctx, &loginHistory); err != nil {
			helper.HandleError(c, http.StatusNotFound, err, "Failed to decode login histories")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "success", gin.H{
			"history": loginHistory,
			"pagination": responses.Pagination{
				Limit: paginationArgs.Limit,
				Skip:  paginationArgs.Skip,
				Count: count,
			},
		})
	}
}

func DeleteLoginHistories() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		userId, err := auth.ExtractTokenID(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Failed to extract user ID")
			return
		}

		var historyIDs models.LoginHistoryIds
		if err := c.BindJSON(&historyIDs); err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Failed to bind JSON")
			return
		}

		var IdsToDelete []primitive.ObjectID
		for _, id := range historyIDs.IDs {
			objId, _ := primitive.ObjectIDFromHex(id)
			IdsToDelete = append(IdsToDelete, objId)
		}

		wc := writeconcern.New(writeconcern.WMajority())
		txnOptions := options.Transaction().SetWriteConcern(wc)
		session, err := configs.DB.StartSession()
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Failed to start database session")
			return
		}
		defer session.EndSession(ctx)

		callback := func(ctx mongo.SessionContext) (interface{}, error) {
			filter := bson.M{"_id": bson.M{"$in": IdsToDelete}, "user_uid": userId}
			result, err := loginHistoryCollection.DeleteMany(ctx, filter)
			if err != nil {
				return nil, err
			}

			return result, nil
		}

		_, err = session.WithTransaction(ctx, callback, txnOptions)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Failed to delete login histories")
			return
		}

		if err := session.CommitTransaction(ctx); err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Failed to commit transaction")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "Login histories deleted successfully", nil)
	}

}

// ////////////////////// START USER PASSWORD RESET //////////////////////////

// PasswordResetEmail - api/send-password-reset?email=borngracedd@gmail.com
func PasswordResetEmail() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		defer cancel()

		currentEmail := strings.ToLower(c.Query("email"))
		var user models.User

		err := userCollection.FindOne(ctx, bson.M{"primary_email": currentEmail}).Decode(&user)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "User with email not found")
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
			helper.HandleError(c, http.StatusBadRequest, err, "Failed to replace password reset token")
			return
		}

		link := fmt.Sprintf("https://khoomi.com/password-reset/?id=%v&token=%v", user.Id.Hex(), token)
		email.SendPasswordResetEmail(user.PrimaryEmail, user.FirstName, link)

		helper.HandleSuccess(c, http.StatusOK, "Password reset email sent successfully", nil)
	}
}

// PasswordReset - api/password-reset/userid?token=..&newpassword=..&id=user_uid
func PasswordReset() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		defer cancel()

		currentId := c.Query("id")
		currentToken := c.Query("token")
		newPassword := c.Query("newpassword")
		var passwordResetData models.UserPasswordResetToken
		var user models.User

		userId, err := primitive.ObjectIDFromHex(currentId)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Invalid userID")
			return
		}

		err = passwordResetTokenCollection.FindOneAndDelete(ctx, bson.M{"user_uid": userId}).Decode(&passwordResetData)
		if err != nil {
			helper.HandleError(c, http.StatusNotFound, err, "Failed to find or delete password reset token")
			return
		}

		now := primitive.NewDateTimeFromTime(time.Now())
		if now.Time().Unix() > passwordResetData.ExpiresAt.Time().Unix() {
			helper.HandleError(c, http.StatusNotFound, nil, "Password reset token has expired. Please restart the reset process")
			return
		}

		if currentToken != passwordResetData.TokenDigest {
			helper.HandleError(c, http.StatusNotFound, nil, "Password reset token is incorrect or expired. Please restart the reset process or use a valid token")
			return
		}

		err = configs.ValidatePassword(newPassword)
		if err != nil {
			helper.HandleError(c, http.StatusNotFound, err, "Invalid new password")
			return
		}

		hashedPassword, err := configs.HashPassword(newPassword)
		if err != nil {
			helper.HandleError(c, http.StatusNotFound, err, "Failed to hash new password")
			return
		}

		filter := bson.M{"_id": passwordResetData.UserId}
		update := bson.M{"$set": bson.M{"auth.password_digest": hashedPassword, "auth.modified_at": now, "auth.email_verified": true}}
		err = userCollection.FindOneAndUpdate(ctx, filter, update).Decode(&user)
		if err != nil {
			helper.HandleError(c, http.StatusNotModified, err, "Failed to update user password")
			return
		}

		email.SendPasswordResetSuccessfulEmail(user.PrimaryEmail, user.FirstName)

		helper.HandleSuccess(c, http.StatusOK, "success", nil)
	}
}

// ////////////////////// START USER THUMBNAIL //////////////////////////

// UploadThumbnail - Upload user profile picture/thumbnail
// api/user/thumbnail?kind=(remote | file)&url=..
func UploadThumbnail() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		defer cancel()

		kind := c.Query("kind")
		currentId, err := auth.ExtractTokenID(c)
		if err != nil {
			helper.HandleError(c, http.StatusUnauthorized, err, "Unauthorized")
			return
		}

		now := time.Now()
		filter := bson.M{"_id": currentId}

		if kind == "remote" {
			url := c.Query("url")
			uploadUrl, err := services.NewMediaUpload().RemoteUpload(models.Url{Url: url})
			if err != nil {
				helper.HandleError(c, http.StatusInternalServerError, err, "Failed to upload remote thumbnail")
				return
			}

			update := bson.M{"$set": bson.M{"thumbnail": uploadUrl, "modified_at": now}}
			_, err = userCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				helper.HandleError(c, http.StatusExpectationFailed, err, "Failed to update user's thumbnail")
				return
			}
		}

		formFile, _, err := c.Request.FormFile("file")
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Failed to retrieve uploaded file")
			return
		}

		uploadUrl, err := services.NewMediaUpload().FileUpload(models.File{File: formFile})
		if err != nil {
			log.Printf("Thumbnail Image upload failed - %v", err.Error())
			helper.HandleError(c, http.StatusInternalServerError, err, "Failed to upload file thumbnail")
			return
		}

		update := bson.M{"$set": bson.M{"thumbnail": uploadUrl, "modified_at": now}}
		_, err = userCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			helper.HandleError(c, http.StatusExpectationFailed, err, "Failed to update user's thumbnail")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "Your thumbnail has been changed successfully.", nil)
	}
}

// DeleteThumbnail - Upload user profile picture/thumbnail
// api/user/thumbnail?name=thumbnail_name&type=jpg
func DeleteThumbnail() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		defer cancel()

		name := c.Query("name")
		kind := c.Query("type")

		myId, err := auth.ExtractTokenID(c)
		if err != nil {
			helper.HandleError(c, http.StatusUnauthorized, err, "Unauthorized")
			return
		}

		_, errOnDelete := helper.ImageDeletionHelper(uploader.DestroyParams{
			PublicID:     name,
			Type:         "upload",
			ResourceType: kind,
			Invalidate:   false,
		})
		if errOnDelete != nil {
			helper.HandleError(c, http.StatusExpectationFailed, errOnDelete, "Failed to delete thumbnail image")
			return
		}

		now := time.Now()
		filter := bson.M{"_id": myId}
		update := bson.M{"$set": bson.M{"thumbnail": nil, "modified_at": now}}
		_, err = userCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			log.Printf("Thumbnail deletion failed: %v", err)
			helper.HandleError(c, http.StatusExpectationFailed, err, "Failed to update user's thumbnail")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "Your thumbnail has been deleted successfully.", nil)
	}
}

// ////////////////////// START USER ADDRESS //////////////////////////

// CreateUserAddress - create new user address
func CreateUserAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		defer cancel()

		var userAddress models.UserAddress

		// Validate the request body
		if err := c.BindJSON(&userAddress); err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Invalid request body")
			return
		}

		// Validate request body
		if validationErr := validate.Struct(&userAddress); validationErr != nil {
			helper.HandleError(c, http.StatusUnprocessableEntity, validationErr, "Validation failed")
			return
		}

		// Extract current user token
		myId, err := auth.ExtractTokenID(c)
		if err != nil {
			helper.HandleError(c, http.StatusUnauthorized, err, "Unauthorized")
			return
		}

		// create user address
		userAddress.UserId = myId
		userAddress.Id = primitive.NewObjectID()
		_, err = userAddressCollection.InsertOne(ctx, userAddress)
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Failed to create user address")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "Address created!", nil)
	}

}

// GetUserAddresses - get user address
func GetUserAddresses() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		defer cancel()

		// Validate user id
		userIdStr := c.Param("userid")
		userId, err := primitive.ObjectIDFromHex(userIdStr)
		if err != nil {
			helper.HandleError(c, http.StatusUnauthorized, err, "Invalid user ID")
			return
		}

		filter := bson.M{"user_id": userId}
		cursor, err := userAddressCollection.Find(ctx, filter)
		if err != nil {
			helper.HandleError(c, http.StatusNotFound, err, "User addresses not found")
			return
		}
		defer func() {
			if err := cursor.Close(ctx); err != nil {
				log.Println("Failed to close cursor:", err)
			}
		}()

		var userAddresses []models.UserAddress
		if err := cursor.All(ctx, &userAddresses); err != nil {
			helper.HandleError(c, http.StatusNotFound, err, "Failed to retrieve user addresses")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "Success", gin.H{"addresses": userAddresses})
	}
}

// UpdateUserAddress - update user address
func UpdateUserAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		defer cancel()

		var userAddress models.UserAddress

		// Validate the request body
		if err := c.BindJSON(&userAddress); err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Invalid request body")
			return
		}

		// Validate request body
		if validationErr := validate.Struct(&userAddress); validationErr != nil {
			helper.HandleError(c, http.StatusUnprocessableEntity, validationErr, "Validation error")
			return
		}

		// Extract current user token
		myId, err := auth.ExtractTokenID(c)
		if err != nil {
			helper.HandleError(c, http.StatusUnauthorized, err, "Unauthorized")
			return
		}

		filter := bson.M{"user_id": myId}
		update := bson.M{
			"$set": bson.M{
				"city":        userAddress.City,
				"state":       userAddress.State,
				"street":      userAddress.Street,
				"postal_code": userAddress.PostalCode,
				"country":     models.CountryNigeria,
			},
		}

		res, err := userAddressCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Failed to update user address")
			return
		}

		if res.ModifiedCount == 0 {
			helper.HandleError(c, http.StatusNotFound, errors.New("user address not found"), "User address not found")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "Address updated", nil)
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
			helper.HandleError(c, http.StatusBadRequest, errBind, "Invalid request body")
			return
		}

		// Validate request body
		if validationErr := validate.Struct(&birthDate); validationErr != nil {
			helper.HandleError(c, http.StatusUnprocessableEntity, validationErr, "Validation error")
			return
		}

		myId, err := auth.ExtractTokenID(c)
		if err != nil {
			helper.HandleError(c, http.StatusUnauthorized, err, "Unauthorized")
			return
		}
		filter := bson.M{"_id": myId}
		update := bson.M{"$set": bson.M{"birthdate.day": birthDate.Day, "birthdate.month": birthDate.Month, "birthdate.year": birthDate.Year}}
		result, errUpdateName := userCollection.UpdateOne(ctx, filter, update)
		if errUpdateName != nil {
			helper.HandleError(c, http.StatusBadRequest, errUpdateName, "Failed to update user birthdate")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "Birthdate updated", result)
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
			helper.HandleError(c, http.StatusUnauthorized, err, "Unauthorized")
			return
		}

		if strings.Contains(field, ".") {
			helper.HandleError(c, http.StatusBadRequest, fmt.Errorf("field '%s' can't contain a '.'", field), "Invalid field")
			return
		}

		notAllowedFields := []string{"role", "login_counts", "modified_at", "created_at", "favorite_shops", "shops", "status", "referred_by_user", "address_id", "transaction_sold_count", "transaction_buy_count", "birthdate", "thumbnail", "auth", "primary_email", "login_name", "_id"}

		for _, n := range notAllowedFields {
			if strings.ToLower(field) == n {
				log.Printf("User (%v) is trying to change their %v", myId.Hex(), n)
				helper.HandleError(c, http.StatusUnauthorized, fmt.Errorf("cannot change field '%s'", n), "Field not allowed")
				return
			}
		}

		filter := bson.M{"_id": myId}
		update := bson.M{"$set": bson.M{field: value}}
		result, errUpdateField := userCollection.UpdateOne(ctx, filter, update)
		if errUpdateField != nil {
			helper.HandleError(c, http.StatusBadRequest, errUpdateField, "Failed to update field")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "Field updated", result)
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
			helper.HandleError(c, http.StatusUnauthorized, err, "Unauthorized")
			return
		}

		filter := bson.M{"_id": myObjectId}
		if action == "add" {
			update := bson.M{"$push": bson.M{"favorite_shops": shop}}
			res, err := userCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				helper.HandleError(c, http.StatusBadRequest, err, "Failed to add favorite shop")
				return
			}

			helper.HandleSuccess(c, http.StatusOK, "Favorite shop added", res)
			return
		}

		if action == "remove" {
			update := bson.M{"$pull": bson.M{"favorite_shops": shop}}
			res, err := userCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				helper.HandleError(c, http.StatusBadRequest, err, "Failed to remove favorite shop")
				return
			}

			helper.HandleSuccess(c, http.StatusOK, "Favorite shop removed", res)
			return
		}

		helper.HandleError(c, http.StatusBadRequest, fmt.Errorf("action '%s' not recognized", action), "Invalid action")
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
			helper.HandleError(c, http.StatusUnauthorized, err, "Invalid listing ID")
			return
		}

		MyId, err := auth.ExtractTokenID(c)
		if err != nil {
			helper.HandleError(c, http.StatusUnauthorized, err, "Unauthorized")
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
			helper.HandleError(c, http.StatusNotModified, err, "Failed to add wishlist item")
			return
		}

		result := fmt.Sprintf("New Wishlist item added with ID %v\n", res.InsertedID)
		helper.HandleSuccess(c, http.StatusOK, "Wishlist item added", result)
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
			helper.HandleError(c, http.StatusUnauthorized, err, "Invalid listing ID")
			return
		}

		MyId, err := auth.ExtractTokenID(c)
		if err != nil {
			helper.HandleError(c, http.StatusUnauthorized, err, "Unauthorized")
			return
		}

		filter := bson.M{"user_id": MyId, "listing_id": listingObjectId}
		res, err := wishListCollection.DeleteOne(ctx, filter)
		if err != nil {
			helper.HandleError(c, http.StatusNotModified, err, "Failed to remove wishlist item")
			return
		}

		result := fmt.Sprintf("Removed %v item from my Wishlist", res.DeletedCount)
		helper.HandleSuccess(c, http.StatusOK, "Wishlist item removed", result)
	}
}

// GetUserWishlist - Get all wishlist items  api/user/wishlist?limit=10&skip=0
func GetUserWishlist() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), UserRequestTimeout*time.Second)
		defer cancel()

		MyId, err := auth.ExtractTokenID(c)
		if err != nil {
			helper.HandleError(c, http.StatusUnauthorized, err, "Unauthorized")
			return
		}

		paginationArgs := services.GetPaginationArgs(c)
		filter := bson.M{"user_id": MyId}
		find := options.Find().SetLimit(int64(paginationArgs.Limit)).SetSkip(int64(paginationArgs.Skip))
		result, err := wishListCollection.Find(ctx, filter, find)
		if err != nil {
			helper.HandleError(c, http.StatusNotFound, err, "Wishlist not found")
			return
		}

		var myWishLists []models.UserWishlist
		if err := result.All(ctx, &myWishLists); err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Internal server error")
			return
		}

		count, err := wishListCollection.CountDocuments(ctx, bson.M{"user_id": MyId})
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Error counting wishlist")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "success", gin.H{
			"wishlist": myWishLists,
			"pagination": responses.Pagination{
				Limit: paginationArgs.Limit,
				Skip:  paginationArgs.Skip,
				Count: count,
			},
		})
	}
}
