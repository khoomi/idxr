package auth

import (
	"errors"
	"khoomi-api-io/api/pkg/util"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const AccessTokenExpirationTime = time.Minute * 15
const RefreshTokenExpirationTime = 7 * 4 * time.Hour

type JWTClaim struct {
	Id        string `json:"id"`
	LoginName string `json:"login_name"`
	Email     string `json:"email"`
	IsSeller  bool   `json:"is_seller"`
	jwt.RegisteredClaims
}

// GetNotBefore implements the Claims interface.
func (c JWTClaim) GetAudience() (jwt.ClaimStrings, error) {
	return c.RegisteredClaims.GetAudience()
}

// GetExpirationTime implements the Claims interface.
func (c JWTClaim) GetExpirationTime() (*jwt.NumericDate, error) {
	return c.RegisteredClaims.GetExpirationTime()
}

// GetNotBefore implements the Claims interface.
func (c JWTClaim) GetNotBefore() (*jwt.NumericDate, error) {
	return c.RegisteredClaims.GetNotBefore()
}

// GetIssuedAt implements the Claims interface.
func (c JWTClaim) GetIssuedAt() (*jwt.NumericDate, error) {
	return c.RegisteredClaims.GetIssuedAt()
}

// GetIssuer implements the Claims interface.
func (c JWTClaim) GetIssuer() (string, error) {
	return c.RegisteredClaims.GetIssuer()
}

// GetSubject implements the Claims interface.
func (c JWTClaim) GetSubject() (string, error) {
	return c.RegisteredClaims.GetSubject()
}

// Generate auth token for new user session
func GenerateJWT(id, email, loginName string, seller bool) (string, int64, error) {
	expirationTime := time.Now().Local().Add(AccessTokenExpirationTime)
	expirationTimeNumericDate := jwt.NewNumericDate(expirationTime)
	jwtKey := util.LoadEnvFor("SECRET")

	claims := JWTClaim{
		Id:        id,
		LoginName: loginName,
		Email:     email,
		IsSeller:  seller,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: expirationTimeNumericDate,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtKey))
	if err != nil {
		return "", 0, err
	}

	return tokenString, expirationTime.Unix(), nil
}

// Generate refresh auth token for new user session.
func GenerateRefreshJWT(id, email, loginName string, seller bool) (string, error) {
	expirationTime := time.Now().Local().Add(RefreshTokenExpirationTime)
	expirationTimeNumericDate := jwt.NewNumericDate(expirationTime)
	jwtKey := util.LoadEnvFor("REFRESH_SECRET")

	claims := JWTClaim{
		Id:        id,
		LoginName: loginName,
		Email:     email,
		IsSeller:  seller,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: expirationTimeNumericDate,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// Validate a signed jwt refresh token and it's expiration time.
func ValidateRefreshToken(signedToken string) (claim JWTClaim, err error) {
	jwtKey := util.LoadEnvFor("REFRESH_SECRET")
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

	exp, _ := claims.GetExpirationTime()
	if exp.Local().Unix() < time.Now().Local().Unix() {
		err = errors.New("token expired")
		return JWTClaim{}, err
	}

	return *claims, nil
}

// Validate a signed jwt auth token and it's expiration time.
func ValidateToken(signedToken string) (JWTClaim, error) {
	jwtKey := util.LoadEnvFor("SECRET")
	token, err := jwt.ParseWithClaims(
		signedToken,
		&JWTClaim{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtKey), nil
		},
	)
	if err != nil {
		return JWTClaim{}, err
	}

	claim, ok := token.Claims.(*JWTClaim)
	if !ok {
		err = errors.New("couldn't parse claims")
		return JWTClaim{}, err
	}
	exp, _ := claim.GetExpirationTime()
	if exp.Local().Unix() < time.Now().Local().Unix() {
		err = errors.New("token expired")
		return JWTClaim{}, err
	}

	return *claim, nil
}

// Extract and Validate jwt auth token.
func InitJwtClaim(c *gin.Context) (JWTClaim, error) {
	tknStr := ExtractToken(c)
	token, err := ValidateToken(tknStr)
	if err != nil {
		return JWTClaim{}, err
	}

	return token, nil
}

// Get user object ID from JWTClaim.
func (j JWTClaim) GetUserObjectId() (primitive.ObjectID, error) {
	userId, err := primitive.ObjectIDFromHex(j.Id)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return userId, nil
}

// Extract authorization token from request header.
func ExtractToken(context *gin.Context) string {
	tokenString := context.GetHeader("Authorization")
	return tokenString
}
