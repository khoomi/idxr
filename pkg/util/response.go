package util

import (
	"log"

	"github.com/gin-gonic/gin"
)

type SuccessResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
	Message string      `json:"message"`
	Status  int         `json:"status"`
}

func HandleSuccess(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, SuccessResponse{
		Status:  statusCode,
		Message: message,
		Data:    data,
		Meta:    nil,
	})
}

func HandleSuccessMeta(c *gin.Context, statusCode int, message string, data, meta interface{}) {
	c.JSON(statusCode, SuccessResponse{
		Status:  statusCode,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

type ErrorResponse struct {
	Error  string `json:"error,omitempty"`
	Status int    `json:"status"`
}

func HandleError(c *gin.Context, statusCode int, err error) {
	log.Printf("error: %v", err)
	c.JSON(statusCode, ErrorResponse{
		Error:  err.Error(),
		Status: statusCode,
	})
}

type UserResponse struct {
	Data    map[string]interface{} `json:"data"`
	Message string                 `json:"message"`
	Status  int                    `json:"status"`
}

type UserResponsePagination struct {
	Data       map[string]interface{} `json:"data"`
	Message    string                 `json:"message"`
	Pagination Pagination             `json:"pagination"`
	Status     int                    `json:"status"`
}

type PaginationArgs struct {
	Sort  string
	Limit int
	Skip  int
}

type Pagination struct {
	Limit int   `json:"limit"`
	Skip  int   `json:"skip"`
	Count int64 `json:"count"`
}
