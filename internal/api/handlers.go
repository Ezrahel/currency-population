package api

import (
	"encoding/json"
	"fmt"
	"image/color"
	"image/png"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/currency-population/internal/database"
	"github.com/currency-population/internal/models"
	"github.com/fogleman/gg"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	countriesAPI = "https://restcountries.com/v2/all?fields=name,capital,region,population,flag,currencies"
	exchangeAPI  = "https://open.er-api.com/v6/latest/USD"
)

func SetupRoutes(r *gin.Engine) {
	r.POST("/countries/refresh", refreshCountries)
	r.GET("/countries", getCountries)
	r.GET("/countries/:name", getCountry)
	r.DELETE("/countries/:name", deleteCountry)
	r.GET("/status", getStatus)
	r.GET("/countries/image", getCountryImage)
}

func refreshCountries(c *gin.Context) {
	// Fetch countries data
	resp, err := http.Get(countriesAPI)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
			Error:   "External data source unavailable",
			Details: map[string]string{"api": "Countries API"},
		})
		return
	}
	defer resp.Body.Close()

	var countries []models.ExternalCountry
	if err := json.NewDecoder(resp.Body).Decode(&countries); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to parse countries data",
		})
		return
	}

	// Fetch exchange rates
	resp, err = http.Get(exchangeAPI)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
			Error:   "External data source unavailable",
			Details: map[string]string{"api": "Exchange Rates API"},
		})
		return
	}
	defer resp.Body.Close()

	var exchangeRates models.ExchangeRates
	if err := json.NewDecoder(resp.Body).Decode(&exchangeRates); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to parse exchange rates data",
		})
		return
	}

	now := time.Now()
	for _, extCountry := range countries {
		country := models.Country{
			Name:            extCountry.Name,
			Capital:         extCountry.Capital,
			Region:          extCountry.Region,
			Population:      extCountry.Population,
			FlagURL:         extCountry.Flag,
			LastRefreshedAt: now,
		}

		if len(extCountry.Currencies) > 0 {
			country.CurrencyCode = extCountry.Currencies[0].Code
			if rate, ok := exchangeRates.Rates[country.CurrencyCode]; ok {
				country.ExchangeRate = &rate
				multiplier := 1000.0 + rand.Float64()*1000.0 // Random between 1000-2000
				gdp := float64(country.Population) * multiplier / rate
				country.EstimatedGDP = &gdp
			}
		}

		// Try to find existing country
		var existingCountry models.Country
		result := database.DB.Where("LOWER(name) = LOWER(?)", country.Name).First(&existingCountry)
		if result.Error == nil {
			// Update existing country
			database.DB.Model(&existingCountry).Updates(country)
		} else if result.Error == gorm.ErrRecordNotFound {
			// Create new country
			database.DB.Create(&country)
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Database error",
			})
			return
		}
	}

	// Generate summary image
	if err := generateSummaryImage(); err != nil {
		// Log the error but don't fail the request
		fmt.Printf("Error generating summary image: %v\n", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Data refreshed successfully"})
}

func getCountries(c *gin.Context) {
	region := c.Query("region")
	currency := c.Query("currency")
	sort := c.Query("sort")

	query := database.DB

	if region != "" {
		query = query.Where("region = ?", region)
	}
	if currency != "" {
		query = query.Where("currency_code = ?", currency)
	}

	if sort == "gdp_desc" {
		query = query.Order("estimated_gdp DESC")
	}

	var countries []models.Country
	if err := query.Find(&countries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Database error",
		})
		return
	}

	c.JSON(http.StatusOK, countries)
}

func getCountry(c *gin.Context) {
	name := c.Param("name")

	var country models.Country
	if err := database.DB.Where("LOWER(name) = LOWER(?)", name).First(&country).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "Country not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Database error",
		})
		return
	}

	c.JSON(http.StatusOK, country)
}

func deleteCountry(c *gin.Context) {
	name := c.Param("name")

	result := database.DB.Where("LOWER(name) = LOWER(?)", name).Delete(&models.Country{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Database error",
		})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Country not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Country deleted successfully"})
}

func getStatus(c *gin.Context) {
	var count int64
	var lastRefreshed models.Country

	if err := database.DB.Model(&models.Country{}).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Database error",
		})
		return
	}

	if err := database.DB.Order("last_refreshed_at DESC").First(&lastRefreshed).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Database error",
			})
			return
		}
	}

	c.JSON(http.StatusOK, models.StatusResponse{
		TotalCountries:  int(count),
		LastRefreshedAt: lastRefreshed.LastRefreshedAt,
	})
}

func generateSummaryImage() error {
	const width = 800
	const height = 600

	dc := gg.NewContext(width, height)

	// Set background
	dc.SetColor(color.White)
	dc.Clear()

	// Draw title
	dc.SetColor(color.Black)
	if err := dc.LoadFontFace("Arial", 24); err != nil {
		return fmt.Errorf("error loading font: %v", err)
	}

	// Get total countries and top 5 by GDP
	var count int64
	database.DB.Model(&models.Country{}).Count(&count)

	var topCountries []models.Country
	database.DB.Order("estimated_gdp DESC").Limit(5).Find(&topCountries)

	// Draw content
	y := 50.0
	dc.DrawString(fmt.Sprintf("Total Countries: %d", count), 50, y)
	y += 50

	dc.DrawString("Top 5 Countries by GDP:", 50, y)
	y += 30

	for _, country := range topCountries {
		gdp := "N/A"
		if country.EstimatedGDP != nil {
			gdp = fmt.Sprintf("$%.2f B", *country.EstimatedGDP/1e9)
		}
		dc.DrawString(fmt.Sprintf("%s: %s", country.Name, gdp), 70, y)
		y += 30
	}

	dc.DrawString(fmt.Sprintf("Last Updated: %s", time.Now().Format(time.RFC3339)), 50, y+30)

	// Create cache directory if it doesn't exist
	if err := os.MkdirAll("cache", 0755); err != nil {
		return fmt.Errorf("error creating cache directory: %v", err)
	}

	// Save the image
	f, err := os.Create(filepath.Join("cache", "summary.png"))
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer f.Close()

	if err := png.Encode(f, dc.Image()); err != nil {
		return fmt.Errorf("error encoding image: %v", err)
	}

	return nil
}

func getCountryImage(c *gin.Context) {
	imagePath := filepath.Join("cache", "summary.png")
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Summary image not found",
		})
		return
	}

	c.File(imagePath)
}
