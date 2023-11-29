package controllers

import (
	"context"
	"errors"
	"fmt"
	"khoomi-api-io/khoomi_api/configs"
	"khoomi-api-io/khoomi_api/email"
	"khoomi-api-io/khoomi_api/helper"
	"khoomi-api-io/khoomi_api/models"
	"khoomi-api-io/khoomi_api/responses"
	"khoomi-api-io/khoomi_api/services"
	"log"
	"net/http"
	"strings"
	"time"

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
		err := ShopCollection.FindOne(ctx, filter).Decode(&shop)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				helper.HandleSuccess(c, http.StatusOK, "Congrats! shop username is available :xD", "")
				return
			}

			helper.HandleError(c, http.StatusInternalServerError, errors.New("internal sever error on checking shop username availability"), "")
			return
		}

		helper.HandleError(c, http.StatusConflict, errors.New("shop username is already taken"), "Shop username is not available")
	}
}

func CreateShop() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		userID, err := configs.ExtractTokenID(c)
		if err != nil {
			helper.HandleError(c, http.StatusUnauthorized, err, "Failed to extract user ID from token")
			return
		}

		loginName, userEmail, err := configs.ExtractTokenLoginNameEmail(c)
		if err != nil {
			log.Println(err)
			helper.HandleError(c, http.StatusUnauthorized, err, "unathorized")
			return
		}

		shopName := c.Request.FormValue("name")
		err = helper.ValidateShopName(shopName)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Invalid shop name")
			return
		}
		shopUserName := c.Request.FormValue("username")
		err = helper.ValidateShopUserName(shopUserName)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Invalid shop username name")
			return
		}
		shopUserName = strings.ToLower(shopUserName)

		shopDescription := c.Request.FormValue("description")
		err = helper.ValidateShopDescription(shopDescription)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Invalid shop description")
			return
		}

		// Logo file handling
		logoFile, _, err := c.Request.FormFile("logo")
		var logoUploadUrl string
		if err == nil {
			logoUploadUrl, err = services.NewMediaUpload().FileUpload(models.File{File: logoFile})
			if err != nil {
				errMsg := fmt.Sprintf("Logo failed to upload - %v", err.Error())
				log.Print(errMsg)
				helper.HandleError(c, http.StatusInternalServerError, err, errMsg)
				return
			}
		} else {
			logoUploadUrl = ""
		}

		// Banner file handling
		bannerFile, _, err := c.Request.FormFile("banner")
		var bannerUploadUrl string
		if err == nil {
			bannerUploadUrl, err = services.NewMediaUpload().FileUpload(models.File{File: bannerFile})
			if err != nil {
				errMsg := fmt.Sprintf("Banner failed to upload - %v", err.Error())
				log.Print(errMsg)
				helper.HandleError(c, http.StatusInternalServerError, err, errMsg)
				return
			}
		} else {
			bannerUploadUrl = ""
		}

		wc := writeconcern.New(writeconcern.WMajority())
		txnOptions := options.Transaction().SetWriteConcern(wc)
		session, err := configs.DB.StartSession()
		if err != nil {
			helper.HandleError(c, http.StatusUnprocessableEntity, err, "Failed to start new session")
			return
		}
		defer session.EndSession(ctx)
		callback := func(ctx mongo.SessionContext) (interface{}, error) {
			now := time.Now()
			slug := slug2.Make(shopUserName)
			policy := models.ShopPolicy{
				PaymentPolicy:  "",
				ShippingPolicy: "",
				RefundPolicy:   "",
				AdditionalInfo: "",
			}
			shopID := primitive.NewObjectID()
			shop := models.Shop{
				ID:                 shopID,
				Name:               shopName,
				Description:        shopDescription,
				Username:           shopUserName,
				UserID:             userID,
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
				RecentReviews:      []models.ShopReview{},
				ReviewsCount:       0,
			}
			_, err := ShopCollection.InsertOne(ctx, shop)
			if err != nil {
				return nil, err
			}

			// Update user profile shop
			filter := bson.M{"_id": userID}
			update := bson.M{"$set": bson.M{"shop_id": shopID, "is_seller": true}}
			result, err := UserCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				return nil, err
			}

			// Create shop about profile.
			shopAboutData := models.ShopAbout{
				ID:        primitive.NewObjectID(),
				ShopID:    shopID,
				Status:    models.ShopAboutStatusDraft,
				Headline:  fmt.Sprintf("Welcome to %v!", shop.Username),
				Story:     fmt.Sprintf("Thank you for visiting our online artisan shop. We are passionate about craftsmanship and dedicated to providing unique, handcrafted items that reflect the creativity and skill of our artisans. Explore our collection and discover the beauty of handmade products that carry a story of craftsmanship and tradition.\n\nAt %v, we believe in the art of creating something special. Each piece in our collection is carefully crafted with attention to detail and a commitment to quality. We aim to connect artisans with appreciative buyers, creating a community that values and supports the artistry behind every creation.\n\nJoin us on this journey of celebrating craftsmanship and supporting talented artisans from around the world. Your purchase not only adds a unique piece to your life but also contributes to the livelihood of skilled individuals who pour their heart and soul into their work.\n\nThank you for being a part of our community. Happy shopping!", shop.Username),
				Instagram: fmt.Sprintf("@%v", shop.Username),
				Facebook:  fmt.Sprintf("@%v", shop.Username),
				X:         fmt.Sprintf("@%v", shop.Username),
			}

			_, err = ShopAboutCollection.InsertOne(ctx, shopAboutData)
			if err != nil {
				helper.HandleError(c, http.StatusInternalServerError, err, "Error creating shop about")
				return nil, err
			}

			return result, nil
		}

		_, err = session.WithTransaction(context.Background(), callback, txnOptions)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Failed to create shop")
			return
		}

		if err := session.CommitTransaction(context.Background()); err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Failed to commit transaction")
			return
		}

		session.EndSession(context.Background())

		// send success shop creation notification
		email.SendNewShopEmail(userEmail, loginName, shopName)

		helper.HandleSuccess(c, http.StatusOK, "Shop creation was successful", "")
	}
}

func UpdateShopInformation() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Error getting shop ID and user ID")
			return
		}

		updateData := bson.M{}

		if shopName := c.Request.FormValue("name"); shopName != "" {
			if err := helper.ValidateShopName(shopName); err != nil {
				helper.HandleError(c, http.StatusBadRequest, err, "Invalid shop name format")
				return
			}
			updateData["name"] = shopName
		}

		if shopUsername := c.Request.FormValue("username"); shopUsername != "" {
			if err := helper.ValidateShopUserName(shopUsername); err != nil {
				helper.HandleError(c, http.StatusBadRequest, err, "Invalid shop username format")
				return
			}
			updateData["username"] = shopUsername
		}

		if description := c.Request.FormValue("description"); description != "" {
			updateData["description"] = description
		}

		if fileHeader, err := c.FormFile("logo_url"); err == nil {
			file, err := fileHeader.Open()
			if err != nil {
				helper.HandleError(c, http.StatusInternalServerError, err, "Failed to retrieve uploaded file")
				return
			}
			defer file.Close()

			uploadUrl, err := services.NewMediaUpload().FileUpload(models.File{File: file})
			if err != nil {
				log.Printf("Logo Image upload failed - %v", err.Error())
				helper.HandleError(c, http.StatusInternalServerError, err, "Failed to upload file logo")
				return
			}
			updateData["logo_url"] = uploadUrl
		} else if err != http.ErrMissingFile {
			helper.HandleError(c, http.StatusInternalServerError, err, "Failed to retrieve uploaded file")
			return
		}

		if fileHeader, err := c.FormFile("banner_url"); err == nil {
			file, err := fileHeader.Open()
			if err != nil {
				helper.HandleError(c, http.StatusInternalServerError, err, "Failed to retrieve uploaded file")
				return
			}
			defer file.Close()

			uploadUrl, err := services.NewMediaUpload().FileUpload(models.File{File: file})
			if err != nil {
				log.Printf("Banner Image upload failed - %v", err.Error())
				helper.HandleError(c, http.StatusInternalServerError, err, "Failed to upload file banner")
				return
			}
			updateData["banner_url"] = uploadUrl
		} else if err != http.ErrMissingFile {
			helper.HandleError(c, http.StatusInternalServerError, err, "Failed to retrieve uploaded file")
			return
		}

		if len(updateData) == 0 {
			helper.HandleError(c, http.StatusOK, errors.New("no update data provided"), "No update data provided")
			return
		}

		updateData["modified_at"] = time.Now()

		filter := bson.M{"_id": shopId, "user_id": myId}
		update := bson.M{"$set": updateData}

		_, err = ShopCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			helper.HandleError(c, http.StatusExpectationFailed, err, "Failed to update user's shop information")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "Shop information updated successfully", nil)
	}
}

func UpdateMyShopStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var payload models.UpdateShopStatusReq
		if err := c.BindJSON(&payload); err != nil {
			log.Println(err)
			helper.HandleError(c, http.StatusBadRequest, err, "Invalid request body")
			return
		}

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Error getting shop ID and user ID")
			return
		}

		filter := bson.M{"_id": shopId, "user_id": myId}
		update := bson.M{"$set": bson.M{"is_live": payload.Status, "modified_at": time.Now()}}
		res, err := ShopCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Error updating shop status")
			return
		}
		if res.ModifiedCount == 0 {
			helper.HandleError(c, http.StatusNotFound, errors.New("no matching documents found"), "No matching documents found")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "Shop status was updated successful", "")
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
				helper.HandleError(c, http.StatusNotFound, err, "Invalid shop ID")
				return
			}
			shopIdentifier = bson.M{"_id": shopObjectID}
		} else {
			// If shopid is a string (e.g., slug)
			shopIdentifier = bson.M{"slug": shopID}
		}

		pipeline := []bson.M{
			{"$match": shopIdentifier},
			{
				"$lookup": bson.M{
					"from":         "UserAddress",
					"localField":   "user_address_id",
					"foreignField": "_id",
					"as":           "address",
				},
			},
			{"$unwind": bson.M{
				"path":                       "$address",
				"preserveNullAndEmptyArrays": true,
			},
			},
			{
				"$lookup": bson.M{
					"from":         "User",
					"localField":   "user_id",
					"foreignField": "_id",
					"as":           "user",
				},
			},
			{"$unwind": bson.M{
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
					"user": bson.M{
						"login_name":             "$user.login_name",
						"first_name":             "$user.first_name",
						"last_name":              "$user.last_name",
						"thumbnail":              "$user.thumbnail",
						"transaction_buy_count":  "$user.transaction_buy_count",
						"transaction_sold_count": "$user.transaction_sold_count",
					},
					"address": bson.M{
						"city":        "$address.city",
						"state":       "$address.state",
						"street":      "$address.street",
						"postal_code": "$address.postal_code",
						"country":     "$address.country",
					},
				},
			},
		}

		cursor, err := ShopCollection.Aggregate(ctx, pipeline)

		var shop models.Shop
		if err != nil {
			log.Println(err)
			helper.HandleError(c, http.StatusInternalServerError, err, "error while retrieving listing")
			return
		}
		if cursor.Next(ctx) {
			if err := cursor.Decode(&shop); err != nil {
				log.Println(err)
				helper.HandleError(c, http.StatusInternalServerError, err, "error while decoding listing")
				return
			}
		} else {
			log.Printf("NotFound, %v %v", shopIdentifier, err)
			helper.HandleError(c, http.StatusNotFound, errors.New("no shop found"), "no shop found")
			return
		}

		if err != nil {
			if err == mongo.ErrNoDocuments {
				helper.HandleError(c, http.StatusNotFound, err, "Shop not found")
				return
			}
			helper.HandleError(c, http.StatusNotFound, err, "error retrieving shop by id")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "Shop retrieved successfully", gin.H{"shop": shop})
	}
}

// GetShopByOwnerUserId - api/users/:userid/shops
func GetShopByOwnerUserId() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		userIDStr := c.Param("userId")
		userID, err := primitive.ObjectIDFromHex(userIDStr)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Invalid user ID")
			return
		}

		var shop models.Shop
		err = ShopCollection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&shop)
		if err != nil {
			helper.HandleError(c, http.StatusNotFound, err, "Shop not found")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "Shop retrieved successfully", gin.H{"shop": shop})
	}
}

// GetShops - api/shops/?limit=50&skip=0
func GetShops() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		paginationArgs := services.GetPaginationArgs(c)

		// Update the query filter to include the status check
		filter := bson.D{{Key: "status", Value: models.ShopStatusActive}}

		find := options.Find().SetLimit(int64(paginationArgs.Limit)).SetSkip(int64(paginationArgs.Skip))
		result, err := ShopCollection.Find(ctx, filter, find)
		if err != nil {
			log.Printf("error finding shop members: %v", err.Error())
			helper.HandleError(c, http.StatusNotFound, err, "error finding shop members")
			return
		}

		var shops []models.Shop
		if err = result.All(ctx, &shops); err != nil {
			log.Printf("error decoding shop members: %v", err.Error())
			helper.HandleError(c, http.StatusNotFound, err, "error decoding shop members")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "Shops retrieved successfully",
			gin.H{"members": shops, "pagination": responses.Pagination{
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
		shops, err := ShopCollection.Find(ctx, bson.M{
			"$or": []bson.M{
				{"shop_name": bson.M{"$regex": primitive.Regex{Pattern: query, Options: "i"}}},
				{"description": bson.M{"$regex": primitive.Regex{Pattern: query, Options: "i"}}},
			},
		}, options.Find().SetSkip(int64(paginationArgs.Skip)).SetLimit(int64(paginationArgs.Limit)))
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Error searching for shops")
			return
		}

		// Count the total number of shops that match the search query
		count, err := ShopCollection.CountDocuments(ctx, bson.M{
			"$or": []bson.M{
				{"shop_name": bson.M{"$regex": primitive.Regex{Pattern: query, Options: "i"}}},
				{"description": bson.M{"$regex": primitive.Regex{Pattern: query, Options: "i"}}},
			},
		})
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Error counting shops")
			return
		}

		// Serialize the shops and return them to the client
		var serializedShops []models.Shop
		for shops.Next(ctx) {
			var shop models.Shop
			if err := shops.Decode(&shop); err != nil {
				helper.HandleError(c, http.StatusInternalServerError, err, "Error decoding shops")
				return
			}
			serializedShops = append(serializedShops, shop)
		}

		helper.HandleSuccess(c, http.StatusOK, "Shops found",
			gin.H{"shops": serializedShops, "pagination": responses.Pagination{
				Limit: paginationArgs.Limit,
				Skip:  paginationArgs.Skip,
				Count: count,
			}})
	}
}

// UpdateShopAnnouncement - api/shops/:shopid/announcement
func UpdateShopAnnouncement() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		now := time.Now()
		var announcement models.ShopAnnouncementRequest
		if err := c.BindJSON(&announcement); err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Error binding JSON")
			return
		}

		if announcement.Announcement == "" {
			helper.HandleError(c, http.StatusBadRequest, errors.New("announcement cannot be empty"), "Invalid announcement")
			return
		}

		if len(announcement.Announcement) > 100 {
			helper.HandleError(c, http.StatusBadRequest, errors.New("announcement is too long"), "Invalid announcement length")
			return
		}

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Error getting shop ID and user ID")
			return
		}

		filter := bson.M{"_id": shopId, "user_id": myId}
		update := bson.M{"$set": bson.M{"announcement": announcement.Announcement, "announcement_modified_at": now, "modified_at": now}}
		res, err := ShopCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Error updating shop announcement")
			return
		}
		if res.ModifiedCount == 0 {
			helper.HandleError(c, http.StatusNotFound, errors.New("no matching documents found"), "No matching documents found")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "Shop announcement updated successfully", nil)
	}
}

func UpdateShopVacation() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var vacation models.ShopVacationRequest
		now := time.Now()
		if err := c.BindJSON(&vacation); err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Error binding JSON")
			return
		}

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Error getting shop ID and user ID")
			return
		}

		filter := bson.M{"_id": shopId, "user_id": myId}
		update := bson.M{"$set": bson.M{"vacation_message": vacation.Message, "is_vacation": vacation.IsVacation, "modified_at": now}}
		res, err := ShopCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Error updating shop vacation")
			return
		}
		if res.ModifiedCount == 0 {
			helper.HandleError(c, http.StatusNotFound, errors.New("no matching documents found"), "No matching documents found")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "Shop vacation updated successfully", nil)
	}
}

func UpdateShopLogo() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Error getting shop ID and user ID")
			return
		}

		logoFile, _, err := c.Request.FormFile("logo")
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Error retrieving logo file")
			return
		}

		logoUploadUrl, err := services.NewMediaUpload().FileUpload(models.File{File: logoFile})
		if err != nil {
			errMsg := fmt.Sprintf("Logo failed to upload - %v", err.Error())
			log.Print(errMsg)
			helper.HandleError(c, http.StatusInternalServerError, errors.New(errMsg), "Error uploading logo")
			return
		}

		now := time.Now()
		filter := bson.M{"_id": shopId, "user_id": myId}
		update := bson.M{"logo_url": logoUploadUrl, "modified_at": now}
		res, err := ShopCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			helper.HandleError(c, http.StatusNotModified, err, "Error updating shop logo")
			return
		}
		if res.ModifiedCount == 0 {
			helper.HandleError(c, http.StatusNotFound, errors.New("no matching documents found"), "No matching documents found")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "Shop logo updated successfully", nil)
	}
}

func UpdateShopBanner() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Error getting shop ID and user ID")
			return
		}

		bannerFile, _, err := c.Request.FormFile("banner")
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Error retrieving banner file")
			return
		}

		bannerFileURL, err := services.NewMediaUpload().FileUpload(models.File{File: bannerFile})
		if err != nil {
			errMsg := fmt.Sprintf("Banner failed to upload - %v", err.Error())
			log.Print(errMsg)
			helper.HandleError(c, http.StatusInternalServerError, errors.New(errMsg), "Error uploading banner")
			return
		}

		now := time.Now()
		filter := bson.M{"_id": shopId, "user_id": myId}
		update := bson.M{"banner_url": bannerFileURL, "modified_at": now}
		res, err := ShopCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			helper.HandleError(c, http.StatusNotModified, err, "Error updating shop banner")
			return
		}
		if res.ModifiedCount == 0 {
			helper.HandleError(c, http.StatusNotFound, errors.New("no matching documents found"), "No matching documents found")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "Shop banner updated successfully", nil)
	}
}

func UpdateShopGallery() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Error getting shop ID and user ID")
			return
		}

		imageFile, _, err := c.Request.FormFile("image")
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Error retrieving image file")
			return
		}

		imageFileURL, err := services.NewMediaUpload().FileUpload(models.File{File: imageFile})
		if err != nil {
			errMsg := fmt.Sprintf("Image failed to upload - %v", err.Error())
			log.Print(errMsg)
			helper.HandleError(c, http.StatusInternalServerError, errors.New(errMsg), "Error uploading image")
			return
		}

		now := time.Now()
		filter := bson.M{"_id": shopId, "user_id": myId}
		update := bson.M{"$push": bson.M{"gallery": imageFileURL}, "modified_at": now}
		res, err := ShopCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			helper.HandleError(c, http.StatusNotModified, err, "Error updating shop gallery")
			return
		}
		if res.ModifiedCount == 0 {
			helper.HandleError(c, http.StatusNotFound, errors.New("no matching documents found"), "No matching documents found")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "Image added to shop gallery successfully", nil)
	}
}

// DeleteFromShopGallery - api/shops/:shopid/favorers?image={image_url}
func DeleteFromShopGallery() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		imageURL := c.Query("image")
		now := time.Now()

		shopID, myID, err := services.MyShopIdAndMyId(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Error getting shop ID and user ID")
			return
		}

		err = VerifyShopOwnership(c, myID, shopID)
		if err != nil {
			log.Printf("You don't have write access to this shop: %s\n", err.Error())
			helper.HandleError(c, http.StatusUnauthorized, err, "You don't have write access to this shop")
			return
		}

		filter := bson.M{"_id": shopID, "user_id": myID}
		update := bson.M{"$pull": bson.M{"gallery": imageURL}, "modified_at": now}
		res, err := ShopCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			helper.HandleError(c, http.StatusNotModified, err, "Error removing image from shop gallery")
			return
		}
		if res.ModifiedCount == 0 {
			helper.HandleError(c, http.StatusNotFound, errors.New("no matching documents found"), "No matching documents found")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "Image removed from shop gallery successfully", nil)
	}
}

// FollowShop - api/shops/:shopid/followers
func FollowShop() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		now := time.Now()

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			log.Println(err)
			helper.HandleError(c, http.StatusBadRequest, err, "Error getting shop ID and user ID")
			return
		}

		// Shop Member session
		wc := writeconcern.New(writeconcern.WMajority())
		txnOptions := options.Transaction().SetWriteConcern(wc)

		session, err := configs.DB.StartSession()
		if err != nil {
			helper.HandleError(c, http.StatusUnprocessableEntity, err, "Unable to start new session")
			return
		}
		defer session.EndSession(ctx)

		callback := func(ctx mongo.SessionContext) (interface{}, error) {
			var user models.User
			err := UserCollection.FindOne(ctx, bson.M{"_id": myId}).Decode(&user)
			if err != nil {
				log.Println(err)
				return nil, err
			}

			var currentShop models.Shop
			err = ShopCollection.FindOne(ctx, bson.M{"_id": shopId}).Decode(&currentShop)
			if err != nil {
				log.Println(err)
				return nil, err
			}

			// Attempt to add member to member collection
			followerId := primitive.NewObjectID()
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
			_, err = ShopFollowerCollection.InsertOne(ctx, shopMemberData)
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
			result, err := ShopCollection.UpdateOne(ctx, filter, update)
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
			helper.HandleError(c, http.StatusBadRequest, err, "Error executing transaction")
			return
		}

		if err := session.CommitTransaction(ctx); err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Error committing transaction")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "You're now a follower of this shop", nil)
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
			helper.HandleError(c, http.StatusBadRequest, err, "Invalid user ID")
			return
		}

		paginationArgs := services.GetPaginationArgs(c)
		filter := bson.M{"shop_id": shopObjectID}
		find := options.Find().SetLimit(int64(paginationArgs.Limit)).SetSkip(int64(paginationArgs.Skip))
		result, err := ShopFollowerCollection.Find(ctx, filter, find)
		if err != nil {
			log.Printf("%v", err)
			helper.HandleError(c, http.StatusNotFound, err, "Error finding shop followers")
			return
		}

		count, err := ShopFollowerCollection.CountDocuments(ctx, bson.M{"shop_id": shopObjectID})
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Error counting shop followers")
			return
		}

		var shopFollowers []models.ShopFollower
		if err = result.All(ctx, &shopFollowers); err != nil {
			helper.HandleError(c, http.StatusNotFound, err, "Error retrieving shop followers")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "Success",
			gin.H{"followers": shopFollowers,
				"pagination": responses.Pagination{
					Limit: paginationArgs.Limit,
					Skip:  paginationArgs.Skip,
					Count: count,
				}})
	}
}

// IsfollowingShop - api/shops/:shopid/followers/is-following
func IsFollowingShop() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Invalid shop or user ID")
			return
		}

		filter := bson.M{"user_id": myId, "shop_id": shopId}
		var follower models.ShopFollower
		err = ShopFollowerCollection.FindOne(ctx, filter).Decode(&follower)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				helper.HandleSuccess(c, http.StatusOK, "Success", gin.H{"is_following": false})
				return
			}

			helper.HandleError(c, http.StatusInternalServerError, err, "Error retrieving shop follower")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "Success", gin.H{"is_following": true})

	}
}

// UnfollowShop - api/shops/:shopid/followers
func UnfollowShop() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Invalid shop or user ID")
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
			filter := bson.M{"shop_id": shopId, "user_id": myId}
			_, err := ShopFollowerCollection.DeleteOne(ctx, filter)
			if err != nil {
				return nil, err
			}

			// attempt to remove member from embedded field in shop
			filter = bson.M{"_id": shopId}
			update := bson.M{"$pull": bson.M{"followers": bson.M{"user_id": myId}}, "$inc": bson.M{"follower_count": -1}}
			result2, err := ShopCollection.UpdateOne(ctx, filter, update)
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
			helper.HandleError(c, http.StatusBadRequest, err, "Failed to leave shop")
			return
		}

		if err := session.CommitTransaction(context.Background()); err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Failed to commit transaction")
			return
		}

		session.EndSession(context.Background())

		helper.HandleSuccess(c, http.StatusOK, "Left shop successfully", nil)
	}
}

// RemoveOtherFollower - api/shops/:shopid/followers/other?userid={user_id to remove}
func RemoveOtherFollower() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		userToBeRemoved := c.Query("userid")
		defer cancel()

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Invalid shop or follower ID")
			return
		}

		// Let's verify shop ownership before attempting to remove follower
		ownershipEerr := VerifyShopOwnership(ctx, myId, shopId)
		if ownershipEerr != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "No Authorised PermissionD")
			return
		}

		userToBeRemovedId, err := primitive.ObjectIDFromHex(userToBeRemoved)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Invalid user ID")
			return
		}

		if shopId == userToBeRemovedId {
			helper.HandleError(c, http.StatusBadRequest, err, "No Authorised PermissionD")
			return
		}

		// Shop follower session
		wc := writeconcern.New(writeconcern.WMajority())
		txnOptions := options.Transaction().SetWriteConcern(wc)

		session, err := configs.DB.StartSession()
		if err != nil {
			panic(err)
		}
		defer session.EndSession(ctx)

		callback := func(ctx mongo.SessionContext) (interface{}, error) {
			// attempt to remove follower from shop follower collection table
			filter := bson.M{"shop_id": shopId, "user_id": userToBeRemovedId}
			_, err := ShopFollowerCollection.DeleteOne(ctx, filter)
			if err != nil {
				return nil, err
			}

			// attempt to remove follower from embedded field in shop
			filter = bson.M{"_id": shopId}
			update := bson.M{"$pull": bson.M{"followers": bson.M{"user_id": userToBeRemovedId}}, "$inc": bson.M{"follower_count": -1}}
			result2, err := ShopCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				return nil, err
			}

			return result2, nil
		}

		_, err = session.WithTransaction(context.Background(), callback, txnOptions)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Failed to remove follower")
			return
		}

		if err := session.CommitTransaction(context.Background()); err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Failed to commit transaction")
			return
		}

		session.EndSession(context.Background())

		helper.HandleSuccess(c, http.StatusOK, "Follower removed successfully", nil)
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
			helper.HandleError(c, http.StatusBadRequest, err, "Invalid shop or member ID")
			return
		}
		loginName, _, err := configs.ExtractTokenLoginNameEmail(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Failed to extract token login name")
			return
		}

		err = c.BindJSON(&shopReviewJson)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Failed to bind JSON")
			return
		}

		// validate request body
		if validationErr := Validate.Struct(&shopReviewJson); validationErr != nil {
			helper.HandleError(c, http.StatusUnprocessableEntity, validationErr, "Validation error")
			return
		}

		// Shop Member session
		wc := writeconcern.New(writeconcern.WMajority())
		txnOptions := options.Transaction().SetWriteConcern(wc)

		session, err := configs.DB.StartSession()
		if err != nil {
			helper.HandleError(c, http.StatusUnprocessableEntity, err, "Unable to start new session")
			return
		}
		defer session.EndSession(ctx)

		callback := func(ctx mongo.SessionContext) (interface{}, error) {
			var userProfile models.User
			err := UserCollection.FindOne(ctx, bson.M{"_id": myId}).Decode(&userProfile)
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
			_, err = ShopReviewCollection.InsertOne(ctx, shopReviewData)
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
			update := bson.M{"$push": bson.M{"recent_reviews": bson.M{"$each": bson.A{embedded}, "$sort": -1, "$slice": -5}}, "$set": bson.M{"modified_at": now}, "$inc": bson.M{"review_counts": 1}}
			result2, err := ShopCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				return nil, err
			}

			return result2, nil
		}

		_, err = session.WithTransaction(context.Background(), callback, txnOptions)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Transaction error")
			return
		}
		if err := session.CommitTransaction(context.Background()); err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Failed to commit transaction")
			return
		}
		session.EndSession(context.Background())

		helper.HandleSuccess(c, http.StatusOK, "Shop creation successfuls", nil)
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
			helper.HandleError(c, http.StatusBadRequest, err, "Invalid shop ID")
			return
		}

		paginationArgs := services.GetPaginationArgs(c)
		filter := bson.M{"shop_id": shopObjectID}
		find := options.Find().SetLimit(int64(paginationArgs.Limit)).SetSkip(int64(paginationArgs.Skip))
		result, err := ShopReviewCollection.Find(ctx, filter, find)
		if err != nil {
			helper.HandleError(c, http.StatusNotFound, err, "Failed to retrieve shop reviews")
			return
		}

		var shopReviews []models.ShopReview
		if err = result.All(ctx, &shopReviews); err != nil {
			helper.HandleError(c, http.StatusNotFound, err, "Failed to decode shop reviews")
			return
		}

		count, err := ShopReviewCollection.CountDocuments(ctx, bson.M{"shop_id": shopObjectID})
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Failed to count shop reviews")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "success", gin.H{
			"reviews": map[string]interface{}{"data": shopReviews},
			"pagination": responses.Pagination{
				Limit: paginationArgs.Limit,
				Skip:  paginationArgs.Skip,
				Count: count,
			},
		})
	}
}

// DeleteMyReview - api/shops/:shopid/members
func DeleteMyReview() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Error retrieving shop ID and user ID")
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
			// Attempt to remove review from review collection table
			filter := bson.M{"shop_id": shopId, "user_id": myId}
			_, err := ShopReviewCollection.DeleteOne(ctx, filter)
			if err != nil {
				return nil, err
			}

			// Attempt to remove member from embedded field in shop
			filter = bson.M{"_id": shopId}
			update := bson.M{"$pull": bson.M{"recent_reviews": bson.M{"user_id": myId}}, "$inc": bson.M{"review_counts": -1}}
			result2, err := ShopCollection.UpdateOne(ctx, filter, update)
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
			helper.HandleError(c, http.StatusBadRequest, err, "Error deleting review")
			return
		}
		if err := session.CommitTransaction(context.Background()); err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Error committing transaction")
			return
		}
		session.EndSession(context.Background())

		helper.HandleSuccess(c, http.StatusOK, "My review was deleted successfully", nil)
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
			helper.HandleError(c, http.StatusBadRequest, err, "Invalid user ID")
			return
		}

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Error retrieving shop ID and user ID")
			return
		}

		err = VerifyShopOwnership(c, myId, shopId)
		if err != nil {
			log.Printf("You don't have write access to this shop: %s\n", err.Error())
			helper.HandleError(c, http.StatusUnauthorized, err, "You don't have write access to this shop")
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
			// Attempt to remove review from review collection table
			filter := bson.M{"shop_id": shopId, "owner_id": myId, "user_id": userToBeRemovedId}
			_, err = ShopFollowerCollection.DeleteOne(ctx, filter)
			if err != nil {
				return nil, err
			}

			// Attempt to remove review from recent review field in shop
			filter = bson.M{"_id": shopId}
			update := bson.M{"$pull": bson.M{"recent_reviews": bson.M{"user_id": userToBeRemovedId}}}
			result2, err := ShopCollection.UpdateOne(ctx, filter, update)
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
			helper.HandleError(c, http.StatusNotFound, err, "Error deleting review")
			return
		}
		if err := session.CommitTransaction(context.Background()); err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Error committing transaction")
			return
		}
		session.EndSession(context.Background())

		helper.HandleSuccess(c, http.StatusOK, "Other user review deleted successfully", nil)
	}
}

// CreateShopAbout - api/shops/:shopid/about
func CreateShopAbout() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Error retrieving shop ID and user ID")
			return
		}

		err = VerifyShopOwnership(c, myId, shopId)
		if err != nil {
			log.Printf("You don't have write access to this shop: %s\n", err.Error())
			helper.HandleError(c, http.StatusUnauthorized, err, "You don't have write access to this shop")
			return
		}

		// Create shop about profile.
		shopAboutData := models.ShopAbout{
			ID:        primitive.NewObjectID(),
			ShopID:    shopId,
			Status:    models.ShopAboutStatusDraft,
			Headline:  "Welcome to My Shop!",
			Story:     "Thank you for visiting our online artisan shop. We are passionate about craftsmanship and dedicated to providing unique, handcrafted items that reflect the creativity and skill of our artisans. Explore our collection and discover the beauty of handmade products that carry a story of craftsmanship and tradition.\n\nAt [My Shop], we believe in the art of creating something special. Each piece in our collection is carefully crafted with attention to detail and a commitment to quality. We aim to connect artisans with appreciative buyers, creating a community that values and supports the artistry behind every creation.\n\nJoin us on this journey of celebrating craftsmanship and supporting talented artisans from around the world. Your purchase not only adds a unique piece to your life but also contributes to the livelihood of skilled individuals who pour their heart and soul into their work.\n\nThank you for being a part of our community. Happy shopping!",
			Instagram: "@MyShop",
			Facebook:  "@MyShop",
			X:         "@MyShop",
		}

		opts := options.Update().SetUpsert(true)
		_, err = ShopAboutCollection.UpdateOne(ctx, bson.M{"shop_id": shopId}, bson.M{"$set": shopAboutData}, opts)
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Error creating shop about")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "Shop about created successfully", nil)
	}
}

// GetShopAbout - api/shops/:shopid/about
func GetShopAbout() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var shopAbout models.ShopAbout

		shopId := c.Param("shopid")
		shopObjectID, err := primitive.ObjectIDFromHex(shopId)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Error parsing shop ID")
			return
		}

		err = ShopAboutCollection.FindOne(ctx, bson.M{"shop_id": shopObjectID}).Decode(&shopAbout)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				helper.HandleError(c, http.StatusNotFound, err, "no document in result")
				return
			}

			helper.HandleError(c, http.StatusInternalServerError, err, "Error retrieving shop about")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "success", gin.H{
			"about": shopAbout,
		})
	}
}

// UpdateShopAbout - api/shops/:shopid/about
func UpdateShopAbout() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var shopAboutJson models.ShopAboutRequest

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Error getting shop ID")
			return
		}

		err = VerifyShopOwnership(c, myId, shopId)
		if err != nil {
			log.Printf("You don't have write access to this shop: %s\n", err.Error())
			helper.HandleError(c, http.StatusUnauthorized, err, "You don't have write access to this shop")
			return
		}

		// bind the request body
		if err := c.BindJSON(&shopAboutJson); err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Error parsing request body")
			return
		}

		// validate request body
		if validationErr := Validate.Struct(&shopAboutJson); validationErr != nil {
			helper.HandleError(c, http.StatusUnprocessableEntity, validationErr, "Validation error")
			return
		}

		filter := bson.M{"shop_id": shopId}
		update := bson.M{
			"$set": bson.M{
				"status":    shopAboutJson.Status,
				"headline":  shopAboutJson.Headline,
				"story":     shopAboutJson.Story,
				"instagram": shopAboutJson.Instagram,
				"facebook":  shopAboutJson.Facebook,
				"x":         shopAboutJson.X,
			},
		}

		res, err := ShopAboutCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Error updating shop about")
			return
		}
		if res.ModifiedCount == 0 {
			helper.HandleSuccess(c, http.StatusNotFound, "success", "No matching documents found")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "Shop about updated successfully", nil)
	}
}

// UpdateShopAboutStatus - api/shops/:shopid/about/status?status=active
func UpdateShopAboutStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		status := c.Query("status")
		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Error getting shop ID")
			return
		}

		err = VerifyShopOwnership(c, myId, shopId)
		if err != nil {
			log.Printf("You don't have write access to this shop: %s\n", err.Error())
			helper.HandleError(c, http.StatusUnauthorized, err, "You don't have write access to this shop")
			return
		}

		if status != "active" && status != "draft" {
			helper.HandleError(c, http.StatusBadRequest, errors.New("status parameter is required and must be either 'active' or 'draft'"), "Invalid status value")
			return
		}

		filter := bson.M{"shop_id": shopId}
		update := bson.M{"$set": bson.M{"status": status}}
		res, err := ShopAboutCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Error updating shop about status")
			return
		}
		if res.ModifiedCount == 0 {
			helper.HandleSuccess(c, http.StatusNotFound, "success", "No matching documents found")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "Shop about status updated successfully", nil)
	}
}

// CreateShopReturnPolicy - api/shops/:shopid/policies
func CreateShopReturnPolicy() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		var shopReturnPolicyJson models.ShopReturnPolicies
		defer cancel()

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Error getting shop ID and user ID")
			return
		}

		if err := c.BindJSON(&shopReturnPolicyJson); err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Error binding JSON")
			return
		}

		if validationErr := Validate.Struct(&shopReturnPolicyJson); validationErr != nil {
			helper.HandleError(c, http.StatusBadRequest, validationErr, "Validation error")
			return
		}

		err = VerifyShopOwnership(c, myId, shopId)
		if err != nil {
			log.Printf("You don't have write access to this shop: %s\n", err.Error())
			helper.HandleError(c, http.StatusUnauthorized, err, "You don't have write access to this shop")
			return
		}

		shopReturnPolicyJson.ID = primitive.NewObjectID()
		shopReturnPolicyJson.ShopId = shopId

		_, err = ShopReturnPolicyCollection.InsertOne(ctx, shopReturnPolicyJson)
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Error creating shop policy")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "Shop policy created successfully", nil)
	}
}

// UpdateShopReturnPolicy - api/shops/:shopid/policies
func UpdateShopReturnPolicy() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		var shopReturnPolicyJson models.ShopReturnPolicies
		defer cancel()

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Error getting shop ID and user ID")
			return
		}

		if err := c.BindJSON(&shopReturnPolicyJson); err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Error binding JSON")
			return
		}

		if validationErr := Validate.Struct(&shopReturnPolicyJson); validationErr != nil {
			helper.HandleError(c, http.StatusBadRequest, validationErr, "Validation error")
			return
		}

		err = VerifyShopOwnership(c, myId, shopId)
		if err != nil {
			log.Printf("You don't have write access to this shop: %s\n", err.Error())
			helper.HandleError(c, http.StatusUnauthorized, err, "You don't have write access to this shop")
			return
		}

		filter := bson.M{"shop_id": shopId}
		update := bson.M{"$set": bson.M{"accepts_return": shopReturnPolicyJson.AcceptsReturn, "accepts_echanges": shopReturnPolicyJson.AcceptsExchanges, "deadline": shopReturnPolicyJson.Deadline}}
		res, err := ShopReturnPolicyCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Error updating shop policy")
			return
		}

		if res.ModifiedCount == 0 {
			helper.HandleError(c, http.StatusNotFound, nil, "No matching documents found")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "Shop policy updated successfully", nil)
	}
}

// DeleteShopReturnPolicy - api/shops/:shopid/policies?id={policy_id}
func DeleteShopReturnPolicy() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		policyIdStr := c.Query("id")
		policyId, err := primitive.ObjectIDFromHex(policyIdStr)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Invalid policy ID")
			return
		}

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Error getting shop ID and user ID")
			return
		}

		err = VerifyShopOwnership(c, myId, shopId)
		if err != nil {
			log.Printf("You don't have write access to this shop: %s\n", err.Error())
			helper.HandleError(c, http.StatusUnauthorized, err, "You don't have write access to this shop")
			return
		}

		filter := bson.M{"_id": policyId, "shop_id": shopId}
		_, err = ShopReturnPolicyCollection.DeleteOne(ctx, filter)
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Error deleting shop policy")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "Shop policy deleted successfully", nil)
	}
}

// GetShopReturnPolicy - api/shops/:shopid/policies?id={policy_id}
func GetShopReturnPolicy() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		policyIdStr := c.Query("id")
		policyId, err := primitive.ObjectIDFromHex(policyIdStr)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Invalid policy ID")
			return
		}

		shopId, _, err := services.MyShopIdAndMyId(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Error getting shop ID and user ID")
			return
		}

		var currentPolicy models.ShopReturnPolicies
		filter := bson.M{"_id": policyId, "shop_id": shopId}
		err = ShopReturnPolicyCollection.FindOne(ctx, filter).Decode(&currentPolicy)
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Error retrieving shop policy")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "success", gin.H{"policy": currentPolicy})
	}
}

// GetShopReturnPolicies - api/shops/:shopid/policies/all
func GetShopReturnPolicies() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shopId, _, err := services.MyShopIdAndMyId(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Error getting shop ID and user ID")
			return
		}

		// Query the database for shops that match the search query
		cursor, err := ShopReturnPolicyCollection.Find(ctx, bson.M{"shop_id": shopId})
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Error searching for shops")
			return
		}
		defer func() {
			if err := cursor.Close(ctx); err != nil {
				log.Println("Failed to close cursor:", err)
			}
		}()

		// Serialize the shop policies and return them to the client
		var policies []models.ShopReturnPolicies
		for cursor.Next(ctx) {
			var policy models.ShopReturnPolicies
			if err := cursor.Decode(&policy); err != nil {
				helper.HandleError(c, http.StatusInternalServerError, err, "Error decoding shops")
				return
			}
			policies = append(policies, policy)
		}

		helper.HandleSuccess(c, http.StatusOK, "success", gin.H{"policies": policies})
	}
}

// CreateShopCompliance - api/shops/:shopid/compliance
func CreateShopComplianceInformation() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		var complianceJson models.ComplianceInformationRequest
		defer cancel()

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Error getting shop ID and user ID")
			return
		}

		if err := c.BindJSON(&complianceJson); err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Error binding JSON")
			return
		}

		if validationErr := Validate.Struct(&complianceJson); validationErr != nil {
			helper.HandleError(c, http.StatusBadRequest, validationErr, "Validation error")
			return
		}

		err = VerifyShopOwnership(c, myId, shopId)
		if err != nil {
			log.Printf("Error verifying if you the shop owner: %s\n", err.Error())
			helper.HandleError(c, http.StatusUnauthorized, err, "shop ownership validation error")
			return
		}

		complianceInformation := models.ComplianceInformation{
			ID:                   primitive.NewObjectID(),
			ShopID:               shopId,
			TermsOfUse:           complianceJson.TermsOfUse,
			IntellectualProperty: complianceJson.IntellectualProperty,
			SellerPolicie:        complianceJson.SellerPolicie,
		}

		_, err = ShopCompliancePolicyCollection.InsertOne(ctx, complianceInformation)
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Error creating shop compliance policy")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "Shop compliance policy created successfully", nil)
	}
}

// GetShopComplianceInformation - api/shops/:shopid/compliance
func GetShopComplianceInformation() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shopId, _, err := services.MyShopIdAndMyId(c)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Error getting shop ID and user ID")
			return
		}

		var complianceInformation models.ComplianceInformation

		err = ShopCompliancePolicyCollection.FindOne(ctx, bson.M{"shop_id": shopId}).Decode(&complianceInformation)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				helper.HandleError(c, http.StatusNotFound, err, "Shop compliance information not found")
				return
			}
			helper.HandleError(c, http.StatusInternalServerError, err, "Error retrieving shop compliance information")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "Shop compliance information created successfully", gin.H{"compliance_information": complianceInformation})
	}
}
