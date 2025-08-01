package controllers

import (
	"context"
	"log"
	"net/http"
	"strings"

	"khoomi-api-io/api/internal/common"
	"khoomi-api-io/api/pkg/models"
	"khoomi-api-io/api/pkg/services"
	"khoomi-api-io/api/pkg/util"

	"github.com/gin-gonic/gin"
	slug2 "github.com/gosimple/slug"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

type CategoryController struct {
	categoryService *services.CategoryService
}

func InitCategoryController(categoryService *services.CategoryService) *CategoryController {
	return &CategoryController{
		categoryService: categoryService,
	}
}

func (cc *CategoryController) CreateCategorySingle(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	var categoryJson models.Category
	if err := c.ShouldBindJSON(&categoryJson); err != nil {
		util.HandleError(c, http.StatusUnprocessableEntity, err)
		return
	}
	if validationErr := common.Validate.Struct(&categoryJson); validationErr != nil {
		util.HandleError(c, http.StatusUnprocessableEntity, validationErr)
		return
	}

	res, err := cc.categoryService.CreateCategory(ctx, categoryJson)
	if err != nil {
		util.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	util.HandleSuccess(c, http.StatusOK, "Category created", res.InsertedID)
}

func (cc *CategoryController) CreateCategoryMulti(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	var categoryJson models.CategoryRequestMulti

	if err := c.ShouldBindJSON(&categoryJson); err != nil {
		util.HandleError(c, http.StatusUnprocessableEntity, err)
		return
	}

	if validationErr := common.Validate.Struct(&categoryJson); validationErr != nil {
		util.HandleError(c, http.StatusUnprocessableEntity, validationErr)
		return
	}

	wc := writeconcern.New(writeconcern.WMajority())
	txnOptions := options.Transaction().SetWriteConcern(wc)
	session, err := util.DB.StartSession()
	if err != nil {
		util.HandleError(c, http.StatusInternalServerError, err)
		return
	}
	defer session.EndSession(ctx)
	callback := func(ctx mongo.SessionContext) (any, error) {
		var categories []interface{}
		for _, category := range categoryJson.Categories {
			categories = append(categories, cc.categoryService.ProcessCategoryForMultiCreate(category))
		}
		res, err := cc.categoryService.CreateManyCategories(ctx, categories)
		if err != nil {
			return nil, err
		}

		return res, nil
	}

	res, err := session.WithTransaction(context.Background(), callback, txnOptions)
	if err != nil {
		util.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	if err := session.CommitTransaction(context.Background()); err != nil {
		util.HandleError(c, http.StatusInternalServerError, err)
		return
	}
	session.EndSession(context.Background())

	util.HandleSuccess(c, http.StatusOK, "All categories created successfully.", res)
}

func (cc *CategoryController) GetAllCategories(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	categories, err := cc.categoryService.GetAllCategories(ctx)
	if err != nil {
		log.Println(err)
		util.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	root := cc.categoryService.BuildCategoryTree(categories)
	util.HandleSuccess(c, http.StatusOK, "Categories retrieved successfully", root)
}

func (cc *CategoryController) GetCategoryChildren(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	categoryID := c.Param("id")

	category, err := cc.categoryService.GetCategoryChildren(ctx, categoryID)
	if err != nil {
		log.Println(err)
		util.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	util.HandleSuccess(c, http.StatusOK, "Category with children retrieved successfully", category)
}

func (cc *CategoryController) GetCategoryAncestor(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	categoryID := c.Param("id")

	// Find the category with the given ID
	category, err := cc.categoryService.GetCategoryByID(ctx, categoryID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			util.HandleError(c, http.StatusNotFound, err)
			return
		}
		util.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	// Traverse up the category tree to find ancestors
	var ancestors []*models.Category
	for category.ParentID != "" {
		// Find the parent category
		parent, err := cc.categoryService.GetCategoryByID(ctx, category.ParentID)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				util.HandleError(c, http.StatusNotFound, err)
				return
			}
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		// Add the parent category to the ancestors list
		ancestors = append([]*models.Category{parent}, ancestors...)

		// Traverse up to the parent category
		*category = *parent
	}

	root := cc.categoryService.BuildCategoryTree(ancestors)
	util.HandleSuccess(c, http.StatusOK, "Category ancestors retrieved successfully", root)
}

func (cc *CategoryController) SearchCategories(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	search := c.Query("s")

	categories, err := cc.categoryService.SearchCategories(ctx, search)
	if err != nil {
		util.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	root := cc.categoryService.BuildCategoryTree(categories)
	util.HandleSuccess(c, http.StatusOK, "Categories found", root)
}

func (cc *CategoryController) DeleteAllCategories(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	res, err := cc.categoryService.DeleteAllCategories(ctx)
	if err != nil {
		util.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	util.HandleSuccess(c, http.StatusOK, "Categories deleted successfully", res)
}

func (cc *CategoryController) CreateCategoryWithImages(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	name := c.PostForm("name")
	description := c.PostForm("description")
	path := c.PostForm("path")
	parentID := c.PostForm("parentId")

	if name == "" {
		util.HandleError(c, http.StatusBadRequest, errors.New("name is required"))
		return
	}

	var imageURL, bannerURL string
	var err error

	imageFile, _, err := c.Request.FormFile("image")
	if err == nil && imageFile != nil {
		defer imageFile.Close()
		imageUpload, err := util.FileUpload(models.File{File: imageFile})
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}
		imageURL = imageUpload.SecureURL
	}

	bannerFile, _, err := c.Request.FormFile("banner")
	if err == nil && bannerFile != nil {
		defer bannerFile.Close()
		bannerUpload, err := util.FileUpload(models.File{File: bannerFile})
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}
		bannerURL = bannerUpload.SecureURL
	}

	category := models.Category{
		ID:          slug2.Make(strings.ToLower(strings.Replace(name, "'", "", 5))),
		Name:        name,
		Description: description,
		Path:        path,
		ParentID:    parentID,
		ImageURL:    imageURL,
		BannerURL:   bannerURL,
	}

	if validationErr := common.Validate.Struct(&category); validationErr != nil {
		util.HandleError(c, http.StatusUnprocessableEntity, validationErr)
		return
	}

	res, err := cc.categoryService.CreateCategory(ctx, category)
	if err != nil {
		util.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	util.HandleSuccess(c, http.StatusOK, "Category created with images", res.InsertedID)
}

func (cc *CategoryController) UpdateCategoryImage(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	categoryID := c.Param("id")
	if categoryID == "" {
		util.HandleError(c, http.StatusBadRequest, errors.New("category id is required"))
		return
	}

	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	var imageURL string
	var err error

	imageFile, _, err := c.Request.FormFile("image")
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, errors.New("image file is required"))
		return
	}
	defer imageFile.Close()

	imageUpload, err := util.FileUpload(models.File{File: imageFile})
	if err != nil {
		util.HandleError(c, http.StatusInternalServerError, err)
		return
	}
	imageURL = imageUpload.SecureURL

	err = cc.categoryService.UpdateCategoryImages(ctx, categoryID, imageURL, "")
	if err != nil {
		util.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	util.HandleSuccess(c, http.StatusOK, "Category image updated successfully", map[string]string{"imageUrl": imageURL})
}

func (cc *CategoryController) UpdateCategoryBanner(c *gin.Context) {
	ctx, cancel := WithTimeout()
	defer cancel()

	categoryID := c.Param("id")
	if categoryID == "" {
		util.HandleError(c, http.StatusBadRequest, errors.New("category id is required"))
		return
	}

	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return
	}

	var bannerURL string
	var err error

	bannerFile, _, err := c.Request.FormFile("banner")
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, errors.New("banner file is required"))
		return
	}
	defer bannerFile.Close()

	bannerUpload, err := util.FileUpload(models.File{File: bannerFile})
	if err != nil {
		util.HandleError(c, http.StatusInternalServerError, err)
		return
	}
	bannerURL = bannerUpload.SecureURL

	err = cc.categoryService.UpdateCategoryImages(ctx, categoryID, "", bannerURL)
	if err != nil {
		util.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	util.HandleSuccess(c, http.StatusOK, "Category banner updated successfully", map[string]string{"bannerUrl": bannerURL})
}
