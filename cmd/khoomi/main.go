package main

import (
	"log"

	"khoomi-api-io/api/internal/routers"
)

func main() {
	log.SetFlags(log.Lshortfile)
	router := routers.InitRoute()
	err := router.Run("0.0.0.0:8080")
	if err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
