package controllers

import (
	"context"
	"github.com/gin-gonic/gin"
	slug2 "github.com/gosimple/slug"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"khoomi-api-io/khoomi_api/auth"
	"khoomi-api-io/khoomi_api/configs"
	"khoomi-api-io/khoomi_api/models"
	"khoomi-api-io/khoomi_api/responses"
	"khoomi-api-io/khoomi_api/services"
	"log"
	"net/http"
	"time"
)

var shopCollection = configs.GetCollection(configs.DB, "Shop")

func CreateShop() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		myObjectId, err := services.GetUserObjectIdFromRequest(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, responses.UserResponse{Status: http.StatusUnauthorized, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		loginName, err := auth.ExtractTokenLoginName(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, responses.UserResponse{Status: http.StatusUnauthorized, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		shopName := c.Request.FormValue("name")
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": "shop name"}})
			return
		}

		err = configs.ValidateShopName(shopName)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": "shop name"}})
			return
		}

		shopNameDescription := c.Request.FormValue("description")
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": "shop description"}})
			return
		}

		logoFile, _, err := c.Request.FormFile("logo")
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": "logo"}})
			return
		}
		logoUploadUrl, err := services.NewMediaUpload().FileUpload(models.File{File: logoFile})
		if err != nil {
			log.Println("Logo failed to upload")
			log.Println(err)
			logoUploadUrl = ""
		}

		bannerFile, _, err := c.Request.FormFile("banner")
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": "banner"}})
			return
		}
		bannerUploadUrl, err := services.NewMediaUpload().FileUpload(models.File{File: bannerFile})
		if err != nil {
			log.Println("Banner failed to upload")
			log.Println(err)
			bannerUploadUrl = ""
		}

		now := time.Now()
		slug := slug2.Make(shopName)
		policy := models.ShopPolicy{
			PaymentPolicy:  "",
			ShippingPolicy: "",
			RefundPolicy:   "",
			AdditionalInfo: "",
		}
		shop := models.Shop{
			ID:                 primitive.NewObjectID(),
			ShopName:           shopName,
			Description:        shopNameDescription,
			LoginName:          loginName,
			UserID:             myObjectId,
			ListingActiveCount: 0,
			Announcement:       "",
			IsVacation:         false,
			VacationMessage:    "",
			Slug:               slug,
			LogoURL:            logoUploadUrl,
			BannerURL:          bannerUploadUrl,
			Gallery:            []string{},
			Favorers:           []string{},
			FavorerCount:       0,
			Members:            []models.EmbeddedShopMember{},
			Status:             models.ShopStatusPendingReview,
			CreatedAt:          now,
			ModifiedAt:         now,
			Policy:             policy,
			RecentReviews:      []models.RecentReviews{},
		}
		res, err := shopCollection.InsertOne(ctx, shop)
		if err != nil {
			c.JSON(http.StatusUnauthorized, responses.UserResponse{Status: http.StatusUnauthorized, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": res}})
	}
}

func UpdateShopAnnouncement() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		var announcement models.ShopAnnouncementRequest
		defer cancel()

		err := c.BindJSON(announcement)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		filter := bson.M{"_id": shopId, "user_id": myId}
		update := bson.M{"announcement": announcement.Announcement, "modified_at": time.Now()}
		res, err := shopCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusNotModified, responses.UserResponse{Status: http.StatusNotModified, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": res}})
	}
}

func UpdateShopVacation() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		var vacation models.ShopVacationRequest
		defer cancel()

		err := c.BindJSON(vacation)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		filter := bson.M{"_id": shopId, "user_id": myId}
		update := bson.M{"vacation_message": vacation.Message, "is_vacation": vacation.IsVacation, "modified_at": time.Now()}
		res, err := shopCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusNotModified, responses.UserResponse{Status: http.StatusNotModified, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": res}})
	}
}

func UpdateShopLogo() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		logoFile, _, err := c.Request.FormFile("logo")
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": "logo"}})
			return
		}
		logoUploadUrl, err := services.NewMediaUpload().FileUpload(models.File{File: logoFile})
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": "logo"}})
		}

		filter := bson.M{"_id": shopId, "user_id": myId}
		update := bson.M{"logo_url": logoUploadUrl, "modified_at": time.Now()}
		res, err := shopCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusNotModified, responses.UserResponse{Status: http.StatusNotModified, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": res}})
	}
}

func UpdateShopBanner() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		bannerFile, _, err := c.Request.FormFile("logo")
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": "logo"}})
			return
		}
		bannerFileUrl, err := services.NewMediaUpload().FileUpload(models.File{File: bannerFile})
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": "logo"}})
			return
		}

		filter := bson.M{"_id": shopId, "user_id": myId}
		update := bson.M{"banner_url": bannerFileUrl, "modified_at": time.Now()}
		res, err := shopCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusNotModified, responses.UserResponse{Status: http.StatusNotModified, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": res}})
	}
}

func UpdateShopGallery() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		bannerFile, _, err := c.Request.FormFile("logo")
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": "logo"}})
			return
		}
		bannerFileUrl, err := services.NewMediaUpload().FileUpload(models.File{File: bannerFile})
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": "logo"}})
			return
		}

		filter := bson.M{"_id": shopId, "user_id": myId}
		update := bson.M{"$push": bson.M{"gallery": bannerFileUrl}, "modified_at": time.Now()}
		res, err := shopCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusNotModified, responses.UserResponse{Status: http.StatusNotModified, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": res}})
	}
}

// DeleteFromShopGallery - api/shop/:shop/favorers?image={image_url}
func DeleteFromShopGallery() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		imageUrl := c.Query("image")
		defer cancel()

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		filter := bson.M{"_id": shopId, "user_id": myId}
		update := bson.M{"$pull": bson.M{"gallery": imageUrl}, "modified_at": time.Now()}
		res, err := shopCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusNotModified, responses.UserResponse{Status: http.StatusNotModified, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": res}})
	}
}

// AddShopFavorers - api/shop/:shop/favorers?userId={userId}
func AddShopFavorer() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		userId := c.Query("image")
		defer cancel()

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		filter := bson.M{"_id": shopId, "user_id": myId}
		update := bson.M{"$push": bson.M{"favorers": userId}, "$inc": bson.M{"favorer_count": 1}, "modified_at": time.Now()}
		res, err := shopCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusNotModified, responses.UserResponse{Status: http.StatusNotModified, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": res}})
	}
}

// AddShopFavorers - api/shop/:shop/favorers?userId={userId}
func RemoeveShopFavorer() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		userId := c.Query("image")
		defer cancel()

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		filter := bson.M{"_id": shopId, "user_id": myId}
		update := bson.M{"pull": bson.M{"favorers": userId}, "$inc": bson.M{"favorer_count": -1}, "modified_at": time.Now()}
		res, err := shopCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusNotModified, responses.UserResponse{Status: http.StatusNotModified, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": res}})
	}
}
