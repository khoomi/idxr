package util

import (
	"log"

	"github.com/gin-gonic/gin"
)

type SuccessResponse struct {
	Data    any    `json:"data,omitempty"`
	Meta    any    `json:"meta,omitempty"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

func HandleSuccess(c *gin.Context, statusCode int, message string, data any) {
	c.JSON(statusCode, SuccessResponse{
		Status:  statusCode,
		Message: message,
		Data:    data,
		Meta:    nil,
	})
}

func HandleSuccessMeta(c *gin.Context, statusCode int, message string, data, meta any) {
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
	Data    map[string]any `json:"data"`
	Message string         `json:"message"`
	Status  int            `json:"status"`
}

type UserResponsePagination struct {
	Data       map[string]any `json:"data"`
	Message    string         `json:"message"`
	Pagination Pagination     `json:"pagination"`
	Status     int            `json:"status"`
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
