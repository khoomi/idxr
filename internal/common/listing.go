package common

import (
	rand2 "crypto/rand"
	"fmt"
	"khoomi-api-io/api/pkg/models"
	"khoomi-api-io/api/pkg/util"
	"math/big"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/cloudinary/cloudinary-go/api/uploader"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	// Title constraints
	MinTitleLength = 5
	MaxTitleLength = 140

	// Description constraints
	MinDescriptionLength = 25
	MaxDescriptionLength = 2000
)

// getFormValue returns the value from formData or default if empty/missing
func getFormValue(formData func(key string) (value string), key, defaultValue string) string {
	if value := formData(key); value != "" {
		return value
	}
	return defaultValue
}

func validateTitle(title string) error {
	length := utf8.RuneCountInString(strings.TrimSpace(title))
	if length < MinTitleLength {
		return fmt.Errorf("title is too short: minimum length is %d characters", MinTitleLength)
	}
	if length > MaxTitleLength {
		return fmt.Errorf("title is too long: maximum length is %d characters", MaxTitleLength)
	}
	return nil
}

// validateDescription checks if the description meets the length requirements
func validateDescription(description string) error {
	length := utf8.RuneCountInString(strings.TrimSpace(description))
	if length < MinDescriptionLength {
		return fmt.Errorf("description is too short: minimum length is %d characters", MinDescriptionLength)
	}
	if length > MaxDescriptionLength {
		return fmt.Errorf("description is too long: maximum length is %d characters", MaxDescriptionLength)
	}
	return nil
}

func GenerateListingCode() string {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	letterChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numberChars := "0123456789"
	letters := make([]byte, 4)
	for i := range letters {
		letters[i] = letterChars[rand.Intn(len(letterChars))]
	}

	numbers := make([]byte, 4)
	for i := range numbers {
		numbers[i] = numberChars[rand.Intn(len(numberChars))]
	}

	productCode := string(letters) + "-" + string(numbers)
	return productCode
}

func GetListingSortingBson(sort string) bson.D {
	value := -1
	var key string

	switch sort {
	case "created_at_asc":
		key = "date.created_at"
	case "created_at_desc":
		key = "date.created_at"
	case "modified_at_asc":
		key = "date.modified_at"
	case "modified_at_desc":
		key = "date.modified_at"
	case "state_updated_at_asc":
		key = "state.updated_at"
	case "state_updated_at_desc":
		key = "state.updated_at"
	case "views_asc":
		key = "views"
	case "views_desc":
		key = "views"
	case "sales_asc":
		key = "financial_information.sales"
	case "sales_desc":
		key = "financial_information.sales"
	case "price_asc":
		key = "inventory.price"
	case "price_desc":
		key = "inventory.price"
	case "rating_desc":
		key = "rating.rating.positive_reviews"
	default:
		key = "date.created_at"
	}

	if strings.Contains(sort, "asc") {
		value = 1
	}
	return bson.D{{Key: key, Value: value}}
}

func parseDimensionUnit(unit string) models.DimensionUnit {
	switch strings.ToLower(unit) {
	case string(models.DimensionUnitINC):
		return models.DimensionUnitINC
	case string(models.DimensionUnitFT):
		return models.DimensionUnitFT
	case string(models.DimensionUnitMM):
		return models.DimensionUnitMM
	case string(models.DimensionUnitCM):
		return models.DimensionUnitCM
	case string(models.DimensionUnitM):
		return models.DimensionUnitM
	default:
		return models.DimensionUnitCM
	}
}

func parseWeightUnit(unit string) models.WeightUnit {
	switch strings.ToLower(unit) {
	case string(models.WeightUnitG):
		return models.WeightUnitG
	case string(models.WeightUnitKG):
		return models.WeightUnitKG
	case string(models.WeightUnitLB):
		return models.WeightUnitLB
	case string(models.WeightUnitOZ):
		return models.WeightUnitOZ
	default:
		return models.WeightUnitKG
	}
}

func MapFormDataToNewListing(formData func(key string) (value string)) (*models.NewListing, error) {
	listing := &models.NewListing{}

	// Map Inventory
	price, err := strconv.ParseFloat(formData("inventory.price"), 64)
	if err != nil {
		return nil, err
	}
	quantity, err := strconv.Atoi(formData("inventory.quantity"))
	if err != nil {
		return nil, err
	}

	sku := formData("inventory.sku")
	if sku == "" {
		newSku := NewSKUGenerator()
		sku = newSku.GenerateSKU()
	}

	listing.Inventory = models.Inventory{
		Price:           price,
		Quantity:        quantity,
		SKU:             formData("inventory.sku"),
		DomesticPricing: false,
		DomesticPrice:   0,
	}

	// Map Measurements
	itemWeight, err := strconv.ParseFloat(formData("measurements.itemWeight"), 64)
	if err != nil {
		itemWeight = 0
	}
	itemLength, err := strconv.ParseFloat(formData("measurements.itemLength"), 64)
	if err != nil {
		itemLength = 0
	}
	itemHeight, err := strconv.ParseFloat(formData("measurements.itemHeight"), 64)
	if err != nil {
		itemHeight = 0
	}
	itemWidth, err := strconv.ParseFloat(formData("measurements.itemWidth"), 64)
	if err != nil {
		itemWidth = 0
	}

	itemWeightUnit := parseWeightUnit(formData("measurements.itemWeightUnit"))
	itemDimensionUnit := parseDimensionUnit(formData("measurements.itemDimensionUnit"))
	listing.Measurements = models.ListingMeasurement{
		ItemWeight:        itemWeight,
		ItemWeightUnit:    itemWeightUnit,
		ItemLength:        itemLength,
		ItemDimensionUnit: itemDimensionUnit,
		ItemHeight:        itemHeight,
		ItemWidth:         itemWidth,
	}

	var tags []string
	if formData("details.tags") != "" {
		tags = strings.Split(formData("details.tags"), ",")
	}

	personalization := formData("details.personalization") == "on"
	var personalizationTextChars int
	if formData("details.personalizationTextChars") != "" {
		personalizationTextChars, _ = strconv.Atoi(formData("details.personalizationTextChars"))
	}

	title := strings.TrimSpace(formData("details.title"))
	if err := validateTitle(title); err != nil {
		return nil, err
	}

	description := strings.TrimSpace(formData("details.description"))
	if err := validateDescription(description); err != nil {
		return nil, err
	}

	listing.ListingDetails = models.NewListingDetails{
		Title: title,
		Category: models.ListingCategory{
			CategoryId:   formData("details.category.categoryId"),
			CategoryName: formData("details.category.categoryName"),
			CategoryPath: formData("details.category.categoryPath"),
		},
		Condition:                getFormValue(formData, "details.condition", "new"),
		Sustainability:           getFormValue(formData, "details.sustainability", "eco-friendly"),
		Description:              description,
		Tags:                     tags,
		WhoMade:                  formData("details.whoMade"),
		WhenMade:                 formData("details.whenMade"),
		Type:                     formData("details.type"),
		Color:                    formData("details.mainColor"),
		OtherColor:               formData("details.otherColor"),
		Personalization:          personalization,
		PersonalizationText:      formData("details.personalizationText"),
		PersonalizationTextChars: personalizationTextChars,
	}

	return listing, nil
}

const (
	MaxFileSize = 10 << 20
	ImageCount  = 5
)

func HandleSequentialImages(c *gin.Context) ([]string, []uploader.UploadResult, error) {
	var (
		uploadedImagesUrl    []string
		uploadedImagesResult []uploader.UploadResult
		wg                   sync.WaitGroup
		mu                   sync.Mutex
		errs                 []error
	)

	for i := 1; i <= ImageCount; i++ {
		wg.Add(1)
		go func(imageNum int) {
			defer wg.Done()

			// Get file from form
			fileField := fmt.Sprintf("image%d", imageNum)
			file, _, err := c.Request.FormFile(fileField)
			if err != nil {
				if err != http.ErrMissingFile {
					mu.Lock()
					errs = append(errs, fmt.Errorf("error getting %s: %w", fileField, err))
					mu.Unlock()
				}
				return
			}
			defer file.Close()

			// Upload the file
			imageUpload, err := util.FileUpload(models.File{File: file})
			if err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("failed to upload %s: %w", fileField, err))
				mu.Unlock()
				return
			}

			// Append results in a thread-safe manner
			mu.Lock()
			uploadedImagesUrl = append(uploadedImagesUrl, imageUpload.SecureURL)
			uploadedImagesResult = append(uploadedImagesResult, imageUpload)
			mu.Unlock()
		}(i)
	}

	// Wait for all uploads to complete
	wg.Wait()

	// Check if any errors occurred
	if len(errs) > 0 {
		// Combine error messages
		errMsg := "Failed to upload some images:"
		for _, err := range errs {
			errMsg += "\n" + err.Error()
		}
		return uploadedImagesUrl, uploadedImagesResult, fmt.Errorf(errMsg)
	}

	var defaultThumbnail = "https://res.cloudinary.com/kh-oo-mi/image/upload/v1705607175/khoomi/mypvl86lihcqvkcqmvbg.jpg"
	// If no images were uploaded, use default thumbnail
	if len(uploadedImagesUrl) == 0 {
		tempImage := uploader.UploadResult{
			SecureURL: defaultThumbnail,
		}
		uploadedImagesUrl = append(uploadedImagesUrl, defaultThumbnail)
		uploadedImagesResult = append(uploadedImagesResult, tempImage)
	}

	return uploadedImagesUrl, uploadedImagesResult, nil
}

type SKUGenerator struct {
	Prefix     string
	Separator  string
	DigitCount int
}

// NewSKUGenerator creates a new SKU generator with default settings
func NewSKUGenerator() *SKUGenerator {
	return &SKUGenerator{
		Prefix:     "",
		Separator:  "-",
		DigitCount: 6,
	}
}

// GenerateSKU generates a new SKU with optional prefix
func (g *SKUGenerator) GenerateSKU() string {
	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(g.DigitCount)), nil)
	n, _ := rand2.Int(rand2.Reader, max)
	number := fmt.Sprintf("%0*d", g.DigitCount, n.Int64())

	if g.Prefix != "" {
		return fmt.Sprintf("%s%s%s", g.Prefix, g.Separator, number)
	}
	return number
}

// GenerateTimedSKU generates a SKU including timestamp components
func (g *SKUGenerator) GenerateTimedSKU() string {
	now := time.Now()
	timeComponent := fmt.Sprintf("%02d%02d", now.Month(), now.Day())

	// Generate random suffix
	suffix, _ := rand2.Int(rand2.Reader, big.NewInt(9999))

	if g.Prefix != "" {
		return fmt.Sprintf("%s%s%s%s%04d",
			g.Prefix, g.Separator,
			timeComponent, g.Separator,
			suffix.Int64())
	}
	return fmt.Sprintf("%s%s%04d", timeComponent, g.Separator, suffix.Int64())
}

// GeneratePatternSKU generates a SKU based on a pattern
func (g *SKUGenerator) GeneratePatternSKU(pattern string) string {
	result := strings.Builder{}

	for _, char := range pattern {
		switch char {
		case 'A': // Random uppercase letter
			result.WriteRune(rune('A' + randInt(26)))
		case 'N': // Random number
			result.WriteRune(rune('0' + randInt(10)))
		case '-': // Separator
			result.WriteRune('-')
		default: // Keep any other character as is
			result.WriteRune(char)
		}
	}

	return result.String()
}

// Helper function to generate random integer
func randInt(max int) int {
	n, _ := rand2.Int(rand2.Reader, big.NewInt(int64(max)))
	return int(n.Int64())
}
