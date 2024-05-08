package controllers

import (
	"context"
	"errors"
	"fmt"
	auth "khoomi-api-io/api/internal/auth"
	"khoomi-api-io/api/internal/common"
	"khoomi-api-io/api/pkg/models"
	"khoomi-api-io/api/pkg/util"
	email "khoomi-api-io/api/web/email"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/api/uploader"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"go.mongodb.org/mongo-driver/x/bsonx"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CurrentUser get current user using userId from request headers.
func ActiveSessionUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
	defer cancel()

	var user models.User
	// Extract user id from request header
	jwt, err := auth.InitJwtClaim(c)
	if err != nil {
		util.HandleError(c, http.StatusNotFound, err, err.Error())
		return
	}

	userId, err := jwt.GetUserObjectId()
	if err != nil {
		log.Printf("User with IP %v tried to gain access with an invalid user ID or token\n", c.ClientIP())
		util.HandleError(c, http.StatusBadRequest, err, "Invalid user ID or token")
		return
	}

	err = common.UserCollection.FindOne(ctx, bson.M{"_id": userId}).Decode(&user)
	if err != nil {
		util.HandleError(c, http.StatusNotFound, err, "User not found")
		return
	}
	user.Auth.PasswordDigest = ""

	user.ConstructUserLinks()
	util.HandleSuccess(c, http.StatusOK, "success", user)
}

// CreateUser creates new user account, and send welcome and verify email notifications.
func CreateUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		var jsonUser models.UserRegistrationBody
		if err := c.BindJSON(&jsonUser); err != nil {
			util.HandleError(c, http.StatusBadRequest, err, "Invalid or missing data in request body")
			return
		}

		if err := common.Validate.Struct(&jsonUser); err != nil {
			util.HandleError(c, http.StatusBadRequest, err, "Invalid or missing data in request body")
			return
		}

		errEmail := util.ValidateEmailAddress(jsonUser.Email)
		if errEmail != nil {
			log.Printf("Invalid email address from user %s with IP %s at %s: %s\n", jsonUser.FirstName, c.ClientIP(), time.Now().Format(time.RFC3339), errEmail.Error())
			util.HandleError(c, http.StatusBadRequest, errEmail, "Invalid email address")
			return
		}

		err := util.ValidatePassword(jsonUser.Password)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err, "Error validating password")
			return
		}

		hashedPassword, errHashPassword := util.HashPassword(jsonUser.Password)
		if errHashPassword != nil {
			util.HandleError(c, http.StatusBadRequest, errHashPassword, "Error hashing password")
			return
		}

		userAuth := models.UserAuthData{
			EmailVerified:  false,
			ModifiedAt:     time.Now(),
			PasswordDigest: hashedPassword,
		}

		jsonUser.Email = strings.ToLower(jsonUser.Email)
		userId := primitive.NewObjectID()
		newUser := bson.M{
			"_id":                         userId,
			"login_name":                  common.GenerateRandomUsername(),
			"primary_email":               jsonUser.Email,
			"first_name":                  jsonUser.FirstName,
			"last_name":                   jsonUser.LastName,
			"auth":                        userAuth,
			"thumbnail":                   common.DefaultUserThumbnail,
			"bio":                         bsonx.Null(),
			"phone":                       bsonx.Null(),
			"birthdate":                   models.UserBirthdate{},
			"is_seller":                   false,
			"transaction_buy_count":       0,
			"transaction_sold_count":      0,
			"referred_by_user":            bsonx.Null(),
			"role":                        models.Regular,
			"status":                      models.Inactive,
			"shop_id":                     bsonx.Null(),
			"favorite_shops":              []string{},
			"created_at":                  now,
			"modified_at":                 now,
			"last_login":                  now,
			"login_counts":                0,
			"last_login_ip":               c.ClientIP(),
			"allow_login_ip_notification": true,
		}

		result, err := common.UserCollection.InsertOne(ctx, newUser)
		if err != nil {
			writeException, ok := err.(mongo.WriteException)
			if ok {
				for _, writeError := range writeException.WriteErrors {
					if writeError.Code == common.MongoDuplicateKeyCode {
						log.Printf("User with email already exists: %s\n", writeError.Message)
						util.HandleError(c, http.StatusBadRequest, writeError, "User with email already exists")
						return
					}
				}
			}

			log.Printf("Mongo Error: Request could not be completed %s\n", err.Error())
			util.HandleError(c, http.StatusInternalServerError, err, "Request could not be completed")
			return
		}
		notification := models.Notification{
			ID:               primitive.NewObjectID(),
			UserID:           userId,
			NewMessage:       true,
			NewFollower:      true,
			ListingExpNotice: true,
			SellerActivity:   true,
			NewsAndFeatures:  true,
		}

		_, err = common.NotificationCollection.InsertOne(ctx, notification)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err, "Error creating notification")
			return
		}

		// Send welcome email.
		email.SendWelcomeEmail(jsonUser.Email, jsonUser.FirstName)

		util.HandleSuccess(c, http.StatusOK, "signup successful", result.InsertedID)
	}
}

// HandleUserAuthentication authenticates new user session while sending necessary notifications depending on the cases. e.g new IP
func HandleUserAuthentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		var jsonUser models.UserLoginBody
		clientIP := c.ClientIP()
		now := time.Now()

		if err := c.BindJSON(&jsonUser); err != nil {
			util.HandleError(c, http.StatusBadRequest, err, "Failed to bind request body")
			return
		}

		if err := common.Validate.Struct(&jsonUser); err != nil {
			log.Println(err)
			util.HandleError(c, http.StatusBadRequest, err, "Invalid or missing data in request body")
			return
		}

		var validUser models.User
		if err := common.UserCollection.FindOne(ctx, bson.M{"primary_email": jsonUser.Email}).Decode(&validUser); err != nil {
			util.HandleError(c, http.StatusNotFound, err, "User not found")
			return
		}

		if errPasswordCheck := util.CheckPassword(validUser.Auth.PasswordDigest, jsonUser.Password); errPasswordCheck != nil {
			util.HandleError(c, http.StatusUnauthorized, errPasswordCheck, "Invalid password")
			return
		}

		wc := writeconcern.New(writeconcern.WMajority())
		txnOptions := options.Transaction().SetWriteConcern(wc)
		session, err := util.DB.StartSession()
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err, "Failed to start database session")
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
			errUpdateLoginCounts := common.UserCollection.FindOneAndUpdate(ctx, filter, update).Decode(&validUser)
			if errUpdateLoginCounts != nil {
				return nil, errUpdateLoginCounts
			}

			doc := models.LoginHistory{
				Id:        primitive.NewObjectID(),
				UserUid:   validUser.Id,
				Date:      now,
				UserAgent: c.Request.UserAgent(),
				IpAddr:    clientIP,
			}
			result, errLoginHistory := common.LoginHistoryCollection.InsertOne(ctx, doc)
			if errLoginHistory != nil {
				return result, errLoginHistory
			}

			return result, nil
		}

		_, err = session.WithTransaction(ctx, callback, txnOptions)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err, "Failed to execute transaction")
			return
		}

		if err := session.CommitTransaction(ctx); err != nil {
			util.HandleError(c, http.StatusInternalServerError, err, "Failed to commit transaction")
			return
		}
		session.EndSession(ctx)

		accessTokenString, accessTokenExp, err := auth.GenerateJWT(validUser.Id.Hex(), validUser.PrimaryEmail, validUser.LoginName, validUser.IsSeller)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err, "Failed to generate JWT")
			return
		}

		refreshTokenString, err := auth.GenerateRefreshJWT(validUser.Id.Hex(), validUser.PrimaryEmail, validUser.LoginName, validUser.IsSeller)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err, "Failed to generate refresh JWT")
			return
		}

		// Send new login IP notification on condition
		if validUser.AllowLoginIpNotification && validUser.LastLoginIp != clientIP {
			email.SendNewIpLoginNotification(validUser.PrimaryEmail, validUser.LoginName, validUser.LastLoginIp, validUser.LastLogin)
		}

		// Send verify email
		// Generate secure and unique token
		token := auth.GenerateSecureToken(8)

		expirationTime := now.Add(common.VERIFICATION_EMAIL_EXPIRATION_TIME)
		verifyEmail := models.UserVerifyEmailToken{
			UserId:      validUser.Id,
			TokenDigest: token,
			CreatedAt:   primitive.NewDateTimeFromTime(now),
			ExpiresAt:   primitive.NewDateTimeFromTime(expirationTime),
		}
		opts := options.Replace().SetUpsert(true)
		filter := bson.M{"user_uid": validUser.Id}
		_, err = common.EmailVerificationTokenCollection.ReplaceOne(ctx, filter, verifyEmail, opts)
		if err != nil {
			log.Printf("error sending verification email for user: %v, error: %v", validUser.PrimaryEmail, err)
		}

		// Send verification email if user's email is not verified.
		if !validUser.Auth.EmailVerified {
			link := fmt.Sprintf("https://khoomi.com/verify-email?token=%v&id=%v", token, validUser.Id.Hex())
			email.SendVerifyEmailNotification(validUser.PrimaryEmail, validUser.FirstName, link)
		}

		util.HandleSuccess(c, http.StatusOK, "Authentication successful", gin.H{
			"_id":              validUser.Id.Hex(),
			"access_token":     accessTokenString,
			"refresh_token":    refreshTokenString,
			"access_token_exp": accessTokenExp,
			"role":             validUser.Role,
			"email":            validUser.PrimaryEmail,
			"first_name":       validUser.FirstName,
			"last_name":        validUser.LastName,
			"login_name":       validUser.LoginName,
			"thumbnail":        validUser.Thumbnail,
			"email_verified":   validUser.Auth.EmailVerified,
			"is_seller":        validUser.IsSeller,
		})
	}
}

// RefreshToken handles auth token refreshments
func RefreshToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		var payload models.RefreshTokenPayload

		if err := c.BindJSON(&payload); err != nil {
			util.HandleError(c, http.StatusBadRequest, err, "Invalid request body")
			return
		}

		refreshClaims, err := auth.ValidateRefreshToken(payload.Token)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err, "Invalid refresh token")
			return
		}

		userObjectId, err := primitive.ObjectIDFromHex(refreshClaims.Id)
		if err != nil {
			log.Println(err)
			util.HandleError(c, http.StatusNotFound, err, "User not found")
			return
		}

		res := common.UserCollection.FindOne(ctx, bson.M{"_id": userObjectId})
		if res.Err() != nil {
			if res.Err() == mongo.ErrNoDocuments {
				util.HandleError(c, http.StatusNotFound, res.Err(), "User not found")
				return
			}
			util.HandleError(c, http.StatusInternalServerError, res.Err(), "Internal Server Error")
			return
		}

		accessToken, accessTokenExp, err := auth.GenerateJWT(refreshClaims.Id, refreshClaims.Email, refreshClaims.LoginName, refreshClaims.IsSeller)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err, "Failed to generate access token")
			return
		}

		newRefreshToken, err := auth.GenerateRefreshJWT(refreshClaims.Id, refreshClaims.Email, refreshClaims.LoginName, refreshClaims.IsSeller)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err, "Failed to generate refresh token")
			return
		}

		util.HandleSuccess(c, http.StatusOK, "token refreshed",
			gin.H{"access_token": accessToken,
				"access_token_exp": accessTokenExp,
				"refresh_token":    newRefreshToken})

	}
}

// Logout - Log user out and invalidate session key
func Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := auth.ExtractToken(c)

		log.Printf("Logging user with ip %v out\n", c.ClientIP())
		_ = auth.InvalidateToken(util.REDIS, token)

		util.HandleSuccess(c, http.StatusOK, "logout successful", nil)

	}
}

// SendDeleteUserAccount -> Delete current user account
func SendDeleteUserAccount() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		auth, err := auth.InitJwtClaim(c)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err, "action is for authorized users")
			return
		}
		userId, err := auth.GetUserObjectId()
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err, "action is for authorized users")
			return
		}

		_, err = common.UserDeletionCollection.InsertOne(ctx, bson.M{"user_id": userId, "created_at": time.Now()})
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err, "error while requesting for account deletion")
			return
		}

		util.HandleSuccess(c, http.StatusOK, "account is now pending deletion", gin.H{"_id": auth.Id})
	}
}

// IsAccountPendingDeletion checks if current user account is pending deletion
func IsAccountPendingDeletion() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		userId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err, "action is for authorized users")
			return
		}

		var account models.AccountDeletionRequested
		err = common.UserDeletionCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&account)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusOK, gin.H{"pending_deletion": false})
				return
			}
			util.HandleError(c, http.StatusInternalServerError, err, "error while checking account deletion status")
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Success", gin.H{"_id": account.ID})
	}
}

// CancelDeleteUserAccount cancels delete user account request.
func CancelDeleteUserAccount() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		userId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err, "action is for authorized users")
			return
		}

		_, err = common.UserDeletionCollection.DeleteOne(ctx, bson.M{"user_id": userId})
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err, "error while cancelling account deletion request")
			return
		}

		util.HandleSuccess(c, http.StatusOK, "account deletion request cancelled", gin.H{"_id": userId.Hex()})
	}
}

// ChangePassword changes active user's password.
func ChangePassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		// Verify current user session
		userId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err, err.Error())
			return
		}

		var newPasswordFromRequest models.NewPasswordRequest

		if err := c.BindJSON(&newPasswordFromRequest); err != nil {
			util.HandleError(c, http.StatusInternalServerError, err, "failed to bind request body")
			return
		}

		var validUser models.User
		if err := common.UserCollection.FindOne(ctx, bson.M{"_id": userId}).Decode(&validUser); err != nil {
			util.HandleError(c, http.StatusUnauthorized, err, "User not found")
			return
		}

		if errPasswordCheck := util.CheckPassword(validUser.Auth.PasswordDigest, newPasswordFromRequest.CurrentPassword); errPasswordCheck != nil {
			util.HandleError(c, http.StatusUnauthorized, errPasswordCheck, "Invalid current password")
			return
		}

		if validationErr := common.Validate.Struct(&newPasswordFromRequest); validationErr != nil {
			util.HandleError(c, http.StatusUnprocessableEntity, validationErr, "invalid or missing data in request body")
			return
		}

		// validate password.
		err = util.ValidatePassword(newPasswordFromRequest.NewPassword)
		if err != nil {
			log.Printf("error validating password: %s\n", err.Error())
			util.HandleError(c, http.StatusExpectationFailed, err, err.Error())
			return
		}

		// hash password before saving to storage.
		hashedPassword, errHashPassword := util.HashPassword(newPasswordFromRequest.NewPassword)
		if errHashPassword != nil {
			log.Printf("error hashing password: %s\n", errHashPassword.Error())
			util.HandleError(c, http.StatusExpectationFailed, errHashPassword, errHashPassword.Error())
			return
		}

		var user models.User
		// change user pasword for userid
		filter := bson.M{"_id": userId}
		update := bson.M{"$set": bson.M{"auth.password_digest": hashedPassword, "modified_at": time.Now()}}
		err = common.UserCollection.FindOneAndUpdate(ctx, filter, update).Decode(&user)

		if err != nil {
			errStr := err.Error()
			log.Printf("user id, %v doesn't belong to a user on Khoomi %v", userId.String(), errStr)
			util.HandleError(c, http.StatusExpectationFailed, errHashPassword, errStr)
			return
		}

		util.HandleSuccess(c, http.StatusOK, "your password has been updated successfuly", user)
	}
}

// GetUser retrieves a user based on their user ID or username.
// It accepts a user ID in the URL path and attempts to retrieve a user with a matching
// ObjectID. If the user ID is not a valid ObjectID, it attempts to find the user by username.
func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		var filter bson.M
		userID := c.Param("userid")
		if primitive.IsValidObjectID(userID) {
			// If shopid is a valid object ID string
			userObjectID, e := primitive.ObjectIDFromHex(userID)
			if e != nil {
				util.HandleError(c, http.StatusBadRequest, e, "invalid user id was provided")
				return
			}

			filter = bson.M{"_id": userObjectID}
		} else {
			filter = bson.M{"username": userID}
		}

		// Query the database to find the user based on the specified field and value
		var user models.User
		err := common.UserCollection.FindOne(ctx, filter).Decode(&user)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err, "User not found")
			return
		}

		user.ConstructUserLinks()
		// Return the user data in the response
		util.HandleSuccess(c, http.StatusOK, "success", user)
	}
}

// SendVerifyEmail sends a verification email notication to a given user's email.
func SendVerifyEmail() gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		emailCurrent := c.Query("email")
		firstName := c.Query("name")

		// Verify current user
		auth_, err := auth.InitJwtClaim(c)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err, "action is for authorized users")
			return
		}
		userId, err := auth_.GetUserObjectId()
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err, "action is for authorized users")
			return
		}

		// Verify current user email
		err = util.ValidateEmailAddress(emailCurrent)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err, "Invalid email address")
			return
		}

		// Generate secure and unique token
		token := auth.GenerateSecureToken(8)

		expirationTime := now.Add(common.VERIFICATION_EMAIL_EXPIRATION_TIME)
		verifyEmail := models.UserVerifyEmailToken{
			UserId:      userId,
			TokenDigest: token,
			CreatedAt:   primitive.NewDateTimeFromTime(now),
			ExpiresAt:   primitive.NewDateTimeFromTime(expirationTime),
		}
		opts := options.Replace().SetUpsert(true)
		filter := bson.M{"user_uid": userId}
		_, err = common.EmailVerificationTokenCollection.ReplaceOne(ctx, filter, verifyEmail, opts)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err, err.Error())
			return
		}

		link := fmt.Sprintf("https://khoomi.com/verify-email?token=%v&id=%v", token, userId)
		email.SendVerifyEmailNotification(emailCurrent, firstName, link)

		util.HandleSuccess(c, http.StatusOK, "Verification email successfully sent", gin.H{"_id": userId.Hex()})
	}
}

// VerifyEmail - api/send-verify-email?email=...&name=user_login_name
func VerifyEmail() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		currentId := c.Query("id")
		currentToken := c.Query("token")
		var emailVerificationData models.UserVerifyEmailToken
		var user models.User

		userId, err := primitive.ObjectIDFromHex(currentId)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err, "Invalid user ID")
			return
		}

		// Get and delete email verification
		err = common.EmailVerificationTokenCollection.FindOneAndDelete(ctx, bson.M{"user_uid": userId}).Decode(&emailVerificationData)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err, "Email verification token not found")
			return
		}

		// Check if verification token has expired
		now := primitive.NewDateTimeFromTime(time.Now())
		if now.Time().Unix() > emailVerificationData.ExpiresAt.Time().Unix() {
			util.HandleError(c, http.StatusNotFound, errors.New("email verification token has expired"), "Email verification token has expired")
			return
		}

		// Check if verification token is correct
		if currentToken != emailVerificationData.TokenDigest {
			util.HandleError(c, http.StatusNotFound, errors.New("incorrect or expired email verification token"), "Incorrect or expired email verification token")
			return
		}

		// Change user email verification status
		filter := bson.M{"_id": emailVerificationData.UserId}
		update := bson.M{"$set": bson.M{"status": "Active", "modified_at": now, "auth.modified_at": now, "auth.email_verified": true}}
		err = common.UserCollection.FindOneAndUpdate(ctx, filter, update).Decode(&user)
		if err != nil {
			util.HandleError(c, http.StatusNotModified, err, "Failed to update user")
			return
		}

		email.SendEmailVerificationSuccessNotification(user.PrimaryEmail, user.FirstName)
		util.HandleSuccess(c, http.StatusOK, "Your email has been verified successfully.", gin.H{"_id": user.Id})
	}
}

// UpdateMyProfile updates the email, thumbnail, first and last name for the current user.
func UpdateMyProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		auth, err := auth.InitJwtClaim(c)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err, "action is for authorized users")
			return
		}
		userId, err := auth.GetUserObjectId()
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err, "action is for authorized users")
			return
		}

		updateData := bson.M{}

		if firstName := c.Request.FormValue("first_name"); firstName != "" {
			if err := common.ValidateNameFormat(firstName); err != nil {
				util.HandleError(c, http.StatusBadRequest, err, "Invalid first name format")
				return
			}
			updateData["first_name"] = firstName
		}

		if lastName := c.Request.FormValue("last_name"); lastName != "" {
			if err := common.ValidateNameFormat(lastName); err != nil {
				util.HandleError(c, http.StatusBadRequest, err, "Invalid last name format")
				return
			}
			updateData["last_name"] = lastName
		}

		if email := c.Request.FormValue("email"); email != "" {
			updateData["primary_email"] = email
		}

		var uploadResult uploader.UploadResult
		if fileHeader, err := c.FormFile("image"); err == nil {
			file, err := fileHeader.Open()
			if err != nil {
				util.HandleError(c, http.StatusInternalServerError, err, "Failed to retrieve uploaded file")
				return
			}
			defer file.Close()

			uploadResult, err = util.FileUpload(models.File{File: file})
			if err != nil {
				log.Printf("Thumbnail Image upload failed - %v", err.Error())
				util.HandleError(c, http.StatusInternalServerError, err, "Failed to upload file thumbnail")
				return
			}
			updateData["thumbnail"] = uploadResult.SecureURL
		} else if err != http.ErrMissingFile {
			util.HandleError(c, http.StatusInternalServerError, err, "Failed to retrieve uploaded file")
			return
		}

		if len(updateData) == 0 {
			// delete media
			_, err = util.DestroyMedia(uploadResult.PublicID)
			log.Println(err)
			// return error
			util.HandleError(c, http.StatusBadRequest, errors.New("no update data provided"), "No update data provided")
			return
		}

		updateData["modified_at"] = time.Now()

		filter := bson.M{"_id": userId}
		update := bson.M{"$set": updateData}

		_, err = common.UserCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			util.HandleError(c, http.StatusExpectationFailed, err, "Failed to update user's profile")
			return
		}

		util.HandleSuccess(c, http.StatusCreated, "Profile updated successfully", gin.H{"_id": auth.Id})
	}
}

// ////////////////////// START USER LOGIN HISTORY //////////////////////////

// GetLoginHistories - Get user login histories (/api/users/:userId/login-history?limit=50&skip=0)
func GetLoginHistories() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		userId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err, err.Error())
			return
		}

		paginationArgs := common.GetPaginationArgs(c)
		filter := bson.M{"user_uid": userId}
		findOptions := options.Find().
			SetLimit(int64(paginationArgs.Limit)).
			SetSkip(int64(paginationArgs.Skip)).
			SetSort(util.GetLoginHistorySortBson(paginationArgs.Sort))

		result, err := common.LoginHistoryCollection.Find(ctx, filter, findOptions)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err, "Failed to find login histories")
			return
		}

		count, err := common.LoginHistoryCollection.CountDocuments(ctx, filter)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err, "Failed to count login histories")
			return
		}

		var loginHistory []models.LoginHistory
		if err = result.All(ctx, &loginHistory); err != nil {
			util.HandleError(c, http.StatusNotFound, err, "Failed to decode login histories")
			return
		}

		util.HandleSuccessMeta(c, http.StatusOK, "success", loginHistory, gin.H{
			"pagination": util.Pagination{
				Limit: paginationArgs.Limit,
				Skip:  paginationArgs.Skip,
				Count: count,
			},
		})
	}
}

// DeleteLoginHistories - Get user login histories (/api/users/:userId/login-history?limit=50&skip=0)
func DeleteLoginHistories() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		userId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err, err.Error())
			return
		}

		var historyIDs models.LoginHistoryIds
		if err := c.BindJSON(&historyIDs); err != nil {
			util.HandleError(c, http.StatusBadRequest, err, "Failed to bind JSON")
			return
		}

		var IdsToDelete []primitive.ObjectID
		for _, id := range historyIDs.IDs {
			objId, _ := primitive.ObjectIDFromHex(id)
			IdsToDelete = append(IdsToDelete, objId)
		}

		wc := writeconcern.New(writeconcern.WMajority())
		txnOptions := options.Transaction().SetWriteConcern(wc)
		session, err := util.DB.StartSession()
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err, "Failed to start database session")
			return
		}
		defer session.EndSession(ctx)

		callback := func(ctx mongo.SessionContext) (interface{}, error) {
			filter := bson.M{"_id": bson.M{"$in": IdsToDelete}, "user_uid": userId}
			result, err := common.LoginHistoryCollection.DeleteMany(ctx, filter)
			if err != nil {
				return nil, err
			}

			return result, nil
		}

		_, err = session.WithTransaction(ctx, callback, txnOptions)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err, "Failed to delete login histories")
			return
		}

		if err := session.CommitTransaction(ctx); err != nil {
			util.HandleError(c, http.StatusInternalServerError, err, "Failed to commit transaction")
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Login histories deleted successfully", userId.Hex())
	}
}

// ////////////////////// START USER PASSWORD RESET //////////////////////////

// PasswordResetEmail - api/send-password-reset?email=borngracedd@gmail.com
func PasswordResetEmail() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		currentEmail := strings.ToLower(c.Query("email"))
		var user models.User

		err := common.UserCollection.FindOne(ctx, bson.M{"primary_email": currentEmail}).Decode(&user)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err, "User with email not found")
			return
		}

		token := auth.GenerateSecureToken(8)
		now := time.Now()
		expirationTime := now.Add(common.VERIFICATION_EMAIL_EXPIRATION_TIME)
		passwordReset := models.UserPasswordResetToken{
			UserId:      user.Id,
			TokenDigest: token,
			CreatedAt:   primitive.NewDateTimeFromTime(now),
			ExpiresAt:   primitive.NewDateTimeFromTime(expirationTime),
		}

		opts := options.Replace().SetUpsert(true)
		filter := bson.M{"user_uid": user.Id}
		_, err = common.PasswordResetTokenCollection.ReplaceOne(ctx, filter, passwordReset, opts)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err, "Failed to replace password reset token")
			return
		}

		link := fmt.Sprintf("https://khoomi.com/password-reset/?id=%v&token=%v", user.Id.Hex(), token)
		email.SendPasswordResetEmail(user.PrimaryEmail, user.FirstName, link)

		util.HandleSuccess(c, http.StatusOK, "Password reset email sent successfully", user.Id.Hex())
	}
}

// PasswordReset - api/password-reset/userid?token=..&newpassword=..&id=user_uid
func PasswordReset() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		currentId := c.Query("id")
		currentToken := c.Query("token")
		newPassword := c.Query("newpassword")
		var passwordResetData models.UserPasswordResetToken
		var user models.User

		userId, err := primitive.ObjectIDFromHex(currentId)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err, "Invalid userID")
			return
		}

		err = common.PasswordResetTokenCollection.FindOneAndDelete(ctx, bson.M{"user_uid": userId}).Decode(&passwordResetData)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err, "Failed to find or delete password reset token")
			return
		}

		now := primitive.NewDateTimeFromTime(time.Now())
		if now.Time().Unix() > passwordResetData.ExpiresAt.Time().Unix() {
			util.HandleError(c, http.StatusNotFound, nil, "Password reset token has expired. Please restart the reset process")
			return
		}

		if currentToken != passwordResetData.TokenDigest {
			util.HandleError(c, http.StatusNotFound, nil, "Password reset token is incorrect or expired. Please restart the reset process or use a valid token")
			return
		}

		err = util.ValidatePassword(newPassword)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err, "Invalid new password")
			return
		}

		hashedPassword, err := util.HashPassword(newPassword)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err, "Failed to hash new password")
			return
		}

		filter := bson.M{"_id": passwordResetData.UserId}
		update := bson.M{"$set": bson.M{"auth.password_digest": hashedPassword, "auth.modified_at": now, "auth.email_verified": true}}
		err = common.UserCollection.FindOneAndUpdate(ctx, filter, update).Decode(&user)
		if err != nil {
			util.HandleError(c, http.StatusNotModified, err, "Failed to update user password")
			return
		}

		email.SendPasswordResetSuccessfulEmail(user.PrimaryEmail, user.FirstName)

		util.HandleSuccess(c, http.StatusOK, "success", user.Id.Hex())
	}
}

// ////////////////////// START USER THUMBNAIL //////////////////////////

// UploadThumbnail - Upload user profile picture/thumbnail
// api/user/thumbnail?remote_addr=..
func UploadThumbnail() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		remoteAddr := c.Query("remote_addr")
		currentId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err, err.Error())
			return
		}

		now := time.Now()
		filter := bson.M{"_id": currentId}

		var uploadResult uploader.UploadResult
		var update bson.M
		if remoteAddr != "" {
			uploadResult, err = util.RemoteUpload(models.Url{Url: remoteAddr})
			if err != nil {
				util.HandleError(c, http.StatusInternalServerError, err, "Failed to upload remote thumbnail")
				return
			}

			update = bson.M{"$set": bson.M{"thumbnail": uploadResult.SecureURL, "modified_at": now}}
		} else {
			formFile, _, err := c.Request.FormFile("file")
			if err != nil {
				util.HandleError(c, http.StatusInternalServerError, err, "Failed to retrieve uploaded file")
				return
			}
			uploadResult, err = util.FileUpload(models.File{File: formFile})
			if err != nil {
				log.Printf("Thumbnail Image upload failed - %v", err.Error())
				util.HandleError(c, http.StatusInternalServerError, err, "Failed to upload file thumbnail")
				return
			}
			update = bson.M{"$set": bson.M{"thumbnail": uploadResult.SecureURL, "modified_at": now}}
		}

		_, err = common.UserCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			// delete media
			_, err = util.DestroyMedia(uploadResult.PublicID)
			log.Println(err)
			// return error
			util.HandleError(c, http.StatusExpectationFailed, err, "Failed to update user's thumbnail")
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Your thumbnail has been changed successfully.", currentId.Hex())
	}
}

// DeleteThumbnail - delete user profile picture/thumbnail
// api/user/thumbnail
func DeleteThumbnail() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		myId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err, err.Error())
			return
		}

		var user models.User
		now := time.Now()
		filter := bson.M{"_id": myId}
		update := bson.M{"$set": bson.M{"thumbnail": nil, "modified_at": now}}
		err = common.UserCollection.FindOneAndUpdate(ctx, filter, update).Decode(&user)
		if err != nil {
			log.Printf("Thumbnail deletion failed: %v", err)
			util.HandleError(c, http.StatusExpectationFailed, err, "Failed to update user's thumbnail")
			return
		}

		filename, extension, err := extractFilenameAndExtension(user.Thumbnail)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err, "Internal server error. Please try again later")
			return
		}

		_, errOnDelete := util.ImageDeletionHelper(uploader.DestroyParams{
			PublicID:     filename,
			Type:         "upload",
			ResourceType: extension,
			Invalidate:   true,
		})
		if errOnDelete != nil {
			util.HandleError(c, http.StatusExpectationFailed, errOnDelete, "Failed to delete thumbnail image")
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Your thumbnail has been deleted successfully.", user.Id.Hex())
	}
}

func extractFilenameAndExtension(urlString string) (filename, extension string, err error) {
	// Parse the URL
	parsedURL, err := url.Parse(urlString)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse URL: %w", err)
	}

	// Extract the filename from the URL path
	filenameWithExtension := filepath.Base(parsedURL.Path)

	// Split the filename and extension
	name := filenameWithExtension[:len(filenameWithExtension)-len(filepath.Ext(filenameWithExtension))]
	ext := filepath.Ext(filenameWithExtension)

	return name, ext, nil
}

// ////////////////////// START USER ADDRESS //////////////////////////

// CreateUserAddress - create new user address
func CreateUserAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		var userAddress models.UserAddressExcerpt

		// Validate the request body
		if err := c.BindJSON(&userAddress); err != nil {
			util.HandleError(c, http.StatusBadRequest, err, "Invalid request body")
			return
		}

		log.Println(userAddress)
		// Validate request body
		if validationErr := common.Validate.Struct(&userAddress); validationErr != nil {
			util.HandleError(c, http.StatusUnprocessableEntity, validationErr, "Validation failed")
			return
		}

		// Extract current user token
		myId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err, err.Error())
			return
		}

		// create user address
		addressId := primitive.NewObjectID()
		userAddressTemp := models.UserAddress{
			Id:                       addressId,
			UserId:                   myId,
			City:                     userAddress.City,
			State:                    userAddress.State,
			Street:                   userAddress.Street,
			PostalCode:               userAddress.PostalCode,
			Country:                  models.CountryNigeria,
			IsDefaultShippingAddress: userAddress.IsDefaultShippingAddress,
		}

		count, err := common.UserAddressCollection.CountDocuments(ctx, bson.M{"user_id": myId})
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err, "Error counting current payment information")
			return
		}

		if count >= 5 {
			util.HandleError(c, http.StatusInsufficientStorage, errors.New("max allowed addresses reached. please delete other address to accommodate a new one"), "max allowed payment information reached")
			return
		}

		if userAddress.IsDefaultShippingAddress {
			// Set IsDefaultShippingAddress to false for other addresses belonging to the user
			err = setOtherAddressesToFalse(ctx, myId, addressId)
			if err != nil {
				util.HandleError(c, http.StatusInternalServerError, err, "Failed to update user addresses")
				return
			}

		}

		_, err = common.UserAddressCollection.InsertOne(ctx, userAddressTemp)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err, "Failed to create user address")
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Address created!", userAddressTemp.Id.Hex())
	}

}

// GetUserAddresses - get user address
func GetUserAddresses() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		// Validate user id
		userIdStr := c.Param("userid")
		userId, err := primitive.ObjectIDFromHex(userIdStr)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err, "Invalid user ID")
			return
		}

		filter := bson.M{"user_id": userId}
		cursor, err := common.UserAddressCollection.Find(ctx, filter)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err, "User addresses not found")
			return
		}
		defer func() {
			if err := cursor.Close(ctx); err != nil {
				log.Println("Failed to close cursor:", err)
			}
		}()

		var userAddresses []models.UserAddress
		if err := cursor.All(ctx, &userAddresses); err != nil {
			util.HandleError(c, http.StatusNotFound, err, "Failed to retrieve user addresses")
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Success", userAddresses)
	}
}

// UpdateUserAddress - update user address
func UpdateUserAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		var userAddress models.UserAddressExcerpt

		// Validate the request body
		if err := c.BindJSON(&userAddress); err != nil {
			util.HandleError(c, http.StatusBadRequest, err, "Invalid request body")
			return
		}

		// Validate request body
		if validationErr := common.Validate.Struct(&userAddress); validationErr != nil {
			util.HandleError(c, http.StatusUnprocessableEntity, validationErr, "Validation error")
			return
		}

		// Extract current address Id
		addressId := c.Param("id")
		addressObjectId, err := primitive.ObjectIDFromHex(addressId)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err, err.Error())
			return
		}

		// Extract current user token
		myId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err, err.Error())
			return
		}

		// Set IsDefaultShippingAddress to false for other addresses belonging to the user
		if userAddress.IsDefaultShippingAddress {
			err = setOtherAddressesToFalse(ctx, myId, addressObjectId)
			if err != nil {
				util.HandleError(c, http.StatusInternalServerError, err, "Failed to update user addresses")
				return
			}
		}

		filter := bson.M{"user_id": myId, "_id": addressObjectId}
		update := bson.M{
			"$set": bson.M{
				"city":                        userAddress.City,
				"state":                       userAddress.State,
				"street":                      userAddress.Street,
				"postal_code":                 userAddress.PostalCode,
				"country":                     models.CountryNigeria,
				"is_default_shipping_address": userAddress.IsDefaultShippingAddress,
			},
		}

		res, err := common.UserAddressCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err, "Failed to update user address")
			return
		}

		if res.ModifiedCount == 0 {
			util.HandleError(c, http.StatusNotFound, errors.New("user address not found"), "User address not found")
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Address updated", addressId)
	}
}

// UpdateUserAddress - update user address
func DeleteUserAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		// Extract current address Id
		addressId := c.Param("id")
		addressObjectId, err := primitive.ObjectIDFromHex(addressId)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err, err.Error())
			return
		}

		// Extract current user token
		myId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err, err.Error())
			return
		}

		// Set IsDefaultShippingAddress to false for other addresses belonging to the user
		err = setOtherAddressesToFalse(ctx, myId, addressObjectId)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err, "Failed to update user addresses")
			return
		}

		filter := bson.M{"user_id": myId, "_id": addressObjectId}
		res, err := common.UserAddressCollection.DeleteOne(ctx, filter)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err, "Failed to delete user address")
			return
		}

		if res.DeletedCount == 0 {
			util.HandleError(c, http.StatusNotFound, errors.New("user address not found"), "User address not found")
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Address deleted", addressId)
	}
}

// SetOtherAddressesToFalse sets IsDefaultShippingAddress to false for other addresses belonging to the user
func setOtherAddressesToFalse(ctx context.Context, userId primitive.ObjectID, addressId primitive.ObjectID) error {
	filter := bson.M{
		"user_id":                     userId,
		"_id":                         bson.M{"$ne": addressId},
		"is_default_shipping_address": true,
	}

	update := bson.M{
		"$set": bson.M{"is_default_shipping_address": false},
	}

	_, err := common.UserAddressCollection.UpdateMany(ctx, filter, update)
	return err
}

// ////////////////////// START USER BIRTHDATE //////////////////////////

// UpdateUserBirthdate - update user birthdate
func UpdateUserBirthdate() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		var birthDate models.UserBirthdate
		defer cancel()

		errBind := c.BindJSON(&birthDate)
		if errBind != nil {
			util.HandleError(c, http.StatusBadRequest, errBind, "Invalid request body")
			return
		}

		// Validate request body
		if validationErr := common.Validate.Struct(&birthDate); validationErr != nil {
			util.HandleError(c, http.StatusUnprocessableEntity, validationErr, "Validation error")
			return
		}

		myId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err, err.Error())
			return
		}
		filter := bson.M{"_id": myId}
		update := bson.M{"$set": bson.M{"birthdate.day": birthDate.Day, "birthdate.month": birthDate.Month, "birthdate.year": birthDate.Year}}
		_, errUpdateName := common.UserCollection.UpdateOne(ctx, filter, update)
		if errUpdateName != nil {
			util.HandleError(c, http.StatusBadRequest, errUpdateName, "Failed to update user birthdate")
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Birthdate updated", myId.Hex())
	}
}

// UpdateUserSingleField - update user single field like Phone, Bio
// api/user/update?field=phone&value=8084051523
func UpdateUserSingleField() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		field := c.Query("field")
		value := c.Query("value")
		defer cancel()

		myId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err, err.Error())
			return
		}

		if strings.Contains(field, ".") {
			util.HandleError(c, http.StatusBadRequest, fmt.Errorf("field '%s' can't contain a '.'", field), "Invalid field")
			return
		}

		notAllowedFields := []string{"role", "login_counts", "modified_at", "created_at", "favorite_shops", "shops", "status", "referred_by_user", "address_id", "transaction_sold_count", "transaction_buy_count", "birthdate", "thumbnail", "auth", "primary_email", "login_name", "_id"}

		for _, n := range notAllowedFields {
			if strings.ToLower(field) == n {
				log.Printf("User (%v) is trying to change their %v", myId.Hex(), n)
				util.HandleError(c, http.StatusUnauthorized, fmt.Errorf("cannot change field '%s'", n), "Field not allowed")
				return
			}
		}

		filter := bson.M{"_id": myId}
		update := bson.M{"$set": bson.M{field: value}}
		result, errUpdateField := common.UserCollection.UpdateOne(ctx, filter, update)
		if errUpdateField != nil {
			util.HandleError(c, http.StatusBadRequest, errUpdateField, "Failed to update field")
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Field updated", result.UpsertedID)
	}
}

// AddRemoveFavoriteShop - update user single field like Phone, Bio
// api/user/update?shopid=phone&value=8084051523
func AddRemoveFavoriteShop() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		shop := c.Query("shopid")
		action := c.Query("action")
		defer cancel()

		myObjectId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err, err.Error())
			return
		}

		filter := bson.M{"_id": myObjectId}
		if action == "add" {
			update := bson.M{"$push": bson.M{"favorite_shops": shop}}
			res, err := common.UserCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				util.HandleError(c, http.StatusBadRequest, err, "Failed to add favorite shop")
				return
			}

			util.HandleSuccess(c, http.StatusOK, "Favorite shop added", res)
			return
		}

		if action == "remove" {
			update := bson.M{"$pull": bson.M{"favorite_shops": shop}}
			res, err := common.UserCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				util.HandleError(c, http.StatusBadRequest, err, "Failed to remove favorite shop")
				return
			}

			util.HandleSuccess(c, http.StatusOK, "Favorite shop removed", res.UpsertedID)
			return
		}

		util.HandleError(c, http.StatusBadRequest, fmt.Errorf("action '%s' not recognized", action), "Invalid action")
	}
}

// AddWishListItem - Add to user wish list
// api/user/:userId/wishlist?listing_id=8084051523
func AddWishListItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		listingId := c.Query("listing_id")
		listingObjectId, err := primitive.ObjectIDFromHex(listingId)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err, "Invalid listing ID")
			return
		}

		MyId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err, err.Error())
			return
		}

		now := time.Now()
		data := models.UserWishlist{
			ID:        primitive.NewObjectID(),
			UserID:    MyId,
			ListingId: listingObjectId,
			CreatedAt: now,
		}
		_, err = common.WishListCollection.InsertOne(ctx, data)
		if err != nil {
			util.HandleError(c, http.StatusNotModified, err, "Failed to add wishlist item")
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Wishlist item added", data.ID.Hex())
	}
}

// RemoveWishListItem - Add to user wish list
// api/user/:userId/wishlist?listing_id=8084051523
func RemoveWishListItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		listingId := c.Query("listing_id")
		listingObjectId, err := primitive.ObjectIDFromHex(listingId)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err, "Invalid listing ID")
			return
		}

		MyId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err, err.Error())
			return
		}

		filter := bson.M{"user_id": MyId, "listing_id": listingObjectId}
		res, err := common.WishListCollection.DeleteOne(ctx, filter)
		if err != nil {
			util.HandleError(c, http.StatusNotModified, err, "Failed to remove wishlist item")
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Wishlist item removed", res.DeletedCount)
	}
}

// GetUserWishlist - Get all wishlist items  api/user/:userId/wishlist?limit=10&skip=0
func GetUserWishlist() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		MyId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err, err.Error())
			return
		}

		paginationArgs := common.GetPaginationArgs(c)
		filter := bson.M{"user_id": MyId}
		find := options.Find().SetLimit(int64(paginationArgs.Limit)).SetSkip(int64(paginationArgs.Skip))
		result, err := common.WishListCollection.Find(ctx, filter, find)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err, "Wishlist not found")
			return
		}

		var myWishLists []models.UserWishlist
		if err := result.All(ctx, &myWishLists); err != nil {
			util.HandleError(c, http.StatusInternalServerError, err, "Internal server error")
			return
		}

		count, err := common.WishListCollection.CountDocuments(ctx, bson.M{"user_id": MyId})
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err, "Error counting wishlist")
			return
		}

		util.HandleSuccessMeta(c, http.StatusOK, "success", myWishLists, gin.H{
			"pagination": util.Pagination{
				Limit: paginationArgs.Limit,
				Skip:  paginationArgs.Skip,
				Count: count,
			},
		})
	}
}

// UpdateSecurityNotificationSetting - GET api/user/:userId/login-notification?set=true
func UpdateSecurityNotificationSetting() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		myID, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err, err.Error())
			return
		}

		set := c.Query("set")
		if set == "" {
			util.HandleError(c, http.StatusBadRequest, errors.New("set can't be empty"), "set can't be empty")
			return
		}

		var setBool bool
		if strings.ToLower(set) == "true" {
			setBool = true
		} else {
			setBool = false
		}

		filter := bson.M{"_id": myID}
		update := bson.M{"$set": bson.M{"allow_login_ip_notification": setBool}}

		res, err := common.UserCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err, "error updating user login notification setting")
			return
		}

		if res.ModifiedCount < 1 {
			util.HandleError(c, http.StatusNotFound, errors.New("no document was modified"), "error updating user login notification setting")
			return
		}

		util.HandleSuccess(c, http.StatusOK, "login notification setting updated successfully.", res.UpsertedID)
	}
}

// GetSecurityNotificationSetting - GET api/user/:userId/login-notification
func GetSecurityNotificationSetting() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		MyId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err, err.Error())
			return
		}

		projection := bson.M{"allow_login_ip_notification": 1}
		options := options.FindOne().SetProjection(projection)

		var result bson.M
		filter := bson.M{"_id": MyId}
		err = common.UserCollection.FindOne(ctx, filter, options).Decode(&result)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err, "error retrieving user login notification setting")
			return
		}

		util.HandleSuccess(c, http.StatusOK, "login notification setting retrieved successfuly.", result)
	}
}
