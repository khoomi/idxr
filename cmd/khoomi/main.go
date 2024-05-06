package main

import (
	routers "khoomi-api-io/api/internal/routers"
)

func main() {
	// Initialize routes
	router := routers.InitRoute()
	err := router.Run("0.0.0.0:8080")
	if err != nil {
		println(err.Error())
		return
	}
}
