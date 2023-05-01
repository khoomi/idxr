package main

import (
	"khoomi-api-io/khoomi_api/configs"
	"khoomi-api-io/khoomi_api/routes"
)

func main() {
	// Initialize database connection
	configs.ConnectDB()
	// User routes
	router := routes.InitRoute()
	err := router.Run("localhost:8080")

	if err != nil {
		println(err.Error())
		return
	}
}
