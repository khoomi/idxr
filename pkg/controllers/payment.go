package controllers

import (
	"log"
	"net/http"

	"khoomi-api-io/api/internal/helpers"
	"khoomi-api-io/api/pkg/models"
	"khoomi-api-io/api/pkg/services"
	"khoomi-api-io/api/pkg/util"

	"github.com/gin-gonic/gin"
)

type PaymentController struct {
	paymentService      services.PaymentService
	userService         services.UserService
	notificationService services.NotificationService
}

// InitPaymentController initializes a new PaymentController with dependencies
func InitPaymentController(paymentService services.PaymentService, userService services.UserService, notificationService services.NotificationService) *PaymentController {
	return &PaymentController{
		paymentService:      paymentService,
		userService:         userService,
		notificationService: notificationService,
	}
}

// CreateSellerPaymentInformation -> POST /shop/:shopId/payment-information/
func (pc *PaymentController) CreateSellerPaymentInformation() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := WithTimeout()
		defer cancel()

		userId, ok := ValidateSellerAccess(c, pc.userService)
		if !ok {
			return
		}

		var paymentInfo models.SellerPaymentInformationRequest
		if !BindJSONAndValidate(c, &paymentInfo) {
			return
		}

		paymentID, err := pc.paymentService.CreateSellerPaymentInformation(ctx, userId, paymentInfo)
		if err != nil {
			log.Println(err)
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		log.Printf("User %v added their payment account information", userId)

		util.HandleSuccess(c, http.StatusOK, "Payment account information created successfully", paymentID.Hex())
	}
}

// GetSellerPaymentInformations -> GET /shop/:shopId/payment-information/
func (pc *PaymentController) GetSellerPaymentInformations() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := WithTimeout()
		defer cancel()

		userId, ok := ValidateSellerAccess(c, pc.userService)
		if !ok {
			return
		}

		paginationArgs := helpers.GetPaginationArgs(c)
		paymentInfos, count, err := pc.paymentService.GetSellerPaymentInformations(ctx, userId, paginationArgs)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}

		HandlePaginationAndResponse(c, paymentInfos, count, paginationArgs, "success")
	}
}

// ChangeDefaultSellerPaymentInformation -> PUT /shop/:shopId/payment-information/:paymentInfoId
func (pc *PaymentController) ChangeDefaultSellerPaymentInformation() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := WithTimeout()
		defer cancel()

		userId, ok := ValidateSellerAccess(c, pc.userService)
		if !ok {
			return
		}

		paymentObjectID, ok := ParseObjectIDParam(c, "paymentInfoId")
		if !ok {
			return
		}

		err := pc.paymentService.ChangeDefaultSellerPaymentInformation(ctx, userId, paymentObjectID)
		if err != nil {
			util.HandleError(c, http.StatusNotModified, err)
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Default payment has been succesfuly changed.", 1)
	}
}

// DeleteSellerPaymentInformation -> DELETE /shop/:shopId/payment-information/:paymentInfoId
func (pc *PaymentController) DeleteSellerPaymentInformation() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := WithTimeout()
		defer cancel()

		paymentObjectID, ok := ParseObjectIDParam(c, "paymentInfoId")
		if !ok {
			return
		}

		userId, ok := ValidateSellerAccess(c, pc.userService)
		if !ok {
			return
		}

		err := pc.paymentService.DeleteSellerPaymentInformation(ctx, userId, paymentObjectID)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Payment information deleted successfully", 1)
	}
}

func (pc *PaymentController) CompletedPaymentOnboarding() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := WithTimeout()
		defer cancel()

		userId, ok := ValidateSellerAccess(c, pc.userService)
		if !ok {
			return
		}

		hasPaymentInfo, err := pc.paymentService.HasSellerPaymentInformation(ctx, userId)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Success", hasPaymentInfo)
	}
}

// / CreatePaymentInformation -> POST /:userId/payment/cards
func (pc *PaymentController) CreatePaymentCard() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := WithTimeout()
		defer cancel()

		userId, ok := ValidateAndGetUserID(c)
		if !ok {
			return
		}

		var cardInfo models.PaymentCardInformationRequest
		if !BindJSONAndValidate(c, &cardInfo) {
			return
		}

		cardID, err := pc.paymentService.CreatePaymentCard(ctx, userId, cardInfo)
		if err != nil {
			log.Println(err)
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		log.Printf("User %v added a card", userId)

		util.HandleSuccess(c, http.StatusOK, "new Card created successfully", cardID.Hex())
	}
}

// / GetPaymentCards-> GET /:userId/payment/cards
func (pc *PaymentController) GetPaymentCards() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := WithTimeout()
		defer cancel()

		userId, ok := ValidateAndGetUserID(c)
		if !ok {
			return
		}

		paginationArgs := helpers.GetPaginationArgs(c)
		paymentInfos, count, err := pc.paymentService.GetPaymentCards(ctx, userId, paginationArgs)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}

		HandlePaginationAndResponse(c, paymentInfos, count, paginationArgs, "success")
	}
}

// / ChangeDefaulterPaymentCard-> PUT /:userId/payment/cards/:id
func (pc *PaymentController) ChangeDefaultPaymentCard() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := WithTimeout()
		defer cancel()

		userId, ok := ValidateAndGetUserID(c)
		if !ok {
			return
		}

		paymentObjectID, ok := ParseObjectIDParam(c, "id")
		if !ok {
			return
		}

		err := pc.paymentService.ChangeDefaultPaymentCard(ctx, userId, paymentObjectID)
		if err != nil {
			util.HandleError(c, http.StatusNotModified, err)
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Default card has been succesfuly changed.", 1)
	}
}

// / DeletePaymentCard-> DELETE /user/:userId/payment/card/:id
func (pc *PaymentController) DeletePaymentCard() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := WithTimeout()
		defer cancel()

		paymentObjectID, ok := ParseObjectIDParam(c, "id")
		if !ok {
			return
		}

		userId, ok := ValidateAndGetUserID(c)
		if !ok {
			return
		}

		err := pc.paymentService.DeletePaymentCard(ctx, userId, paymentObjectID)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}

		util.HandleSuccess(c, http.StatusOK, "card deleted successfully", 1)
	}
}
