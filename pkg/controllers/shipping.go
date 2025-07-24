package controllers

import (
	"log"
	"net/http"

	"khoomi-api-io/api/internal/common"
	"khoomi-api-io/api/pkg/models"
	"khoomi-api-io/api/pkg/services"
	"khoomi-api-io/api/pkg/util"

	"github.com/gin-gonic/gin"
)

type ShippingController struct {
	shippingService     services.ShippingService
	shopService         services.ShopService
	notificationService services.NotificationService
}

// InitShippingController initializes a new ShippingController with dependencies
func InitShippingController(shippingService services.ShippingService, shopService services.ShopService, notificationService services.NotificationService) *ShippingController {
	return &ShippingController{
		shippingService:     shippingService,
		shopService:         shopService,
		notificationService: notificationService,
	}
}

func (sc *ShippingController) CreateShopShippingProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := WithTimeout()
		defer cancel()

		shopIdObj, ok := ParseObjectIDParam(c, "shopid")
		if !ok {
			return
		}

		userID, ok := ValidateAndGetUserID(c)
		if !ok {
			return
		}

		err := sc.shopService.VerifyShopOwnership(c, userID, shopIdObj)
		if err != nil {
			log.Printf("Error you the shop owner: %s\n", err.Error())
			util.HandleError(c, http.StatusForbidden, err)
			return
		}

		var shippingJson models.ShopShippingProfileRequest
		if !BindJSONAndValidate(c, &shippingJson) {
			return
		}

		shippingID, err := sc.shippingService.CreateShopShippingProfile(ctx, userID, shopIdObj, shippingJson)
		if err != nil {
			log.Println(err)
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		util.HandleSuccess(c, http.StatusOK, "document inserted", shippingID)
	}
}

func (sc *ShippingController) GetShopShippingProfileInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := WithTimeout()
		defer cancel()

		profileId, ok := ParseObjectIDParam(c, "id")
		if !ok {
			return
		}

		shippingProfile, err := sc.shippingService.GetShopShippingProfile(ctx, profileId)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		util.HandleSuccess(c, http.StatusOK, "success", shippingProfile)
	}
}

func (sc *ShippingController) GetShopShippingProfileInfos() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := WithTimeout()
		defer cancel()

		shopIDObject, ok := ParseObjectIDParam(c, "shopid")
		if !ok {
			return
		}

		paginationArgs := common.GetPaginationArgs(c)
		shippingProfiles, count, err := sc.shippingService.GetShopShippingProfiles(ctx, shopIDObject, paginationArgs)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		HandlePaginationAndResponse(c, shippingProfiles, count, paginationArgs, "success")
	}
}

func (s *ShippingController) UpdateShopShippingProfileInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := WithTimeout()
		defer cancel()

		shopId, ok := ParseObjectIDParam(c, "shopid")
		if !ok {
			return
		}

		userID, ok := ValidateAndGetUserID(c)
		if !ok {
			return
		}

		err := s.shopService.VerifyShopOwnership(ctx, userID, shopId)

		if err != nil {
			log.Printf("Error you the shop owner: %s\n", err.Error())
			util.HandleError(c, http.StatusForbidden, err)
			return
		}

		var shippingJson models.ShopShippingProfileRequest
		if !BindJSONAndValidate(c, &shippingJson) {
			return
		}

		res, err := s.shippingService.UpdateShippingProfile(ctx, shopId, shippingJson)

		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		util.HandleSuccess(c, http.StatusOK, "document inserted", res)
	}
}
