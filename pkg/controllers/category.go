package controllers

import (
	"context"
	"log"
	"net/http"
	"strings"

	"khoomi-api-io/api/internal/common"
	"khoomi-api-io/api/pkg/models"
	"khoomi-api-io/api/pkg/util"

	"github.com/gin-gonic/gin"
	slug2 "github.com/gosimple/slug"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

var ListingCategoryCollection = util.GetCollection(util.DB, "ListingCategory")

func CreateCategorySingle() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		var categoryJson models.Category

		// Bind and validate the request body
		if err := c.ShouldBindJSON(&categoryJson); err != nil {
			util.HandleError(c, http.StatusUnprocessableEntity, err)
			return
		}

		// Validate the request body
		if validationErr := common.Validate.Struct(&categoryJson); validationErr != nil {
			util.HandleError(c, http.StatusUnprocessableEntity, validationErr)
			return
		}

		res, err := ListingCategoryCollection.InsertOne(ctx, categoryJson)
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Category created", res.InsertedID)
	}
}

func CreateCategoryMulti() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		var categoryJson models.CategoryRequestMulti

		// Bind and validate the request body
		if err := c.ShouldBindJSON(&categoryJson); err != nil {
			util.HandleError(c, http.StatusUnprocessableEntity, err)
			return
		}

		// Validate the request body
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
			var categories []any
			for _, category := range categoryJson.Categories {
				categories = append(categories, models.Category{
					ID:          slug2.Make(strings.ToLower(strings.Replace(category.Name, "'", "", 5))),
					Name:        category.Name,
					Description: category.Description,
					Path:        category.Path,
					ParentID:    slug2.Make(strings.ToLower(strings.Replace(category.ParentID, "'", "", 5))),
				})
			}
			// Create many categories
			res, err := ListingCategoryCollection.InsertMany(ctx, categories)
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
}

// GetAllCategories - /api/categories?path=jewelry
func GetAllCategories() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		var categories []*models.Category

		find := options.Find().SetSort(bson.M{"path": 1})
		result, err := ListingCategoryCollection.Find(ctx, bson.D{}, find)
		if err != nil {
			log.Println(err)
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		if err = result.All(ctx, &categories); err != nil {
			log.Println(err)
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		root := BuildCategoryTree(categories)
		util.HandleSuccess(c, http.StatusOK, "Categories retrieved successfully",
			root)
	}
}

// GetCategoryChildren - /api/categories?path=jewelry
func GetCategoryChildren() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		categoryID := c.Param("id")

		// find all the children of the category
		var children []*models.Category
		filter := bson.M{"parent_id": categoryID}
		result, err := ListingCategoryCollection.Find(ctx, filter)
		if err != nil {
			log.Println(err)
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}
		if err = result.All(ctx, &children); err != nil {
			log.Println(err)
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Category children retrieved successfully", children)
	}
}

// GetCategoryAncestor - /api/categories?path=jewelry
func GetCategoryAncestor() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		categoryID := c.Param("id")

		// Find the category with the given ID
		var category models.Category
		err := ListingCategoryCollection.FindOne(ctx, bson.M{"_id": categoryID}).Decode(&category)
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
			// If this category has no parent, it's the root category
			// Find the parent category
			var parent models.Category
			err = ListingCategoryCollection.FindOne(ctx, bson.M{"_id": category.ParentID}).Decode(&parent)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					util.HandleError(c, http.StatusNotFound, err)
					return
				}
				util.HandleError(c, http.StatusInternalServerError, err)
				return
			}

			// Add the parent category to the ancestors list
			ancestors = append([]*models.Category{&parent}, ancestors...)

			// Traverse up to the parent category
			category = parent
		}

		root := BuildCategoryTree(ancestors)
		util.HandleSuccess(c, http.StatusOK, "Category ancestors retrieved successfully", root)
	}
}

// SearchCategories - /api/categories?s=jewelry
func SearchCategories() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		search := c.Query("s")

		// Query the database for categories that match the search query
		result, err := ListingCategoryCollection.Find(ctx, bson.M{
			"$or": []bson.M{
				{"name": bson.M{"$regex": primitive.Regex{Pattern: search, Options: "i"}}},
				{"path": bson.M{"$regex": primitive.Regex{Pattern: search, Options: "i"}}},
				{"parent": bson.M{"$regex": primitive.Regex{Pattern: search, Options: "i"}}},
			},
		}, options.Find())
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		// Serialize the categories and return them to the client
		var serializedCategories []*models.Category
		for result.Next(ctx) {
			var category models.Category
			if err := result.Decode(&category); err != nil {
				util.HandleError(c, http.StatusInternalServerError, err)
				return
			}
			serializedCategories = append(serializedCategories, &category)
		}

		root := BuildCategoryTree(serializedCategories)
		util.HandleSuccess(c, http.StatusOK, "Categories found", root)
	}
}

func DeleteAllCategories() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), common.REQ_TIMEOUT_SECS)
		defer cancel()

		res, err := ListingCategoryCollection.DeleteMany(ctx, bson.M{})
		if err != nil {
			util.HandleError(c, http.StatusInternalServerError, err)
			return
		}

		util.HandleSuccess(c, http.StatusOK, "Categories deleted successfully", res)
	}
}

func BuildCategoryTree(categories []*models.Category) []*models.Category {
	categoryMap := make(map[string]*models.Category)
	var rootCategories []*models.Category

	for _, category := range categories {
		if category.ParentID == "" {
			rootCategories = append(rootCategories, category)
		} else {
			parentCategory, ok := categoryMap[category.ParentID]
			if !ok {
				parentCategory = &models.Category{ID: category.ParentID, Children: []*models.Category{}}
				categoryMap[category.ParentID] = parentCategory
			}
			parentCategory.Children = append(parentCategory.Children, category)
		}
		categoryMap[category.ID] = category
	}

	return rootCategories
}
