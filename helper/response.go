package helper

import (
	"log"

	"github.com/gin-gonic/gin"
)

type SuccessResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func HandleSuccess(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, SuccessResponse{
		Status:  statusCode,
		Message: message,
		Data:    data,
	})
}

type ErrorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

func HandleError(c *gin.Context, statusCode int, err error, message string) {
	log.Println(err, "—", message)
	c.JSON(statusCode, ErrorResponse{
		Status:  statusCode,
		Message: message,
		Error:   err.Error(),
	})
}
