package controllers

import (
	"context"
	"github.com/gin-gonic/gin"
	slug2 "github.com/gosimple/slug"
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
	"strings"
	"time"
)

var listingCategoryCollection = configs.GetCollection(configs.DB, "ListingCategory")

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

		_, err := listingCategoryCollection.InsertOne(ctx, categoryJson)
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
					ID:          slug2.Make(strings.ToLower(strings.Replace(category.Name, "'", "", 5))),
					Name:        category.Name,
					Description: category.Description,
					Path:        category.Path,
					ParentID:    slug2.Make(strings.ToLower(strings.Replace(category.ParentID, "'", "", 5))),
				})
			}
			// create many categories
			res, err := listingCategoryCollection.InsertMany(ctx, categories)
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

		c.JSON(http.StatusOK, responses.UserResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": "All categories created successfully."}})
	}
}

// GetAllCategories - /api/categories?path=jewelry
func GetAllCategories() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		var categories []*models.Category
		defer cancel()

		find := options.Find().SetSort(bson.M{"path": 1})
		result, err := listingCategoryCollection.Find(ctx, bson.D{}, find)
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
		result, err := listingCategoryCollection.Find(ctx, filter)
		if err != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}
		if err = result.All(ctx, &children); err != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponsePagination{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": children}})
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
		err := listingCategoryCollection.FindOne(ctx, bson.M{"_id": categoryID}).Decode(&category)
		if err != nil {
			c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
			return
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
			err = listingCategoryCollection.FindOne(ctx, bson.M{"_id": category.ParentID}).Decode(&parent)
			if err != nil {
				c.JSON(http.StatusNotFound, responses.UserResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"error": err.Error()}})
				return
			}

			// Add the parent category to the ancestors list
			ancestors = append([]*models.Category{&parent}, ancestors...)

			// Traverse up to the parent category
			category = parent
		}

		root := BuildCategoryTree(ancestors)
		c.JSON(http.StatusOK, responses.UserResponsePagination{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": root}})
	}
}

// SearchCategories - /api/categories?s=jewelry
func SearchCategories() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		search := c.Query("s")

		// Query the database for catgories that match the search query
		result, err := listingCategoryCollection.Find(ctx, bson.M{
			"$or": []bson.M{
				{"name": bson.M{"$regex": primitive.Regex{Pattern: search, Options: "i"}}},
				{"path": bson.M{"$regex": primitive.Regex{Pattern: search, Options: "i"}}},
				{"parent": bson.M{"$regex": primitive.Regex{Pattern: search, Options: "i"}}},
			},
		}, options.Find())
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "Error searching for shops", Data: map[string]interface{}{"error": err.Error()}})
			return
		}

		// Serialize the categories and return them to the client
		var serializedCategories []*models.Category
		for result.Next(ctx) {
			var category models.Category
			if err := result.Decode(&category); err != nil {
				c.JSON(http.StatusInternalServerError, responses.UserResponse{Status: http.StatusInternalServerError, Message: "Error decoding shops", Data: map[string]interface{}{"error": err.Error()}})
				return
			}
			serializedCategories = append(serializedCategories, &category)
		}

		root := BuildCategoryTree(serializedCategories)
		c.JSON(http.StatusOK, responses.UserResponsePagination{Status: http.StatusOK, Message: "Shops found", Data: map[string]interface{}{"data": root}})
	}
}

func DeleteAllCategories() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		res, err := listingCategoryCollection.DeleteMany(ctx, bson.M{})
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
