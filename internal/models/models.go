package models

import "time"

type Country struct {
	ID              uint      `json:"id" gorm:"primary_key"`
	Name            string    `json:"name" gorm:"unique;not null"`
	Capital         string    `json:"capital"`
	Region          string    `json:"region"`
	Population      int64     `json:"population" gorm:"not null"`
	CurrencyCode    string    `json:"currency_code" gorm:"not null"`
	ExchangeRate    *float64  `json:"exchange_rate"`
	EstimatedGDP    *float64  `json:"estimated_gdp"`
	FlagURL         string    `json:"flag_url"`
	LastRefreshedAt time.Time `json:"last_refreshed_at"`
}

type ExternalCountry struct {
	Name       string `json:"name"`
	Capital    string `json:"capital"`
	Region     string `json:"region"`
	Population int64  `json:"population"`
	Flag       string `json:"flag"`
	Currencies []struct {
		Code string `json:"code"`
	} `json:"currencies"`
}

type ExchangeRates struct {
	Rates map[string]float64 `json:"rates"`
}

type StatusResponse struct {
	TotalCountries  int       `json:"total_countries"`
	LastRefreshedAt time.Time `json:"last_refreshed_at"`
}

type ErrorResponse struct {
	Error   string            `json:"error"`
	Details map[string]string `json:"details,omitempty"`
}
