package services

import (
	"context"
	"fmt"
	"khoomi-api-io/khoomi_api/configs"
	"khoomi-api-io/khoomi_api/models"
	"khoomi-api-io/khoomi_api/responses"
	"math/rand"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var userCollection = configs.GetCollection(configs.DB, "User")

func GetUserById(ctx context.Context, id primitive.ObjectID) (models.User, error) {
	var user models.User
	err := userCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

func GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	var user models.User
	err := userCollection.FindOne(ctx, bson.M{"primary_email": email}).Decode(&user)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

func GetPaginationArgs(c *gin.Context) responses.PaginationArgs {

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	skip, _ := strconv.Atoi(c.DefaultQuery("skip", "0"))

	return responses.PaginationArgs{
		Limit: limit,
		Skip:  skip,
	}
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