package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func Ping(context *gin.Context) {
	time := time.Now().Local()
	context.JSON(http.StatusOK, gin.H{"message": "pong", "local_time": time})
}
