package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"khoomi-api-io/api/internal"
	"khoomi-api-io/api/internal/auth"
	"khoomi-api-io/api/internal/common"
	"khoomi-api-io/api/internal/helpers"
	"khoomi-api-io/api/internal/validators"
	"khoomi-api-io/api/pkg/models"
	"khoomi-api-io/api/pkg/util"

	googleAuthIDTokenVerifier "github.com/futurenda/google-auth-id-token-verifier"
	"github.com/gin-gonic/gin"

	"github.com/cloudinary/cloudinary-go/api/uploader"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

type userService struct {
	emailService                       EmailService
	userCollection                     *mongo.Collection
	loginHistoryCollection             *mongo.Collection
	passwordResetTokenCollection       *mongo.Collection
	emailVerificationTokenCollection   *mongo.Collection
	wishListCollection                 *mongo.Collection
	userDeletionCollection             *mongo.Collection
	userNotificationSettingsCollection *mongo.Collection
	shopCollection                     *mongo.Collection
	listingCollection                  *mongo.Collection
	listingReviewCollection            *mongo.Collection
	userAddressCollection              *mongo.Collection
	userCartCollection                 *mongo.Collection
	userFavoriteListingCollection      *mongo.Collection
	userFavoriteShopCollection         *mongo.Collection
	sellerPaymentInfoCollection        *mongo.Collection
	userPaymentCardsCollection         *mongo.Collection
	shopFollowerCollection             *mongo.Collection
	sellerVerificationCollection       *mongo.Collection
}

func NewUserService() UserService {
	return &userService{
		emailService:                       NewEmailService(),
		userCollection:                     util.GetCollection(util.DB, "User"),
		loginHistoryCollection:             util.GetCollection(util.DB, "UserLoginHistory"),
		passwordResetTokenCollection:       util.GetCollection(util.DB, "UserPasswordResetToken"),
		emailVerificationTokenCollection:   util.GetCollection(util.DB, "UserEmailVerificationToken"),
		wishListCollection:                 util.GetCollection(util.DB, "UserWishList"),
		userDeletionCollection:             util.GetCollection(util.DB, "UserDeletionRequest"),
		userNotificationSettingsCollection: util.GetCollection(util.DB, "UserNotificationSetting"),
		shopCollection:                     util.GetCollection(util.DB, "Shop"),
		listingCollection:                  util.GetCollection(util.DB, "Listing"),
		listingReviewCollection:            util.GetCollection(util.DB, "ListingReview"),
		userAddressCollection:              util.GetCollection(util.DB, "UserAddress"),
		userCartCollection:                 util.GetCollection(util.DB, "UserCart"),
		userFavoriteListingCollection:      util.GetCollection(util.DB, "UserFavoriteListing"),
		userFavoriteShopCollection:         util.GetCollection(util.DB, "UserFavoriteShop"),
		sellerPaymentInfoCollection:        util.GetCollection(util.DB, "SellerPaymentInformation"),
		userPaymentCardsCollection:         util.GetCollection(util.DB, "UserPaymentCards"),
		shopFollowerCollection:             util.GetCollection(util.DB, "ShopFollower"),
		sellerVerificationCollection:       util.GetCollection(util.DB, "SellerVerification"),
	}
}

func (s *userService) CreateUser(ctx context.Context, req models.CreateUserRequest, clientIP string) (primitive.ObjectID, error) {
	now := time.Now()

	errEmail := util.ValidateEmailAddress(req.Email)
	if errEmail != nil {
		log.Printf("Invalid email address from user %s with IP %s at %s: %s\n", req.FirstName, clientIP, time.Now().Format(time.RFC3339), errEmail.Error())
		return primitive.NilObjectID, errEmail
	}

	userEmail := strings.ToLower(req.Email)

	err := util.ValidatePassword(req.Password)
	if err != nil {
		return primitive.NilObjectID, err
	}

	hashedPassword, errHashPassword := util.HashPassword(req.Password)
	if errHashPassword != nil {
		return primitive.NilObjectID, errHashPassword
	}

	userId := primitive.NewObjectID()
	userAuth := models.UserAuthData{
		EmailVerified:  false,
		ModifiedAt:     time.Now(),
		PasswordDigest: hashedPassword,
	}

	newUser := bson.M{
		"_id":                         userId,
		"login_name":                  GenerateRandomUsername(),
		"primary_email":               userEmail,
		"first_name":                  req.FirstName,
		"last_name":                   req.LastName,
		"auth":                        userAuth,
		"thumbnail":                   common.DEFAULT_USER_THUMBNAIL,
		"bio":                         bsonx.Null(),
		"phone":                       bsonx.Null(),
		"birthdate":                   nil,
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
		"last_login_ip":               clientIP,
		"allow_login_ip_notification": true,
		"review_count":                0,
		"seller_onboarding_level":     models.OnboardingLevelBuyer}

	callback := func(ctx mongo.SessionContext) (any, error) {
		result, err := s.userCollection.InsertOne(ctx, newUser)
		if err != nil {
			writeException, ok := err.(mongo.WriteException)
			if ok {
				for _, writeError := range writeException.WriteErrors {
					if writeError.Code == common.MONGO_DUPLICATE_KEY_CODE {
						log.Printf("User with email already exists: %s\n", writeError.Message)
						return nil, writeError
					}
				}
			}
			log.Printf("Mongo Error: Request could not be completed %s\n", err.Error())
			return nil, err
		}

		notification := models.UserNotificationSettings{
			ID:                   primitive.NewObjectID(),
			UserID:               userId,
			NewMessage:           true,
			NewFollower:          true,
			NewsAndFeatures:      true,
			EmailEnabled:         true,
			SMSEnabled:           true,
			PushEnabled:          true,
			OrderUpdates:         true,
			PaymentConfirmations: true,
			DeliveryUpdates:      true,
			CreatedAt:            time.Time{},
			ModifiedAt:           time.Time{},
		}

		_, err = s.userNotificationSettingsCollection.InsertOne(ctx, notification)
		if err != nil {
			return nil, err
		}

		return result, nil
	}

	_, err = ExecuteTransaction(ctx, callback)
	if err != nil {
		return primitive.NilObjectID, err
	}

	s.emailService.SendWelcomeEmail(userEmail, req.FirstName)

	return userId, nil
}

func (s *userService) CreateUserFromGoogle(ctx context.Context, claim any, clientIP string) (primitive.ObjectID, error) {
	claimSet, ok := claim.(*googleAuthIDTokenVerifier.ClaimSet)
	if !ok {
		return primitive.NilObjectID, errors.New("invalid claim type")
	}

	now := time.Now()
	firstName := claimSet.Name
	lastName := claimSet.FamilyName
	userEmail := claimSet.Email
	thumbnail := common.DEFAULT_USER_THUMBNAIL
	if claimSet.Picture != "" {
		thumbnail = claimSet.Picture
	}

	userAuth := models.UserAuthData{
		EmailVerified:  claimSet.EmailVerified,
		ModifiedAt:     time.Now(),
		PasswordDigest: "",
	}

	userId := primitive.NewObjectID()
	newUser := bson.M{
		"_id":                         userId,
		"login_name":                  GenerateRandomUsername(),
		"primary_email":               userEmail,
		"first_name":                  firstName,
		"last_name":                   lastName,
		"auth":                        userAuth,
		"thumbnail":                   thumbnail,
		"bio":                         bsonx.Null(),
		"phone":                       bsonx.Null(),
		"birthdate":                   nil,
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
		"last_login_ip":               clientIP,
		"allow_login_ip_notification": true,
		"review_count":                0,
	}

	result, err := s.userCollection.InsertOne(ctx, newUser)
	if err != nil {
		writeException, ok := err.(mongo.WriteException)
		if ok {
			for _, writeError := range writeException.WriteErrors {
				if writeError.Code == common.MONGO_DUPLICATE_KEY_CODE {
					log.Printf("User with email already exists: %s\n", writeError.Message)
					return primitive.NilObjectID, writeError
				}
			}
		}
		log.Printf("Mongo Error: Request could not be completed %s\n", err.Error())
		return primitive.NilObjectID, err
	}

	notification := models.UserNotificationSettings{
		ID:                   primitive.NewObjectID(),
		UserID:               userId,
		EmailEnabled:         true,
		SMSEnabled:           true,
		PushEnabled:          true,
		Promotional:          true,
		SupportMessage:       true,
		NewMessage:           true,
		NewFollower:          true,
		NewsAndFeatures:      true,
		OrderUpdates:         true,
		PaymentConfirmations: true,
		DeliveryUpdates:      true,
		CreatedAt:            time.Time{},
		ModifiedAt:           time.Time{},
	}

	_, err = s.userNotificationSettingsCollection.InsertOne(ctx, notification)
	if err != nil {
		return primitive.NilObjectID, err
	}

	// Send welcome email
	s.emailService.SendWelcomeEmail(userEmail, firstName)

	return result.InsertedID.(primitive.ObjectID), nil
}

func (s *userService) AuthenticateUser(ctx context.Context, gCtx *gin.Context, req models.UserAuthRequest, clientIP, userAgent string) (*models.User, string, error) {
	now := time.Now()

	var validUser models.User
	if err := s.userCollection.FindOne(ctx, bson.M{"primary_email": req.Email}).Decode(&validUser); err != nil {
		return nil, "", err
	}

	if errPasswordCheck := util.CheckPassword(validUser.Auth.PasswordDigest, req.Password); errPasswordCheck != nil {
		return nil, "", errPasswordCheck
	}

	wc := writeconcern.New(writeconcern.WMajority())
	txnOptions := options.Transaction().SetWriteConcern(wc)
	session, err := util.DB.StartSession()
	if err != nil {
		return nil, "", err
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
		errUpdateLoginCounts := s.userCollection.FindOneAndUpdate(ctx, filter, update).Decode(&validUser)
		if errUpdateLoginCounts != nil {
			return nil, errUpdateLoginCounts
		}

		doc := models.LoginHistory{
			Id:        primitive.NewObjectID(),
			UserUid:   validUser.Id,
			Date:      now,
			UserAgent: userAgent,
			IpAddr:    clientIP,
		}
		result, errLoginHistory := s.loginHistoryCollection.InsertOne(ctx, doc)
		if errLoginHistory != nil {
			return result, errLoginHistory
		}

		return result, nil
	}

	_, err = session.WithTransaction(ctx, callback, txnOptions)
	if err != nil {
		return nil, "", err
	}

	if err := session.CommitTransaction(ctx); err != nil {
		return nil, "", err
	}

	// Send new login IP notification on condition
	if validUser.AllowLoginIpNotification && validUser.LastLoginIp != clientIP {
		s.emailService.SendNewIpLoginNotification(validUser.PrimaryEmail, validUser.LoginName, validUser.LastLoginIp, validUser.LastLogin)
	}

	// Generate secure and unique token for email verification
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
	_, err = s.emailVerificationTokenCollection.ReplaceOne(ctx, filter, verifyEmail, opts)
	if err != nil {
		log.Printf("error sending verification email for user: %v, error: %v", validUser.PrimaryEmail, err)
	}

	sessionId, err := auth.SetSession(gCtx, validUser.Id, validUser.PrimaryEmail, validUser.LoginName)
	if err != nil {
		return nil, "", err
	}

	return &validUser, sessionId, nil
}

func (s *userService) AuthenticateGoogleUser(ctx context.Context, gCtx *gin.Context, idToken, clientIP, userAgent string) (*models.User, string, error) {
	v := googleAuthIDTokenVerifier.Verifier{}
	googleClientID := util.LoadEnvFor("GOOGLE_CLIENT_ID")
	err := v.VerifyIDToken(idToken, []string{googleClientID})
	if err != nil {
		return nil, "", err
	}

	claimSet, err := googleAuthIDTokenVerifier.Decode(idToken)
	if err != nil {
		return nil, "", errors.New("cannot decode token")
	}

	var validUser models.User
	// If no user is found with email, create new user
	if err := s.userCollection.FindOne(ctx, bson.M{"primary_email": claimSet.Email}).Decode(&validUser); err != nil {
		_, err := s.CreateUserFromGoogle(ctx, claimSet, clientIP)
		if err != nil {
			return nil, "", errors.New("error setting up user")
		}
	}

	if err := s.userCollection.FindOne(ctx, bson.M{"primary_email": claimSet.Email}).Decode(&validUser); err != nil {
		return nil, "", err
	}

	now := time.Now()
	wc := writeconcern.New(writeconcern.WMajority())
	txnOptions := options.Transaction().SetWriteConcern(wc)
	session, err := util.DB.StartSession()
	if err != nil {
		return nil, "", err
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
		errUpdateLoginCounts := s.userCollection.FindOneAndUpdate(ctx, filter, update).Decode(&validUser)
		if errUpdateLoginCounts != nil {
			return nil, errUpdateLoginCounts
		}

		doc := models.LoginHistory{
			Id:        primitive.NewObjectID(),
			UserUid:   validUser.Id,
			Date:      now,
			UserAgent: userAgent,
			IpAddr:    clientIP,
		}
		result, errLoginHistory := s.loginHistoryCollection.InsertOne(ctx, doc)
		if errLoginHistory != nil {
			return result, errLoginHistory
		}

		return result, nil
	}

	_, err = session.WithTransaction(ctx, callback, txnOptions)
	if err != nil {
		return nil, "", err
	}

	if err := session.CommitTransaction(ctx); err != nil {
		return nil, "", err
	}

	// Send new login IP notification on condition
	if validUser.AllowLoginIpNotification && validUser.LastLoginIp != clientIP {
		s.emailService.SendNewIpLoginNotification(validUser.PrimaryEmail, validUser.LoginName, validUser.LastLoginIp, validUser.LastLogin)
	}

	// Generate secure and unique token for email verification
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
	_, err = s.emailVerificationTokenCollection.ReplaceOne(ctx, filter, verifyEmail, opts)
	if err != nil {
		log.Printf("error sending verification email for user: %v, error: %v", validUser.PrimaryEmail, err)
	}

	sessionId, err := auth.SetSession(gCtx, validUser.Id, validUser.PrimaryEmail, validUser.LoginName)
	if err != nil {
		return nil, "", err
	}

	return &validUser, sessionId, nil
}

func (s *userService) GetUserByID(ctx context.Context, userID primitive.ObjectID) (*models.User, error) {
	var user models.User
	err := s.userCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return nil, err
	}
	user.Auth.PasswordDigest = ""
	user.ConstructUserLinks()
	return &user, nil
}

func (s *userService) GetUser(ctx context.Context, userIdentifier string) (*models.User, error) {
	var filter bson.M
	if primitive.IsValidObjectID(userIdentifier) {
		userObjectID, e := primitive.ObjectIDFromHex(userIdentifier)
		if e != nil {
			return nil, e
		}
		filter = bson.M{"_id": userObjectID}
	} else {
		filter = bson.M{"login_name": userIdentifier}
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
			"gender":                      1,
			"allow_login_ip_notification": 1,
			"review_count":                1,
			"seller_onboarding_level":     1,
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
	cursor, err := s.userCollection.Aggregate(ctx, userPipeline)

	var user models.User
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, err
		}
		return nil, err
	}
	if cursor.Next(ctx) {
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("no user found")
	}

	user.ConstructUserLinks()
	return &user, nil
}

func (s *userService) UpdateUserProfile(ctx context.Context, userID primitive.ObjectID, req models.UpdateUserProfileRequest) error {
	updateData := bson.M{}

	if req.FirstName != "" {
		if err := validators.ValidateNameFormat(req.FirstName); err != nil {
			return err
		}
		updateData["first_name"] = req.FirstName
	}

	if req.LastName != "" {
		if err := validators.ValidateNameFormat(req.LastName); err != nil {
			return err
		}
		updateData["last_name"] = req.LastName
	}

	if req.Email != "" {
		if err := util.ValidateEmailAddress(req.Email); err != nil {
			return err
		}
		updateData["primary_email"] = strings.ToLower(req.Email)
	}

	if req.Gender != "" {
		updateData["gender"] = req.Gender
	}

	if req.Dob != nil {
		updateData["birthdate"] = req.Dob
	}

	if req.Phone != "" {
		updateData["phone"] = req.Phone
	}

	if len(updateData) == 0 {
		return errors.New("no update data provided")
	}

	updateData["modified_at"] = time.Now()

	filter := bson.M{"_id": userID}
	update := bson.M{"$set": updateData}

	_, err := s.userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	internal.PublishCacheMessage(ctx, internal.CacheInvalidateUser, userID.Hex())
	return nil
}

func (s *userService) UpdateUserSingleField(ctx context.Context, userID primitive.ObjectID, field, value string) error {
	if strings.Contains(field, ".") {
		return fmt.Errorf("field '%s' can't contain a '.'", field)
	}

	notAllowedFields := []string{"role", "login_counts", "modified_at", "created_at", "favorite_shops", "shops", "status", "referred_by_user", "address_id", "transaction_sold_count", "transaction_buy_count", "birthdate", "thumbnail", "auth", "primary_email", "login_name", "_id"}

	for _, n := range notAllowedFields {
		if strings.ToLower(field) == n {
			log.Printf("User (%v) is trying to change their %v", userID.Hex(), n)
			return fmt.Errorf("cannot change field '%s'", n)
		}
	}

	filter := bson.M{"_id": userID}
	update := bson.M{"$set": bson.M{field: value}}
	_, err := s.userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	internal.PublishCacheMessage(ctx, internal.CacheInvalidateUser, userID.Hex())
	return nil
}

func (s *userService) UpdateUserBirthdate(ctx context.Context, userID primitive.ObjectID, birthdate *time.Time) error {
	filter := bson.M{"_id": userID}
	var update bson.M

	if birthdate == nil {
		update = bson.M{"$unset": bson.M{"birthdate": ""}}
	} else {
		if birthdate.After(time.Now()) {
			return errors.New("birthdate cannot be in the future")
		}

		minAge := time.Now().AddDate(-13, 0, 0)
		if birthdate.After(minAge) {
			return errors.New("user must be at least 13 years old")
		}

		maxAge := time.Now().AddDate(-150, 0, 0)
		if birthdate.Before(maxAge) {
			return errors.New("invalid birthdate")
		}

		update = bson.M{"$set": bson.M{"birthdate": birthdate}}
	}

	_, err := s.userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	internal.PublishCacheMessage(ctx, internal.CacheInvalidateUser, userID.Hex())
	return nil
}

func (s *userService) ChangePassword(ctx context.Context, userID primitive.ObjectID, req models.PasswordChangeRequest) error {
	var validUser models.User
	if err := s.userCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&validUser); err != nil {
		return err
	}

	if errPasswordCheck := util.CheckPassword(validUser.Auth.PasswordDigest, req.CurrentPassword); errPasswordCheck != nil {
		return errPasswordCheck
	}

	if validationErr := common.Validate.Struct(&req); validationErr != nil {
		return validationErr
	}

	err := util.ValidatePassword(req.NewPassword)
	if err != nil {
		return err
	}

	hashedPassword, errHashPassword := util.HashPassword(req.NewPassword)
	if errHashPassword != nil {
		return errHashPassword
	}

	filter := bson.M{"_id": userID}
	update := bson.M{"$set": bson.M{"auth.password_digest": hashedPassword, "modified_at": time.Now()}}
	_, err = s.userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (s *userService) SendPasswordResetEmail(ctx context.Context, email string) error {
	currentEmail := strings.ToLower(email)
	var user models.User

	err := s.userCollection.FindOne(ctx, bson.M{"primary_email": currentEmail}).Decode(&user)
	if err != nil {
		return err
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
	_, err = s.passwordResetTokenCollection.ReplaceOne(ctx, filter, passwordReset, opts)
	if err != nil {
		return err
	}

	link := fmt.Sprintf("https://khoomi.com/reset/?id=%v&token=%v&email=%v", user.Id.Hex(), token, user.PrimaryEmail)
	s.emailService.SendPasswordResetEmail(user.PrimaryEmail, user.FirstName, link)

	return nil
}

// TODO: Password reset is failin due to document not found.. Debug when I have somoe time since there're more important stuffs to handle!
func (s *userService) ResetPassword(ctx context.Context, userID primitive.ObjectID, token, newPassword string) error {
	var passwordResetData models.UserPasswordResetToken
	var user models.User

	err := s.passwordResetTokenCollection.FindOneAndDelete(ctx, bson.M{"user_uid": userID}).Decode(&passwordResetData)
	if err != nil {
		return err
	}

	now := primitive.NewDateTimeFromTime(time.Now())
	if now.Time().Unix() > passwordResetData.ExpiresAt.Time().Unix() {
		return errors.New("password reset token has expired. Please restart the reset process")
	}

	if token != passwordResetData.TokenDigest {
		return errors.New("password reset token is incorrect or expired. Please restart the reset process or use a valid token")
	}

	err = util.ValidatePassword(newPassword)
	if err != nil {
		return err
	}

	hashedPassword, err := util.HashPassword(newPassword)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": passwordResetData.UserId}
	update := bson.M{"$set": bson.M{"auth.password_digest": hashedPassword, "auth.modified_at": now, "auth.email_verified": true}}
	err = s.userCollection.FindOneAndUpdate(ctx, filter, update).Decode(&user)
	if err != nil {
		return err
	}

	s.emailService.SendPasswordResetSuccessfulEmail(user.PrimaryEmail, user.FirstName)
	return nil
}

func (s *userService) SendVerificationEmail(ctx context.Context, userID primitive.ObjectID, email, firstName string) error {
	now := time.Now()

	err := util.ValidateEmailAddress(email)
	if err != nil {
		return err
	}

	token := auth.GenerateSecureToken(8)
	expirationTime := now.Add(common.VERIFICATION_EMAIL_EXPIRATION_TIME)
	verifyEmail := models.UserVerifyEmailToken{
		UserId:      userID,
		TokenDigest: token,
		CreatedAt:   primitive.NewDateTimeFromTime(now),
		ExpiresAt:   primitive.NewDateTimeFromTime(expirationTime),
	}
	opts := options.Replace().SetUpsert(true)
	filter := bson.M{"user_uid": userID}
	_, err = s.emailVerificationTokenCollection.ReplaceOne(ctx, filter, verifyEmail, opts)
	if err != nil {
		return err
	}

	link := fmt.Sprintf("https://khoomi.com/verify-email?token=%v&id=%v", token, userID)
	s.emailService.SendVerifyEmailNotification(email, firstName, link)

	return nil
}

func (s *userService) VerifyEmail(ctx context.Context, userID primitive.ObjectID, token string) error {
	var emailVerificationData models.UserVerifyEmailToken
	var user models.User

	err := s.emailVerificationTokenCollection.FindOneAndDelete(ctx, bson.M{"user_uid": userID}).Decode(&emailVerificationData)
	if err != nil {
		return err
	}

	now := primitive.NewDateTimeFromTime(time.Now())
	if now.Time().Unix() > emailVerificationData.ExpiresAt.Time().Unix() {
		return errors.New("email verification token has expired")
	}

	if token != emailVerificationData.TokenDigest {
		return errors.New("incorrect or expired email verification token")
	}

	filter := bson.M{"_id": emailVerificationData.UserId}
	update := bson.M{"$set": bson.M{"status": "Active", "modified_at": now, "auth.modified_at": now, "auth.email_verified": true}}
	err = s.userCollection.FindOneAndUpdate(ctx, filter, update).Decode(&user)
	if err != nil {
		return err
	}

	s.emailService.SendEmailVerificationSuccessNotification(user.PrimaryEmail, user.FirstName)
	return nil
}

func (s *userService) RefreshUserSession(ctx context.Context, userID primitive.ObjectID) (*models.User, error) {
	var user models.User
	err := s.userCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *userService) RequestAccountDeletion(ctx context.Context, userID primitive.ObjectID) error {
	_, err := s.userDeletionCollection.InsertOne(ctx, bson.M{"user_id": userID, "created_at": time.Now()})
	if err != nil {
		return err
	}

	internal.PublishCacheMessage(ctx, internal.CacheInvalidateUserDeletion, userID.Hex())
	return nil
}

func (s *userService) CancelAccountDeletion(ctx context.Context, userID primitive.ObjectID) error {
	_, err := s.userDeletionCollection.DeleteOne(ctx, bson.M{"user_id": userID})
	if err != nil {
		return err
	}

	internal.PublishCacheMessage(ctx, internal.CacheInvalidateUserDeletion, userID.Hex())
	return nil
}

func (s *userService) IsAccountPendingDeletion(ctx context.Context, userID primitive.ObjectID) (bool, error) {
	var account models.AccountDeletionRequested
	err := s.userDeletionCollection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&account)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *userService) GetLoginHistories(ctx context.Context, userID primitive.ObjectID, pagination util.PaginationArgs) ([]models.LoginHistory, int64, error) {
	filter := bson.M{"user_uid": userID}
	findOptions := options.Find().
		SetLimit(int64(pagination.Limit)).
		SetSkip(int64(pagination.Skip)).
		SetSort(util.GetLoginHistorySortBson(pagination.Sort))
	cursor, err := s.loginHistoryCollection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var loginHistory []models.LoginHistory
	if err = cursor.All(ctx, &loginHistory); err != nil {
		return nil, 0, err
	}

	count, err := s.loginHistoryCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return loginHistory, count, nil
}

func (s *userService) DeleteLoginHistories(ctx context.Context, userID primitive.ObjectID, historyIDs []string) error {
	var IdsToDelete []primitive.ObjectID
	for _, id := range historyIDs {
		objId, _ := primitive.ObjectIDFromHex(id)
		IdsToDelete = append(IdsToDelete, objId)
	}

	wc := writeconcern.New(writeconcern.WMajority())
	txnOptions := options.Transaction().SetWriteConcern(wc)
	session, err := util.DB.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	callback := func(ctx mongo.SessionContext) (any, error) {
		filter := bson.M{"_id": bson.M{"$in": IdsToDelete}, "user_uid": userID}
		result, err := s.loginHistoryCollection.DeleteMany(ctx, filter)
		if err != nil {
			return nil, err
		}
		return result, nil
	}

	_, err = session.WithTransaction(ctx, callback, txnOptions)
	if err != nil {
		return err
	}

	if err := session.CommitTransaction(ctx); err != nil {
		return err
	}

	internal.PublishCacheMessage(ctx, internal.CacheInvalidateUserLoginHistories, userID.Hex())
	return nil
}

func (s *userService) UploadThumbnail(ctx context.Context, userID primitive.ObjectID, file any, remoteAddr string) error {
	now := time.Now()
	filter := bson.M{"_id": userID}

	var uploadResult uploader.UploadResult
	var update bson.M
	var err error

	if remoteAddr != "" {
		uploadResult, err = util.RemoteUpload(models.Url{Url: remoteAddr})
		if err != nil {
			return err
		}
		update = bson.M{"$set": bson.M{"thumbnail": uploadResult.SecureURL, "modified_at": now}}
	} else {
		// Handle file upload - this would need to be implemented based on file type
		// For now, assuming the file handling is done externally
		return errors.New("file upload not implemented in service")
	}

	_, err = s.userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		// delete media on error
		_, delErr := util.DestroyMedia(uploadResult.PublicID)
		log.Println(delErr)
		return err
	}

	internal.PublishCacheMessage(ctx, internal.CacheInvalidateUser, userID.Hex())
	return nil
}

func (s *userService) DeleteThumbnail(ctx context.Context, userID primitive.ObjectID, url string) error {
	now := time.Now()
	filter := bson.M{"_id": userID}
	update := bson.M{"$set": bson.M{"thumbnail": nil, "modified_at": now}}

	var user models.User
	err := s.userCollection.FindOneAndUpdate(ctx, filter, update).Decode(&user)
	if err != nil {
		return err
	}

	filename, _, err := helpers.ExtractFilenameAndExtension(url)
	if err != nil {
		return err
	}

	_, errOnDelete := util.ImageDeletionHelper(uploader.DestroyParams{
		PublicID:     filename,
		Type:         "upload",
		ResourceType: "image",
		Invalidate:   true,
	})
	if errOnDelete != nil {
		return errOnDelete
	}

	internal.PublishCacheMessage(ctx, internal.CacheInvalidateUser, userID.Hex())
	return nil
}

func (s *userService) AddWishlistItem(ctx context.Context, userID, listingID primitive.ObjectID) (primitive.ObjectID, error) {
	now := time.Now()
	data := models.UserWishlist{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		ListingId: listingID,
		CreatedAt: now,
	}
	_, err := s.wishListCollection.InsertOne(ctx, data)
	if err != nil {
		return primitive.NilObjectID, err
	}

	internal.PublishCacheMessage(ctx, internal.CacheInvalidateShopCompliance, userID.Hex())
	return data.ID, nil
}

func (s *userService) RemoveWishlistItem(ctx context.Context, userID, listingID primitive.ObjectID) error {
	filter := bson.M{"user_id": userID, "listing_id": listingID}
	_, err := s.wishListCollection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	internal.PublishCacheMessage(ctx, internal.CacheInvalidateShopCompliance, userID.Hex())
	return nil
}

func (s *userService) GetUserWishlist(ctx context.Context, userID primitive.ObjectID, pagination util.PaginationArgs) ([]models.UserWishlist, int64, error) {
	filter := bson.M{"user_id": userID}
	find := options.Find().SetLimit(int64(pagination.Limit)).SetSkip(int64(pagination.Skip))
	cursor, err := s.wishListCollection.Find(ctx, filter, find)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var myWishLists []models.UserWishlist
	if err := cursor.All(ctx, &myWishLists); err != nil {
		return nil, 0, err
	}

	count, err := s.wishListCollection.CountDocuments(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, 0, err
	}

	return myWishLists, count, nil
}

func (s *userService) UpdateSecurityNotificationSetting(ctx context.Context, userID primitive.ObjectID, enabled bool) error {
	filter := bson.M{"_id": userID}
	update := bson.M{"$set": bson.M{"allow_login_ip_notification": enabled}}

	res, err := s.userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if res.ModifiedCount < 1 {
		return errors.New("no document was modified")
	}

	internal.PublishCacheMessage(ctx, internal.CacheInvalidateUserWishlist, userID.Hex())
	return nil
}

func (s *userService) GetSecurityNotificationSetting(ctx context.Context, userID primitive.ObjectID) (bool, error) {
	projection := bson.M{"allow_login_ip_notification": 1}
	options := options.FindOne().SetProjection(projection)

	var result bson.M
	filter := bson.M{"_id": userID}
	err := s.userCollection.FindOne(ctx, filter, options).Decode(&result)
	if err != nil {
		return false, err
	}

	enabled, ok := result["allow_login_ip_notification"].(bool)
	if !ok {
		return false, errors.New("invalid notification setting type")
	}

	return enabled, nil
}

// GetUserByEmail retrieves a user by their email address
func (s *userService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := s.userCollection.FindOne(ctx, bson.M{"primary_email": email}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// IsSeller checks if the specified user is a seller in the database
func (s *userService) IsSeller(ctx context.Context, userID primitive.ObjectID) (bool, error) {
	err := s.userCollection.FindOne(ctx, bson.M{"_id": userID, "is_seller": true}).Err()
	if err == mongo.ErrNoDocuments {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}

// DeleteUser deletes all personal user data while preserving marketplace integrity
func (s *userService) DeleteUser(ctx context.Context, userID primitive.ObjectID) (*DeleteUserResult, error) {
	log.Printf("Starting user deletion process for user ID: %s", userID.Hex())

	// Verify user exists
	var user models.User
	err := s.userCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to find user: %v", err)
	}

	result := &DeleteUserResult{}

	// Execute deletion in transaction
	transactionResult, err := ExecuteTransaction(ctx, func(ctx mongo.SessionContext) (any, error) {
		// Step 1: Anonymize shops owned by the user
		shopsAnonymized, err := s.anonymizeUserShops(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to anonymize shops: %v", err)
		}
		result.ShopsAnonymized = shopsAnonymized

		// Step 2: Anonymize listings owned by the user
		listingsAnonymized, err := s.anonymizeUserListings(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to anonymize listings: %v", err)
		}
		result.ListingsAnonymized = listingsAnonymized

		// Step 3: Anonymize reviews written by the user
		reviewsAnonymized, err := s.anonymizeUserReviews(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to anonymize reviews: %v", err)
		}
		result.ReviewsAnonymized = reviewsAnonymized

		// Step 4: Remove user from shop followers
		err = s.removeUserFromShopFollowers(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to remove user from shop followers: %v", err)
		}

		// Step 5: Delete personal data
		addressesDeleted, err := s.deleteUserAddresses(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to delete user addresses: %v", err)
		}
		result.AddressesDeleted = addressesDeleted

		paymentInfoDeleted, err := s.deleteSellerPaymentInfo(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to delete seller payment info: %v", err)
		}
		result.PaymentInfoDeleted = paymentInfoDeleted

		paymentCardsDeleted, err := s.deleteUserPaymentCards(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to delete user payment cards: %v", err)
		}
		result.PaymentCardsDeleted = paymentCardsDeleted

		cartItemsDeleted, err := s.deleteUserCartItems(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to delete cart items: %v", err)
		}
		result.CartItemsDeleted = cartItemsDeleted

		wishlistDeleted, err := s.deleteUserWishlist(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to delete wishlist: %v", err)
		}
		result.WishlistDeleted = wishlistDeleted

		favoritesDeleted, err := s.deleteUserFavorites(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to delete favorites: %v", err)
		}
		result.FavoritesDeleted = favoritesDeleted

		loginHistoriesDeleted, err := s.deleteUserLoginHistories(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to delete login histories: %v", err)
		}
		result.LoginHistoriesDeleted = loginHistoriesDeleted

		tokensDeleted, err := s.deleteUserTokens(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to delete tokens: %v", err)
		}
		result.TokensDeleted = tokensDeleted

		notificationsDeleted, err := s.deleteUserNotifications(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to delete notifications: %v", err)
		}
		result.NotificationsDeleted = notificationsDeleted

		// Step 6: Delete seller verification data
		_, err = s.sellerVerificationCollection.DeleteMany(ctx, bson.M{"user_id": userID})
		if err != nil {
			return nil, fmt.Errorf("failed to delete seller verification: %v", err)
		}

		// Step 7: Delete user deletion requests
		_, err = s.userDeletionCollection.DeleteMany(ctx, bson.M{"user_id": userID})
		if err != nil {
			return nil, fmt.Errorf("failed to delete user deletion requests: %v", err)
		}

		// Step 8: Finally delete the user profile
		deleteResult, err := s.userCollection.DeleteOne(ctx, bson.M{"_id": userID})
		if err != nil {
			return nil, fmt.Errorf("failed to delete user profile: %v", err)
		}
		result.UserDeleted = deleteResult.DeletedCount > 0

		return result, nil
	})

	if err != nil {
		log.Printf("User deletion transaction failed for user %s: %v", userID.Hex(), err)
		return nil, err
	}

	log.Printf("User deletion completed successfully for user %s", userID.Hex())
	return transactionResult.(*DeleteUserResult), nil
}

// Helper methods for user deletion

func (s *userService) anonymizeUserShops(ctx mongo.SessionContext, userID primitive.ObjectID) (int, error) {
	// Update shops to anonymize owner information
	filter := bson.M{"user_id": userID}
	update := bson.M{
		"$set": bson.M{
			"user.first_name": "Former",
			"user.last_name":  "Seller",
			"user.login_name": "deleted_user_" + userID.Hex()[:8],
			"user.thumbnail":  common.DEFAULT_USER_THUMBNAIL,
		},
	}

	result, err := s.shopCollection.UpdateMany(ctx, filter, update)
	if err != nil {
		return 0, err
	}
	return int(result.ModifiedCount), nil
}

func (s *userService) anonymizeUserListings(ctx mongo.SessionContext, userID primitive.ObjectID) (int, error) {
	// Note: We don't need to anonymize listings as they don't contain personal user info
	// The shop reference will already be anonymized from the previous step
	count, err := s.listingCollection.CountDocuments(ctx, bson.M{"user_id": userID})
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (s *userService) anonymizeUserReviews(ctx mongo.SessionContext, userID primitive.ObjectID) (int, error) {
	// Anonymize reviews written by the user
	filter := bson.M{"user_id": userID}
	update := bson.M{
		"$set": bson.M{
			"user.first_name": "Former",
			"user.last_name":  "Customer",
			"user.login_name": "deleted_user",
			"user.thumbnail":  common.DEFAULT_USER_THUMBNAIL,
		},
	}

	result, err := s.listingReviewCollection.UpdateMany(ctx, filter, update)
	if err != nil {
		return 0, err
	}
	return int(result.ModifiedCount), nil
}

func (s *userService) removeUserFromShopFollowers(ctx mongo.SessionContext, userID primitive.ObjectID) error {
	// Remove user from all shop followers lists
	_, err := s.shopFollowerCollection.DeleteMany(ctx, bson.M{"user_id": userID})
	return err
}

func (s *userService) deleteUserAddresses(ctx mongo.SessionContext, userID primitive.ObjectID) (int, error) {
	result, err := s.userAddressCollection.DeleteMany(ctx, bson.M{"user_id": userID})
	if err != nil {
		return 0, err
	}
	return int(result.DeletedCount), nil
}

func (s *userService) deleteSellerPaymentInfo(ctx mongo.SessionContext, userID primitive.ObjectID) (int, error) {
	result, err := s.sellerPaymentInfoCollection.DeleteMany(ctx, bson.M{"user_id": userID})
	if err != nil {
		return 0, err
	}
	return int(result.DeletedCount), nil
}

func (s *userService) deleteUserPaymentCards(ctx mongo.SessionContext, userID primitive.ObjectID) (int, error) {
	result, err := s.userPaymentCardsCollection.DeleteMany(ctx, bson.M{"user_id": userID})
	if err != nil {
		return 0, err
	}
	return int(result.DeletedCount), nil
}

func (s *userService) deleteUserCartItems(ctx mongo.SessionContext, userID primitive.ObjectID) (int, error) {
	result, err := s.userCartCollection.DeleteMany(ctx, bson.M{"user_id": userID})
	if err != nil {
		return 0, err
	}
	return int(result.DeletedCount), nil
}

func (s *userService) deleteUserWishlist(ctx mongo.SessionContext, userID primitive.ObjectID) (int, error) {
	result, err := s.wishListCollection.DeleteMany(ctx, bson.M{"user_id": userID})
	if err != nil {
		return 0, err
	}
	return int(result.DeletedCount), nil
}

func (s *userService) deleteUserFavorites(ctx mongo.SessionContext, userID primitive.ObjectID) (int, error) {
	// Delete favorite listings
	result1, err := s.userFavoriteListingCollection.DeleteMany(ctx, bson.M{"user_id": userID})
	if err != nil {
		return 0, err
	}

	// Delete favorite shops
	result2, err := s.userFavoriteShopCollection.DeleteMany(ctx, bson.M{"user_id": userID})
	if err != nil {
		return int(result1.DeletedCount), err
	}

	return int(result1.DeletedCount + result2.DeletedCount), nil
}

func (s *userService) deleteUserLoginHistories(ctx mongo.SessionContext, userID primitive.ObjectID) (int, error) {
	result, err := s.loginHistoryCollection.DeleteMany(ctx, bson.M{"user_uid": userID})
	if err != nil {
		return 0, err
	}
	return int(result.DeletedCount), nil
}

func (s *userService) deleteUserTokens(ctx mongo.SessionContext, userID primitive.ObjectID) (int, error) {
	// Delete password reset tokens
	result1, err := s.passwordResetTokenCollection.DeleteMany(ctx, bson.M{"user_uid": userID})
	if err != nil {
		return 0, err
	}

	// Delete email verification tokens
	result2, err := s.emailVerificationTokenCollection.DeleteMany(ctx, bson.M{"user_uid": userID})
	if err != nil {
		return int(result1.DeletedCount), err
	}

	return int(result1.DeletedCount + result2.DeletedCount), nil
}

func (s *userService) deleteUserNotifications(ctx mongo.SessionContext, userID primitive.ObjectID) (int, error) {
	result, err := s.userNotificationSettingsCollection.DeleteMany(ctx, bson.M{"user_id": userID})
	if err != nil {
		return 0, err
	}
	return int(result.DeletedCount), nil
}

// CreateNotificationSettings implements UserService.
func (s *userService) CreateNotificationSettings(ctx context.Context, userID primitive.ObjectID, req models.UserNotificationSettings) (primitive.ObjectID, error) {
	result, err := s.userNotificationSettingsCollection.InsertOne(ctx, req)
	if err != nil {
		return primitive.NilObjectID, err
	}

	internal.PublishCacheMessage(ctx, internal.CacheInvalidateUserNotificationSettings, userID.Hex())
	return result.InsertedID.(primitive.ObjectID), nil
}

// GetNotificationSettings implements UserService.
func (s *userService) GetNotificationSettings(ctx context.Context, userID primitive.ObjectID) (models.UserNotificationSettings, error) {

	var notificationSettings models.UserNotificationSettings
	err := common.NotificationCollection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&notificationSettings)
	if err != nil {
		log.Printf("No notification settings configured for user %v, creating new notification profile", userID)
		if err == mongo.ErrNoDocuments {
			notificationSettings := models.UserNotificationSettings{
				ID:                   primitive.NewObjectID(),
				UserID:               userID,
				NewMessage:           true,
				NewFollower:          true,
				NewsAndFeatures:      true,
				EmailEnabled:         true,
				SMSEnabled:           true,
				PushEnabled:          true,
				OrderUpdates:         true,
				PaymentConfirmations: true,
				DeliveryUpdates:      true,
				CreatedAt:            time.Time{},
				ModifiedAt:           time.Time{},
			}

			_, err = s.userNotificationSettingsCollection.InsertOne(ctx, notificationSettings)
			if err != nil {
				return models.UserNotificationSettings{}, err
			}

			return notificationSettings, nil
		}

		return models.UserNotificationSettings{}, err
	}

	return notificationSettings, nil
}

// UpdateNotificationSettings implements UserService.
func (s *userService) UpdateNotificationSettings(ctx context.Context, userID primitive.ObjectID, field, value string) error {
	updateBool := false
	if value == "true" {
		updateBool = true
	}

	var update bson.M
	switch field {
	case "new_message":
		{
			update = bson.M{"$set": bson.M{"new_message": updateBool}}
			break
		}
	case "new_follower":
		{
			update = bson.M{"$set": bson.M{"new_follower": updateBool}}
			break
		}
	case "delivery_updates":
		{
			update = bson.M{"$set": bson.M{"delivery_updates": updateBool}}
			break
		}
	case "payment_confirmations":
		{
			update = bson.M{"$set": bson.M{"payment_confirmations": updateBool}}
			break
		}
	case "order_updates":
		{
			update = bson.M{"$set": bson.M{"order_updates": updateBool}}
			break
		}
	case "news_and_features":
		{
			update = bson.M{"$set": bson.M{"news_and_features": updateBool}}
			break
		}
	case "sms_enabled":
		{
			update = bson.M{"$set": bson.M{"sms_enabled": updateBool}}
			break
		}
	case "push_enabled":
		{
			update = bson.M{"$set": bson.M{"push_enabled": updateBool}}
			break
		}
	case "email_enabled":
		{
			update = bson.M{"$set": bson.M{"email_enabled": updateBool}}
			break
		}
	case "support_message":
		{
			update = bson.M{"$set": bson.M{"support_message": updateBool}}
			break
		}
	case "promotional":
		{
			update = bson.M{"$set": bson.M{"promotional": updateBool}}
			break
		}
	default:
		{
			errorMsg := fmt.Sprintf("Invalid update field %v", field)
			return errors.New(errorMsg)
		}
	}

	filter := bson.M{"user_id": userID}
	_, err := s.userNotificationSettingsCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	internal.PublishCacheMessage(ctx, internal.CacheInvalidateUserNotificationSettings, userID.Hex())
	return nil
}
