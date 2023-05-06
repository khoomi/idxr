package controllers

import (
	"context"
	"errors"
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
	"strconv"
	"time"
)

var shopCollection = configs.GetCollection(configs.DB, "Shop")
var shopMemberCollection = configs.GetCollection(configs.DB, "ShopMember")
var shopReviewCollection = configs.GetCollection(configs.DB, "ShopReview")

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
			Members:            []models.ShopMember{},
			Status:             models.ShopStatusPendingReview,
			CreatedAt:          now,
			ModifiedAt:         now,
			Policy:             policy,
			RecentReviews:      []models.ShopReview{},
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

// AddShopFavorer - api/shop/:shop/favorers?userId={userId}
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

// RemoveShopFavorer - api/shop/:shop/favorers?userId={userId}
func RemoveShopFavorer() gin.HandlerFunc {
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

// JoinShopMembers - api/shop/:shop/members
func JoinShopMembers() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		var shopMember models.ShopMemberFromRequest
		defer cancel()

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}
		loginName, err := auth.ExtractTokenLoginName(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
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
			c.JSON(http.StatusUnprocessableEntity, responses.UserResponse{Status: http.StatusUnprocessableEntity, Message: "error", Data: map[string]interface{}{"error": validationErr.Error(), "field": ""}})
			return
		}

		// Shop Member session
		wc := writeconcern.New(writeconcern.WMajority())
		txnOptions := options.Transaction().SetWriteConcern(wc)

		session, err := configs.DB.StartSession()
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, responses.UserResponse{Status: http.StatusUnprocessableEntity, Message: "error", Data: map[string]interface{}{"error": "Unable  to start new session", "field": ""}})
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
			update := bson.M{"$push": bson.M{"members": bson.M{"$each": bson.A{inner}, "$sort": -1, "$slice": -5}}, "$set": bson.M{"modified_at": time.Now()}}
			result2, err := shopCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				return nil, err
			}

			return result2, nil
		}

		res, err := session.WithTransaction(context.Background(), callback, txnOptions)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": res}})
	}
}

// GetShopMembers - api/shop/:shop/members?limit=50&skip=0&sort=created_at
func GetShopMembers() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shopId := c.Param("shopid")
		shopOBjectID, err := primitive.ObjectIDFromHex(shopId)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		sort := c.Query("sort")
		limit := c.Query("limit")
		limitInt, err := strconv.Atoi(limit)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": "Expected an integer for 'limit'", "field": "limit"}})
			return
		}

		skip := c.Query("skip")
		skipInt, err := strconv.Atoi(skip)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": "Expected an integer for 'skip'", "field": "skip"}})
			return
		}

		filter := bson.M{"shop_id": shopOBjectID}
		find := options.Find().SetLimit(int64(limitInt)).SetSkip(int64(skipInt)).SetSort(bson.M{sort: 1})
		result, err := shopMemberCollection.Find(ctx, filter, find)
		if err != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		var shopMembers []models.ShopMember
		if err = result.All(ctx, &shopMembers); err != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponsePagination{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": shopMembers}, Pagination: responses.Pagination{
			Limit: limitInt,
			Skip:  skipInt,
			Sort:  sort,
			Total: len(shopMembers),
		}})

	}
}

// LeaveShopMembers - api/shop/:shop/members
func LeaveShopMembers() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
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

			return result2, nil
		}

		res, err := session.WithTransaction(context.Background(), callback, txnOptions)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": res}})
	}
}

// RemoveOtherMember - api/shop/:shop/members/other?userid={user_id to remeove}
func RemoveOtherMember() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		userToBeRemoved := c.Query("userid")
		defer cancel()

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
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

			return result2, nil
		}

		res, err := session.WithTransaction(context.Background(), callback, txnOptions)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": res}})
	}
}

// CreateShopReview - api/shop/:shop/reviews
func CreateShopReview() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		var shopReviewJson models.ShopReviewRequest
		defer cancel()

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}
		loginName, err := auth.ExtractTokenLoginName(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		err = c.BindJSON(&shopReviewJson)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": "bind"}})
			return
		}

		// validate request body
		if validationErr := validate.Struct(&shopReviewJson); validationErr != nil {
			c.JSON(http.StatusUnprocessableEntity, responses.UserResponse{Status: http.StatusUnprocessableEntity, Message: "error", Data: map[string]interface{}{"error": validationErr.Error(), "field": ""}})
			return
		}

		// Shop Member section
		wc := writeconcern.New(writeconcern.WMajority())
		txnOptions := options.Transaction().SetWriteConcern(wc)

		session, err := configs.DB.StartSession()
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, responses.UserResponse{Status: http.StatusUnprocessableEntity, Message: "error", Data: map[string]interface{}{"error": "Unable  to start new session", "field": ""}})
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
				CreatedAt:    time.Now(),
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
			update := bson.M{"$push": bson.M{"recent_reviews": bson.M{"$each": bson.A{embedded}, "$sort": -1, "$slice": -5}}, "$set": bson.M{"modified_at": time.Now()}}
			result2, err := shopCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				return nil, err
			}

			return result2, nil
		}

		res, err := session.WithTransaction(context.Background(), callback, txnOptions)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": res}})
	}
}

// GetShopReviews - api/shop/:shop/reviews?limit=50&skip=0&sort=created_at
func GetShopReviews() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shopId := c.Param("shopid")
		shopOBjectID, err := primitive.ObjectIDFromHex(shopId)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		sort := c.Query("sort")
		limit := c.Query("limit")
		limitInt, err := strconv.Atoi(limit)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": "Expected an integer for 'limit'", "field": "limit"}})
			return
		}

		skip := c.Query("skip")
		skipInt, err := strconv.Atoi(skip)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": "Expected an integer for 'skip'", "field": "skip"}})
			return
		}

		filter := bson.M{"shop_id": shopOBjectID}
		find := options.Find().SetLimit(int64(limitInt)).SetSkip(int64(skipInt)).SetSort(bson.M{sort: 1})
		result, err := shopReviewCollection.Find(ctx, filter, find)
		if err != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		var shopReviews []models.ShopReview
		if err = result.All(ctx, &shopReviews); err != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponsePagination{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": shopReviews}, Pagination: responses.Pagination{
			Limit: limitInt,
			Skip:  skipInt,
			Sort:  sort,
			Total: len(shopReviews),
		}})

	}
}

// DeleteMyReview - api/shop/:shop/members
func DeleteMyReview() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
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

			return result2, nil
		}

		res, err := session.WithTransaction(context.Background(), callback, txnOptions)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": res}})
	}
}

// DeleteOtherReview - api/shop/:shop/reviews/other?userid={user_id to remove}
func DeleteOtherReview() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		userToBeRemoved := c.Query("userid")
		defer cancel()

		shopId, myId, err := services.MyShopIdAndMyId(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		userToBeRemovedId, err := primitive.ObjectIDFromHex(userToBeRemoved)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": "userid"}})
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

			return result2, nil
		}

		res, err := session.WithTransaction(context.Background(), callback, txnOptions)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error(), "field": ""}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": res}})
	}
}
