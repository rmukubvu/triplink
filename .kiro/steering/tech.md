# Tech Stack

## Core Technologies

- **Language**: Go 1.23.3
- **Web Framework**: Fiber v2 (Express-like HTTP framework for Go)
- **Database**: PostgreSQL 15 with GORM ORM
- **Authentication**: JWT tokens with golang-jwt/jwt/v4
- **Password Hashing**: bcrypt (golang.org/x/crypto)
- **API Documentation**: Swagger/OpenAPI with swaggo/swag
- **Containerization**: Docker & Docker Compose

## Key Dependencies

- `github.com/gofiber/fiber/v2` - HTTP web framework
- `gorm.io/gorm` & `gorm.io/driver/postgres` - ORM and PostgreSQL driver
- `github.com/golang-jwt/jwt/v4` - JWT authentication
- `golang.org/x/crypto` - Cryptographic functions
- `github.com/gofiber/swagger` - Swagger UI integration

## Development Commands

### Database Setup
```bash
# Start PostgreSQL database
docker-compose up -d
# or
docker compose up -d

# Check database status
docker-compose ps

# Stop database
docker-compose down
```

### Application Commands
```bash
# Run development server
go run main.go

# Build application
go build -o triplink

# Run built binary
./triplink
```

### Database Access
```bash
# Connect to PostgreSQL directly
docker exec -it triplink_postgres psql -U postgres -d triplink
```

## Database Configuration

- **Host**: localhost:5432
- **Database**: triplink
- **User**: postgres
- **Password**: password
- **Connection String**: `postgresql://postgres:password@localhost:5432/triplink`

## API Documentation

Swagger documentation is auto-generated and available when the server is running. The API uses standard REST conventions with JSON payloads.