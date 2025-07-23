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
	"khoomi-api-io/api/pkg/util"
	"khoomi-api-io/api/pkg/services"

	"github.com/cloudinary/cloudinary-go/api/uploader"
	"github.com/gin-gonic/gin"
	slug2 "github.com/gosimple/slug"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

// CheckShopNameAvailability -> Check if a given shop name is available or not.
// /api/shop/check/:shop_username
func CheckShopNameAvailability() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shop_name := c.Param("username")
		var shop models.Shop

		filter := bson.M{"username": shop_name}
		err := common.ShopCollection.FindOne(ctx, filter).Decode(&shop)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				util.HandleSuccess(c, http.StatusOK, "Congrats! shop username is available :xD", true)
				return
			}

			util.HandleError(c, http.StatusInternalServerError, errors.New("internal sever error on checking shop username availability"))
			return
		}

		util.HandleSuccess(c, http.StatusOK, "shop username is already taken", false)
	}
}

func CreateShop() gin.HandlerFunc {
	return CreateShopWithEmailService(nil)
}

func CreateShopWithEmailService(emailService services.EmailService) gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		session_, err := auth.GetSessionAuto(c)
		if err != nil {
			util.HandleError(c, http.StatusUnauthorized, err)
			return
		}

		loginName, userEmail := session_.LoginName, session_.Email
		shopName := c.Request.FormValue("name")
		err = util.ValidateShopName(shopName)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}
		shopUserName := c.Request.FormValue("username")
		err = util.ValidateShopUserName(shopUserName)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}
		shopUserName = strings.ToLower(shopUserName)

		shopDescription := c.Request.FormValue("description")
		err = util.ValidateShopDescription(shopDescription)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		// Logo file handling
		logoFile, _, err := c.Request.FormFile("logo")
		logoUploadUrl := common.DEFAULT_LOGO
		var logoUploadResult uploader.UploadResult
		if err == nil {
			logoUploadResult, err = util.FileUpload(models.File{File: logoFile})
			if err != nil {
				errMsg := fmt.Sprintf("Logo failed to upload - %v", err.Error())
				util.HandleError(c, http.StatusInternalServerError, errors.New(errMsg))
				return
			}
			logoUploadUrl = logoUploadResult.SecureURL
		} else {
			logoUploadResult = uploader.UploadResult{}
		}

		// Banner file handling
		bannerFile, _, err := c.Request.FormFile("banner")
		var bannerUploadUrl string
		var bannerUploadResult uploader.UploadResult

		if err == nil {
			bannerUploadResult, err = util.FileUpload(models.File{File: bannerFile})
			if err != nil {
				errMsg := fmt.Sprintf("Banner failed to upload - %v", err.Error())
				util.HandleError(c, http.StatusInternalServerError, errors.New(errMsg))
				return
			}
			bannerUploadUrl = bannerUploadResult.SecureURL
		} else {
			bannerUploadUrl = common.DEFAULT_THUMBNAIL
			bannerUploadResult = uploader.UploadResult{}
		}

		shopID := primitive.NewObjectID()
		wc := writeconcern.New(writeconcern.WMajority())
		txnOptions := options.Transaction().SetWriteConcern(wc)
		session, err := util.DB.StartSession()
		if err != nil {
			util.HandleError(c, http.StatusUnprocessableEntity, err)
			return
		}
		defer session.EndSession(ctx)
		callback := func(ctx mongo.SessionContext) (any, error) {
			slug := slug2.Make(shopUserName)
			policy := models.ShopPolicy{
				PaymentPolicy:  "",
				ShippingPolicy: "",
				RefundPolicy:   "",
				AdditionalInfo: "",
			}

			shopRating := models.Rating{
				AverageRating:  0.0,
				ReviewCount:    0,
				FiveStarCount:  0,
				FourStarCount:  0,
				ThreeStarCount: 0,
				TwoStarCount:   0,
				OneStarCount:   0,
			}
			shopAboutData := models.ShopAbout{
				Headline:  fmt.Sprintf("Welcome to %v!", shopUserName),
				Story:     fmt.Sprintf("Thank you for visiting our online artisan shop. We are passionate about craftsmanship and dedicated to providing unique, handcrafted items that reflect the creativity and skill of our artisans. Explore our collection and discover the beauty of handmade products that carry a story of craftsmanship and tradition.\n\nAt %v, we believe in the art of creating something special. Each piece in our collection is carefully crafted with attention to detail and a commitment to quality. We aim to connect artisans with appreciative buyers, creating a community that values and supports the artistry behind every creation.\n\nJoin us on this journey of celebrating craftsmanship and supporting talented artisans from around the world. Your purchase not only adds a unique piece to your life but also contributes to the livelihood of skilled individuals who pour their heart and soul into their work.\n\nThank you for being a part of our community. Happy shopping!", shopUserName),
				Instagram: fmt.Sprintf("@%v", shopUserName),
				Facebook:  fmt.Sprintf("@%v", shopUserName),
				X:         fmt.Sprintf("@%v", shopUserName),
			}
			shop := models.Shop{
				ID:                 shopID,
				Name:               shopName,
				Description:        shopDescription,
				Username:           shopUserName,
				UserID:             session_.UserId,
				ListingActiveCount: 0,
				Announcement:       "",
				IsVacation:         false,
				VacationMessage:    "",
				Slug:               slug,
				LogoURL:            logoUploadUrl,
				BannerURL:          bannerUploadUrl,
				Gallery:            []string{},
				FollowerCount:      0,
				Followers:          []models.ShopFollower{},
				Status:             models.ShopStatusActive,
				IsLive:             true,
				CreatedAt:          now,
				ModifiedAt:         now,
				Policy:             policy,
				ReviewsCount:       0,
				Rating:             shopRating,
				About:              shopAboutData,
			}
			_, err := common.ShopCollection.InsertOne(ctx, shop)
			if err != nil {
				return nil, err
			}

			// Update user profile shop
			filter := bson.M{"_id": session_.UserId}
			update := bson.M{"$set": bson.M{"shop_id": shopID, "is_seller": true, "modified_at": now}}
			result, err := common.UserCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				return nil, err
			}

			return result, nil
		}

		_, err = session.WithTransaction(context.Background(), callback, txnOptions)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}
		err = session.CommitTransaction(context.Background())
		if err != nil {
			// delete media
			util.DestroyMedia(logoUploadResult.PublicID)
			util.DestroyMedia(bannerUploadResult.PublicID)
			// return error
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		session.EndSession(context.Background())

		if emailService == nil {
			emailService = services.NewEmailService()
		}
		
		// send success shop creation notification
		emailService.SendNewShopEmail(userEmail, loginName, shopName)

		internal.PublishCacheMessage(c, internal.CacheInvalidateShop, shopID.Hex())
		util.HandleSuccess(c, http.StatusOK, shopID.Hex(), shopID.Hex())
	}
}

func UpdateShopInformation() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shopId, myId, err := common.MyShopIdAndMyId(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		updateData := bson.M{}

		if shopName := c.Request.FormValue("name"); shopName != "" {
			if err := util.ValidateShopName(shopName); err != nil {
				util.HandleError(c, http.StatusBadRequest, err)
				return
			}
			updateData["name"] = shopName
		}

		if shopUsername := c.Request.FormValue("username"); shopUsername != "" {
			if err := util.ValidateShopUserName(shopUsername); err != nil {
				util.HandleError(c, http.StatusBadRequest, err)
				return
			}
			updateData["username"] = shopUsername
		}

		if description := c.Request.FormValue("description"); description != "" {
			updateData["description"] = description
		}

		var logoUploadResult uploader.UploadResult
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
			updateData["logo_url"] = logoUploadResult.SecureURL
		} else if err != http.ErrMissingFile {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		var bannerUploadResult uploader.UploadResult
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
			updateData["banner_url"] = bannerUploadResult.SecureURL
		} else if err != http.ErrMissingFile {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		if len(updateData) == 0 {
			// delete media
			_, err := util.DestroyMedia(logoUploadResult.PublicID)
			log.Println(err)
			_, err = util.DestroyMedia(bannerUploadResult.PublicID)
			log.Println(err)
			// return error

			util.HandleError(c, http.StatusOK, errors.New("no update data provided"))
			return
		}

		updateData["modified_at"] = time.Now()

		filter := bson.M{"_id": shopId, "user_id": myId}
		update := bson.M{"$set": updateData}

		_, err = common.ShopCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			// delete media
			_, err = util.DestroyMedia(logoUploadResult.PublicID)
			log.Println(err)
			_, err = util.DestroyMedia(bannerUploadResult.PublicID)
			log.Println(err)
			// return error

			util.HandleError(c, http.StatusExpectationFailed, err)
			return
		}

		internal.PublishCacheMessage(c, internal.CacheInvalidateShop, shopId.Hex())

		util.HandleSuccess(c, http.StatusOK, "Shop information updated successfully", shopId.Hex())
	}
}

func UpdateMyShopStatus() gin.HandlerFunc {
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

		filter := bson.M{"_id": shopId, "user_id": myId}
		update := bson.M{"$set": bson.M{"is_live": payload.Status, "modified_at": time.Now()}}
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

		util.HandleSuccess(c, http.StatusOK, "Shop status was updated successful", shopId.Hex())
	}
}

func UpdateShopAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now()
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
	}
}

// GetShop - api/shops/:shopid
func GetShop() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var err error
		var shopIdentifier bson.M

		shopID := c.Param("shopid")
		if primitive.IsValidObjectID(shopID) {
			// If shopid is a valid object ID string
			shopObjectID, err := primitive.ObjectIDFromHex(shopID)
			if err != nil {
				util.HandleError(c, http.StatusNotFound, err)
				return
			}
			shopIdentifier = bson.M{"_id": shopObjectID}
		} else {
			// If shopid is a string (e.g., slug)
			shopIdentifier = bson.M{"slug": shopID}
		}

		shopPipeline := []bson.M{
			{"$match": shopIdentifier},
			{
				"$lookup": bson.M{
					"from":         "User",
					"localField":   "user_id",
					"foreignField": "_id",
					"as":           "user",
				},
			},
			{
				"$unwind": bson.M{
					"path":                       "$user",
					"preserveNullAndEmptyArrays": true,
				},
			},
			{
				"$project": bson.M{
					"_id":                      1,
					"name":                     1,
					"description":              1,
					"user_id":                  1,
					"username":                 1,
					"user_address_id":          1,
					"listing_active_count":     1,
					"announcement":             1,
					"announcement_modified_at": 1,
					"is_vacation":              1,
					"vacation_message":         1,
					"slug":                     1,
					"logo_url":                 1,
					"banner_url":               1,
					"gallery":                  1,
					"follower_count":           1,
					"followers":                1,
					"status":                   1,
					"is_live":                  1,
					"created_at":               1,
					"modified_at":              1,
					"policy":                   1,
					"recent_reviews":           1,
					"reviews_count":            1,
					"sales_message":            1,
					"rating":                   1,
					"address":                  1,
					"about":                    1,
					"user": bson.M{
						"login_name":             "$user.login_name",
						"first_name":             "$user.first_name",
						"last_name":              "$user.last_name",
						"thumbnail":              "$user.thumbnail",
						"transaction_buy_count":  "$user.transaction_buy_count",
						"transaction_sold_count": "$user.transaction_sold_count",
					},
				},
			},
		}
		cursor, err := common.ShopCollection.Aggregate(ctx, shopPipeline)

		var shop models.Shop
		if err != nil {
			if err == mongo.ErrNoDocuments {
				util.HandleError(c, http.StatusNotFound, err)
				return
			}

			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}
		if cursor.Next(ctx) {
			if err := cursor.Decode(&shop); err != nil {
				util.HandleError(c, http.StatusInternalServerError, err)
				return
			}
		} else {
			log.Printf("NotFound, %v %v", shopIdentifier, err)
			util.HandleError(c, http.StatusNotFound, errors.New("no shop found"))
			return
		}
		shop.ConstructShopLinks()

		withCategory := c.Query("category")
		if len(withCategory) > 0 {
			listingPipeline := []bson.M{
				{"$match": bson.M{"shop_id": shop.ID}},
				{"$group": bson.M{"_id": "$details.category.category_name", "count": bson.M{"$sum": 1}}},
				{"$project": bson.M{"name": "$_id", "count": 1, "_id": 0, "path": "$details.category.category_path"}},
			}

			cursor, err = common.ListingCollection.Aggregate(ctx, listingPipeline)
			if err != nil {
				if err != mongo.ErrNoDocuments {
					util.HandleError(c, http.StatusInternalServerError, err)
					return
				}
			}

			var shopCategories []models.ShopCategory
			if cursor.Next(ctx) {
				var shopCategory models.ShopCategory
				if err := cursor.Decode(&shopCategory); err != nil {
					util.HandleError(c, http.StatusInternalServerError, err)
					return
				}

				shopCategories = append(shopCategories, shopCategory)
			}

			shop.Categories = shopCategories
		}

		util.HandleSuccess(c, http.StatusOK, "Shop retrieved successfully", shop)
	}
}

// GetShopByOwnerUserId - api/users/:userid/shops
func GetShopByOwnerUserId() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		userIDStr := c.Param("userid")
		userID, err := primitive.ObjectIDFromHex(userIDStr)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		var shop models.Shop
		err = common.ShopCollection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&shop)
		if err != nil {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Shop retrieved successfully", shop)
	}
}

// GetShops - api/shops/?limit=50&skip=0
func GetShops() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		paginationArgs := common.GetPaginationArgs(c)

		// Update the query filter to include the status check
		filter := bson.D{{Key: "status", Value: models.ShopStatusActive}}

		find := options.Find().SetLimit(int64(paginationArgs.Limit)).SetSkip(int64(paginationArgs.Skip))
		result, err := common.ShopCollection.Find(ctx, filter, find)
		if err != nil {
			log.Printf("error finding shop members: %v", err.Error())
			util.HandleError(c, http.StatusNotFound, err)
			return
		}

		var shops []models.Shop
		if err = result.All(ctx, &shops); err != nil {
			log.Printf("error decoding shop members: %v", err.Error())
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
func SearchShops() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		query := c.Query("q")
		paginationArgs := common.GetPaginationArgs(c)

		// Query the database for shops that match the search query
		shops, err := common.ShopCollection.Find(ctx, bson.M{
			"$or": []bson.M{
				{"shop_name": bson.M{"$regex": primitive.Regex{Pattern: query, Options: "i"}}},
				{"description": bson.M{"$regex": primitive.Regex{Pattern: query, Options: "i"}}},
			},
		}, options.Find().SetSkip(int64(paginationArgs.Skip)).SetLimit(int64(paginationArgs.Limit)))
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		// Count the total number of shops that match the search query
		count, err := common.ShopCollection.CountDocuments(ctx, bson.M{
			"$or": []bson.M{
				{"shop_name": bson.M{"$regex": primitive.Regex{Pattern: query, Options: "i"}}},
				{"description": bson.M{"$regex": primitive.Regex{Pattern: query, Options: "i"}}},
			},
		})
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		// Serialize the shops and return them to the client
		var serializedShops []models.Shop
		for shops.Next(ctx) {
			var shop models.Shop
			if err := shops.Decode(&shop); err != nil {
				util.HandleError(c, http.StatusInternalServerError, err)
				return
			}
			serializedShops = append(serializedShops, shop)
		}

		util.HandleSuccessMeta(c, http.StatusOK, "Shops found", serializedShops,
			gin.H{"pagination": util.Pagination{
				Limit: paginationArgs.Limit,
				Skip:  paginationArgs.Skip,
				Count: count,
			}})
	}
}

func UpdateShopField() gin.HandlerFunc {
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
func UpdateShopAnnouncement() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQUEST_TIMEOUT_SECS)
		defer cancel()

		now := time.Now()
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

		filter := bson.M{"_id": shopId, "user_id": myId}
		update := bson.M{"$set": bson.M{"announcement": announcement.Announcement, "announcement_modified_at": now, "modified_at": now}}
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

		util.HandleSuccess(c, http.StatusOK, "Shop announcement updated successfully", res.UpsertedID)
	}
}

func UpdateShopVacation() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var vacation models.ShopVacationRequest
		now := time.Now()
		if err := c.BindJSON(&vacation); err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		shopId, myId, err := common.MyShopIdAndMyId(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		filter := bson.M{"_id": shopId, "user_id": myId}
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
	}
}

func UpdateShopLogo() gin.HandlerFunc {
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

		now := time.Now()
		filter := bson.M{"_id": shopId, "user_id": myId}
		update := bson.M{"$set": bson.M{"logo_url": logoUploadResult.SecureURL, "modified_at": now}}
		res, err := common.ShopCollection.UpdateOne(ctx, filter, update)
		if err != nil || res.ModifiedCount == 0 {
			// delete media
			_, err = util.DestroyMedia(logoUploadResult.PublicID)
			log.Println(err)
			// return error
			util.HandleError(c, http.StatusNotModified, err)
			return
		}

		internal.PublishCacheMessage(c, internal.CacheInvalidateShop, shopId.Hex())

		util.HandleSuccess(c, http.StatusOK, "Shop logo updated successfully", res.UpsertedID)
	}
}

func UpdateShopBanner() gin.HandlerFunc {
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

		now := time.Now()
		filter := bson.M{"_id": shopId, "user_id": myId}
		update := bson.M{"banner_url": bannerUploadResult.SecureURL, "modified_at": now}
		res, err := common.ShopCollection.UpdateOne(ctx, filter, update)
		if err != nil || res.ModifiedCount == 0 {
			// delete media
			_, err = util.DestroyMedia(bannerUploadResult.PublicID)
			log.Println(err)
			// return error
			util.HandleError(c, http.StatusNotModified, err)
			return
		}

		internal.PublishCacheMessage(c, internal.CacheInvalidateShop, shopId.Hex())

		util.HandleSuccess(c, http.StatusOK, "Shop banner updated successfully", res.UpsertedID)
	}
}

func UpdateShopGallery() gin.HandlerFunc {
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

		now := time.Now()
		filter := bson.M{"_id": shopId, "user_id": myId}
		update := bson.M{"$push": bson.M{"gallery": imageUploadResult.SecureURL}, "modified_at": now}
		res, err := common.ShopCollection.UpdateOne(ctx, filter, update)
		if err != nil || res.ModifiedCount == 0 {
			// delete media
			_, err = util.DestroyMedia(imageUploadResult.PublicID)
			log.Println(err)
			// return error
			util.HandleError(c, http.StatusNotModified, err)
			return
		}

		internal.PublishCacheMessage(c, internal.CacheInvalidateShop, shopId.Hex())

		util.HandleSuccess(c, http.StatusOK, "Image added to shop gallery successfully", res.UpsertedCount)
	}
}

// DeleteFromShopGallery - api/shops/:shopid/favorers?image={image_url}
func DeleteFromShopGallery() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		imageURL := c.Query("image")
		now := time.Now()

		shopID, myID, err := common.MyShopIdAndMyId(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		filter := bson.M{"_id": shopID, "user_id": myID}
		update := bson.M{"$pull": bson.M{"gallery": imageURL}, "modified_at": now}
		res, err := common.ShopCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			util.HandleError(c, http.StatusNotModified, err)
			return
		}
		if res.ModifiedCount == 0 {
			util.HandleError(c, http.StatusNotFound, errors.New("no matching documents found"))
			return
		}

		internal.PublishCacheMessage(c, internal.CacheInvalidateShop, shopID.Hex())

		util.HandleSuccess(c, http.StatusOK, "Image removed from shop gallery successfully", res.UpsertedID)
	}
}

// FollowShop - api/shops/:shopid/followers
func FollowShop() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		now := time.Now()

		shopId, myId, err := common.MyShopIdAndMyId(c)
		if err != nil {
			log.Println(err)
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		// Shop Member session
		wc := writeconcern.New(writeconcern.WMajority())
		txnOptions := options.Transaction().SetWriteConcern(wc)

		session, err := util.DB.StartSession()
		if err != nil {
			util.HandleError(c, http.StatusUnprocessableEntity, err)
			return
		}
		defer session.EndSession(ctx)
		followerId := primitive.NewObjectID()
		callback := func(ctx mongo.SessionContext) (any, error) {
			var user models.User
			err := common.UserCollection.FindOne(ctx, bson.M{"_id": myId}).Decode(&user)
			if err != nil {
				log.Println(err)
				return nil, err
			}

			var currentShop models.Shop
			err = common.ShopCollection.FindOne(ctx, bson.M{"_id": shopId}).Decode(&currentShop)
			if err != nil {
				log.Println(err)
				return nil, err
			}

			// Attempt to add member to member collection
			shopMemberData := models.ShopFollower{
				Id:        followerId,
				UserId:    myId,
				ShopId:    shopId,
				LoginName: user.LoginName,
				FirstName: user.FirstName,
				LastName:  user.LastName,
				Thumbnail: user.Thumbnail,
				IsOwner:   currentShop.UserID == myId,
				JoinedAt:  time.Now(),
			}
			_, err = common.ShopFollowerCollection.InsertOne(ctx, shopMemberData)
			if err != nil {
				log.Println(err)
				return nil, err
			}

			// Attempt to add follower to follower field in shop
			inner := models.ShopFollowerExcerpt{
				Id:        followerId,
				UserId:    myId,
				LoginName: user.LoginName,
				FirstName: user.FirstName,
				LastName:  user.LastName,
				Thumbnail: user.Thumbnail,
				IsOwner:   currentShop.UserID == myId,
			}
			filter := bson.M{"_id": shopId, "followers": bson.M{"$not": bson.M{"$elemMatch": bson.M{"user_id": &user.Id}}}}
			update := bson.M{
				"$push": bson.M{
					"followers": bson.M{
						"$each":  bson.A{inner},
						"$sort":  -1,
						"$slice": -5,
					},
				},
				"$set": bson.M{"modified_at": now},
				"$inc": bson.M{"follower_count": 1},
			}
			result, err := common.ShopCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				return nil, err
			}

			if result.ModifiedCount == 0 {
				return nil, errors.New("no matching documents found")
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

		internal.PublishCacheMessage(c, internal.CacheInvalidateShop, shopId.Hex())

		util.HandleSuccess(c, http.StatusOK, "You're now a follower of this shop", followerId.Hex())
	}
}

// GetShopFollowers - api/shops/:shopid/followers?limit=50&skip=0
func GetShopFollowers() gin.HandlerFunc {
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
		filter := bson.M{"shop_id": shopObjectID}
		find := options.Find().SetLimit(int64(paginationArgs.Limit)).SetSkip(int64(paginationArgs.Skip))
		result, err := common.ShopFollowerCollection.Find(ctx, filter, find)
		if err != nil {
			log.Printf("%v", err)
			util.HandleError(c, http.StatusNotFound, err)
			return
		}

		count, err := common.ShopFollowerCollection.CountDocuments(ctx, bson.M{"shop_id": shopObjectID})
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		var shopFollowers []models.ShopFollower
		if err = result.All(ctx, &shopFollowers); err != nil {
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
func IsFollowingShop() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shopId, myId, err := common.MyShopIdAndMyId(c)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, err)
			return
		}

		filter := bson.M{"user_id": myId, "shop_id": shopId}
		var follower models.ShopFollower
		err = common.ShopFollowerCollection.FindOne(ctx, filter).Decode(&follower)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				util.HandleSuccess(c, http.StatusOK, "Success", false)
				return
			}

			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Success", true)
	}
}

// UnfollowShop - api/shops/:shopid/followers
func UnfollowShop() gin.HandlerFunc {
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
func RemoveOtherFollower() gin.HandlerFunc {
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
		ownershipEerr := common.VerifyShopOwnership(ctx, myId, shopId)
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
func UpdateShopAbout() gin.HandlerFunc {
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

		err = common.VerifyShopOwnership(c, myId, shopId)
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
func CreateShopReturnPolicy() gin.HandlerFunc {
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

		err = common.VerifyShopOwnership(c, myId, shopId)
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
func UpdateShopReturnPolicy() gin.HandlerFunc {
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

		err = common.VerifyShopOwnership(c, myId, shopId)
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
func DeleteShopReturnPolicy() gin.HandlerFunc {
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

		err = common.VerifyShopOwnership(c, myId, shopId)
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
func GetShopReturnPolicy() gin.HandlerFunc {
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
func GetShopReturnPolicies() gin.HandlerFunc {
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
func CreateShopComplianceInformation() gin.HandlerFunc {
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

		err = common.VerifyShopOwnership(c, myId, shopId)
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
func GetShopComplianceInformation() gin.HandlerFunc {
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
