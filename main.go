package main

import (
	"khoomi-api-io/khoomi_api2/configs"
	"khoomi-api-io/khoomi_api2/routes"
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
