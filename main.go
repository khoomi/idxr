package main

import (
	"khoomi-api-io/khoomi_api/configs"
	"khoomi-api-io/khoomi_api/controllers"
	"khoomi-api-io/khoomi_api/email"
	"khoomi-api-io/khoomi_api/routes"
)

func main() {
	// Initialize email worker pool
	controllers.EmailPool = email.KhoomiEmailWorkerPoolInstance(5)
	controllers.EmailPool.Start()
	defer controllers.EmailPool.Stop()

	// Initialize database connection
	configs.ConnectDB()
	
// Initialize routes 
	router := routes.InitRoute()
	err := router.Run("localhost:8080")

	if err != nil {
		println(err.Error())
		return
	}
}
