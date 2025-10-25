package database

import (
	"fmt"
	"log"
	"os"

	"github.com/currency-population/internal/models"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(envPath string) {
	if envPath != "" {
		if err := godotenv.Load(envPath); err != nil {
			log.Printf("Warning: Error loading .env file from %s: %v", envPath, err)
		}
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	// Auto Migrate the schema
	err = db.AutoMigrate(&models.Country{})
	if err != nil {
		log.Fatal("Failed to migrate database: ", err)
	}

	DB = db
}
