package common

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"time"

	"khoomi-api-io/api/pkg/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetUserById(ctx context.Context, id primitive.ObjectID) (models.User, error) {
	var user models.User
	err := UserCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

func GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	var user models.User
	err := UserCollection.FindOne(ctx, bson.M{"primary_email": email}).Decode(&user)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

// IsSeller checks if the specified user is a seller in the database. It returns true if the user is a seller,
// and false otherwise, along with an error in case of a database access issue.
func IsSeller(c *gin.Context, userId primitive.ObjectID) (bool, error) {
	err := UserCollection.FindOne(c, bson.M{"_id": userId, "is_seller": true}).Err()
	if err == mongo.ErrNoDocuments {
		// User not found or not a seller
		return false, nil
	} else if err != nil {
		// Other error occurred
		return false, err
	}

	// User is a seller
	return true, nil
}

func GenerateRandomUsername() string {
	// Create a private random generator with a seeded source
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

	// List of adjectives and nouns
	adjectives := []string{
		"fluffy", "sunny", "breezy", "whisper", "dazzle", "sparkle", "mystic", "shimmer",
		"twinkle", "dreamy", "enchant", "radiant", "brave", "vibrant", "gloomy", "chilly",
		"gentle", "witty", "fierce", "graceful", "dashing", "dapper", "elegant", "quirky",
		"clever", "cheerful", "joyful", "lively", "charming", "silly", "jovial", "playful",
	}

	nouns := []string{
		"cat", "sun", "wind", "whisper", "glitter", "moon", "star", "wave", "glimmer", "rainbow",
		"cloud", "butterfly", "mountain", "river", "ocean", "tree", "flower", "bird", "song",
		"dream", "adventure", "journey", "fantasy", "harmony", "paradise", "magic", "serenity",
		"wonder", "delight", "treasure", "triumph", "inspiration", "smile", "laughter",
	}

	// Randomly select an adjective and noun
	adjective := adjectives[r.Intn(len(adjectives))]
	noun := nouns[r.Intn(len(nouns))]

	// Generate a random number between 100 and 999
	number := r.Intn(900) + 100

	// Combine the adjective, noun, and number to form the username
	username := fmt.Sprintf("%s%s%d", adjective, noun, number)

	return username
}

// validateNameFormat checks if the provided name follows the required naming rule.
func ValidateNameFormat(name string) error {
	validName, err := regexp.MatchString("([A-Z][a-zA-Z]*)", name)
	if err != nil {
		return err
	}
	if !validName {
		return errors.New("name should follow the naming rule")
	}
	return nil
}
