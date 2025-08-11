package controllers

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"khoomi-api-io/api/internal/auth"
	"khoomi-api-io/api/internal/common"
	"khoomi-api-io/api/internal/helpers"
	"khoomi-api-io/api/pkg/models"
	"khoomi-api-io/api/pkg/services"
	"khoomi-api-io/api/pkg/util"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserController struct {
	userService         services.UserService
	notificationService services.NotificationService
}

func InitUserController(userService services.UserService, notificationService services.NotificationService) *UserController {
	return &UserController{
		userService:         userService,
		notificationService: notificationService,
	}
}

// ActiveSessionUser get current user using userId from request headers.
func (uc *UserController) ActiveSessionUser(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	session, err := auth.GetSessionAuto(c)
	if err != nil {
		util.HandleError(c, http.StatusNotFound, err)
		return
	}

	user, err := uc.userService.GetUserByID(ctx, session.UserId)
	if err != nil {
		util.HandleError(c, http.StatusNotFound, err)
		return
	}

	util.HandleSuccess(c, http.StatusOK, "success", user)
}

// CreateUser creates new user account
func (uc *UserController) CreateUser(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	var jsonUser models.CreateUserRequest
	if err := c.BindJSON(&jsonUser); err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	if err := common.Validate.Struct(&jsonUser); err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	req := models.CreateUserRequest{
		Email:     jsonUser.Email,
		FirstName: jsonUser.FirstName,
		LastName:  jsonUser.LastName,
		Password:  jsonUser.Password,
	}

	userID, err := uc.userService.CreateUser(ctx, req, c.ClientIP())
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	util.HandleSuccess(c, http.StatusOK, "signup successful", userID)
}

// HandleUserAuthentication authenticates user session
func (uc *UserController) HandleUserAuthentication(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	var jsonUser models.UserLoginBody
	if err := c.BindJSON(&jsonUser); err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	if err := common.Validate.Struct(&jsonUser); err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	req := models.UserAuthRequest{
		Email:    jsonUser.Email,
		Password: jsonUser.Password,
	}

	validUser, sessionId, err := uc.userService.AuthenticateUser(ctx, c, req, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		util.HandleError(c, http.StatusUnauthorized, err)
		return
	}

	userData := map[string]any{
		"userId":        validUser.Id.Hex(),
		"role":          validUser.Role,
		"email":         validUser.PrimaryEmail,
		"firstName":     validUser.FirstName,
		"lastName":      validUser.LastName,
		"loginName":     validUser.LoginName,
		"thumbnail":     validUser.Thumbnail,
		"emailVerified": validUser.Auth.EmailVerified,
		"isSeller":      validUser.IsSeller,
		"lastLogin":     validUser.LastLogin,
	}
	util.HandleSuccess(c, http.StatusOK, "Authentication successful",
		gin.H{
			"user":      userData,
			"sessionId": sessionId,
		})
}

func (uc *UserController) HandleUserGoogleAuthentication(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	var body struct {
		IDToken string `json:"idToken"`
	}

	if err := c.BindJSON(&body); err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	if err := common.Validate.Struct(&body); err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	validUser, sessionId, err := uc.userService.AuthenticateGoogleUser(ctx, c, body.IDToken, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
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

// GetMyActiveSession returns current active session given a session id
func (uc *UserController) GetMyActiveSession(c *gin.Context) {
	_, cancel := WithTimeout()
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

// RefreshToken handles auth token refreshments
func (uc *UserController) RefreshToken(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	session, err := auth.GetSessionAuto(c)
	if err != nil {
		auth.DeleteSession(c)
		util.HandleError(c, http.StatusUnauthorized, errors.New("unauthorized request"))
		return
	}

	if session.Expired() {
		auth.DeleteSession(c)
		util.HandleError(c, http.StatusUnauthorized, errors.New("unauthorized request"))
		return
	}

	user, err := uc.userService.RefreshUserSession(ctx, session.UserId)
	if err != nil {
		auth.DeleteSession(c)
		util.HandleError(c, http.StatusUnauthorized, errors.New("user not found"))
		return
	}

	auth.DeleteSession(c)

	newSessionId, err := auth.SetSession(c, session.UserId, user.PrimaryEmail, user.LoginName)
	if err != nil {
		util.HandleError(c, http.StatusInternalServerError, errors.New("failed to create new session"))
		return
	}

	util.HandleSuccess(c, http.StatusOK, "token refreshed successfully", gin.H{"sessionId": newSessionId})
}

// Logout - Log user out and invalidate session key
func (uc *UserController) Logout(c *gin.Context) {
	auth.DeleteSession(c)
	util.HandleSuccess(c, http.StatusOK, "logout successful", nil)
}

// SendDeleteUserAccount -> Delete current user account
func (uc *UserController) SendDeleteUserAccount(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	session_, err := auth.GetSessionAuto(c)
	if err != nil {
		util.HandleError(c, http.StatusUnauthorized, err)
		return
	}

	err = uc.userService.RequestAccountDeletion(ctx, session_.UserId)
	if err != nil {
		util.HandleError(c, http.StatusUnauthorized, err)
		return
	}

	util.HandleSuccess(c, http.StatusOK, "account is now pending deletion", gin.H{"_id": session_.UserId})
}

// IsAccountPendingDeletion checks if current user account is pending deletion
func (uc *UserController) IsAccountPendingDeletion(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	userId, err := auth.ValidateUserID(c)
	if err != nil {
		util.HandleError(c, http.StatusUnauthorized, err)
		return
	}

	pending, err := uc.userService.IsAccountPendingDeletion(ctx, userId)
	if err != nil {
		util.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	util.HandleSuccess(c, http.StatusOK, "Success", gin.H{"pendingDeletion": pending})
}

// CancelDeleteUserAccount cancels delete user account request.
func (uc *UserController) CancelDeleteUserAccount(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	userId, err := auth.ValidateUserID(c)
	if err != nil {
		util.HandleError(c, http.StatusUnauthorized, err)
		return
	}

	err = uc.userService.CancelAccountDeletion(ctx, userId)
	if err != nil {
		util.HandleError(c, http.StatusUnauthorized, err)
		return
	}

	util.HandleSuccess(c, http.StatusOK, "account deletion request cancelled", gin.H{"_id": userId.Hex()})
}

// ChangePassword changes active user's password.
func (uc *UserController) ChangePassword(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	userId, err := auth.ValidateUserID(c)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	var newPasswordFromRequest models.PasswordChangeRequest
	if err := c.Bind(&newPasswordFromRequest); err != nil {
		util.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	req := models.PasswordChangeRequest{
		CurrentPassword: newPasswordFromRequest.CurrentPassword,
		NewPassword:     newPasswordFromRequest.NewPassword,
	}

	err = uc.userService.ChangePassword(ctx, userId, req)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	util.HandleSuccess(c, http.StatusOK, "your password has been updated successfully", userId)
}

// GetUser retrieves a user based on their user ID or username.
// It accepts a user ID in the URL path and attempts to retrieve a user with a matching
// ObjectID. If the user ID is not a valid ObjectID, it attempts to find the user by username.
func (uc *UserController) GetUser(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	userID := c.Param("userid")
	user, err := uc.userService.GetUser(ctx, userID)
	if err != nil {
		util.HandleError(c, http.StatusNotFound, err)
		return
	}

	// Return the user data in the response
	util.HandleSuccess(c, http.StatusOK, "success", user)
}

// SendVerifyEmail sends a verification email notification to a given user's email.
func (uc *UserController) SendVerifyEmail(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	emailCurrent := c.Query("email")
	firstName := c.Query("name")

	// Verify current user
	session, err := auth.GetSessionAuto(c)
	if err != nil {
		util.HandleError(c, http.StatusUnauthorized, err)
		return
	}

	err = uc.userService.SendVerificationEmail(ctx, session.UserId, emailCurrent, firstName)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	util.HandleSuccess(c, http.StatusOK, "Verification email successfully sent", gin.H{"_id": session.UserId.Hex()})
}

func (uc *UserController) VerifyEmail(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	currentId := c.Query("id")
	currentToken := c.Query("token")

	userId, err := primitive.ObjectIDFromHex(currentId)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	err = uc.userService.VerifyEmail(ctx, userId, currentToken)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	util.HandleSuccess(c, http.StatusOK, "Your email has been verified successfully.", gin.H{"_id": userId})
}

// UpdateMyProfile updates the email, thumbnail, first and last name for the current user.
func (uc *UserController) UpdateMyProfile(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	session_, err := auth.GetSessionAuto(c)
	if err != nil {
		util.HandleError(c, http.StatusUnauthorized, err)
		return
	}

	var req models.UpdateUserProfileRequest
	if !BindJSONAndValidate(c, &req) {
		return
	}

	err = uc.userService.UpdateUserProfile(ctx, session_.UserId, req)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	util.HandleSuccess(c, http.StatusOK, "Profile updated successfully", gin.H{"_id": session_.UserId})
}

// ////////////////////// START USER LOGIN HISTORY //////////////////////////

// GetLoginHistories - Get user login histories (/api/users/:userId/login-history?limit=50&skip=0)
func (uc *UserController) GetLoginHistories(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	userId, err := auth.ValidateUserID(c)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	paginationArgs := helpers.GetPaginationArgs(c)
	loginHistory, count, err := uc.userService.GetLoginHistories(ctx, userId, paginationArgs)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
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

// DeleteLoginHistories - Delete user login histories
func (uc *UserController) DeleteLoginHistories(c *gin.Context) {
	ctx, cancel := WithTimeout()
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

	err = uc.userService.DeleteLoginHistories(ctx, userId, historyIDs.IDs)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	util.HandleSuccess(c, http.StatusOK, "Login histories deleted successfully", userId.Hex())
}

// ////////////////////// START USER PASSWORD RESET //////////////////////////

// PasswordResetEmail - api/send-password-reset?email=borngracedd@gmail.com
func (uc *UserController) PasswordResetEmail(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	currentEmail := c.Query("email")
	err := uc.userService.SendPasswordResetEmail(ctx, currentEmail)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	util.HandleSuccess(c, http.StatusOK, "Password reset email sent successfully", nil)
}

// PasswordReset - api/password-reset/userid?token=..&newpassword=..&id=user_uid
func (uc *UserController) PasswordReset(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	currentId := c.Query("id")
	currentToken := c.Query("token")
	newPassword := c.Query("newpassword")

	// Validate required parameters
	if currentId == "" || currentToken == "" || newPassword == "" {
		util.HandleError(c, http.StatusBadRequest, errors.New("id, token, and newpassword are required"))
		return
	}

	userId, err := primitive.ObjectIDFromHex(currentId)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	err = uc.userService.ResetPassword(ctx, userId, currentToken, newPassword)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	util.HandleSuccess(c, http.StatusOK, "success", userId.Hex())
}

// ////////////////////// START USER THUMBNAIL //////////////////////////

// UploadThumbnail - Upload user profile picture/thumbnail
// api/user/thumbnail?remote_addr=..
func (uc *UserController) UploadThumbnail(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	currentId, err := auth.ValidateUserID(c)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	remoteAddr := c.Query("remote_addr")
	var file any
	if remoteAddr == "" {
		fileHeader, err := c.FormFile("thumbnail")
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}
		file, err = fileHeader.Open()
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}
		defer file.(io.ReadCloser).Close()
	}

	err = uc.userService.UploadThumbnail(ctx, currentId, file, remoteAddr)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	util.HandleSuccess(c, http.StatusOK, "Your thumbnail has been changed successfully.", currentId.Hex())
}

// DeleteThumbnail - delete user profile picture/thumbnail
// api/user/thumbnail
func (uc *UserController) DeleteThumbnail(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	myId, err := auth.ValidateUserID(c)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	url := c.Param("url")
	if url == "" {
		util.HandleError(c, http.StatusBadRequest, errors.New("url parameter is required"))
		return
	}

	err = uc.userService.DeleteThumbnail(ctx, myId, url)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	util.HandleSuccess(c, http.StatusOK, "Your thumbnail has been deleted successfully.", myId.Hex())
}

// ////////////////////// START USER BIRTHDATE //////////////////////////
// UpdateUserBirthdate - update user birthdate
func (uc *UserController) UpdateUserBirthdate(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	var req struct {
		Birthdate string `json:"birthdate"`
	}

	if err := c.BindJSON(&req); err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	myId, err := auth.ValidateUserID(c)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	var birthdatePtr *time.Time
	if req.Birthdate != "" {
		birthdate, err := time.Parse("2006-01-02", req.Birthdate)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, errors.New("invalid date format, use YYYY-MM-DD"))
			return
		}
		birthdatePtr = &birthdate
	}

	err = uc.userService.UpdateUserBirthdate(ctx, myId, birthdatePtr)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	util.HandleSuccess(c, http.StatusOK, "Birthdate updated", myId.Hex())
}

// UpdateUserSingleField - update user single field like Phone, Bio
// api/user/update?field=phone&value=8084051523
func (uc *UserController) UpdateUserSingleField(c *gin.Context) {
	ctx, cancel := WithTimeout()
	field := c.Query("field")
	value := c.Query("value")
	defer cancel()

	myId, err := auth.ValidateUserID(c)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	err = uc.userService.UpdateUserSingleField(ctx, myId, field, value)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	util.HandleSuccess(c, http.StatusOK, "Field updated", nil)
}

// AddWishListItem - Add to user wish list
// api/user/:userId/wishlist?listing_id=8084051523
func (uc *UserController) AddWishListItem(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	listingId := c.Query("listing_id")
	listingObjectId, err := primitive.ObjectIDFromHex(listingId)
	if err != nil {
		util.HandleError(c, http.StatusUnauthorized, err)
		return
	}

	myId, err := auth.ValidateUserID(c)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	wishlistID, err := uc.userService.AddWishlistItem(ctx, myId, listingObjectId)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	util.HandleSuccess(c, http.StatusOK, "Wishlist item added", wishlistID.Hex())
}

// RemoveWishListItem - Add to user wish list
// api/user/:userId/wishlist?listing_id=8084051523
func (uc *UserController) RemoveWishListItem(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	listingId := c.Query("listing_id")
	listingObjectId, err := primitive.ObjectIDFromHex(listingId)
	if err != nil {
		util.HandleError(c, http.StatusUnauthorized, err)
		return
	}

	myId, err := auth.ValidateUserID(c)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	err = uc.userService.RemoveWishlistItem(ctx, myId, listingObjectId)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	util.HandleSuccess(c, http.StatusOK, "Wishlist item removed", nil)
}

// GetUserWishlist - Get all wishlist items  api/user/:userId/wishlist?limit=10&skip=0
func (uc *UserController) GetUserWishlist(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	MyId, err := auth.ValidateUserID(c)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	paginationArgs := helpers.GetPaginationArgs(c)
	myWishLists, count, err := uc.userService.GetUserWishlist(ctx, MyId, paginationArgs)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
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

func (uc *UserController) CreateUserNotificationSettings() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := WithTimeout()
		defer cancel()

		userId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		var notificationRequest models.UserNotificationSettingsRequest
		if err := c.BindJSON(&notificationRequest); err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		notification := models.UserNotificationSettings{
			ID:                   primitive.NewObjectID(),
			UserID:               userId,
			NewMessage:           notificationRequest.NewMessage,
			NewFollower:          notificationRequest.NewFollower,
			NewsAndFeatures:      notificationRequest.NewsAndFeatures,
			EmailEnabled:         notificationRequest.EmailEnabled,
			SMSEnabled:           notificationRequest.SMSEnabled,
			PushEnabled:          notificationRequest.PushEnabled,
			OrderUpdates:         notificationRequest.OrderUpdates,
			PaymentConfirmations: notificationRequest.PaymentConfirmations,
			DeliveryUpdates:      notificationRequest.DeliveryUpdates,
			CreatedAt:            time.Time{},
			ModifiedAt:           time.Time{},
		}

		insertedID, err := uc.userService.CreateNotificationSettings(ctx, userId, notification)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Notification created successfully", insertedID)
	}
}

func (uc *UserController) GetUserNotificationSettings() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := WithTimeout()
		defer cancel()

		userId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		settings, err := uc.userService.GetNotificationSettings(ctx, userId)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Notification settings retrieved successfully", settings)
	}
}

func (uc *UserController) UpdateUserNotificationSettings() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := WithTimeout()
		defer cancel()

		userId, err := auth.ValidateUserID(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		field := c.Query("name")
		value := c.Query("value")
		updateErr := uc.userService.UpdateNotificationSettings(ctx, userId, field, value)
		if updateErr != nil {
			util.HandleError(c, http.StatusBadRequest, updateErr)
			return
		}

		// TODO: maybe return UpdateResult from UpdateUserNotificationSettings call and return to user here.
		util.HandleSuccess(c, http.StatusOK, "Notification settings updated successfully", userId)
	}
}

// UpdateSecurityNotificationSetting - GET api/user/:userId/login-notification?set=true
func (uc *UserController) UpdateSecurityNotificationSetting(c *gin.Context) {
	ctx, cancel := WithTimeout()
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

	setBool, err := strconv.ParseBool(set)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, errors.New("invalid boolean value"))
		return
	}

	err = uc.userService.UpdateSecurityNotificationSetting(ctx, myID, setBool)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	util.HandleSuccess(c, http.StatusOK, "login notification setting updated successfully.", nil)
}

// GetSecurityNotificationSetting - GET api/user/:userId/login-notification
func (uc *UserController) GetSecurityNotificationSetting(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	MyId, err := auth.ValidateUserID(c)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	enabled, err := uc.userService.GetSecurityNotificationSetting(ctx, MyId)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	util.HandleSuccess(c, http.StatusOK, "login notification setting retrieved successfully.", gin.H{"allow_login_ip_notification": enabled})
}

// DeleteUser - DELETE api/admin/users/delete (Admin Only)
func (uc *UserController) DeleteUser(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	// Get target user ID from query parameter
	targetUserID := c.Query("userId")
	if targetUserID == "" {
		util.HandleError(c, http.StatusBadRequest, errors.New("target userId is required"))
		return
	}

	// Convert to ObjectID
	userObjectID, err := primitive.ObjectIDFromHex(targetUserID)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, errors.New("invalid userId format"))
		return
	}

	// Add confirmation parameter requirement
	confirmation := c.Query("confirm")
	if confirmation != "true" {
		util.HandleError(c, http.StatusBadRequest, errors.New("account deletion requires confirmation parameter: ?confirm=true"))
		return
	}

	// Execute user deletion
	result, err := uc.userService.DeleteUser(ctx, userObjectID)
	if err != nil {
		util.HandleError(c, http.StatusInternalServerError, fmt.Errorf("failed to delete user account: %v", err))
		return
	}

	util.HandleSuccess(c, http.StatusOK, "User account and all personal data deleted successfully", result)
}

// GetUserNotifications - GET /api/users/:userid/notifications
func (uc *UserController) GetUserNotifications(c *gin.Context) {

	ctx, cancel := WithTimeout()
	defer cancel()

	userID, err := auth.ValidateUserID(c)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}
	paginationArgs := helpers.GetPaginationArgs(c)
	var filters models.NotificationFilters
	if types := c.QueryArray("type"); len(types) > 0 {
		for _, t := range types {
			filters.Types = append(filters.Types, models.NotificationType(t))
		}
	}

	if priorities := c.QueryArray("priority"); len(priorities) > 0 {
		for _, p := range priorities {
			filters.Priorities = append(filters.Priorities, models.NotificationPriority(p))
		}
	}

	if isReadStr := c.Query("isRead"); isReadStr != "" {
		isRead := isReadStr == "true"
		filters.IsRead = &isRead
	}

	notifications, count, err := uc.notificationService.GetUserNotifications(ctx, userID, filters, paginationArgs)
	if err != nil {
		util.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	util.HandleSuccessMeta(c, http.StatusOK, "Notifications retrieved successfully", notifications, gin.H{
		"pagination": util.Pagination{
			Limit: paginationArgs.Limit,
			Skip:  paginationArgs.Skip,
			Count: count,
		},
	})
}

// GetUnreadNotifications - GET /api/users/:userid/notifications/unread
func (uc *UserController) GetUnreadNotifications(c *gin.Context) {

	ctx, cancel := WithTimeout()
	defer cancel()

	userID, err := auth.ValidateUserID(c)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	paginationArgs := helpers.GetPaginationArgs(c)

	notifications, count, err := uc.notificationService.GetUnreadNotifications(ctx, userID, paginationArgs)
	if err != nil {
		util.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	util.HandleSuccessMeta(c, http.StatusOK, "Unread notifications retrieved successfully", notifications, gin.H{
		"pagination": util.Pagination{
			Limit: paginationArgs.Limit,
			Skip:  paginationArgs.Skip,
			Count: count,
		},
		"unreadCount": count,
	})
}

// MarkNotificationAsRead - PUT /api/users/:userid/notifications/:notificationid/read
func (uc *UserController) MarkNotificationAsRead(c *gin.Context) {

	ctx, cancel := WithTimeout()
	defer cancel()

	userID, err := auth.ValidateUserID(c)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	notificationIDStr := c.Param("notificationid")
	notificationID, err := primitive.ObjectIDFromHex(notificationIDStr)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, errors.New("invalid notification ID"))
		return
	}

	err = uc.notificationService.MarkNotificationAsRead(ctx, userID, notificationID)
	if err != nil {
		util.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	util.HandleSuccess(c, http.StatusOK, "Notification marked as read", nil)
}

// MarkAllNotificationsAsRead - PUT /api/users/:userid/notifications/read-all
func (uc *UserController) MarkAllNotificationsAsRead(c *gin.Context) {

	ctx, cancel := WithTimeout()
	defer cancel()

	userID, err := auth.ValidateUserID(c)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	count, err := uc.notificationService.MarkAllNotificationsAsRead(ctx, userID)
	if err != nil {
		util.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	util.HandleSuccess(c, http.StatusOK, fmt.Sprintf("Marked %d notifications as read", count), gin.H{
		"markedCount": count,
	})
}

// DeleteNotification - DELETE /api/users/:userid/notifications/:notificationid
func (uc *UserController) DeleteNotification(c *gin.Context) {

	ctx, cancel := WithTimeout()
	defer cancel()

	userID, err := auth.ValidateUserID(c)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	notificationIDStr := c.Param("notificationid")
	notificationID, err := primitive.ObjectIDFromHex(notificationIDStr)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, errors.New("invalid notification ID"))
		return
	}

	err = uc.notificationService.DeleteNotification(ctx, userID, notificationID)
	if err != nil {
		util.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	util.HandleSuccess(c, http.StatusOK, "Notification deleted successfully", nil)
}

// GetNotificationStats - GET /api/users/:userid/notifications/stats
func (uc *UserController) GetNotificationStats(c *gin.Context) {

	ctx, cancel := WithTimeout()
	defer cancel()

	userID, err := auth.ValidateUserID(c)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	stats, err := uc.notificationService.GetNotificationStats(ctx, userID)
	if err != nil {
		util.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	util.HandleSuccess(c, http.StatusOK, "Notification statistics retrieved", stats)
}

// GetUnreadNotificationCount - GET /api/users/:userid/notifications/count
func (uc *UserController) GetUnreadNotificationCount(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	userID, err := auth.ValidateUserID(c)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	count, err := uc.notificationService.GetUnreadNotificationCount(ctx, userID)
	if err != nil {
		util.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	util.HandleSuccess(c, http.StatusOK, "Unread notification count retrieved", gin.H{
		"unreadCount": count,
	})
}
