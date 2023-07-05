package configs

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectDB() (client *mongo.Client) {
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

	log.Println("Connected to MongoDB")
	return
}

// DB client instance
var DB = ConnectDB()

// GetCollection Get collection from Db
func GetCollection(client *mongo.Client, name string) (collection *mongo.Collection) {
	collection = client.Database("khoomi").Collection(name)
	return
}

func ConnectRedis() *redis.Client {
	// Connect to Redis
	addr, err := redis.ParseURL(LoadEnvFor("REDIS_URL"))
	if err != nil {
		log.Fatal(err)
	}

	client := redis.NewClient(addr)

	log.Println("Connected to Redis")
	return client
}

var REDIS = ConnectRedis()
