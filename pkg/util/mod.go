package util

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Initialize env vars
func LoadEnvFor(v string) (x string) {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Unable to load .env file")
	}

	x = os.Getenv(v)
	return
}

// Initialize db connection
func ConnectDB() (client *mongo.Client) {
	log.Println("starting MongoDB connection..")
	client, err := mongo.NewClient(options.Client().ApplyURI(LoadEnvFor("DATABASE_URL")))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// try to ping the database
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("MongoDB connection successful")
	return
}

// DB client instance
var DB = ConnectDB()

// GetCollection Get collection from Db
func GetCollection(client *mongo.Client, name string) (collection *mongo.Collection) {
	collection = client.Database("khoomi").Collection(name)
	return
}

// Initialize redis connection
func ConnectRedis() *redis.Client {
	// Connect to Redis
	log.Println("starting redis connection..")
	addr, err := redis.ParseURL(LoadEnvFor("REDIS_URL"))
	if err != nil {
		log.Fatal(err)
	}

	client := redis.NewClient(addr)

	log.Println("redis connection successful..")
	return client
}

var REDIS = ConnectRedis()
