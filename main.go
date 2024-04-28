package main

import (
	"khoomi-api-io/khoomi_api/routes"
)

func main() {
	// Initialize routes
	router := routes.InitRoute()
	err := router.Run("0.0.0.0:8080")
	if err != nil {
		println(err.Error())
		return
	}
}
