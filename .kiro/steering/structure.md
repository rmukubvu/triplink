
# Project Structure

## Directory Organization

```
triplink/
├── backend/               # Go backend API server
│   ├── main.go           # Application entry point
│   ├── go.mod            # Go module dependencies
│   ├── go.sum            # Dependency checksums
│   ├── docker-compose.yml # PostgreSQL container setup
│   ├── README.md         # Project documentation
│   ├── api_schema.json   # API schema reference
│   ├── auth/             # Authentication logic
│   │   └── auth.go
│   ├── database/         # Database connection and configuration
│   │   └── database.go
│   ├── models/           # Data models and database schemas
│   │   └── models.go
│   ├── handlers/         # HTTP request handlers (controllers)
│   │   ├── user_handler.go
│   │   ├── trip_handler.go
│   │   ├── load_handler.go
│   │   ├── quote_handler.go
│   │   ├── vehicle_handler.go
│   │   ├── message_handler.go
│   │   ├── transaction_handler.go
│   │   ├── review_handler.go
│   │   ├── notification_handler.go
│   │   ├── manifest_handler.go
│   │   └── customs_handler.go
│   ├── routes/           # Route definitions and middleware
│   ├── services/         # Business logic services
│   ├── test/             # Backend tests
│   └── docs/             # Auto-generated Swagger documentation
│       ├── docs.go
│       ├── swagger.json
│       └── swagger.yaml
├── triplink-web-react/    # React web application
│   ├── src/
│   │   ├── components/   # Reusable React components
│   │   ├── pages/        # Page components
│   │   ├── services/     # API service functions
│   │   ├── types/        # TypeScript type definitions
│   │   ├── contexts/     # React contexts
│   │   ├── hooks/        # Custom React hooks
│   │   ├── utils/        # Utility functions
│   │   └── assets/       # Static assets
│   ├── public/
│   ├── package.json
│   └── README.md
└── triplink-mobile/       # React Native mobile application
    ├── src/
    │   ├── components/   # Reusable React Native components
    │   ├── screens/      # Screen components
    │   ├── services/     # API and device services
    │   ├── navigation/   # Navigation configuration
    │   ├── state/        # State management (Redux)
    │   ├── utils/        # Utility functions
    │   └── assets/       # Static assets
    ├── __tests__/        # Mobile app tests
    ├── android/          # Android-specific code
    ├── ios/              # iOS-specific code
    ├── package.json
    └── README.md
```

## Architecture Patterns

### MVC-Style Organization
- **Models** (`models/`): GORM-based data structures with database relationships
- **Handlers** (`handlers/`): Business logic and HTTP request processing
- **Routes** (`routes/`): URL routing and middleware configuration

### Key Conventions

#### Models
- All models extend `BaseModel` with standard fields (ID, CreatedAt, UpdatedAt, DeletedAt)
- Use GORM tags for database constraints and JSON serialization
- Password fields use `json:"-"` to exclude from responses
- Relationships defined with foreign keys and GORM associations

#### Handlers
- Each domain has its own handler file (e.g., `user_handler.go`, `trip_handler.go`)
- Functions follow naming pattern: `Register`, `Login`, `CreateTrip`, etc.
- Use Swagger comments for API documentation generation
- Standard error response structure with `ErrorResponse` type
- JWT authentication handled via cookies and tokens

#### Database
- Single `database.go` file handles connection and auto-migration
- Global `DB` variable for database access across handlers
- PostgreSQL connection with GORM ORM

#### API Structure
- RESTful endpoints under `/api/` prefix
- JSON request/response format
- Swagger documentation auto-generated from handler comments
- Standard HTTP status codes for responses

## File Naming Conventions
- Snake_case for handler files: `user_handler.go`, `trip_handler.go`
- Singular names for model files: `models.go`
- Lowercase package names matching directory names