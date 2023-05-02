package auth

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"khoomi-api-io/khoomi_api/configs"
	"time"
)

type JWTClaim struct {
	LoginName string `json:"login_name"`
	Email     string `json:"email"`
	jwt.StandardClaims
}

func GenerateJWT(email, loginName string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := JWTClaim{
		LoginName: loginName,
		Email:     email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	tokenString, err := token.SigningString()
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ValidateToken(signedToken string) (err error) {
	jwtKey := configs.LoadEnvFor("SECRET")

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
