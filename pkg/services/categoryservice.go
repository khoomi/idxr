package services

import (
	"context"
	"strings"

	"khoomi-api-io/api/pkg/models"
	"khoomi-api-io/api/pkg/util"

	slug2 "github.com/gosimple/slug"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CategoryService struct {
	categoryCollection *mongo.Collection
}

func NewCategoryService() *CategoryService {
	return &CategoryService{
		categoryCollection: util.GetCollection(util.DB, "ListingCategory"),
	}
}

func (s *CategoryService) CreateCategory(ctx context.Context, category models.Category) (*mongo.InsertOneResult, error) {
	return s.categoryCollection.InsertOne(ctx, category)
}

func (s *CategoryService) CreateManyCategories(ctx mongo.SessionContext, categories []interface{}) (*mongo.InsertManyResult, error) {
	return s.categoryCollection.InsertMany(ctx, categories)
}

func (s *CategoryService) GetAllCategories(ctx context.Context) ([]*models.Category, error) {
	var categories []*models.Category
	find := options.Find().SetSort(bson.M{"path": 1})
	result, err := s.categoryCollection.Find(ctx, bson.D{}, find)
	if err != nil {
		return nil, err
	}

	if err = result.All(ctx, &categories); err != nil {
		return nil, err
	}

	return categories, nil
}

func (s *CategoryService) GetCategoryChildren(ctx context.Context, categoryID string) (*models.Category, error) {
	category, err := s.GetCategoryByID(ctx, categoryID)
	if err != nil {
		return nil, err
	}

	var children []*models.Category
	filter := bson.M{"parent_id": categoryID}
	result, err := s.categoryCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if err = result.All(ctx, &children); err != nil {
		return nil, err
	}

	category.Children = children

	return category, nil
}

func (s *CategoryService) GetCategoryByID(ctx context.Context, categoryID string) (*models.Category, error) {
	var category models.Category
	err := s.categoryCollection.FindOne(ctx, bson.M{"_id": categoryID}).Decode(&category)
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (s *CategoryService) SearchCategories(ctx context.Context, search string) ([]*models.Category, error) {
	var categories []*models.Category
	result, err := s.categoryCollection.Find(ctx, bson.M{
		"$or": []bson.M{
			{"name": bson.M{"$regex": primitive.Regex{Pattern: search, Options: "i"}}},
			{"path": bson.M{"$regex": primitive.Regex{Pattern: search, Options: "i"}}},
			{"parent": bson.M{"$regex": primitive.Regex{Pattern: search, Options: "i"}}},
		},
	}, options.Find())
	if err != nil {
		return nil, err
	}

	for result.Next(ctx) {
		var category models.Category
		if err := result.Decode(&category); err != nil {
			return nil, err
		}
		categories = append(categories, &category)
	}

	return categories, nil
}

func (s *CategoryService) DeleteAllCategories(ctx context.Context) (*mongo.DeleteResult, error) {
	return s.categoryCollection.DeleteMany(ctx, bson.M{})
}

func (s *CategoryService) UpdateCategoryImages(ctx context.Context, categoryID string, imageURL, bannerURL string) error {
	updateDoc := bson.M{}
	if imageURL != "" {
		updateDoc["image_url"] = imageURL
	}
	if bannerURL != "" {
		updateDoc["banner_url"] = bannerURL
	}

	if len(updateDoc) == 0 {
		return nil
	}

	_, err := s.categoryCollection.UpdateOne(ctx, bson.M{"_id": categoryID}, bson.M{"$set": updateDoc})
	return err
}

func (s *CategoryService) ProcessCategoryForMultiCreate(category models.CategoryRequest) models.Category {
	return models.Category{
		ID:          slug2.Make(strings.ToLower(strings.Replace(category.Name, "'", "", 5))),
		Name:        category.Name,
		Description: category.Description,
		Path:        category.Path,
		ParentID:    slug2.Make(strings.ToLower(strings.Replace(category.ParentID, "'", "", 5))),
		ImageURL:    category.ImageURL,
		BannerURL:   category.BannerURL,
	}
}

func (s *CategoryService) BuildCategoryTree(categories []*models.Category) []*models.Category {
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
