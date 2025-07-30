package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"khoomi-api-io/api/internal"
	auth "khoomi-api-io/api/internal/auth"
	"khoomi-api-io/api/internal/common"
	"khoomi-api-io/api/pkg/models"
	"khoomi-api-io/api/pkg/services"
	"khoomi-api-io/api/pkg/util"

	"github.com/cloudinary/cloudinary-go/api/uploader"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ListingController struct {
	listingService      services.ListingService
	shopService         services.ShopService
	notificationService services.NotificationService
}

// InitListingController initializes a new ListingController with dependencies
func InitListingController(listingService services.ListingService, shopService services.ShopService, notificationService services.NotificationService) *ListingController {
	return &ListingController{
		listingService:      listingService,
		shopService:         shopService,
		notificationService: notificationService,
	}
}

func (lc *ListingController) CreateListing(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	shopId, myId, err := common.MyShopIdAndMyId(c)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	session, err := auth.GetSessionAuto(c)
	if err != nil {
		util.HandleError(c, http.StatusUnauthorized, err)
		return
	}

	loginName, loginEmail := session.LoginName, session.Email
	listingJson := c.PostForm("listing")
	if listingJson == "" {
		util.HandleError(c, http.StatusBadRequest, errors.New("missing listing payload"))
		return
	}

	var newListing models.NewListing
	if err := json.Unmarshal([]byte(listingJson), &newListing); err != nil {
		util.HandleError(c, http.StatusBadRequest, fmt.Errorf("invalid JSON: %v", err))
		return
	}

	if validationErr := common.Validate.Struct(newListing); validationErr != nil {
		util.HandleError(c, http.StatusBadRequest, validationErr)
		return
	}

	mainImage, _, err := c.Request.FormFile("mainImage")
	var mainImageUploadUrl uploader.UploadResult
	if err == nil {
		mainImageUploadUrl, err = util.FileUpload(models.File{File: mainImage})
		if err != nil {
			errMsg := fmt.Sprintf("Logo failed to upload - %v", err.Error())
			util.HandleError(c, http.StatusInternalServerError, errors.New(errMsg))
			return
		}
	} else {
		mainImageUploadUrl = uploader.UploadResult{}
		mainImageUploadUrl.SecureURL = common.DEFAULT_THUMBNAIL
	}

	uploadedImagesUrl, uploadedImagesResult, err := common.HandleSequentialImages(c)
	if err != nil {
		util.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	req := services.CreateListingRequest{
		UserID:          myId,
		ShopID:          shopId,
		LoginName:       loginName,
		LoginEmail:      loginEmail,
		NewListing:      newListing,
		MainImageURL:    mainImageUploadUrl.SecureURL,
		ImagesURLs:      uploadedImagesUrl,
		ImagesResults:   make([]any, len(uploadedImagesResult)),
		MainImageResult: mainImageUploadUrl,
	}

	for i, result := range uploadedImagesResult {
		req.ImagesResults[i] = result
	}

	listingID, err := lc.listingService.CreateListing(ctx, req)
	if err != nil {
		if mainImageUploadUrl.PublicID != "" {
			util.DestroyMedia(mainImageUploadUrl.PublicID)
		}
		for _, url := range uploadedImagesResult {
			util.DestroyMedia(url.PublicID)
		}

		util.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	internal.PublishCacheMessage(c, internal.CacheInvalidateShopListings, shopId.Hex())

	util.HandleSuccess(c, http.StatusOK, "Listing was created successfully", listingID)
}

func (lc *ListingController) GetListing(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	listingId := c.Param("listingid")
	listing, err := lc.listingService.GetListing(ctx, listingId)
	if err != nil {
		if err.Error() == "no listing found" {
			util.HandleError(c, http.StatusNotFound, err)
		} else {
			util.HandleError(c, http.StatusInternalServerError, err)
		}
		return
	}

	util.HandleSuccess(c, http.StatusOK, "Success", listing)
}

func (lc *ListingController) GetListings(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	paginationArgs := common.GetPaginationArgs(c)
	filters := lc.listingService.GetListingFilters(c)
	sort := lc.listingService.GetListingSortingBson(paginationArgs.Sort)

	listings, count, err := lc.listingService.GetListings(ctx, paginationArgs, filters, sort)
	if err != nil {
		util.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	util.HandleSuccessMeta(c, http.StatusOK, "success", listings, gin.H{
		"pagination": util.Pagination{
			Limit: paginationArgs.Limit,
			Skip:  paginationArgs.Skip,
			Count: count,
		},
	})
}

func (lc *ListingController) GetMyListingsSummary(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	shopId, myId, err := common.MyShopIdAndMyId(c)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	paginationArgs := common.GetPaginationArgs(c)
	sort := lc.listingService.GetListingSortingBson(paginationArgs.Sort)

	listings, count, err := lc.listingService.GetMyListingsSummary(ctx, shopId, myId, paginationArgs, sort)
	if err != nil {
		util.HandleError(c, http.StatusNotFound, err)
		return
	}

	util.HandleSuccessMeta(c, http.StatusOK, "success", listings, gin.H{
		"pagination": util.Pagination{
			Limit: paginationArgs.Limit,
			Skip:  paginationArgs.Skip,
			Count: count,
		},
	})
}

// GetShopListings - Get single shop listings.
func (lc *ListingController) GetShopListings(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	shopId := c.Param("shopid")
	shopObjectId, err := primitive.ObjectIDFromHex(shopId)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	filters := lc.listingService.GetListingFilters(c)
	paginationArgs := common.GetPaginationArgs(c)
	sort := lc.listingService.GetListingSortingBson(paginationArgs.Sort)

	listings, count, err := lc.listingService.GetShopListings(ctx, shopObjectId, paginationArgs, filters, sort)
	if err != nil {
		util.HandleError(c, http.StatusNotFound, err)
		return
	}

	util.HandleSuccessMeta(c, http.StatusOK, "success", listings, gin.H{
		"pagination": util.Pagination{
			Limit: paginationArgs.Limit,
			Skip:  paginationArgs.Skip,
			Count: count,
		},
	})
}

func (lc *ListingController) HasUserCreatedListingOnboarding(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	_, userId, err := common.MyShopIdAndMyId(c)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	hasListing, err := lc.listingService.HasUserCreatedListing(ctx, userId)
	if err != nil {
		log.Printf("error retrieving user listing: %v", err)
		util.HandleError(c, http.StatusNotFound, err)
		return
	}

	if !hasListing {
		util.HandleError(c, http.StatusNotFound, errors.New("User has no listings"))
		return
	}

	util.HandleSuccess(c, http.StatusOK, "Success", hasListing)
}

func (lc *ListingController) DeleteListings(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	shopId, myId, err := common.MyShopIdAndMyId(c)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	listingIDs := c.QueryArray("ids")
	if len(listingIDs) < 1 {
		util.HandleError(c, http.StatusBadRequest, errors.New("no listing IDs provided"))
		return
	}

	var objectIDs []primitive.ObjectID
	for _, id := range listingIDs {
		objectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			util.HandleError(c, http.StatusBadRequest, fmt.Errorf("invalid listing ID: %s", id))
			return
		}
		objectIDs = append(objectIDs, objectID)
	}

	result, err := lc.listingService.DeleteListings(ctx, myId, shopId, objectIDs, services.NewReviewService())
	if err != nil {
		util.HandleError(c, http.StatusInternalServerError, fmt.Errorf("failed to delete listings: %v", err))
		return
	}

	internal.PublishCacheMessage(c, internal.CacheInvalidateListings, shopId.Hex())

	message := fmt.Sprintf("Deleted %d listings with %d reviews", len(result.DeletedListings), result.DeletedReviews)
	if len(result.NotDeletedListings) > 0 {
		message += fmt.Sprintf(", %d failed to delete", len(result.NotDeletedListings))
	}

	util.HandleSuccess(c, http.StatusOK, message, result)
}

func (lc *ListingController) ChangeListingState(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	shopId, myId, err := common.MyShopIdAndMyId(c)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	listingIDs := c.QueryArray("ids")
	if len(listingIDs) < 1 {
		util.HandleError(c, http.StatusBadRequest, errors.New("no listing IDs provided"))
		return
	}
	log.Println(listingIDs)

	newStatus := models.ListingStateType(c.Query("status"))

	result, err := lc.listingService.UpdateListingState(ctx, myId, listingIDs, newStatus)
	if err != nil {
		util.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	internal.PublishCacheMessage(c, internal.CacheInvalidateListings, shopId.Hex())

	util.HandleSuccess(c, http.StatusOK, "Listing(s) state update", gin.H{"updated": result.UpdatedListings, "not_updated": result.NotUpdatedListings})
}

func (lc *ListingController) UpdateListing(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	listingId := c.Param("listingid")
	listingObjectId, err := primitive.ObjectIDFromHex(listingId)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, fmt.Errorf("invalid listing ID: %s", listingId))
		return
	}

	session, err := auth.GetSessionAuto(c)
	if err != nil {
		util.HandleError(c, http.StatusUnauthorized, err)
		return
	}

	// Parse listing update data
	listingJson := c.PostForm("listing")
	if listingJson == "" {
		util.HandleError(c, http.StatusBadRequest, errors.New("missing listing update payload"))
		return
	}

	var updateListing models.UpdateListing
	if err := json.Unmarshal([]byte(listingJson), &updateListing); err != nil {
		util.HandleError(c, http.StatusBadRequest, fmt.Errorf("invalid JSON: %v", err))
		return
	}

	// Validate the update data
	if validationErr := common.Validate.Struct(updateListing); validationErr != nil {
		util.HandleError(c, http.StatusBadRequest, validationErr)
		return
	}

	// Handle main image
	keepMainImage := c.PostForm("keepMainImage") == "true"
	var newMainImageURL *string
	var mainImageResult uploader.UploadResult

	mainImage, _, err := c.Request.FormFile("mainImage")
	if err == nil {
		// Upload new main image
		mainImageResult, err = util.FileUpload(models.File{File: mainImage})
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, fmt.Errorf("failed to upload main image: %v", err))
			return
		}
		newMainImageURL = &mainImageResult.SecureURL
		keepMainImage = true // If new image uploaded, we're keeping a main image
	}

	// Handle new images
	newImagesURLs, newImagesResults, err := common.HandleSequentialImages(c)
	if err != nil {
		// Cleanup main image if already uploaded
		if mainImageResult.PublicID != "" {
			util.DestroyMedia(mainImageResult.PublicID)
		}
		util.HandleError(c, http.StatusInternalServerError, fmt.Errorf("failed to upload images: %v", err))
		return
	}

	// Get images to remove
	removeImages := c.PostFormArray("removeImages")

	// Get image order (optional)
	var imageOrder []string
	imageOrderJson := c.PostForm("imageOrder")
	if imageOrderJson != "" {
		if err := json.Unmarshal([]byte(imageOrderJson), &imageOrder); err != nil {
			// Cleanup uploaded images
			if mainImageResult.PublicID != "" {
				util.DestroyMedia(mainImageResult.PublicID)
			}
			for _, result := range newImagesResults {
				util.DestroyMedia(result.PublicID)
			}
			util.HandleError(c, http.StatusBadRequest, fmt.Errorf("invalid image order: %v", err))
			return
		}
	}

	// Get shop ID for the listing
	var listing struct {
		ShopID primitive.ObjectID `bson:"shop_id"`
	}
	err = common.ListingCollection.FindOne(ctx, bson.M{"_id": listingObjectId}).Decode(&listing)
	if err != nil {
		// Cleanup uploaded images
		if mainImageResult.PublicID != "" {
			util.DestroyMedia(mainImageResult.PublicID)
		}
		for _, result := range newImagesResults {
			util.DestroyMedia(result.PublicID)
		}
		util.HandleError(c, http.StatusNotFound, errors.New("listing not found"))
		return
	}

	// Build update request
	req := services.UpdateListingRequest{
		ListingID:       listingObjectId,
		UserID:          session.UserId,
		ShopID:          listing.ShopID,
		UpdatedListing:  updateListing,
		NewMainImageURL: newMainImageURL,
		KeepMainImage:   keepMainImage,
		ImagesToAdd:     newImagesURLs,
		ImagesToRemove:  removeImages,
		ImageOrder:      imageOrder,
		MainImageResult: mainImageResult,
		NewImageResults: make([]any, len(newImagesResults)),
	}

	// Convert upload results to []any
	for i, result := range newImagesResults {
		req.NewImageResults[i] = result
	}

	// Call service to update listing
	err = lc.listingService.UpdateListing(ctx, req)
	if err != nil {
		util.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	// Invalidate cache
	internal.PublishCacheMessage(c, internal.CacheInvalidateShopListings, listing.ShopID.Hex())
	internal.PublishCacheMessage(c, internal.CacheInvalidateListing, listingId)

	util.HandleSuccess(c, http.StatusOK, "Listing updated successfully", gin.H{"listing_id": listingObjectId})
}
