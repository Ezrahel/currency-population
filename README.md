# Country Currency & Exchange Rate API

This is a RESTful API built with Go that fetches country data and exchange rates, stores them in a MySQL database, and provides various endpoints for data access and management.

## Features

- Fetch and store country data from external APIs
- Calculate estimated GDP based on population and exchange rates
- Filter countries by region and currency
- Sort countries by GDP
- Generate and serve summary images
- Full CRUD operations
- Error handling and validation
- MySQL database integration

## Prerequisites

- Go 1.16 or later
- MySQL 5.7 or later
- GCC (for image processing library)

## Setup

1. Clone the repository:
```bash
git clone https://github.com/yourusername/currency-population.git
cd currency-population
```

2. Install dependencies:
```bash
go mod download
```

3. Create a MySQL database:
```sql
CREATE DATABASE country_data;
```

4. Configure environment variables:
Copy the `.env.example` file to `.env` and update the values:
```
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=country_data
APP_PORT=8080
```

5. Run the application:
```bash
go run cmd/api/main.go
```

## API Endpoints

### Refresh Country Data
- `POST /countries/refresh`
  - Fetches fresh data from external APIs and updates the database
  - Generates a new summary image

### Get Countries
- `GET /countries`
  - Query parameters:
    - `region`: Filter by region (e.g., ?region=Africa)
    - `currency`: Filter by currency code (e.g., ?currency=USD)
    - `sort`: Sort by GDP (e.g., ?sort=gdp_desc)

### Get Single Country
- `GET /countries/:name`
  - Get details for a specific country by name

### Delete Country
- `DELETE /countries/:name`
  - Remove a country record from the database

### Get Status
- `GET /status`
  - Returns total number of countries and last refresh timestamp

### Get Summary Image
- `GET /countries/image`
  - Returns a generated image with summary statistics

## Error Handling

The API returns consistent JSON error responses:

- 404: `{ "error": "Country not found" }`
- 400: `{ "error": "Validation failed" }`
- 500: `{ "error": "Internal server error" }`
- 503: `{ "error": "External data source unavailable" }`

## Dependencies

- [gin-gonic/gin](https://github.com/gin-gonic/gin) - Web framework
- [gorm.io/gorm](https://gorm.io) - ORM library
- [joho/godotenv](https://github.com/joho/godotenv) - Environment variable management
- [fogleman/gg](https://github.com/fogleman/gg) - 2D graphics library

## Development

To run the project in development mode with hot reload:

```bash
go install github.com/cosmtrek/air@latest
air
```

## Testing

To run the tests:

```bash
go test ./...
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.