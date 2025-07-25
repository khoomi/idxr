package controllers

import (
	"context"
	"errors"
	"log"
	"net/http"

	"khoomi-api-io/api/internal/auth"
	"khoomi-api-io/api/internal/common"
	"khoomi-api-io/api/pkg/services"
	"khoomi-api-io/api/pkg/util"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ControllerContext wraps the standard context with useful utilities for controllers
type ControllerContext struct {
	Ctx    context.Context
	Cancel context.CancelFunc
	UserID primitive.ObjectID
}

// WithTimeout creates a context with the standard request timeout
func WithTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), common.REQUEST_TIMEOUT_SECS)
}

// ValidateAndGetUserID validates user ID and handles errors automatically
func ValidateAndGetUserID(c *gin.Context) (primitive.ObjectID, bool) {
	userID, err := auth.ValidateUserID(c)
	if err != nil {
		util.HandleError(c, http.StatusUnauthorized, err)
		return primitive.NilObjectID, false
	}
	return userID, true
}

// ValidateAndGetUserIDWithCustomError validates user ID with custom error status
func ValidateAndGetUserIDWithCustomError(c *gin.Context, errorStatus int) (primitive.ObjectID, bool) {
	userID, err := auth.ValidateUserID(c)
	if err != nil {
		util.HandleError(c, errorStatus, err)
		return primitive.NilObjectID, false
	}
	return userID, true
}

// SetupControllerContext creates a standard controller context with timeout and user validation
func SetupControllerContext(c *gin.Context) (*ControllerContext, bool) {
	ctx, cancel := WithTimeout()

	userID, ok := ValidateAndGetUserID(c)
	if !ok {
		cancel()
		return nil, false
	}

	return &ControllerContext{
		Ctx:    ctx,
		Cancel: cancel,
		UserID: userID,
	}, true
}

// SetupControllerContextWithoutAuth creates a controller context without user validation
func SetupControllerContextWithoutAuth() *ControllerContext {
	ctx, cancel := WithTimeout()

	return &ControllerContext{
		Ctx:    ctx,
		Cancel: cancel,
	}
}

// Cleanup should be called to release resources (typically deferred)
func (cc *ControllerContext) Cleanup() {
	if cc.Cancel != nil {
		cc.Cancel()
	}
}

// ParseObjectIDParam parses an ObjectID from URL parameter and handles errors
func ParseObjectIDParam(c *gin.Context, paramName string) (primitive.ObjectID, bool) {
	idString := c.Param(paramName)
	if idString == "" {
		util.HandleError(c, http.StatusBadRequest, gin.Error{Err: gin.Error{}.Err, Type: gin.ErrorTypePublic})
		return primitive.NilObjectID, false
	}

	objectID, err := primitive.ObjectIDFromHex(idString)
	if err != nil {
		util.HandleError(c, http.StatusBadRequest, err)
		return primitive.NilObjectID, false
	}

	return objectID, true
}

// BindJSONAndValidate binds JSON and handles validation errors
func BindJSONAndValidate(c *gin.Context, obj any) bool {
	if err := c.ShouldBind(obj); err != nil {
		log.Printf("JSON binding error: %v", err)
		util.HandleError(c, http.StatusBadRequest, err)
		return false
	}

	log.Println("After binding:", obj)

	if err := common.Validate.Struct(obj); err != nil {
		log.Printf("Validation error: %v", err)
		util.HandleError(c, http.StatusBadRequest, err)
		return false
	}

	return true
}

// HandlePaginationAndResponse is a utility for common pagination responses
func HandlePaginationAndResponse(c *gin.Context, data any, count int64, paginationArgs util.PaginationArgs, message string) {
	util.HandleSuccessMeta(c, http.StatusOK, message, data, gin.H{
		"pagination": util.Pagination{
			Limit: paginationArgs.Limit,
			Skip:  paginationArgs.Skip,
			Count: count,
		},
	})
}

// ValidateSellerAccess validates user ID and checks if user is a seller
func ValidateSellerAccess(c *gin.Context, userService services.UserService) (primitive.ObjectID, bool) {
	userID, ok := ValidateAndGetUserID(c)
	if !ok {
		return primitive.NilObjectID, false
	}

	isSeller, err := userService.IsSeller(c, userID)
	if err != nil {
		util.HandleError(c, http.StatusInternalServerError, err)
		return primitive.NilObjectID, false
	}

	if !isSeller {
		util.HandleError(c, http.StatusUnauthorized, gin.Error{Err: errors.New("only sellers can perform this action")})
		return primitive.NilObjectID, false
	}

	return userID, true
}
