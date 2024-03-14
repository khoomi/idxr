package configs

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func LoadEnvFor(v string) (x string) {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Unable to load .env file")
	}

	x = os.Getenv(v)
	return
}

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

type JWTClaim struct {
	Id        string `json:"id"`
	LoginName string `json:"login_name"`
	Email     string `json:"email"`
	IsSeller  bool   `json:"is_seller"`
	jwt.StandardClaims
}

func GenerateJWT(id, email, loginName string, seller bool) (string, int64, error) {
	expirationTime := time.Now().Add(15 * time.Minute)
	jwtKey := LoadEnvFor("SECRET")

	claims := JWTClaim{
		Id:        id,
		LoginName: loginName,
		Email:     email,
		IsSeller:  seller,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtKey))
	if err != nil {
		return "", 0, err
	}

	return tokenString, expirationTime.Unix(), nil
}

func GenerateRefreshJWT(id, email, loginName string, seller bool) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour * 7)
	jwtKey := LoadEnvFor("REFRESH_SECRET")

	claims := JWTClaim{
		Id:        id,
		LoginName: loginName,
		Email:     email,
		IsSeller:  seller,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ValidateToken(signedToken string) (err error) {
	jwtKey := LoadEnvFor("SECRET")
	token, err := jwt.ParseWithClaims(
		signedToken,
		&JWTClaim{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtKey), nil
		},
	)
	if err != nil {
		return
	}
	claims, ok := token.Claims.(*JWTClaim)
	if !ok {
		err = errors.New("couldn't parse claims")
		return
	}
	if claims.ExpiresAt < time.Now().Local().Unix() {
		err = errors.New("token expired")
		return
	}
	return
}

func ValidateRefreshToken(signedToken string) (claim JWTClaim, err error) {
	jwtKey := LoadEnvFor("REFRESH_SECRET")
	token, err := jwt.ParseWithClaims(
		signedToken,
		&JWTClaim{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtKey), nil
		},
	)
	if err != nil {
		return
	}
	claims, ok := token.Claims.(*JWTClaim)
	if !ok {
		err = errors.New("couldn't parse claims")
		return JWTClaim{}, err
	}
	if claims.ExpiresAt < time.Now().Local().Unix() {
		err = errors.New("token expired")
		return JWTClaim{}, err
	}

	return *claims, nil
}

func ExtractToken(context *gin.Context) string {
	tokenString := context.GetHeader("Authorization")
	return tokenString
}

func ExtractTokenID(c *gin.Context) (primitive.ObjectID, error) {
	tokenString := ExtractToken(c)
	jwtKey := LoadEnvFor("SECRET")
	token, err := jwt.ParseWithClaims(
		tokenString,
		&JWTClaim{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtKey), nil
		},
	)
	if err != nil {
		return primitive.NilObjectID, err
	}

	claims, ok := token.Claims.(*JWTClaim)
	if !ok {
		err = errors.New("couldn't parse claims")
		return primitive.NilObjectID, err
	}

	Id, err := primitive.ObjectIDFromHex(claims.Id)
	if err != nil {
		err = errors.New("invalid user id")
		return primitive.NilObjectID, err
	}

	return Id, nil
}

func ExtractTokenLoginNameEmail(c *gin.Context) (string, string, error) {
	tokenString := ExtractToken(c)
	jwtKey := LoadEnvFor("SECRET")
	token, err := jwt.ParseWithClaims(
		tokenString,
		&JWTClaim{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtKey), nil
		},
	)
	if err != nil {
		return "", "", err
	}

	claims, ok := token.Claims.(*JWTClaim)
	if !ok {
		err = errors.New("couldn't parse claims")
		return "", "", err
	}

	return claims.LoginName, claims.Email, nil
}

func ValidateUserID(c *gin.Context) (primitive.ObjectID, error) {
	myID, err := ExtractTokenID(c)
	if err != nil {
		errMsg := fmt.Sprintf("unauthorized: User ID not found in authentication token - %v", err.Error())
		log.Println(errMsg)
		return primitive.NilObjectID, errors.New(errMsg)
	}

	userID := c.Param("userId")
	if userID != myID.Hex() {
		errMsg := fmt.Sprintln("unauthorized: User ID in the URL path doesn't match with currently authenticated user")
		log.Println(errMsg)
		return primitive.NilObjectID, errors.New(errMsg)
	}

	return myID, nil
}

func InvalidateToken(db *redis.Client, tokenString string) error {
	// Add the token to the blacklist with an expiration time of 24 hours
	_, err := db.Set(context.Background(), tokenString, true, 24*time.Hour).Result()
	if err != nil {
		return err
	}

	return nil
}

func IsTokenValid(db *redis.Client, tokenString string) bool {
	// Check if the token is in the blacklist
	_, err := db.Get(context.Background(), tokenString).Result()
	if err == redis.Nil {
		// Token is not in the blacklist, so it's valid
		return true
	}
	if err != nil {
		// Error while checking the blacklist
		log.Printf("Error while checking blacklist: %s", err)
		return false
	}

	// Token is in the blacklist, so it's invalid
	return false
}
