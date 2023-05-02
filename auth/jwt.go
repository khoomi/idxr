package auth

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"khoomi-api-io/khoomi_api/configs"
	"log"
	"time"
)

type JWTClaim struct {
	Id        string `json:"id"`
	LoginName string `json:"login_name"`
	Email     string `json:"email"`
	jwt.StandardClaims
}

func GenerateJWT(id, email, loginName string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	jwtKey := configs.LoadEnvFor("SECRET")

	claims := JWTClaim{
		Id:        id,
		LoginName: loginName,
		Email:     email,
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
	jwtKey := configs.LoadEnvFor("SECRET")

	log.Println(signedToken)
	token, err := jwt.ParseWithClaims(
		signedToken,
		&JWTClaim{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtKey), nil
		},
	)
	if err != nil {
		log.Println(err)
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

func ExtractToken(context *gin.Context) string {
	tokenString := context.GetHeader("Authorization")
	return tokenString
}

func ExtractTokenID(c *gin.Context) (string, error) {
	tokenString := ExtractToken(c)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(configs.LoadEnvFor("SECRET")), nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		res := fmt.Sprintf("%v", claims["id"])
		return res, nil
	}

	return "", nil
}
