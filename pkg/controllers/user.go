package controllers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"khoomi-api-io/api/internal/common"
	"khoomi-api-io/api/pkg/models"
	"khoomi-api-io/api/pkg/util"

	auth "khoomi-api-io/api/internal/auth"

	googleAuthIDTokenVerifier "github.com/futurenda/google-auth-id-token-verifier"

	email "khoomi-api-io/api/web/email"

	"github.com/cloudinary/cloudinary-go/api/uploader"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"go.mongodb.org/mongo-driver/x/bsonx"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func createUserImpl(c *gin.Context, claim *googleAuthIDTokenVerifier.ClaimSet, jsonUser models.UserRegistrationBody) *mongo.InsertOneResult {
	ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
	defer cancel()
	now := time.Now()

	firstName := ""
	lastName := ""
	userEmail := ""
	thumbnail := common.DefaultUserThumbnail
	userAuth := models.UserAuthData{}

	if claim == nil {
		errEmail := util.ValidateEmailAddress(jsonUser.Email)
		if errEmail != nil {
			log.Printf("Invalid email address from user %s with IP %s at %s: %s\n", jsonUser.FirstName, c.ClientIP(), time.Now().Format(time.RFC3339), errEmail.Error())
			util.HandleError(c, http.StatusBadRequest, errEmail)
			return nil
		}
		userEmail = strings.ToLower(jsonUser.Email)

		err := util.ValidatePassword(jsonUser.Password)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return nil
		}

		hashedPassword, errHashPassword := util.HashPassword(jsonUser.Password)
		if errHashPassword != nil {
			util.HandleError(c, http.StatusBadRequest, errHashPassword)
			return nil
		}

		firstName = jsonUser.FirstName
		lastName = jsonUser.LastName
		userAuth = models.UserAuthData{
			EmailVerified:  false,
			ModifiedAt:     time.Now(),
			PasswordDigest: hashedPassword,
		}
	} else {
		fmt.Println("HERE")
		firstName = claim.Name
		lastName = claim.FamilyName
		userEmail = claim.Email
		userAuth = models.UserAuthData{
			EmailVerified:  claim.EmailVerified,
			ModifiedAt:     time.Now(),
			PasswordDigest: "",
		}
		if claim.Picture != "" {
			thumbnail = claim.Picture
		}
	}

	userId := primitive.NewObjectID()
	newUser := bson.M{
		"_id":                         userId,
		"login_name":                  common.GenerateRandomUsername(),
		"primary_email":               userEmail,
		"first_name":                  firstName,
		"last_name":                   lastName,
		"auth":                        userAuth,
		"thumbnail":                   thumbnail,
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
		"review_count":                0,
	}

	result, err := common.UserCollection.InsertOne(ctx, newUser)
	if err != nil {
		writeException, ok := err.(mongo.WriteException)
		if ok {
			for _, writeError := range writeException.WriteErrors {
				if writeError.Code == common.MongoDuplicateKeyCode {
					log.Printf("User with email already exists: %s\n", writeError.Message)
					util.HandleError(c, http.StatusBadRequest, writeError)
					return nil
				}
			}
		}

		log.Printf("Mongo Error: Request could not be completed %s\n", err.Error())
		util.HandleError(c, http.StatusInternalServerError, err)
		return nil
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
		util.HandleError(c, http.StatusInternalServerError, err)
		return nil
	}

	// Send welcome email.
	email.SendWelcomeEmail(userEmail, jsonUser.FirstName)

	return result
}

// CurrentUser get current user using userId from request headers.
func ActiveSessionUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
	defer cancel()

	var user models.User
	// Extract user id from request header
	session, err := auth.GetSessionAuto(c)
	if err != nil {
		util.HandleError(c, http.StatusNotFound, err)
		return
	}

	err = common.UserCollection.FindOne(ctx, bson.M{"_id": session.UserId}).Decode(&user)
	if err != nil {
		util.HandleError(c, http.StatusNotFound, err)
		return
	}
	user.Auth.PasswordDigest = ""

	user.ConstructUserLinks()
	util.HandleSuccess(c, http.StatusOK, "success", user)
}

// CreateUser creates new user account, and send welcome and verify email notifications.
func CreateUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var jsonUser models.UserRegistrationBody
		if err := c.BindJSON(&jsonUser); err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		if err := common.Validate.Struct(&jsonUser); err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		result := createUserImpl(c, nil, jsonUser)
		if result == nil {
			return
		}

		util.HandleSuccess(c, http.StatusOK, "signup successful", result.InsertedID)
	}
}

// HandleUserAuthentication authenticates new user session while sending necessary notifications depending on the cases. e.g new IP
func HandleUserAuthentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		clientIP := c.ClientIP()
		now := time.Now()

		var jsonUser models.UserLoginBody
		if err := c.BindJSON(&jsonUser); err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		if err := common.Validate.Struct(&jsonUser); err != nil {
			log.Println(err)
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		var validUser models.User
		if err := common.UserCollection.FindOne(ctx, bson.M{"primary_email": jsonUser.Email}).Decode(&validUser); err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}

		if errPasswordCheck := util.CheckPassword(validUser.Auth.PasswordDigest, jsonUser.Password); errPasswordCheck != nil {
			util.HandleError(c, http.StatusUnauthorized, errPasswordCheck)
			return
		}

		wc := writeconcern.New(writeconcern.WMajority())
		txnOptions := options.Transaction().SetWriteConcern(wc)
		session, err := util.DB.StartSession()
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}
		defer session.EndSession(ctx)

		callback := func(ctx mongo.SessionContext) (any, error) {
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
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		if err := session.CommitTransaction(ctx); err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}
		session.EndSession(ctx)

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
		// if !validUser.Auth.EmailVerified {
		// 	link := fmt.Sprintf("https://khoomi.com/verify-email?token=%v&id=%v", token, validUser.Id.Hex())
		// 	email.SendVerifyEmailNotification(validUser.PrimaryEmail, validUser.FirstName, link)
		// }

		sessionId, err := auth.SetSession(c, validUser.Id, validUser.PrimaryEmail, validUser.LoginName)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		userData := map[string]any{
			"userId":        validUser.Id.Hex(),
			"role":          validUser.Role,
			"email":         validUser.PrimaryEmail,
			"FirstName":     validUser.FirstName,
			"lastName":      validUser.LastName,
			"loginName":     validUser.LoginName,
			"thumbnail":     validUser.Thumbnail,
			"emailverified": validUser.Auth.EmailVerified,
			"isSeller":      validUser.IsSeller,
		}
		util.HandleSuccess(c, http.StatusOK, "Authentication successful",
			gin.H{
				"user":      userData,
				"sessionId": sessionId,
			})
	}
}

func HandleUserGoogleAuthentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		var body struct {
			IDToken string `json:"idToken"`
		}

		fmt.Println("HERE4")
		clientIP := c.ClientIP()
		now := time.Now()

		if err := c.BindJSON(&body); err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		if err := common.Validate.Struct(&body); err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		v := googleAuthIDTokenVerifier.Verifier{}
		token := util.LoadEnvFor("GOOGLE_CLIENT_ID")
		err := v.VerifyIDToken(body.IDToken, []string{
			token,
		})
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		claimSet, err := googleAuthIDTokenVerifier.Decode(body.IDToken)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, errors.New("cannot decode token"))
			return
		}

		var validUser models.User
		// If no user is found with email, create new user
		if err := common.UserCollection.FindOne(ctx, bson.M{"primary_email": claimSet.Email}).Decode(&validUser); err != nil {
			result := createUserImpl(c, claimSet, models.UserRegistrationBody{})
			if result == nil {
				util.HandleError(c, http.StatusBadRequest, errors.New("error setting up user"))
				return
			}
		}

		if err := common.UserCollection.FindOne(ctx, bson.M{"primary_email": claimSet.Email}).Decode(&validUser); err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return

		}

		wc := writeconcern.New(writeconcern.WMajority())
		txnOptions := options.Transaction().SetWriteConcern(wc)
		session, err := util.DB.StartSession()
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}
		defer session.EndSession(ctx)

		callback := func(ctx mongo.SessionContext) (any, error) {
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
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		if err := session.CommitTransaction(ctx); err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}
		session.EndSession(ctx)

		// Send new login IP notification on condition
		if validUser.AllowLoginIpNotification && validUser.LastLoginIp != clientIP {
			email.SendNewIpLoginNotification(validUser.PrimaryEmail, validUser.LoginName, validUser.LastLoginIp, validUser.LastLogin)
		}

		// Send verify email
		// Generate secure and unique token
		// token := auth.GenerateSecureToken(8)

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
		// if !validUser.Auth.EmailVerified {
		// 	link := fmt.Sprintf("https://khoomi.com/verify-email?token=%v&id=%v", token, validUser.Id.Hex())
		// 	email.SendVerifyEmailNotification(validUser.PrimaryEmail, validUser.FirstName, link)
		// }

		sessionId, err := auth.SetSession(c, validUser.Id, validUser.PrimaryEmail, validUser.LoginName)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, errors.New("failed to set session"))
			return
		}

		userData := map[string]any{
			"userId":        validUser.Id.Hex(),
			"role":          validUser.Role,
			"email":         validUser.PrimaryEmail,
			"FirstName":     validUser.FirstName,
			"lastName":      validUser.LastName,
			"loginName":     validUser.LoginName,
			"thumbnail":     validUser.Thumbnail,
			"emailverified": validUser.Auth.EmailVerified,
			"isSeller":      validUser.IsSeller,
		}
		util.HandleSuccess(c, http.StatusOK, "Authentication successful",
			gin.H{
				"user":      userData,
				"sessionId": sessionId,
			})
	}
}

// GetMyActiveSession returns current active sesison given a session id
func GetMyActiveSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		session, err := auth.GetSessionAuto(c)
		if err != nil {
			// Delete old session
			auth.DeleteSession(c)
			util.HandleError(c, http.StatusUnauthorized, errors.New("unauthorized request"))
			return
		}

		util.HandleSuccess(c, http.StatusOK, "success", gin.H{"session": session})
	}
}

// RefreshToken handles auth token refreshments
func RefreshToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		session, err := auth.GetSessionAuto(c)
		if err != nil {
			// Delete old session
			auth.DeleteSession(c)
			util.HandleError(c, http.StatusUnauthorized, errors.New("unauthorized request"))
			return
		}

		if session.Expired() {
			auth.DeleteSession(c)
			util.HandleError(c, http.StatusUnauthorized, errors.New("unauthorized request"))
			return
		}

		// Delete old session
		auth.DeleteSession(c)

		// Set new session
		// TODO: request for user data e.g email and login_name to refresh session
		// err = auth.SetSession(c, session.UserId)
		// if err != nil {
		// 	util.HandleError(c, http.StatusUnauthorized, err, "Internal server error occurred")
		// 	return
		// }

		util.HandleSuccess(c, http.StatusOK, "success", gin.H{})
	}
}

// Logout - Log user out and invalidate session key
func Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth.DeleteSession(c)
		util.HandleSuccess(c, http.StatusOK, "logout successful", nil)
	}
}

// SendDeleteUserAccount -> Delete current user account
func SendDeleteUserAccount() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		session_, err := auth.GetSessionAuto(c)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err)
			return
		}
		_, err = common.UserDeletionCollection.InsertOne(ctx, bson.M{"user_id": session_.UserId, "created_at": time.Now()})
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err)
			return
		}

		util.HandleSuccess(c, http.StatusOK, "account is now pending deletion", gin.H{"_id": session_.UserId})
	}
}

// IsAccountPendingDeletion checks if current user account is pending deletion
func IsAccountPendingDeletion() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		userId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err)
			return
		}

		var account models.AccountDeletionRequested
		err = common.UserDeletionCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&account)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusOK, gin.H{"pendingDeletion": false})
				return
			}
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Success", gin.H{"pendingDeletion": true})
	}
}

// CancelDeleteUserAccount cancels delete user account request.
func CancelDeleteUserAccount() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		userId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err)
			return
		}

		_, err = common.UserDeletionCollection.DeleteOne(ctx, bson.M{"user_id": userId})
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err)
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
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		var newPasswordFromRequest models.NewPasswordRequest
		if err := c.Bind(&newPasswordFromRequest); err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}
		log.Printf("%v", newPasswordFromRequest)
		var validUser models.User
		if err := common.UserCollection.FindOne(ctx, bson.M{"_id": userId}).Decode(&validUser); err != nil {
			util.HandleError(c, http.StatusUnauthorized, err)
			return
		}

		if errPasswordCheck := util.CheckPassword(validUser.Auth.PasswordDigest, newPasswordFromRequest.CurrentPassword); errPasswordCheck != nil {
			util.HandleError(c, http.StatusUnauthorized, errPasswordCheck)
			return
		}

		if validationErr := common.Validate.Struct(&newPasswordFromRequest); validationErr != nil {
			util.HandleError(c, http.StatusUnprocessableEntity, validationErr)
			return
		}

		// validate password.
		err = util.ValidatePassword(newPasswordFromRequest.NewPassword)
		if err != nil {
			log.Printf("error validating password: %s\n", err.Error())
			util.HandleError(c, http.StatusExpectationFailed, err)
			return
		}

		// hash password before saving to storage.
		hashedPassword, errHashPassword := util.HashPassword(newPasswordFromRequest.NewPassword)
		if errHashPassword != nil {
			log.Printf("error hashing password: %s\n", errHashPassword.Error())
			util.HandleError(c, http.StatusExpectationFailed, errHashPassword)
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
			util.HandleError(c, http.StatusExpectationFailed, errHashPassword)
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
				util.HandleError(c, http.StatusBadRequest, e)
				return
			}

			filter = bson.M{"_id": userObjectID}
		} else {
			filter = bson.M{"login_name": userID}
		}

		// Query the database to find the user based on the specified field and value
		userPipeline := []bson.M{
			{"$match": filter},
			{"$lookup": bson.M{
				"from":         "Shop",
				"localField":   "shop_id",
				"foreignField": "_id",
				"as":           "shopDoc",
			}},
			{"$unwind": bson.M{
				"path":                       "$shopDoc",
				"preserveNullAndEmptyArrays": true,
			}},

			{"$project": bson.M{
				"_id":                         1,
				"last_login":                  1,
				"modified_at":                 1,
				"created_at":                  1,
				"auth":                        1,
				"thumbnail":                   1,
				"login_name":                  1,
				"bio":                         1,
				"phone":                       1,
				"last_name":                   1,
				"primary_email":               1,
				"first_name":                  1,
				"status":                      1,
				"referred_by_user":            1,
				"role":                        1,
				"favorite_shops":              1,
				"birthdate":                   1,
				"transaction_buy_count":       1,
				"transaction_sold_count":      1,
				"shop_id":                     1,
				"is_seller":                   1,
				"allow_login_ip_notification": 1,
				"review_count":                1,
				"shop": bson.M{
					"id":                 "$shopDoc._id",
					"name":               "$shopDoc.name",
					"slug":               "$shopDoc.slug",
					"username":           "$shopDoc.username",
					"logoUrl":            "$shopDoc.logo_url",
					"bannerUrl":          "$shopDoc.banner_url",
					"status":             "$shopDoc.status",
					"createdAt":          "$shopDoc.created_at",
					"listingActiveCount": "$shopDoc.listing_active_count",
					"followerCount":      "$shopDoc.follower_count",
					"reviewsCount":       "$shopDoc.reviews_count",
				},
			}},
		}
		cursor, err := common.UserCollection.Aggregate(ctx, userPipeline)

		var user models.User
		if err != nil {
			if err == mongo.ErrNoDocuments {
				util.HandleError(c, http.StatusNotFound, err)
				return
			}

			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}
		if cursor.Next(ctx) {
			if err := cursor.Decode(&user); err != nil {
				util.HandleError(c, http.StatusInternalServerError, err)
				return
			}
		} else {
			log.Printf("NotFound, %v %v", userID, err)
			util.HandleError(c, http.StatusNotFound, errors.New("no user found"))
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
		session, err := auth.GetSessionAuto(c)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err)
			return
		}
		// Verify current user email
		err = util.ValidateEmailAddress(emailCurrent)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		// Generate secure and unique token
		token := auth.GenerateSecureToken(8)

		expirationTime := now.Add(common.VERIFICATION_EMAIL_EXPIRATION_TIME)
		verifyEmail := models.UserVerifyEmailToken{
			UserId:      session.UserId,
			TokenDigest: token,
			CreatedAt:   primitive.NewDateTimeFromTime(now),
			ExpiresAt:   primitive.NewDateTimeFromTime(expirationTime),
		}
		opts := options.Replace().SetUpsert(true)
		filter := bson.M{"user_uid": session.UserId}
		_, err = common.EmailVerificationTokenCollection.ReplaceOne(ctx, filter, verifyEmail, opts)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		link := fmt.Sprintf("https://khoomi.com/verify-email?token=%v&id=%v", token, session.UserId)
		email.SendVerifyEmailNotification(emailCurrent, firstName, link)

		util.HandleSuccess(c, http.StatusOK, "Verification email successfully sent", gin.H{"_id": session.UserId.Hex()})
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
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		// Get and delete email verification
		err = common.EmailVerificationTokenCollection.FindOneAndDelete(ctx, bson.M{"user_uid": userId}).Decode(&emailVerificationData)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}

		// Check if verification token has expired
		now := primitive.NewDateTimeFromTime(time.Now())
		if now.Time().Unix() > emailVerificationData.ExpiresAt.Time().Unix() {
			util.HandleError(c, http.StatusNotFound, errors.New("email verification token has expired"))
			return
		}

		// Check if verification token is correct
		if currentToken != emailVerificationData.TokenDigest {
			util.HandleError(c, http.StatusNotFound, errors.New("incorrect or expired email verification token"))
			return
		}

		// Change user email verification status
		filter := bson.M{"_id": emailVerificationData.UserId}
		update := bson.M{"$set": bson.M{"status": "Active", "modified_at": now, "auth.modified_at": now, "auth.email_verified": true}}
		err = common.UserCollection.FindOneAndUpdate(ctx, filter, update).Decode(&user)
		if err != nil {
			util.HandleError(c, http.StatusNotModified, err)
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

		session_, err := auth.GetSessionAuto(c)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err)
			return
		}

		updateData := bson.M{}
		if firstName := c.Request.FormValue("firstName"); firstName != "" {
			if err := common.ValidateNameFormat(firstName); err != nil {
				util.HandleError(c, http.StatusBadRequest, err)
				return
			}
			updateData["first_name"] = firstName
		}

		if lastName := c.Request.FormValue("lastName"); lastName != "" {
			if err := common.ValidateNameFormat(lastName); err != nil {
				util.HandleError(c, http.StatusBadRequest, err)
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
				util.HandleError(c, http.StatusInternalServerError, err)
				return
			}
			defer file.Close()
			uploadResult, err = util.FileUpload(models.File{File: file})
			if err != nil {
				util.HandleError(c, http.StatusInternalServerError, fmt.Errorf("thumbnail Image upload failed - %v", err.Error()))
				return
			}
			updateData["thumbnail"] = uploadResult.SecureURL
		} else if err != http.ErrMissingFile {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		if len(updateData) == 0 {
			// delete media
			_, err = util.DestroyMedia(uploadResult.PublicID)
			log.Println(err)
			// return error
			util.HandleError(c, http.StatusBadRequest, errors.New("no update data provided"))
			return
		}

		updateData["modified_at"] = time.Now()

		filter := bson.M{"_id": session_.UserId}
		update := bson.M{"$set": updateData}

		_, err = common.UserCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			util.HandleError(c, http.StatusExpectationFailed, err)
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Profile updated successfully", gin.H{"_id": session_.UserId})
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
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		paginationArgs := common.GetPaginationArgs(c)
		filter := bson.M{"user_uid": userId}
		findOptions := options.Find().
			SetLimit(int64(paginationArgs.Limit)).
			SetSkip(int64(paginationArgs.Skip)).
			SetSort(util.GetLoginHistorySortBson(paginationArgs.Sort))
		cursor, err := common.LoginHistoryCollection.Find(ctx, filter, findOptions)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}
		defer cursor.Close(ctx)

		var loginHistory []models.LoginHistory
		if err = cursor.All(ctx, &loginHistory); err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}

		count, err := common.LoginHistoryCollection.CountDocuments(ctx, filter)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
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
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		var historyIDs models.LoginHistoryIds
		if err := c.BindJSON(&historyIDs); err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
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
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}
		defer session.EndSession(ctx)

		callback := func(ctx mongo.SessionContext) (any, error) {
			filter := bson.M{"_id": bson.M{"$in": IdsToDelete}, "user_uid": userId}
			result, err := common.LoginHistoryCollection.DeleteMany(ctx, filter)
			if err != nil {
				return nil, err
			}

			return result, nil
		}

		_, err = session.WithTransaction(ctx, callback, txnOptions)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		if err := session.CommitTransaction(ctx); err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
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

		log.Println(currentEmail)

		err := common.UserCollection.FindOne(ctx, bson.M{"primary_email": currentEmail}).Decode(&user)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
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
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		link := fmt.Sprintf("https://khoomi.com/reset/?id=%v&token=%v&email=%v", user.Id.Hex(), token, user.PrimaryEmail)
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
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		err = common.PasswordResetTokenCollection.FindOneAndDelete(ctx, bson.M{"user_uid": userId}).Decode(&passwordResetData)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}

		now := primitive.NewDateTimeFromTime(time.Now())
		if now.Time().Unix() > passwordResetData.ExpiresAt.Time().Unix() {
			util.HandleError(c, http.StatusNotFound, errors.New("password reset token has expired. Please restart the reset process"))
			return
		}

		if currentToken != passwordResetData.TokenDigest {
			util.HandleError(c, http.StatusNotFound, errors.New("password reset token is incorrect or expired. Please restart the reset process or use a valid token"))
			return
		}

		err = util.ValidatePassword(newPassword)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}

		hashedPassword, err := util.HashPassword(newPassword)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}

		filter := bson.M{"_id": passwordResetData.UserId}
		update := bson.M{"$set": bson.M{"auth.password_digest": hashedPassword, "auth.modified_at": now, "auth.email_verified": true}}
		err = common.UserCollection.FindOneAndUpdate(ctx, filter, update).Decode(&user)
		if err != nil {
			util.HandleError(c, http.StatusNotModified, err)
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

		currentId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		now := time.Now()
		filter := bson.M{"_id": currentId}

		var uploadResult uploader.UploadResult
		var update bson.M
		remoteAddr := c.Query("remote_addr")
		if remoteAddr != "" {
			uploadResult, err = util.RemoteUpload(models.Url{Url: remoteAddr})
			if err != nil {
				util.HandleError(c, http.StatusInternalServerError, err)
				return
			}

			update = bson.M{"$set": bson.M{"thumbnail": uploadResult.SecureURL, "modified_at": now}}
		} else {
			form, err := c.MultipartForm()
			if err != nil {
				fmt.Printf("MultipartForm error: %v\n", err)
			} else {
				fmt.Printf("Form fields: %v\n", form.Value)
				fmt.Printf("Form files: %v\n", form.File)
			}

			file, err := c.FormFile("thumbnail")
			if err != nil {
				fmt.Printf("Form error: %v\n", err)                                // Debug log
				fmt.Printf("Available form fields: %v\n", c.Request.MultipartForm) // Debug log
				util.HandleError(c, http.StatusBadRequest, err)
				return
			}

			src, err := file.Open()
			if err != nil {
				util.HandleError(c, http.StatusInternalServerError, err)
				return
			}
			defer src.Close()

			uploadResult, err = util.FileUpload(models.File{File: src})
			if err != nil {
				util.HandleError(c, http.StatusInternalServerError, err)
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
			util.HandleError(c, http.StatusExpectationFailed, err)
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
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		url := c.Param("url")
		if url == "" {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		var user models.User
		now := time.Now()
		filter := bson.M{"_id": myId}
		update := bson.M{"$set": bson.M{"thumbnail": nil, "modified_at": now}}
		err = common.UserCollection.FindOneAndUpdate(ctx, filter, update).Decode(&user)
		if err != nil {
			log.Printf("Thumbnail deletion failed: %v", err)
			util.HandleError(c, http.StatusExpectationFailed, err)
			return
		}

		filename, _, err := extractFilenameAndExtension(url)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		_, errOnDelete := util.ImageDeletionHelper(uploader.DestroyParams{
			PublicID:     filename,
			Type:         "upload",
			ResourceType: "image",
			Invalidate:   true,
		})
		if errOnDelete != nil {
			util.HandleError(c, http.StatusExpectationFailed, errOnDelete)
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
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		log.Println(userAddress)
		// Validate request body
		if validationErr := common.Validate.Struct(&userAddress); validationErr != nil {
			util.HandleError(c, http.StatusUnprocessableEntity, validationErr)
			return
		}

		// Extract current user token
		myId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		// create user address
		addressId := primitive.NewObjectID()
		userAddressTemp := models.UserAddress{
			Id:         addressId,
			UserId:     myId,
			City:       userAddress.City,
			State:      userAddress.State,
			Street:     userAddress.Street,
			PostalCode: userAddress.PostalCode,
			Country:    models.CountryNigeria,
			IsDefault:  userAddress.IsDefault,
		}

		count, err := common.UserAddressCollection.CountDocuments(ctx, bson.M{"user_id": myId})
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		if count >= 5 {
			util.HandleError(c, http.StatusInsufficientStorage, errors.New("max allowed addresses reached. please delete other address to accommodate a new one"))
			return
		}

		if userAddress.IsDefault {
			// Set IsDefaultShippingAddress to false for other addresses belonging to the user
			err = setOtherAddressesToFalse(ctx, myId, addressId)
			if err != nil {
				util.HandleError(c, http.StatusInternalServerError, err)
				return
			}

		}

		_, err = common.UserAddressCollection.InsertOne(ctx, userAddressTemp)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
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
			util.HandleError(c, http.StatusUnauthorized, err)
			return
		}

		filter := bson.M{"user_id": userId}
		cursor, err := common.UserAddressCollection.Find(ctx, filter)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}
		defer cursor.Close(ctx)

		var userAddresses []models.UserAddress
		if err := cursor.All(ctx, &userAddresses); err != nil {
			util.HandleError(c, http.StatusNotFound, err)
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
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		// Validate request body
		if validationErr := common.Validate.Struct(&userAddress); validationErr != nil {
			util.HandleError(c, http.StatusUnprocessableEntity, validationErr)
			return
		}

		// Extract current address Id
		addressId := c.Param("id")
		addressObjectId, err := primitive.ObjectIDFromHex(addressId)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}
		fmt.Println(addressObjectId)
		// Extract current user token
		myId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		// Set IsDefaultShippingAddress to false for other addresses belonging to the user
		if userAddress.IsDefault {
			err = setOtherAddressesToFalse(ctx, myId, addressObjectId)
			if err != nil {
				util.HandleError(c, http.StatusInternalServerError, err)
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
				"is_default_shipping_address": userAddress.IsDefault,
			},
		}

		_, err = common.UserAddressCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Address updated", addressId)
	}
}

// / ChangeDefaultAddress -> PUT /:userId/address/:addressId/default
func ChangeDefaultAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		userId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err)
			return
		}

		addressID := c.Param("id")
		if addressID == "" {
			util.HandleError(c, http.StatusBadRequest, errors.New("no address id was provided"))
			return
		}

		addressObjectID, err := primitive.ObjectIDFromHex(addressID)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, errors.New("bad address id"))
			return
		}

		// Set all other payment information records to is_default=false
		_, err = common.UserAddressCollection.UpdateMany(ctx, bson.M{"user_id": userId, "_id": bson.M{"$ne": addressObjectID}}, bson.M{"$set": bson.M{"is_default_shipping_address": false}})
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		filter := bson.M{"user_id": userId, "_id": addressObjectID}
		insertRes, insertErr := common.UserAddressCollection.UpdateOne(ctx, filter, bson.M{"$set": bson.M{"is_default_shipping_address": true}})
		if insertErr != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Default address has been succesfuly changed.", insertRes.ModifiedCount)
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
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		// Extract current user token
		myId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		// Set IsDefaultShippingAddress to false for other addresses belonging to the user
		err = setOtherAddressesToFalse(ctx, myId, addressObjectId)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		filter := bson.M{"user_id": myId, "_id": addressObjectId}
		res, err := common.UserAddressCollection.DeleteOne(ctx, filter)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		if res.DeletedCount == 0 {
			util.HandleError(c, http.StatusNotFound, errors.New("user address not found"))
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
			util.HandleError(c, http.StatusBadRequest, errBind)
			return
		}

		// Validate request body
		if validationErr := common.Validate.Struct(&birthDate); validationErr != nil {
			util.HandleError(c, http.StatusUnprocessableEntity, validationErr)
			return
		}

		myId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}
		filter := bson.M{"_id": myId}
		update := bson.M{"$set": bson.M{"birthdate.day": birthDate.Day, "birthdate.month": birthDate.Month, "birthdate.year": birthDate.Year}}
		_, errUpdateName := common.UserCollection.UpdateOne(ctx, filter, update)
		if errUpdateName != nil {
			util.HandleError(c, http.StatusBadRequest, errUpdateName)
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
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		if strings.Contains(field, ".") {
			util.HandleError(c, http.StatusBadRequest, fmt.Errorf("field '%s' can't contain a '.'", field))
			return
		}

		notAllowedFields := []string{"role", "login_counts", "modified_at", "created_at", "favorite_shops", "shops", "status", "referred_by_user", "address_id", "transaction_sold_count", "transaction_buy_count", "birthdate", "thumbnail", "auth", "primary_email", "login_name", "_id"}

		for _, n := range notAllowedFields {
			if strings.ToLower(field) == n {
				log.Printf("User (%v) is trying to change their %v", myId.Hex(), n)
				util.HandleError(c, http.StatusUnauthorized, fmt.Errorf("cannot change field '%s'", n))
				return
			}
		}

		filter := bson.M{"_id": myId}
		update := bson.M{"$set": bson.M{field: value}}
		result, errUpdateField := common.UserCollection.UpdateOne(ctx, filter, update)
		if errUpdateField != nil {
			util.HandleError(c, http.StatusBadRequest, errUpdateField)
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
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		filter := bson.M{"_id": myObjectId}
		if action == "add" {
			update := bson.M{"$push": bson.M{"favorite_shops": shop}}
			res, err := common.UserCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				util.HandleError(c, http.StatusBadRequest, err)
				return
			}

			util.HandleSuccess(c, http.StatusOK, "Favorite shop added", res)
			return
		}

		if action == "remove" {
			update := bson.M{"$pull": bson.M{"favorite_shops": shop}}
			res, err := common.UserCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				util.HandleError(c, http.StatusBadRequest, err)
				return
			}

			util.HandleSuccess(c, http.StatusOK, "Favorite shop removed", res.UpsertedID)
			return
		}

		util.HandleError(c, http.StatusBadRequest, fmt.Errorf("action '%s' not recognized", action))
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
			util.HandleError(c, http.StatusUnauthorized, err)
			return
		}

		MyId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
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
			util.HandleError(c, http.StatusNotModified, err)
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
			util.HandleError(c, http.StatusUnauthorized, err)
			return
		}

		MyId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		filter := bson.M{"user_id": MyId, "listing_id": listingObjectId}
		res, err := common.WishListCollection.DeleteOne(ctx, filter)
		if err != nil {
			util.HandleError(c, http.StatusNotModified, err)
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
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		paginationArgs := common.GetPaginationArgs(c)
		filter := bson.M{"user_id": MyId}
		find := options.Find().SetLimit(int64(paginationArgs.Limit)).SetSkip(int64(paginationArgs.Skip))
		cursor, err := common.WishListCollection.Find(ctx, filter, find)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}
		defer cursor.Close(ctx)
		var myWishLists []models.UserWishlist
		if err := cursor.All(ctx, &myWishLists); err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		count, err := common.WishListCollection.CountDocuments(ctx, bson.M{"user_id": MyId})
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
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
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		set := c.Query("set")
		if set == "" {
			util.HandleError(c, http.StatusBadRequest, errors.New("set can't be empty"))
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
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		if res.ModifiedCount < 1 {
			util.HandleError(c, http.StatusNotFound, errors.New("no document was modified"))
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
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		projection := bson.M{"allow_login_ip_notification": 1}
		options := options.FindOne().SetProjection(projection)

		var result bson.M
		filter := bson.M{"_id": MyId}
		err = common.UserCollection.FindOne(ctx, filter, options).Decode(&result)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}

		util.HandleSuccess(c, http.StatusOK, "login notification setting retrieved successfuly.", result)
	}
}
