package main

import (
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/currency-population/internal/api"
	"github.com/currency-population/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Find and load environment variables
	envPaths := []string{
		".env",          // Try current directory first
		"../../.env",    // Try project root
		"../../../.env", // Try one level up (in case running from binary location)
	}

	var envPath string
	for _, path := range envPaths {
		if _, err := os.Stat(path); err == nil {
			envPath = path
			break
		}
	}

	if envPath == "" {
		log.Fatal("Could not find .env file in any of the expected locations")
	}

	if err := godotenv.Load(envPath); err != nil {
		log.Fatal("Error loading .env file: ", err)
	}

	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Initialize database
	database.InitDB(envPath)

	// Setup Gin router
	r := gin.Default()

	// Setup routes
	api.SetupRoutes(r)

	// Get port from environment variable
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
