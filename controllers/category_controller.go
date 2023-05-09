package controllers

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"khoomi-api-io/khoomi_api/configs"
	"khoomi-api-io/khoomi_api/models"
	"khoomi-api-io/khoomi_api/responses"
	"log"
	"net/http"
	"time"
)

var ListingCategoryCollection = configs.GetCollection(configs.DB, "ListingCategory")

func CreateCategorySingle() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		var categoryJson models.Category
		defer cancel()

		// Validate the request body
		if err := c.BindJSON(&categoryJson); err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		// Validate request body
		if validationErr := validate.Struct(&categoryJson); validationErr != nil {
			c.JSON(http.StatusUnprocessableEntity, responses.UserResponse{Status: http.StatusUnprocessableEntity, Message: "error", Data: map[string]interface{}{"error": validationErr.Error()}})
			return
		}

		_, err := ListingCategoryCollection.InsertOne(ctx, categoryJson)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, responses.UserResponse{Status: http.StatusUnprocessableEntity, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": "Category created."}})
	}
}

func CreateCategoryMulti() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		var categoryJson models.CategoryRequestMulti
		defer cancel()

		// Validate the request body
		if err := c.BindJSON(&categoryJson); err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		// Validate request body
		if validationErr := validate.Struct(&categoryJson); validationErr != nil {
			c.JSON(http.StatusUnprocessableEntity, responses.UserResponse{Status: http.StatusUnprocessableEntity, Message: "error", Data: map[string]interface{}{"error": validationErr.Error()}})
			return
		}

		wc := writeconcern.New(writeconcern.WMajority())
		txnOptions := options.Transaction().SetWriteConcern(wc)
		session, err := configs.DB.StartSession()
		if err != nil {
			defer func() {
				if r := recover(); r != nil {
					log.Println("Recovered from panic:", r)
				}
				c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"error": "Failed to start database session"}})
			}()
			panic("Failed to start database session: " + err.Error())
		}
		defer session.EndSession(ctx)
		callback := func(ctx mongo.SessionContext) (interface{}, error) {
			var categories []interface{}
			for _, category := range categoryJson.Categories {
				categories = append(categories, models.Category{
					ID:          category.Name,
					Name:        category.Name,
					Description: category.Description,
					Path:        category.Path,
					ParentID:    category.ParentID,
				})
			}
			// create many categories
			res, err := ListingCategoryCollection.InsertMany(ctx, categories)
			if err != nil {
				return nil, err
			}

			return res, nil
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

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": "All ategories created successfully."}})
	}
}

// GetAllCategories - /api/categories?path=jewelry
func GetAllCategories() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		var categories []*models.Category
		defer cancel()

		path := c.Query("path")
		if path == "" {
			find := options.Find().SetSort(bson.M{"path": 1})
			result, err := ListingCategoryCollection.Find(ctx, bson.D{}, find)
			if err != nil {
				c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
				return
			}
			if err = result.All(ctx, &categories); err != nil {
				c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
				return
			}

			root := BuildCategoryTree(categories)
			c.JSON(http.StatusOK, responses.UserResponsePagination{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": root}})
			return
		}

		pattern := fmt.Sprintf("^/%s/", path)
		result, err := ListingCategoryCollection.Find(ctx, bson.M{"path": bson.M{
			"$regex": primitive.Regex{Pattern: pattern},
		}})
		if err != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}
		if err = result.All(ctx, &categories); err != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		root := BuildCategoryTree(categories)
		c.JSON(http.StatusOK, responses.UserResponsePagination{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": root}})
	}
}

func DeleteAllCategories() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		res, err := ListingCategoryCollection.DeleteMany(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponsePagination{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": res}})
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
