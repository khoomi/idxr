package helper

import (
	"log"

	"github.com/gin-gonic/gin"
)

type SuccessResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
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
	Error string `json:"error,omitempty"`
}

func HandleError(c *gin.Context, statusCode int, err error, message string) {
	log.Println(err)
	c.JSON(statusCode, ErrorResponse{
		Error: message,
	})
}

type UserResponse struct {
	Status  int                    `json:"status"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

type UserResponsePagination struct {
	Status     int                    `json:"status"`
	Message    string                 `json:"message"`
	Data       map[string]interface{} `json:"data"`
	Pagination Pagination             `json:"pagination"`
}

type PaginationArgs struct {
	Limit int
	Skip  int
	Sort  string
}

type Pagination struct {
	Limit int   `json:"limit"`
	Skip  int   `json:"skip"`
	Count int64 `json:"count"`
}
