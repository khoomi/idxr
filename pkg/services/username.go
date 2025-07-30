package services

import (
	"fmt"
	"math/rand"
	"time"
)

// GenerateRandomUsername generates a random username using adjectives and nouns
func GenerateRandomUsername() string {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

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

	adjective := adjectives[r.Intn(len(adjectives))]
	noun := nouns[r.Intn(len(nouns))]

	number := r.Intn(900) + 100

	username := fmt.Sprintf("%s%s%d", adjective, noun, number)

	return username
}