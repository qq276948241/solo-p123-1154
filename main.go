package main

import (
	"log"
	"pet-clinic-backend/config"
	"pet-clinic-backend/database"
)

func main() {
	config.LoadConfig()

	database.InitMySQL()
	database.InitRedis()

	r := SetupRouter()

	log.Printf("Server starting on port %s", config.AppConfig.ServerPort)
	if err := r.Run(":" + config.AppConfig.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
