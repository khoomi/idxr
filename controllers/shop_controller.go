package controllers

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	slug2 "github.com/gosimple/slug"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
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
var shopAboutCollection = configs.GetCollection(configs.DB, "ShopAbout")
var shopMemberCollection = configs.GetCollection(configs.DB, "ShopMember")
var shopReviewCollection = configs.GetCollection(configs.DB, "ShopReview")

func CreateShop() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		userID, err := auth.ExtractTokenID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, responses.UserResponse{Status: http.StatusUnauthorized, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		loginName, err := auth.ExtractTokenLoginName(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, responses.UserResponse{Status: http.StatusUnauthorized, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		shopName := c.Request.FormValue("name")
		err = configs.ValidateShopName(shopName)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		shopNameDescription := c.Request.FormValue("description")

		logoFile, _, err := c.Request.FormFile("logo")
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}
		logoUploadUrl, err := services.NewMediaUpload().FileUpload(models.File{File: logoFile})
		if err != nil {
			errMsg, _ := fmt.Printf("Logo failed to upload - %v", err.Error())
			log.Print(errMsg)
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"error": errMsg}})
			return
		}

		bannerFile, _, err := c.Request.FormFile("banner")
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}
		bannerUploadUrl, err := services.NewMediaUpload().FileUpload(models.File{File: bannerFile})
		if err != nil {
			errMsg, _ := fmt.Printf("Banner failed to upload - %v", err.Error())
			log.Print(errMsg)
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"error": errMsg}})
			return
		}

		wc := writeconcern.New(writeconcern.WMajority())
		txnOptions := options.Transaction().SetWriteConcern(wc)
		session, err := configs.DB.StartSession()
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, responses.UserResponse{Status: http.StatusUnprocessableEntity, Message: "error", Data: map[string]interface{}{"error": "Unable  to start new session"}})
		}
		defer session.EndSession(ctx)

		callback := func(ctx mongo.SessionContext) (interface{}, error) {
			now := time.Now()
			slug := slug2.Make(shopName)
			policy := models.ShopPolicy{
				PaymentPolicy:  "",
				ShippingPolicy: "",
				RefundPolicy:   "",
				AdditionalInfo: "",
			}
			shopId := primitive.NewObjectID()
			shop := models.Shop{
				ID:                 shopId,
				ShopName:           shopName,
				Description:        shopNameDescription,
				LoginName:          loginName,
				UserID:             userID,
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
				Members:            []models.ShopMember{},
				Status:             models.ShopStatusPendingReview,
				CreatedAt:          now,
				ModifiedAt:         now,
				Policy:             policy,
				RecentReviews:      []models.ShopReview{},
			}
			_, err := shopCollection.InsertOne(ctx, shop)
			if err != nil {
				return nil, err
			}

			// update user profile shop
			filter := bson.M{"_id": userID}
			update := bson.M{"$push": bson.M{"shops": shopId}}
			result, err := userCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				return nil, err
			}

			return result, nil
		}

		_, err = session.WithTransaction(context.Background(), callback, txnOptions)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}
		if err := session.CommitTransaction(context.Background()); err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}
		session.EndSession(context.Background())

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": "Shop creation was successful"}})
	}
}

// GetShop - api/shops/:shopid
func GetShop() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shopId := c.Param("shopid")
		shopObjectID, err := primitive.ObjectIDFromHex(shopId)
		if err != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		var shop models.Shop
		err = shopCollection.FindOne(ctx, bson.M{"_id": shopObjectID}).Decode(&shop)
		if err != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": shop}})

	}
}

// GetShops - api/shops/?limit=50&skip=0
func GetShops() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		paginationArgs := services.GetPaginationArgs(c)
		find := options.Find().SetLimit(int64(paginationArgs.Limit)).SetSkip(int64(paginationArgs.Skip)).SetSort(bson.M{paginationArgs.Sort: paginationArgs.Order})
		result, err := shopMemberCollection.Find(ctx, bson.D{}, find)
		if err != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		var shopMembers []models.ShopMember
		if err = result.All(ctx, &shopMembers); err != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponsePagination{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": shopMembers}, Pagination: responses.Pagination{
			Limit: paginationArgs.Limit,
			Skip:  paginationArgs.Skip,
		}})

	}
}

// SearchShops - api/shops/:shopid/search?q=khoomi&limit=50&skip=0
func SearchShops() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		query := c.Query("q")
		paginationArgs := services.GetPaginationArgs(c)

		// Query the database for shops that match the search query
		shops, err := shopCollection.Find(ctx, bson.M{
			"$or": []bson.M{
				{"shop_name": bson.M{"$regex": primitive.Regex{Pattern: query, Options: "i"}}},
				{"description": bson.M{"$regex": primitive.Regex{Pattern: query, Options: "i"}}},
			},
		}, options.Find().SetSkip(int64(paginationArgs.Skip)).SetLimit(int64(paginationArgs.Limit)))
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "Error searching for shops", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		// Count the total number of shops that match the search query
		count, err := shopCollection.CountDocuments(ctx, bson.M{
			"$or": []bson.M{
				{"shop_name": bson.M{"$regex": primitive.Regex{Pattern: query, Options: "i"}}},
				{"description": bson.M{"$regex": primitive.Regex{Pattern: query, Options: "i"}}},
			},
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "Error counting shops", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		// Serialize the shops and return them to the client
		var serializedShops []models.Shop
		for shops.Next(ctx) {
			var shop models.Shop
			if err := shops.Decode(&shop); err != nil {
				c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "Error decoding shops", Data: map[string]interface{}{"error": err.Error()}})
				return
			}
			serializedShops = append(serializedShops, shop)

			c.JSON(http.StatusOK, responses.UserResponsePagination{Status: http.StatusOK, Message: "Shops found", Data: map[string]interface{}{"data": serializedShops}, Pagination: responses.Pagination{
				Limit: paginationArgs.Limit,
				Skip:  paginationArgs.Skip,
				Count: count,
			}})
		}

	}
}

func UpdateShopAnnouncement() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		var announcement models.ShopAnnouncementRequest
		defer cancel()

		err := c.BindJSON(announcement)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}
		if announcement.Announcement == "" {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": "announcement cannot be empty"}})
			return
		}

		if len(announcement.Announcement) > 100 {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": "announcement is too long"}})
			return
		}

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		filter := bson.M{"_id": shopId, "user_id": myId}
		update := bson.M{"announcement": announcement.Announcement, "modified_at": time.Now()}
		res, err := shopCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusNotModified, responses.UserResponse{Status: http.StatusNotModified, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}
		if res.ModifiedCount == 0 {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "success", Data: map[string]interface{}{"data": "no matching documents found"}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": "Shop announcement updated successfully"}})
	}
}

func UpdateShopVacation() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		var vacation models.ShopVacationRequest
		now := time.Now()
		defer cancel()

		err := c.BindJSON(&vacation)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		filter := bson.M{"_id": shopId, "user_id": myId}
		update := bson.M{"$set": bson.M{"vacation_message": vacation.Message, "is_vacation": vacation.IsVacation, "modified_at": now}}
		res, err := shopCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusNotModified, responses.UserResponse{Status: http.StatusNotModified, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}
		if res.ModifiedCount == 0 {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "success", Data: map[string]interface{}{"data": "no matching documents found"}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": "Shop vacation updated successfully"}})
	}
}

func UpdateShopLogo() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		now := time.Now()
		defer cancel()

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		logoFile, _, err := c.Request.FormFile("logo")
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}
		logoUploadUrl, err := services.NewMediaUpload().FileUpload(models.File{File: logoFile})
		if err != nil {
			errMsg, _ := fmt.Printf("Logo failed to upload - %v", err.Error())
			log.Print(errMsg)
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"error": errMsg}})
			return
		}

		filter := bson.M{"_id": shopId, "user_id": myId}
		update := bson.M{"logo_url": logoUploadUrl, "modified_at": now}
		res, err := shopCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusNotModified, responses.UserResponse{Status: http.StatusNotModified, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}
		if res.ModifiedCount == 0 {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "success", Data: map[string]interface{}{"data": "no matching documents found"}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": "Shop logo updated successfully"}})
	}
}

func UpdateShopBanner() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		now := time.Now()
		defer cancel()

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		bannerFile, _, err := c.Request.FormFile("logo")
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}
		bannerFileUrl, err := services.NewMediaUpload().FileUpload(models.File{File: bannerFile})
		if err != nil {
			errMsg, _ := fmt.Printf("Banner failed to upload - %v", err.Error())
			log.Print(errMsg)
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"error": errMsg}})
			return
		}

		filter := bson.M{"_id": shopId, "user_id": myId}
		update := bson.M{"banner_url": bannerFileUrl, "modified_at": now}
		res, err := shopCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusNotModified, responses.UserResponse{Status: http.StatusNotModified, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}
		if res.ModifiedCount == 0 {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "success", Data: map[string]interface{}{"data": "no matching documents found"}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": "Shop banner has been updated successfully"}})
	}
}

func UpdateShopGallery() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		now := time.Now()
		defer cancel()

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		imageFile, _, err := c.Request.FormFile("logo")
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}
		imageFileUrl, err := services.NewMediaUpload().FileUpload(models.File{File: imageFile})
		if err != nil {
			errMsg, _ := fmt.Printf("Image failed to upload - %v", err.Error())
			log.Print(errMsg)
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"error": errMsg}})
			return
		}

		filter := bson.M{"_id": shopId, "user_id": myId}
		update := bson.M{"$push": bson.M{"gallery": imageFileUrl}, "modified_at": now}
		res, err := shopCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusNotModified, responses.UserResponse{Status: http.StatusNotModified, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}
		if res.ModifiedCount == 0 {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "success", Data: map[string]interface{}{"data": "no matching documents found"}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": "Image added to shop gallery successfully"}})
	}
}

// DeleteFromShopGallery - api/shops/:shopid/favorers?image={image_url}
func DeleteFromShopGallery() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		imageUrl := c.Query("image")
		now := time.Now()
		defer cancel()

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		filter := bson.M{"_id": shopId, "user_id": myId}
		update := bson.M{"$pull": bson.M{"gallery": imageUrl}, "modified_at": now}
		res, err := shopCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusNotModified, responses.UserResponse{Status: http.StatusNotModified, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}
		if res.ModifiedCount == 0 {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "success", Data: map[string]interface{}{"data": "no matching documents found"}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": "Image removed from shop gallery successfully"}})
	}
}

// AddShopFavorer - api/shops/:shopid/favorers?userId={userId}
func AddShopFavorer() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		userId := c.Query("image")
		now := time.Now()
		defer cancel()

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		filter := bson.M{"_id": shopId, "user_id": myId}
		update := bson.M{"$push": bson.M{"favorers": userId}, "$inc": bson.M{"favorer_count": 1}, "modified_at": now}
		res, err := shopCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusNotModified, responses.UserResponse{Status: http.StatusNotModified, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}
		if res.ModifiedCount == 0 {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "success", Data: map[string]interface{}{"data": "no matching documents found"}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": "User is now a favorer of this shop"}})
	}
}

// RemoveShopFavorer - api/shops/:shopid/favorers?userId={userId}
func RemoveShopFavorer() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		userId := c.Query("image")
		now := time.Now()
		defer cancel()

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		filter := bson.M{"_id": shopId, "user_id": myId}
		update := bson.M{"pull": bson.M{"favorers": userId}, "$inc": bson.M{"favorer_count": -1}, "modified_at": now}
		res, err := shopCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusNotModified, responses.UserResponse{Status: http.StatusNotModified, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}
		if res.ModifiedCount == 0 {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "success", Data: map[string]interface{}{"data": "no matching documents found"}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": "User is no longer a favorer of this shop"}})
	}
}

// JoinShopMembers - api/shops/:shopid/members
func JoinShopMembers() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		var shopMember models.ShopMemberFromRequest
		now := time.Now()
		defer cancel()

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}
		loginName, err := auth.ExtractTokenLoginName(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		err = c.BindJSON(&shopMember)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": "bind"}})
			return
		}
		newMemberObjectId, err := primitive.ObjectIDFromHex(shopMember.MemberId)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": "member_id"}})
			return
		}

		// validate request body
		if validationErr := validate.Struct(&shopMember); validationErr != nil {
			c.JSON(http.StatusUnprocessableEntity, responses.UserResponse{Status: http.StatusUnprocessableEntity, Message: "error", Data: map[string]interface{}{"error": validationErr.Error()}})
			return
		}

		// Shop Member session
		wc := writeconcern.New(writeconcern.WMajority())
		txnOptions := options.Transaction().SetWriteConcern(wc)

		session, err := configs.DB.StartSession()
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, responses.UserResponse{Status: http.StatusUnprocessableEntity, Message: "error", Data: map[string]interface{}{"error": "Unable  to start new session"}})
		}
		defer session.EndSession(ctx)

		callback := func(ctx mongo.SessionContext) (interface{}, error) {
			var currentShop models.Shop
			err := shopCollection.FindOne(ctx, bson.M{"_id": shopId}).Decode(&currentShop)
			if err != nil {
				log.Println(err)
				return nil, err
			}

			// attempt to add member to member collection
			shopMemberData := models.ShopMember{
				Id:        primitive.NewObjectID(),
				MemberId:  newMemberObjectId,
				ShopId:    shopId,
				LoginName: loginName,
				Thumbnail: shopMember.Thumbnail,
				IsOwner:   currentShop.UserID == myId,
				OwnerId:   currentShop.UserID,
				JoinedAt:  time.Now(),
			}
			_, err = shopMemberCollection.InsertOne(ctx, shopMemberData)
			if err != nil {
				log.Println(err)
				return nil, err
			}

			// attempt to add member to member field in shop
			inner := models.ShopMemberEmbedded{
				MemberId:  myId,
				LoginName: loginName,
				Thumbnail: shopMember.Thumbnail,
				IsOwner:   currentShop.UserID == myId,
			}
			filter := bson.M{"_id": shopId, "members": bson.M{"$not": bson.M{"$elemMatch": bson.M{"member_id": &shopMember.MemberId}}}}
			update := bson.M{"$push": bson.M{"members": bson.M{"$each": bson.A{inner}, "$sort": -1, "$slice": -5}}, "$set": bson.M{"modified_at": now}}
			result2, err := shopCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				return nil, err
			}

			if result2.ModifiedCount == 0 {
				return nil, errors.New("no matching documents found")
			}

			return result2, nil
		}

		_, err = session.WithTransaction(context.Background(), callback, txnOptions)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}
		if err := session.CommitTransaction(context.Background()); err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}
		session.EndSession(context.Background())

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": "You're now a member of this shop"}})
	}
}

// GetShopMembers - api/shops/:shopid/members?limit=50&skip=0
func GetShopMembers() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shopId := c.Param("shopid")
		shopObjectID, err := primitive.ObjectIDFromHex(shopId)
		if err != nil {
			log.Printf("Invalid user id %v", shopId)
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		paginationArgs := services.GetPaginationArgs(c)
		filter := bson.M{"shop_id": shopObjectID}
		find := options.Find().SetLimit(int64(paginationArgs.Limit)).SetSkip(int64(paginationArgs.Skip))
		result, err := shopMemberCollection.Find(ctx, filter, find)
		if err != nil {
			log.Printf("%v", err)
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		count, err := shopMemberCollection.CountDocuments(ctx, bson.M{"shop_id": shopObjectID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "Error counting shops", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		var shopMembers []models.ShopMember
		if err = result.All(ctx, &shopMembers); err != nil {
			log.Printf("%v", err)
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponsePagination{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": shopMembers}, Pagination: responses.Pagination{
			Limit: paginationArgs.Limit,
			Skip:  paginationArgs.Skip,
			Count: count,
		}})

	}
}

// LeaveShopMembers - api/shops/:shopid/members
func LeaveShopMembers() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		// Shop Member session
		wc := writeconcern.New(writeconcern.WMajority())
		txnOptions := options.Transaction().SetWriteConcern(wc)

		session, err := configs.DB.StartSession()
		if err != nil {
			panic(err)
		}
		defer session.EndSession(ctx)

		callback := func(ctx mongo.SessionContext) (interface{}, error) {
			// attempt to remove member from member collection table
			filter := bson.M{"shop_id": shopId, "member_id": myId}
			_, err := shopMemberCollection.DeleteOne(ctx, filter)
			if err != nil {
				return nil, err
			}

			// attempt to remove member from embedded field in shop
			filter = bson.M{"_id": shopId}
			update := bson.M{"$pull": bson.M{"members": bson.M{"member_id": myId}}}
			result2, err := shopCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				return nil, err
			}

			if result2.ModifiedCount == 0 {
				return nil, errors.New("no matching documents found")
			}

			return result2, nil
		}
		_, err = session.WithTransaction(context.Background(), callback, txnOptions)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}
		if err := session.CommitTransaction(context.Background()); err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}
		session.EndSession(context.Background())

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": "Left shop successfully"}})
	}
}

// RemoveOtherMember - api/shops/:shopid/members/other?userid={user_id to remove}
func RemoveOtherMember() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		userToBeRemoved := c.Query("userid")
		defer cancel()

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		userToBeRemovedId, err := primitive.ObjectIDFromHex(userToBeRemoved)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": "userid"}})
			return
		}

		// Shop Member section
		wc := writeconcern.New(writeconcern.WMajority())
		txnOptions := options.Transaction().SetWriteConcern(wc)

		session, err := configs.DB.StartSession()
		if err != nil {
			panic(err)
		}
		defer session.EndSession(ctx)

		callback := func(ctx mongo.SessionContext) (interface{}, error) {
			// attempt to remove member from member collection table
			filter := bson.M{"shop_id": shopId, "owner_id": myId, "member_id": userToBeRemovedId}
			_, err := shopMemberCollection.DeleteOne(ctx, filter)
			if err != nil {
				return nil, err
			}

			// attempt to remove member from embedded field in shop
			filter = bson.M{"_id": shopId}
			update := bson.M{"$pull": bson.M{"members": bson.M{"member_id": userToBeRemovedId}}}
			result2, err := shopCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				return nil, err
			}

			if result2.ModifiedCount == 0 {
				return nil, errors.New("no matching documents found")
			}

			return result2, nil
		}
		_, err = session.WithTransaction(context.Background(), callback, txnOptions)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}
		if err := session.CommitTransaction(context.Background()); err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}
		session.EndSession(context.Background())

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": "Member removed successfully"}})
	}
}

// CreateShopReview - api/shops/:shopid/reviews
func CreateShopReview() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		var shopReviewJson models.ShopReviewRequest
		now := time.Now()
		defer cancel()

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}
		loginName, err := auth.ExtractTokenLoginName(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		err = c.BindJSON(&shopReviewJson)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": "bind"}})
			return
		}

		// validate request body
		if validationErr := validate.Struct(&shopReviewJson); validationErr != nil {
			c.JSON(http.StatusUnprocessableEntity, responses.UserResponse{Status: http.StatusUnprocessableEntity, Message: "error", Data: map[string]interface{}{"error": validationErr.Error()}})
			return
		}

		// Shop Member session
		wc := writeconcern.New(writeconcern.WMajority())
		txnOptions := options.Transaction().SetWriteConcern(wc)

		session, err := configs.DB.StartSession()
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, responses.UserResponse{Status: http.StatusUnprocessableEntity, Message: "error", Data: map[string]interface{}{"error": "Unable  to start new session"}})
		}
		defer session.EndSession(ctx)

		callback := func(ctx mongo.SessionContext) (interface{}, error) {
			var userProfile models.User
			err := userCollection.FindOne(ctx, bson.M{"_id": myId}).Decode(&userProfile)
			if err != nil {
				log.Println(err)
				return nil, err
			}

			// attempt to add review to review collection
			shopReviewData := models.ShopReview{
				Id:           primitive.NewObjectID(),
				UserId:       myId,
				ShopId:       shopId,
				Review:       shopReviewJson.Review,
				ReviewAuthor: loginName,
				Thumbnail:    userProfile.Thumbnail,
				CreatedAt:    now,
				Status:       models.ShopReviewStatusApproved,
			}
			_, err = shopReviewCollection.InsertOne(ctx, shopReviewData)
			if err != nil {
				log.Println(err)
				return nil, err
			}

			// attempt to add member to review field in shop
			embedded := models.EmbeddedShopReview{
				UserId:       myId,
				ShopId:       shopId,
				Review:       shopReviewJson.Review,
				ReviewAuthor: loginName,
				Thumbnail:    userProfile.Thumbnail,
			}
			filter := bson.M{"_id": shopId, "recent_reviews": bson.M{"$not": bson.M{"$elemMatch": bson.M{"user_id": myId}}}}
			update := bson.M{"$push": bson.M{"recent_reviews": bson.M{"$each": bson.A{embedded}, "$sort": -1, "$slice": -5}}, "$set": bson.M{"modified_at": now}}
			result2, err := shopCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				return nil, err
			}

			return result2, nil
		}

		_, err = session.WithTransaction(context.Background(), callback, txnOptions)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}
		if err := session.CommitTransaction(context.Background()); err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}
		session.EndSession(context.Background())

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": "Shop creation successful"}})
	}
}

// GetShopReviews - api/shops/:shopid/reviews?limit=50&skip=0
func GetShopReviews() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shopId := c.Param("shopid")
		shopObjectID, err := primitive.ObjectIDFromHex(shopId)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		paginationArgs := services.GetPaginationArgs(c)
		filter := bson.M{"shop_id": shopObjectID}
		find := options.Find().SetLimit(int64(paginationArgs.Limit)).SetSkip(int64(paginationArgs.Skip))
		result, err := shopReviewCollection.Find(ctx, filter, find)
		if err != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		var shopReviews []models.ShopReview
		if err = result.All(ctx, &shopReviews); err != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		count, err := shopReviewCollection.CountDocuments(ctx,
			bson.M{
				"shop_id": shopObjectID,
			})
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "Error counting shops", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponsePagination{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": shopReviews}, Pagination: responses.Pagination{
			Limit: paginationArgs.Limit,
			Skip:  paginationArgs.Skip,
			Count: count,
		}})

	}
}

// DeleteMyReview - api/shops/:shopid/members
func DeleteMyReview() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		// Shop Member session
		wc := writeconcern.New(writeconcern.WMajority())
		txnOptions := options.Transaction().SetWriteConcern(wc)

		session, err := configs.DB.StartSession()
		if err != nil {
			panic(err)
		}
		defer session.EndSession(ctx)

		callback := func(ctx mongo.SessionContext) (interface{}, error) {
			// attempt to remove review from review collection table
			filter := bson.M{"shop_id": shopId, "user_id": myId}
			_, err := shopReviewCollection.DeleteOne(ctx, filter)
			if err != nil {
				return nil, err
			}

			// attempt to remove member from embedded field in shop
			filter = bson.M{"_id": shopId}
			update := bson.M{"$pull": bson.M{"recent_reviews": bson.M{"user_id": myId}}}
			result2, err := shopCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				return nil, err
			}

			if result2.ModifiedCount == 0 {
				return nil, errors.New("no matching documents found")
			}

			return result2, nil
		}

		_, err = session.WithTransaction(context.Background(), callback, txnOptions)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}
		if err := session.CommitTransaction(context.Background()); err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}
		session.EndSession(context.Background())

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": "My review was deleted successfully"}})
	}
}

// DeleteOtherReview - api/shops/:shopid/reviews/other?userid={user_id to remove}
func DeleteOtherReview() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		userToBeRemoved := c.Query("userid")
		userToBeRemovedId, err := primitive.ObjectIDFromHex(userToBeRemoved)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		// Shop review session
		wc := writeconcern.New(writeconcern.WMajority())
		txnOptions := options.Transaction().SetWriteConcern(wc)
		session, err := configs.DB.StartSession()
		if err != nil {
			panic(err)
		}
		defer session.EndSession(ctx)

		callback := func(ctx mongo.SessionContext) (interface{}, error) {
			var currentShop models.Shop
			err := shopCollection.FindOne(ctx, bson.M{"_id": shopId}).Decode(&currentShop)
			if err != nil {
				log.Println(err)
				return nil, err
			}

			if currentShop.UserID != myId {
				return nil, errors.New("this is not your shop")
			}

			// attempt to remove review from review collection table
			filter := bson.M{"shop_id": shopId, "owner_id": myId, "user_id": userToBeRemovedId}
			_, err = shopMemberCollection.DeleteOne(ctx, filter)
			if err != nil {
				return nil, err
			}

			// attempt to remove review from recent review field in shop
			filter = bson.M{"_id": shopId}
			update := bson.M{"$pull": bson.M{"recent_reviews": bson.M{"user_id": userToBeRemovedId}}}
			result2, err := shopCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				return nil, err
			}

			if result2.ModifiedCount == 0 {
				return nil, errors.New("no matching documents found")
			}

			return result2, nil
		}

		_, err = session.WithTransaction(context.Background(), callback, txnOptions)
		if err != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}
		if err := session.CommitTransaction(context.Background()); err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}
		session.EndSession(context.Background())

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": "Other user  review deleted successfully"}})
	}
}

// CreateShopAbout - api/shops/:shopid/about
func CreateShopAbout() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		var shopAboutJson models.ShopAboutRequest
		defer cancel()

		shopId, _, err := services.MyShopIdAndMyId(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		// bind the request body
		if err := c.BindJSON(&shopAboutJson); err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		// validate request body
		if validationErr := validate.Struct(&shopAboutJson); validationErr != nil {
			c.JSON(http.StatusUnprocessableEntity, responses.UserResponse{Status: http.StatusUnprocessableEntity, Message: "error", Data: map[string]interface{}{"error": validationErr.Error()}})
			return
		}

		shopAboutData := models.ShopAbout{
			ID:                    primitive.NewObjectID(),
			ShopID:                shopId,
			Status:                shopAboutJson.Status,
			RelatedLinks:          shopAboutJson.RelatedLinks,
			StoryLeadingParagraph: shopAboutJson.StoryLeadingParagraph,
			StoryHeadline:         shopAboutJson.StoryHeadline,
		}

		_, err = shopAboutCollection.InsertOne(ctx, shopAboutData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": "Shop about created successfully"}})
	}
}

// GetShopAbout - api/shops/:shopid/about
func GetShopAbout() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		var shopAbout models.ShopAbout
		defer cancel()

		shopId := c.Param("shopid")
		shopObjectID, err := primitive.ObjectIDFromHex(shopId)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		err = shopAboutCollection.FindOne(ctx, bson.M{"shop_id": shopObjectID}).Decode(&shopAbout)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": shopAbout}})
	}
}

// UpdateShopAbout - api/shops/:shopid/about
func UpdateShopAbout() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		var shopAboutJson models.ShopAboutRequest
		defer cancel()

		shopId, _, err := services.MyShopIdAndMyId(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		// bind the request body
		if err := c.BindJSON(&shopAboutJson); err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		// validate request body
		if validationErr := validate.Struct(&shopAboutJson); validationErr != nil {
			c.JSON(http.StatusUnprocessableEntity, responses.UserResponse{Status: http.StatusUnprocessableEntity, Message: "error", Data: map[string]interface{}{"error": validationErr.Error()}})
			return
		}

		filter := bson.M{"shop_id": shopId}
		update := bson.M{"$set": bson.M{"status": shopAboutJson.Status, "related_links": shopAboutJson.RelatedLinks, "story_leading_paragraph": shopAboutJson.StoryLeadingParagraph, "story_headline": shopAboutJson.StoryHeadline}}
		res, err := shopAboutCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}
		if res.ModifiedCount == 0 {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "success", Data: map[string]interface{}{"data": "no matching documents found"}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": "Shop about updated successfully"}})
	}
}

// UpdateShopAboutStatus - api/shops/:shopid/about/status?status=active
func UpdateShopAboutStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		status := c.Query("status")
		shopId, _, err := services.MyShopIdAndMyId(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		filter := bson.M{"shop_id": shopId}
		update := bson.M{"$set": bson.M{"status": status}}
		res, err := shopAboutCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"error": "status parameter is required and must be either 'active' or 'draft"}})
			return
		}

		if res.ModifiedCount == 0 {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "success", Data: map[string]interface{}{"data": "no matching documents found"}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": "Shop about status updated successfully"}})
	}
}
