package controllers

import (
	"context"
	"net/http"

	"khoomi-api-io/api/internal/auth"
	"khoomi-api-io/api/internal/common"
	"khoomi-api-io/api/pkg/models"
	"khoomi-api-io/api/pkg/services"
	"khoomi-api-io/api/pkg/util"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// VerificationController handles verification-related HTTP requests
type VerificationController struct {
	verificationService  services.VerificationService
	notificationService  services.NotificationService
}

// InitVerificationController creates a new VerificationController with dependency injection
func InitVerificationController(verificationService services.VerificationService, notificationService services.NotificationService) *VerificationController {
	return &VerificationController{
		verificationService: verificationService,
		notificationService: notificationService,
	}
}

// CreateSellerVerificationProfile creates a new seller verification profile
func (vc *VerificationController) CreateSellerVerificationProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQUEST_TIMEOUT_SECS)
		defer cancel()

		shopId := c.Param("shopid")
		shopIdObj, err := primitive.ObjectIDFromHex(shopId)
		if err != nil {
			util.HandleError(c, http.StatusUnprocessableEntity, err)
			return
		}

		// Get authenticated user session
		session_, err := auth.GetSessionAuto(c)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err)
			return
		}

		var verificationJson models.CreateSellerVerificationRequest
		err = c.BindJSON(&verificationJson)
		if err != nil {
			util.HandleError(c, http.StatusUnprocessableEntity, err)
			return
		}

		// Validate request body
		if validationErr := common.Validate.Struct(&verificationJson); validationErr != nil {
			util.HandleError(c, http.StatusUnprocessableEntity, validationErr)
			return
		}

		// Create verification profile using service
		verificationID, err := vc.verificationService.CreateSellerVerificationProfile(ctx, session_.UserId, shopIdObj, verificationJson)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		util.HandleSuccess(c, http.StatusCreated, "Seller verification profile created successfully", verificationID)
	}
}

// GetSellerVerificationProfile retrieves a seller verification profile
func (vc *VerificationController) GetSellerVerificationProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQUEST_TIMEOUT_SECS)
		defer cancel()

		shopId := c.Param("shopid")
		shopIdObj, err := primitive.ObjectIDFromHex(shopId)
		if err != nil {
			util.HandleError(c, http.StatusUnprocessableEntity, err)
			return
		}

		// Get authenticated user session
		session_, err := auth.GetSessionAuto(c)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err)
			return
		}

		// Get verification profile using service
		verificationProfile, err := vc.verificationService.GetSellerVerificationProfile(ctx, session_.UserId, shopIdObj)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Seller verification profile retrieved successfully", verificationProfile)
	}
}
