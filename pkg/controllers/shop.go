package controllers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"khoomi-api-io/api/internal"
	auth "khoomi-api-io/api/internal/auth"
	"khoomi-api-io/api/internal/common"
	"khoomi-api-io/api/pkg/models"
	"khoomi-api-io/api/pkg/services"
	"khoomi-api-io/api/pkg/util"

	"github.com/cloudinary/cloudinary-go/api/uploader"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

type ShopController struct {
	shopService         services.ShopService
	notificationService services.NotificationService
}

// InitShopController initializes a new ShopController with dependencies
func InitShopController(shopService services.ShopService, notificationService services.NotificationService) *ShopController {
	return &ShopController{
		shopService:         shopService,
		notificationService: notificationService,
	}
}

// CheckShopNameAvailability -> Check if a given shop name is available or not.
// /api/shop/check/:shop_username
func (sc *ShopController) CheckShopNameAvailability() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shopName := c.Param("username")

		isAvailable, err := sc.shopService.CheckShopNameAvailability(ctx, shopName)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, errors.New("internal server error on checking shop username availability"))
			return
		}

		if isAvailable {
			util.HandleSuccess(c, http.StatusOK, "Congrats! shop username is available :xD", true)
		} else {
			util.HandleSuccess(c, http.StatusOK, "shop username is already taken", false)
		}
	}
}

func (sc *ShopController) CreateShop(emailService services.EmailService) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		session_, err := auth.GetSessionAuto(c)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err)
			return
		}

		loginName, userEmail := session_.LoginName, session_.Email
		shopName := c.Request.FormValue("name")
		shopUserName := strings.ToLower(c.Request.FormValue("username"))
		shopDescription := c.Request.FormValue("description")

		if err := util.ValidateShopName(shopName); err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}
		if err := util.ValidateShopUserName(shopUserName); err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}
		if err := util.ValidateShopDescription(shopDescription); err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		// Logo file handling
		logoFile, _, err := c.Request.FormFile("logo")
		logoUploadUrl := ""
		var logoUploadResult uploader.UploadResult
		if err == nil {
			logoUploadResult, err = util.FileUpload(models.File{File: logoFile})
			if err != nil {
				errMsg := fmt.Sprintf("Logo failed to upload - %v", err.Error())
				util.HandleError(c, http.StatusInternalServerError, errors.New(errMsg))
				return
			}
			logoUploadUrl = logoUploadResult.SecureURL
		}

		// Banner file handling
		bannerFile, _, err := c.Request.FormFile("banner")
		bannerUploadUrl := ""
		var bannerUploadResult uploader.UploadResult
		if err == nil {
			bannerUploadResult, err = util.FileUpload(models.File{File: bannerFile})
			if err != nil {
				errMsg := fmt.Sprintf("Banner failed to upload - %v", err.Error())
				util.HandleError(c, http.StatusInternalServerError, errors.New(errMsg))
				return
			}
			bannerUploadUrl = bannerUploadResult.SecureURL
		}

		req := services.CreateShopRequest{
			Name:        shopName,
			Username:    shopUserName,
			Description: shopDescription,
			LogoFile:    logoUploadUrl,
			BannerFile:  bannerUploadUrl,
		}

		shopID, err := sc.shopService.CreateShop(ctx, session_.UserId, req)
		if err != nil {
			// delete media on error
			if logoUploadResult.PublicID != "" {
				util.DestroyMedia(logoUploadResult.PublicID)
			}
			if bannerUploadResult.PublicID != "" {
				util.DestroyMedia(bannerUploadResult.PublicID)
			}
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		if emailService == nil {
			emailService = services.NewEmailService()
		}

		// send success shop creation notification
		emailService.SendNewShopEmail(userEmail, loginName, shopName)

		internal.PublishCacheMessage(c, internal.CacheInvalidateShop, shopID.Hex())
		util.HandleSuccess(c, http.StatusOK, shopID.Hex(), shopID.Hex())
	}
}

func (sc *ShopController) UpdateShopInformation() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shopId, myId, err := common.MyShopIdAndMyId(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		shopName := c.Request.FormValue("name")
		shopUsername := c.Request.FormValue("username")
		description := c.Request.FormValue("description")

		// Validate form data first before doing expensive file uploads
		if shopName != "" {
			if err := util.ValidateShopName(shopName); err != nil {
				util.HandleError(c, http.StatusBadRequest, err)
				return
			}
		}
		if shopUsername != "" {
			if err := util.ValidateShopUserName(shopUsername); err != nil {
				util.HandleError(c, http.StatusBadRequest, err)
				return
			}
		}
		if description != "" {
			if err := util.ValidateShopDescription(description); err != nil {
				util.HandleError(c, http.StatusBadRequest, err)
				return
			}
		}

		// Only do file uploads after validation passes
		var logoUploadResult uploader.UploadResult
		logoUploadUrl := ""
		if fileHeader, err := c.FormFile("logo_url"); err == nil {
			file, err := fileHeader.Open()
			if err != nil {
				util.HandleError(c, http.StatusInternalServerError, err)
				return
			}
			defer file.Close()

			logoUploadResult, err = util.FileUpload(models.File{File: file})
			if err != nil {
				log.Printf("Logo Image upload failed - %v", err.Error())
				util.HandleError(c, http.StatusInternalServerError, err)
				return
			}
			logoUploadUrl = logoUploadResult.SecureURL
		} else if err != http.ErrMissingFile {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		var bannerUploadResult uploader.UploadResult
		bannerUploadUrl := ""
		if fileHeader, err := c.FormFile("banner_url"); err == nil {
			file, err := fileHeader.Open()
			if err != nil {
				util.HandleError(c, http.StatusInternalServerError, err)
				return
			}
			defer file.Close()

			bannerUploadResult, err = util.FileUpload(models.File{File: file})
			if err != nil {
				log.Printf("Banner Image upload failed - %v", err.Error())
				util.HandleError(c, http.StatusInternalServerError, err)
				return
			}
			bannerUploadUrl = bannerUploadResult.SecureURL
		} else if err != http.ErrMissingFile {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		req := services.UpdateShopRequest{
			Name:        shopName,
			Username:    shopUsername,
			Description: description,
			LogoFile:    logoUploadUrl,
			BannerFile:  bannerUploadUrl,
		}

		err = sc.shopService.UpdateShopInformation(ctx, shopId, myId, req)
		if err != nil {
			// delete media on error
			if logoUploadResult.PublicID != "" {
				util.DestroyMedia(logoUploadResult.PublicID)
			}
			if bannerUploadResult.PublicID != "" {
				util.DestroyMedia(bannerUploadResult.PublicID)
			}

			if err.Error() == "no update data provided" {
				util.HandleError(c, http.StatusOK, err)
			} else {
				util.HandleError(c, http.StatusExpectationFailed, err)
			}
			return
		}

		internal.PublishCacheMessage(c, internal.CacheInvalidateShop, shopId.Hex())

		util.HandleSuccess(c, http.StatusOK, "Shop information updated successfully", shopId.Hex())
	}
}

func (sc *ShopController) UpdateMyShopStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var payload models.UpdateShopStatusReq
		if err := c.BindJSON(&payload); err != nil {
			log.Println(err)
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		shopId, myId, err := common.MyShopIdAndMyId(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		err = sc.shopService.UpdateShopStatus(ctx, shopId, myId, payload.Status)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		internal.PublishCacheMessage(c, internal.CacheInvalidateShop, shopId.Hex())

		util.HandleSuccess(c, http.StatusOK, "Shop status was updated successful", shopId.Hex())
	}
}

func (sc *ShopController) UpdateShopAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var payload models.ShopAddress
		if err := c.BindJSON(&payload); err != nil {
			log.Println(err)
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		shopId, myId, err := common.MyShopIdAndMyId(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		err = sc.shopService.UpdateShopAddress(ctx, shopId, myId, payload)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		internal.PublishCacheMessage(c, internal.CacheInvalidateShop, shopId.Hex())

		util.HandleSuccess(c, http.StatusOK, "Shop address was updated successful", shopId.Hex())
	}
}

// GetShop - api/shops/:shopid
func (sc *ShopController) GetShop() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shopID := c.Param("shopid")
		withCategory := len(c.Query("category")) > 0

		shop, err := sc.shopService.GetShop(ctx, shopID, withCategory)
		if err != nil {
			if err.Error() == "no shop found" {
				util.HandleError(c, http.StatusNotFound, err)
			} else {
				util.HandleError(c, http.StatusInternalServerError, err)
			}
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Shop retrieved successfully", *shop)
	}
}

// GetShopByOwnerUserId - api/users/:userid/shops
func (sc *ShopController) GetShopByOwnerUserId() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		userIDStr := c.Param("userid")
		userID, err := primitive.ObjectIDFromHex(userIDStr)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		shop, err := sc.shopService.GetShopByOwnerUserId(ctx, userID)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Shop retrieved successfully", *shop)
	}
}

// GetShops - api/shops/?limit=50&skip=0
func (sc *ShopController) GetShops() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		paginationArgs := common.GetPaginationArgs(c)

		shops, err := sc.shopService.GetShops(ctx, paginationArgs)
		if err != nil {
			log.Printf("error finding shops: %v", err.Error())
			util.HandleError(c, http.StatusNotFound, err)
			return
		}

		util.HandleSuccessMeta(c, http.StatusOK, "Shops retrieved successfully", shops, gin.H{"pagination": util.Pagination{
			Limit: paginationArgs.Limit,
			Skip:  paginationArgs.Skip,
		}})
	}
}

// SearchShops - api/shops/:shopid/search?q=khoomi&limit=50&skip=0
func (sc *ShopController) SearchShops() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		query := c.Query("q")
		paginationArgs := common.GetPaginationArgs(c)

		shops, count, err := sc.shopService.SearchShops(ctx, query, paginationArgs)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		util.HandleSuccessMeta(c, http.StatusOK, "Shops found", shops,
			gin.H{"pagination": util.Pagination{
				Limit: paginationArgs.Limit,
				Skip:  paginationArgs.Skip,
				Count: count,
			}})
	}
}

func (sc *ShopController) UpdateShopField() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQUEST_TIMEOUT_SECS)
		defer cancel()
		now := time.Now()

		shopId, myId, err := common.MyShopIdAndMyId(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		field := c.Query("field")
		if common.IsEmptyString(field) {
			util.HandleError(c, http.StatusBadRequest, errors.New("field query parameter is required"))
			return
		}

		filter := bson.M{"_id": shopId, "user_id": myId}

		switch field {
		case "banner":
			{
				action := c.Query("action")
				if common.IsEmptyString(action) {
					util.HandleError(c, http.StatusBadRequest, errors.New("action query parameter is required"))
					return
				}
				switch action {
				case "update":
					{
						contentType := c.GetHeader("Content-Type")
						if !strings.HasPrefix(contentType, "multipart/form-data") {
							util.HandleError(c, http.StatusBadRequest, errors.New("request must be multipart/form-data"))
							return
						}

						form, err := c.MultipartForm()
						if err != nil {
							log.Printf("Failed to parse multipart form: %v", err)
							util.HandleError(c, http.StatusBadRequest, fmt.Errorf("failed to parse multipart form: %v", err))
							return
						}

						log.Printf("Available form files: %+v", form.File)

						bannerFile, err := c.FormFile("banner")
						if err != nil {
							log.Printf("FormFile error for 'banner': %v", err)
							util.HandleError(c, http.StatusBadRequest, fmt.Errorf("failed to get banner file: %v", err))
							return
						}

						src, err := bannerFile.Open()
						if err != nil {
							util.HandleError(c, http.StatusInternalServerError, err)
							return
						}
						defer src.Close()

						bannerUploadResult, err := util.FileUpload(models.File{File: src})
						if err != nil {
							errMsg := fmt.Sprintf("Banner failed to upload - %v", err.Error())
							util.HandleError(c, http.StatusInternalServerError, errors.New(errMsg))
							return
						}

						update := bson.M{"banner_url": bannerUploadResult.SecureURL, "modified_at": now}
						res, err := common.ShopCollection.UpdateOne(ctx, filter, update)
						if err != nil || res.ModifiedCount == 0 {
							_, err = util.DestroyMedia(bannerUploadResult.PublicID)
							util.HandleError(c, http.StatusNotModified, err)
							return
						}

						internal.PublishCacheMessage(c, internal.CacheInvalidateShop, shopId.Hex())

						util.HandleSuccess(c, http.StatusOK, "Shop banner updated successfully", res.UpsertedID)
						return
					}
				case "delete":
					{
						update := bson.M{"$set": bson.M{"banner_url": common.DEFAULT_THUMBNAIL, "modified_at": now}}
						res, err := common.ShopCollection.UpdateOne(ctx, filter, update)
						if err != nil || res.ModifiedCount == 0 {
							util.HandleError(c, http.StatusNotModified, err)
							return
						}

						internal.PublishCacheMessage(c, internal.CacheInvalidateShop, shopId.Hex())

						util.HandleSuccess(c, http.StatusOK, "Shop logo updated successfully", res.UpsertedID)
						return
					}
				}
			}
		case "logo":
			{
				action := c.Query("action")
				if common.IsEmptyString(action) {
					util.HandleError(c, http.StatusBadRequest, errors.New("action query parameter is required"))
					return
				}
				switch action {
				case "update":
					{
						contentType := c.GetHeader("Content-Type")
						if !strings.HasPrefix(contentType, "multipart/form-data") {
							util.HandleError(c, http.StatusBadRequest, errors.New("request must be multipart/form-data"))
							return
						}

						form, err := c.MultipartForm()
						if err != nil {
							log.Printf("Failed to parse multipart form: %v", err)
							util.HandleError(c, http.StatusBadRequest, fmt.Errorf("failed to parse multipart form: %v", err))
							return
						}

						log.Printf("Available form files: %+v", form.File)

						logoFile, err := c.FormFile("logo")
						if err != nil {
							log.Printf("FormFile error for 'logo': %v", err)
							util.HandleError(c, http.StatusBadRequest, fmt.Errorf("failed to get logo file: %v", err))
							return
						}

						src, err := logoFile.Open()
						if err != nil {
							util.HandleError(c, http.StatusInternalServerError, err)
							return
						}
						defer src.Close()

						logoUploadResult, err := util.FileUpload(models.File{File: src})
						if err != nil {
							errMsg := fmt.Sprintf("Logo failed to upload - %v", err.Error())
							util.HandleError(c, http.StatusInternalServerError, errors.New(errMsg))
							return
						}

						update := bson.M{"$set": bson.M{"logo_url": logoUploadResult.SecureURL, "modified_at": now}}
						res, err := common.ShopCollection.UpdateOne(ctx, filter, update)
						if err != nil || res.ModifiedCount == 0 {
							_, err = util.DestroyMedia(logoUploadResult.PublicID)
							util.HandleError(c, http.StatusNotModified, err)
							return
						}

						internal.PublishCacheMessage(c, internal.CacheInvalidateShop, shopId.Hex())

						util.HandleSuccess(c, http.StatusOK, "Shop logo updated successfully", res.UpsertedID)
						return
					}
				case "delete":
					{
						update := bson.M{"$set": bson.M{"logo_url": common.DEFAULT_LOGO, "modified_at": now}}
						res, err := common.ShopCollection.UpdateOne(ctx, filter, update)
						if err != nil || res.ModifiedCount == 0 {
							util.HandleError(c, http.StatusNotModified, err)
							return
						}

						internal.PublishCacheMessage(c, internal.CacheInvalidateShop, shopId.Hex())

						util.HandleSuccess(c, http.StatusOK, "Shop logo updated successfully", res.UpsertedID)
						return
					}
				}
			}
		case "address":
			{
				var payload models.ShopAddress
				if err := c.Bind(&payload); err != nil {
					log.Println(err)
					util.HandleError(c, http.StatusBadRequest, err)
					return
				}

				shopId, myId, err := common.MyShopIdAndMyId(c)
				if err != nil {
					util.HandleError(c, http.StatusBadRequest, err)
					return
				}

				payload.ModifiedAt = now
				filter := bson.M{"_id": shopId, "user_id": myId}
				update := bson.M{"$set": bson.M{"address": payload, "modified_at": now}}
				res, err := common.ShopCollection.UpdateOne(ctx, filter, update)
				if err != nil {
					util.HandleError(c, http.StatusInternalServerError, err)
					return
				}
				if res.ModifiedCount == 0 {
					util.HandleError(c, http.StatusNotModified, errors.New("unknown error while trying to update shop"))
					return
				}

				internal.PublishCacheMessage(c, internal.CacheInvalidateShop, shopId.Hex())

				util.HandleSuccess(c, http.StatusOK, "Shop address was updated successful", shopId.Hex())
				return
			}
		case "vacation":
			{
				var vacation models.ShopVacationRequest
				if err := c.BindJSON(&vacation); err != nil {
					util.HandleError(c, http.StatusBadRequest, err)
					return
				}

				update := bson.M{"$set": bson.M{"vacation_message": vacation.Message, "is_vacation": vacation.IsVacation, "modified_at": now}}
				res, err := common.ShopCollection.UpdateOne(ctx, filter, update)
				if err != nil {
					util.HandleError(c, http.StatusInternalServerError, err)
					return
				}
				if res.ModifiedCount == 0 {
					util.HandleError(c, http.StatusNotFound, errors.New("no matching documents found"))
					return
				}

				internal.PublishCacheMessage(c, internal.CacheInvalidateShop, shopId.Hex())

				util.HandleSuccess(c, http.StatusOK, "Shop vacation updated successfully", res.UpsertedID)
				return
			}
		case "basic":
			{
				var basic models.ShopBasicInformationRequest
				if err := c.BindJSON(&basic); err != nil {
					util.HandleError(c, http.StatusBadRequest, err)
					return
				}

				err = util.ValidateShopName(basic.Name)
				if err != nil {
					util.HandleError(c, http.StatusBadRequest, err)
					return
				}

				err = util.ValidateShopDescription(basic.Description)
				if err != nil {
					util.HandleError(c, http.StatusBadRequest, err)
					return
				}

				update := bson.M{"$set": bson.M{"name": basic.Name, "is_live": basic.IsLive, "description": basic.Description, "sales_message": basic.SalesMessage, "announcement": basic.Announcement, "modified_at": now}}
				res, err := common.ShopCollection.UpdateOne(ctx, filter, update)
				if err != nil {
					util.HandleError(c, http.StatusInternalServerError, err)
					return
				}
				if res.ModifiedCount == 0 {
					util.HandleError(c, http.StatusNotFound, errors.New("no matching documents found"))
					return
				}

				internal.PublishCacheMessage(c, internal.CacheInvalidateShop, shopId.Hex())

				util.HandleSuccess(c, http.StatusOK, "Shop information updated successfully", res.UpsertedID)
				return
			}
		case "policy":
			{
				var payload models.ShopPolicy
				if err := c.Bind(&payload); err != nil {
					log.Println(err)
					util.HandleError(c, http.StatusBadRequest, err)
					return
				}

				shopId, myId, err := common.MyShopIdAndMyId(c)
				if err != nil {
					util.HandleError(c, http.StatusBadRequest, err)
					return
				}

				filter := bson.M{"_id": shopId, "user_id": myId}
				update := bson.M{"$set": bson.M{"policy": payload, "modified_at": now}}
				res, err := common.ShopCollection.UpdateOne(ctx, filter, update)
				if err != nil {
					util.HandleError(c, http.StatusInternalServerError, err)
					return
				}
				if res.ModifiedCount == 0 {
					util.HandleError(c, http.StatusNotModified, errors.New("unknown error while trying to update shop"))
					return
				}

				internal.PublishCacheMessage(c, internal.CacheInvalidateShop, shopId.Hex())

				util.HandleSuccess(c, http.StatusOK, "Shop policy was updated successful", shopId.Hex())

				return
			}
		default:
			util.HandleError(c, http.StatusBadRequest, errors.New("unsupported field. supported fields: banner, logo, address, vacation"))
			return
		}
	}
}

// UpdateShopAnnouncement - api/shops/:shopid/announcement
func (sc *ShopController) UpdateShopAnnouncement() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQUEST_TIMEOUT_SECS)
		defer cancel()

		var announcement models.ShopAnnouncementRequest
		if err := c.BindJSON(&announcement); err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		if announcement.Announcement == "" {
			util.HandleError(c, http.StatusBadRequest, errors.New("announcement cannot be empty"))
			return
		}

		if len(announcement.Announcement) > 100 {
			util.HandleError(c, http.StatusBadRequest, errors.New("announcement is too long"))
			return
		}

		shopId, myId, err := common.MyShopIdAndMyId(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		err = sc.shopService.UpdateShopAnnouncement(ctx, shopId, myId, announcement.Announcement)
		if err != nil {
			if err.Error() == "no matching documents found" {
				util.HandleError(c, http.StatusNotFound, err)
			} else {
				util.HandleError(c, http.StatusInternalServerError, err)
			}
			return
		}

		internal.PublishCacheMessage(c, internal.CacheInvalidateShop, shopId.Hex())

		util.HandleSuccess(c, http.StatusOK, "Shop announcement updated successfully", nil)
	}
}

func (sc *ShopController) UpdateShopVacation() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var vacation models.ShopVacationRequest
		if err := c.BindJSON(&vacation); err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		shopId, myId, err := common.MyShopIdAndMyId(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		err = sc.shopService.UpdateShopVacation(ctx, shopId, myId, vacation)
		if err != nil {
			if err.Error() == "no matching documents found" {
				util.HandleError(c, http.StatusNotFound, err)
			} else {
				util.HandleError(c, http.StatusInternalServerError, err)
			}
			return
		}

		internal.PublishCacheMessage(c, internal.CacheInvalidateShop, shopId.Hex())

		util.HandleSuccess(c, http.StatusOK, "Shop vacation updated successfully", nil)
	}
}

func (sc *ShopController) UpdateShopLogo() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shopId, myId, err := common.MyShopIdAndMyId(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		logoFile, _, err := c.Request.FormFile("logo")
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		logoUploadResult, err := util.FileUpload(models.File{File: logoFile})
		if err != nil {
			errMsg := fmt.Sprintf("Logo failed to upload - %v", err.Error())
			util.HandleError(c, http.StatusInternalServerError, errors.New(errMsg))
			return
		}

		err = sc.shopService.UpdateShopLogo(ctx, shopId, myId, logoUploadResult.SecureURL)
		if err != nil {
			// delete media on error
			util.DestroyMedia(logoUploadResult.PublicID)
			util.HandleError(c, http.StatusNotModified, err)
			return
		}

		internal.PublishCacheMessage(c, internal.CacheInvalidateShop, shopId.Hex())

		util.HandleSuccess(c, http.StatusOK, "Shop logo updated successfully", nil)
	}
}

func (sc *ShopController) UpdateShopBanner() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shopId, myId, err := common.MyShopIdAndMyId(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		bannerFile, _, err := c.Request.FormFile("banner")
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		bannerUploadResult, err := util.FileUpload(models.File{File: bannerFile})
		if err != nil {
			errMsg := fmt.Sprintf("Banner failed to upload - %v", err.Error())
			util.HandleError(c, http.StatusInternalServerError, errors.New(errMsg))
			return
		}

		err = sc.shopService.UpdateShopBanner(ctx, shopId, myId, bannerUploadResult.SecureURL)
		if err != nil {
			// delete media on error
			util.DestroyMedia(bannerUploadResult.PublicID)
			util.HandleError(c, http.StatusNotModified, err)
			return
		}

		internal.PublishCacheMessage(c, internal.CacheInvalidateShop, shopId.Hex())

		util.HandleSuccess(c, http.StatusOK, "Shop banner updated successfully", nil)
	}
}

func (sc *ShopController) UpdateShopGallery() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shopId, myId, err := common.MyShopIdAndMyId(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		imageFile, _, err := c.Request.FormFile("image")
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		imageUploadResult, err := util.FileUpload(models.File{File: imageFile})
		if err != nil {
			errMsg := fmt.Sprintf("Image failed to upload - %v", err.Error())
			util.HandleError(c, http.StatusInternalServerError, errors.New(errMsg))
			return
		}

		err = sc.shopService.UpdateShopGallery(ctx, shopId, myId, imageUploadResult.SecureURL)
		if err != nil {
			// delete media on error
			util.LogError("Error uploading gallery", err)
			util.DestroyMedia(imageUploadResult.PublicID)
			util.HandleError(c, http.StatusNotModified, err)
			return
		}

		internal.PublishCacheMessage(c, internal.CacheInvalidateShop, shopId.Hex())

		util.HandleSuccess(c, http.StatusOK, "Image added to shop gallery successfully", nil)
	}
}

// DeleteFromShopGallery - api/shops/:shopid/favorers?image={image_url}
func (sc *ShopController) DeleteFromShopGallery() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		imageURL := c.Query("image")

		shopID, myID, err := common.MyShopIdAndMyId(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		err = sc.shopService.DeleteFromShopGallery(ctx, shopID, myID, imageURL)
		if err != nil {
			if err.Error() == "no matching documents found" {
				util.HandleError(c, http.StatusNotFound, err)
			} else {
				util.HandleError(c, http.StatusNotModified, err)
			}
			return
		}

		internal.PublishCacheMessage(c, internal.CacheInvalidateShop, shopID.Hex())

		util.HandleSuccess(c, http.StatusOK, "Image removed from shop gallery successfully", nil)
	}
}

// FollowShop - api/shops/:shopid/followers
func (sc *ShopController) FollowShop() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shopId, myId, err := common.MyShopIdAndMyId(c)
		if err != nil {
			log.Println(err)
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		followerId, err := sc.shopService.FollowShop(ctx, myId, shopId)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		internal.PublishCacheMessage(c, internal.CacheInvalidateShop, shopId.Hex())

		util.HandleSuccess(c, http.StatusOK, "You're now a follower of this shop", followerId.Hex())
	}
}

// GetShopFollowers - api/shops/:shopid/followers?limit=50&skip=0
func (sc *ShopController) GetShopFollowers() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shopId := c.Param("shopid")
		shopObjectID, err := primitive.ObjectIDFromHex(shopId)
		if err != nil {
			log.Printf("Invalid user id %v", shopId)
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		paginationArgs := common.GetPaginationArgs(c)

		shopFollowers, count, err := sc.shopService.GetShopFollowers(ctx, shopObjectID, paginationArgs)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}

		util.HandleSuccessMeta(c, http.StatusOK, "Success", shopFollowers,
			gin.H{
				"pagination": util.Pagination{
					Limit: paginationArgs.Limit,
					Skip:  paginationArgs.Skip,
					Count: count,
				},
			})
	}
}

// IsfollowingShop - api/shops/:shopid/followers/following
func (sc *ShopController) IsFollowingShop() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shopId, myId, err := common.MyShopIdAndMyId(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		isFollowing, err := sc.shopService.IsFollowingShop(ctx, myId, shopId)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Success", isFollowing)
	}
}

// UnfollowShop - api/shops/:shopid/followers
func (sc *ShopController) UnfollowShop() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shopId, myId, err := common.MyShopIdAndMyId(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		// Shop Member session
		wc := writeconcern.New(writeconcern.WMajority())
		txnOptions := options.Transaction().SetWriteConcern(wc)

		session, err := util.DB.StartSession()
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, fmt.Errorf("failed to start session: %v", err))
			return
		}
		defer session.EndSession(ctx)

		callback := func(ctx mongo.SessionContext) (any, error) {
			// attempt to remove member from member collection table
			filter := bson.M{"shop_id": shopId, "user_id": myId}
			_, err := common.ShopFollowerCollection.DeleteOne(ctx, filter)
			if err != nil {
				return nil, err
			}

			// attempt to remove member from embedded field in shop
			filter = bson.M{"_id": shopId}
			update := bson.M{"$pull": bson.M{"followers": bson.M{"user_id": myId}}, "$inc": bson.M{"follower_count": -1}}
			result2, err := common.ShopCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				return nil, err
			}

			if result2.ModifiedCount == 0 {
				return nil, errors.New("no matching documents found")
			}

			return result2, nil
		}

		_, err = session.WithTransaction(ctx, callback, txnOptions)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		if err := session.CommitTransaction(context.Background()); err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		session.EndSession(context.Background())

		internal.PublishCacheMessage(c, internal.CacheInvalidateShop, shopId.Hex())

		util.HandleSuccess(c, http.StatusOK, "Left shop successfully", shopId.Hex())
	}
}

// RemoveOtherFollower - api/shops/:shopid/followers/other?userid={user_id to remove}
func (sc *ShopController) RemoveOtherFollower() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		userToBeRemoved := c.Query("userid")
		defer cancel()

		shopId, myId, err := common.MyShopIdAndMyId(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		// Let's verify shop ownership before attempting to remove follower
		ownershipEerr := sc.shopService.VerifyShopOwnership(ctx, myId, shopId)
		if ownershipEerr != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		userToBeRemovedId, err := primitive.ObjectIDFromHex(userToBeRemoved)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		if myId == userToBeRemovedId {
			util.HandleError(c, http.StatusBadRequest, errors.New("cannot remove yourself"))
			return
		}

		// Shop follower session
		wc := writeconcern.New(writeconcern.WMajority())
		txnOptions := options.Transaction().SetWriteConcern(wc)

		session, err := util.DB.StartSession()
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, fmt.Errorf("failed to start session: %v", err))
			return
		}
		defer session.EndSession(ctx)

		callback := func(ctx mongo.SessionContext) (any, error) {
			// attempt to remove follower from shop follower collection table
			filter := bson.M{"shop_id": shopId, "user_id": userToBeRemovedId}
			_, err := common.ShopFollowerCollection.DeleteOne(ctx, filter)
			if err != nil {
				return nil, err
			}

			// attempt to remove follower from embedded field in shop
			filter = bson.M{"_id": shopId}
			update := bson.M{"$pull": bson.M{"followers": bson.M{"user_id": userToBeRemovedId}}, "$inc": bson.M{"follower_count": -1}}
			result2, err := common.ShopCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				return nil, err
			}

			return result2, nil
		}

		_, err = session.WithTransaction(context.Background(), callback, txnOptions)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		if err := session.CommitTransaction(context.Background()); err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		session.EndSession(context.Background())

		internal.PublishCacheMessage(c, internal.CacheInvalidateShop, shopId.Hex())

		util.HandleSuccess(c, http.StatusOK, "Follower removed successfully", userToBeRemovedId.Hex())
	}
}

// UpdateShopAbout - api/shops/:shopid/about
func (sc *ShopController) UpdateShopAbout() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var shopAboutJson models.ShopAbout

		shopId, myId, err := common.MyShopIdAndMyId(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		if err := c.BindJSON(&shopAboutJson); err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		if validationErr := common.Validate.Struct(&shopAboutJson); validationErr != nil {
			util.HandleError(c, http.StatusBadRequest, validationErr)
			return
		}

		err = sc.shopService.VerifyShopOwnership(c, myId, shopId)
		if err != nil {
			log.Printf("You don't have write access to this shop: %s\n", err.Error())
			util.HandleError(c, http.StatusUnauthorized, err)
			return
		}

		filter := bson.M{"_id": shopId}
		update := bson.M{
			"$set": bson.M{"about": shopAboutJson},
		}
		res, err := common.ShopCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}
		if res.ModifiedCount == 0 {
			util.HandleSuccess(c, http.StatusNotFound, "success", "No matching documents found")
			return
		}

		internal.PublishCacheMessage(c, internal.CacheInvalidateShopAbout, shopId.Hex())

		util.HandleSuccess(c, http.StatusOK, "Shop about updated successfully", res.UpsertedID)
	}
}

// CreateShopReturnPolicy - api/shops/:shopid/policies
func (sc *ShopController) CreateShopReturnPolicy() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		var shopReturnPolicyJson models.ShopReturnPolicies
		defer cancel()

		shopId, myId, err := common.MyShopIdAndMyId(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		if err := c.BindJSON(&shopReturnPolicyJson); err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		if validationErr := common.Validate.Struct(&shopReturnPolicyJson); validationErr != nil {
			util.HandleError(c, http.StatusBadRequest, validationErr)
			return
		}

		err = sc.shopService.VerifyShopOwnership(c, myId, shopId)
		if err != nil {
			log.Printf("You don't have write access to this shop: %s\n", err.Error())
			util.HandleError(c, http.StatusUnauthorized, err)
			return
		}

		shopReturnPolicyJson.ID = primitive.NewObjectID()
		shopReturnPolicyJson.ShopId = shopId

		_, err = common.ShopReturnPolicyCollection.InsertOne(ctx, shopReturnPolicyJson)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		internal.PublishCacheMessage(c, internal.CacheInvalidateShopPolicy, shopId.Hex())

		util.HandleSuccess(c, http.StatusOK, "Shop policy created successfully", shopReturnPolicyJson.ID.Hex())
	}
}

// UpdateShopReturnPolicy - api/shops/:shopid/policies
func (sc *ShopController) UpdateShopReturnPolicy() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		var shopReturnPolicyJson models.ShopReturnPolicies
		defer cancel()

		shopId, myId, err := common.MyShopIdAndMyId(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		if err := c.BindJSON(&shopReturnPolicyJson); err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		if validationErr := common.Validate.Struct(&shopReturnPolicyJson); validationErr != nil {
			util.HandleError(c, http.StatusBadRequest, validationErr)
			return
		}

		err = sc.shopService.VerifyShopOwnership(c, myId, shopId)
		if err != nil {
			log.Printf("You don't have write access to this shop: %s\n", err.Error())
			util.HandleError(c, http.StatusUnauthorized, err)
			return
		}

		filter := bson.M{"shop_id": shopId}
		update := bson.M{"$set": bson.M{"accepts_return": shopReturnPolicyJson.AcceptsReturn, "accepts_echanges": shopReturnPolicyJson.AcceptsExchanges, "deadline": shopReturnPolicyJson.Deadline}}
		res, err := common.ShopReturnPolicyCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		if res.ModifiedCount == 0 {
			util.HandleError(c, http.StatusNotFound, nil)
			return
		}

		internal.PublishCacheMessage(c, internal.CacheInvalidateShopPolicy, shopId.Hex())

		util.HandleSuccess(c, http.StatusOK, "Shop policy updated successfully", res.UpsertedID)
	}
}

// DeleteShopReturnPolicy - api/shops/:shopid/policies?id={policy_id}
func (sc *ShopController) DeleteShopReturnPolicy() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		policyIdStr := c.Param("policyid")
		policyId, err := primitive.ObjectIDFromHex(policyIdStr)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		shopId, myId, err := common.MyShopIdAndMyId(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		err = sc.shopService.VerifyShopOwnership(c, myId, shopId)
		if err != nil {
			log.Printf("You don't have write access to this shop: %s\n", err.Error())
			util.HandleError(c, http.StatusUnauthorized, err)
			return
		}

		filter := bson.M{"_id": policyId, "shop_id": shopId}
		res, err := common.ShopReturnPolicyCollection.DeleteOne(ctx, filter)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		internal.PublishCacheMessage(c, internal.CacheInvalidateShopPolicy, shopId.Hex())
		util.HandleSuccess(c, http.StatusOK, "Shop policy deleted successfully", res.DeletedCount)
	}
}

// GetShopReturnPolicy - api/shops/:shopid/policies?id={policy_id}
func (sc *ShopController) GetShopReturnPolicy() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		policyIdStr := c.Param("policyid")
		policyId, err := primitive.ObjectIDFromHex(policyIdStr)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		shopId, _, err := common.MyShopIdAndMyId(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		var currentPolicy models.ShopReturnPolicies
		filter := bson.M{"_id": policyId, "shop_id": shopId}
		err = common.ShopReturnPolicyCollection.FindOne(ctx, filter).Decode(&currentPolicy)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		util.HandleSuccess(c, http.StatusOK, "success", currentPolicy)
	}
}

// GetShopReturnPolicies - api/shops/:shopid/policies/all
func (sc *ShopController) GetShopReturnPolicies() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shopId, _, err := common.MyShopIdAndMyId(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		// Query the database for shops that match the search query
		cursor, err := common.ShopReturnPolicyCollection.Find(ctx, bson.M{"shop_id": shopId})
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}
		defer cursor.Close(ctx)

		// Serialize the shop policies and return them to the client
		var policies []models.ShopReturnPolicies
		for cursor.Next(ctx) {
			var policy models.ShopReturnPolicies
			if err := cursor.Decode(&policy); err != nil {
				util.HandleError(c, http.StatusInternalServerError, err)
				return
			}
			policies = append(policies, policy)
		}

		util.HandleSuccess(c, http.StatusOK, "success", policies)
	}
}

// CreateShopCompliance - api/shops/:shopid/compliance
func (sc *ShopController) CreateShopComplianceInformation() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		var complianceJson models.ComplianceInformationRequest
		defer cancel()

		shopId, myId, err := common.MyShopIdAndMyId(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		if err := c.BindJSON(&complianceJson); err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		if validationErr := common.Validate.Struct(&complianceJson); validationErr != nil {
			util.HandleError(c, http.StatusBadRequest, validationErr)
			return
		}

		err = sc.shopService.VerifyShopOwnership(c, myId, shopId)
		if err != nil {
			log.Printf("Error verifying if you the shop owner: %s\n", err.Error())
			util.HandleError(c, http.StatusUnauthorized, err)
			return
		}

		complianceInformation := models.ComplianceInformation{
			ID:                   primitive.NewObjectID(),
			ShopID:               shopId,
			TermsOfUse:           complianceJson.TermsOfUse,
			IntellectualProperty: complianceJson.IntellectualProperty,
			SellerPolicie:        complianceJson.SellerPolicie,
		}

		_, err = common.ShopCompliancePolicyCollection.InsertOne(ctx, complianceInformation)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		internal.PublishCacheMessage(c, internal.CacheInvalidateShopCompliance, shopId.Hex())

		util.HandleSuccess(c, http.StatusOK, "Shop compliance policy created successfully", nil)
	}
}

// GetShopComplianceInformation - api/shops/:shopid/compliance
func (sc *ShopController) GetShopComplianceInformation() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shopId, _, err := common.MyShopIdAndMyId(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		var complianceInformation models.ComplianceInformation

		err = common.ShopCompliancePolicyCollection.FindOne(ctx, bson.M{"shop_id": shopId}).Decode(&complianceInformation)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				util.HandleError(c, http.StatusNotFound, err)
				return
			}
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Shop compliance information created successfully", gin.H{"compliance_information": complianceInformation})
	}
}
