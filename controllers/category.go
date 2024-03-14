package controllers

import (
	"context"
	configs "khoomi-api-io/khoomi_api/config"
	"khoomi-api-io/khoomi_api/helper"
	"khoomi-api-io/khoomi_api/models"
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

var ListingCategoryCollection = configs.GetCollection(configs.DB, "ListingCategory")

func CreateCategorySingle() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var categoryJson models.Category

		// Bind and validate the request body
		if err := c.ShouldBindJSON(&categoryJson); err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Invalid request body")
		}

		// Validate the request body
		if validationErr := Validate.Struct(&categoryJson); validationErr != nil {
			helper.HandleError(c, http.StatusUnprocessableEntity, validationErr, "Validation error")
		}

		_, err := ListingCategoryCollection.InsertOne(ctx, categoryJson)
		if err != nil {
			helper.HandleError(c, http.StatusUnprocessableEntity, err, "Error creating category")
			return
		}

		helper.HandleSuccess(c, http.StatusOK, "Category created", "Category created.")
	}
}

func CreateCategoryMulti() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var categoryJson models.CategoryRequestMulti

		// Bind and validate the request body
		if err := c.ShouldBindJSON(&categoryJson); err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Invalid request body")
		}

		// Validate the request body
		if validationErr := Validate.Struct(&categoryJson); validationErr != nil {
			helper.HandleError(c, http.StatusUnprocessableEntity, validationErr, "Validation error")
		}

		wc := writeconcern.New(writeconcern.WMajority())
		txnOptions := options.Transaction().SetWriteConcern(wc)
		session, err := configs.DB.StartSession()
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Failed to start database session")
		}
		defer session.EndSession(ctx)
		callback := func(ctx mongo.SessionContext) (interface{}, error) {
			var categories []interface{}
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

		_, err = session.WithTransaction(context.Background(), callback, txnOptions)
		if err != nil {
			helper.HandleError(c, http.StatusBadRequest, err, "Error creating categories")
		}

		if err := session.CommitTransaction(context.Background()); err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Failed to commit transaction")
		}
		session.EndSession(context.Background())

		helper.HandleSuccess(c, http.StatusOK, "All categories created successfully.", "")
	}

}

// GetAllCategories - /api/categories?path=jewelry
func GetAllCategories() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var categories []*models.Category

		find := options.Find().SetSort(bson.M{"path": 1})
		result, err := ListingCategoryCollection.Find(ctx, bson.D{}, find)
		if err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Failed to retrieve categories")
		}

		if err = result.All(ctx, &categories); err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Failed to decode categories")
		}

		root := BuildCategoryTree(categories)
		helper.HandleSuccess(c, http.StatusOK, "Categories retrieved successfully", gin.H{
			"categories": root,
		})
	}
}

// GetCategoryChildren - /api/categories?path=jewelry
func GetCategoryChildren() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		categoryID := c.Param("id")

		// find all the children of the category
		var children []*models.Category
		filter := bson.M{"parent_id": categoryID}
		result, err := ListingCategoryCollection.Find(ctx, filter)
		if err != nil {
			helper.HandleError(c, http.StatusNotFound, err, "Failed to retrieve category children")
		}
		if err = result.All(ctx, &children); err != nil {
			helper.HandleError(c, http.StatusInternalServerError, err, "Failed to decode category children")
		}

		helper.HandleSuccess(c, http.StatusOK, "Category children retrieved successfully", gin.H{
			"categories": children,
		})
	}
}

// GetCategoryAncestor - /api/categories?path=jewelry
func GetCategoryAncestor() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		categoryID := c.Param("id")

		// Find the category with the given ID
		var category models.Category
		err := ListingCategoryCollection.FindOne(ctx, bson.M{"_id": categoryID}).Decode(&category)
		if err != nil {
			helper.HandleError(c, http.StatusNotFound, err, "Failed to retrieve category")
		}

		// Traverse up the category tree to find ancestors
		var ancestors []*models.Category
		for {
			// If this category has no parent, it's the root category
			if category.ParentID == "" {
				break
			}

			// Find the parent category
			var parent models.Category
			err = ListingCategoryCollection.FindOne(ctx, bson.M{"_id": category.ParentID}).Decode(&parent)
			if err != nil {
				helper.HandleError(c, http.StatusNotFound, err, "Failed to retrieve category parent")
			}

			// Add the parent category to the ancestors list
			ancestors = append([]*models.Category{&parent}, ancestors...)

			// Traverse up to the parent category
			category = parent
		}

		root := BuildCategoryTree(ancestors)
		helper.HandleSuccess(c, http.StatusOK, "Category ancestors retrieved successfully", gin.H{
			"categories": root,
		})
	}
}

// SearchCategories - /api/categories?s=jewelry
func SearchCategories() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
			helper.HandleError(c, http.StatusInternalServerError, err, "Error searching for categories")
		}

		// Serialize the categories and return them to the client
		var serializedCategories []*models.Category
		for result.Next(ctx) {
			var category models.Category
			if err := result.Decode(&category); err != nil {
				helper.HandleError(c, http.StatusInternalServerError, err, "Error decoding categories")
			}
			serializedCategories = append(serializedCategories, &category)
		}

		root := BuildCategoryTree(serializedCategories)
		helper.HandleSuccess(c, http.StatusOK, "Categories found", gin.H{
			"categories": root,
		})
	}
}

func DeleteAllCategories() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		res, err := ListingCategoryCollection.DeleteMany(ctx, bson.M{})
		if err != nil {
			helper.HandleError(c, http.StatusNotFound, err, "Failed to delete categories")
		}

		helper.HandleSuccess(c, http.StatusOK, "Categories deleted successfully", res)
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
