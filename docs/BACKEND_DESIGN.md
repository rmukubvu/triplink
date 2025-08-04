# TripLink Backend - System Design Document

## Overview

TripLink is a comprehensive logistics and freight management platform that connects carriers and shippers for efficient cargo transportation. The backend is built using Go with the Fiber web framework, PostgreSQL database, and follows a clean architecture pattern with clear separation of concerns.

## Architecture

### High-Level Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Client    │    │  Mobile Client  │    │  External APIs  │
│  (React SPA)    │    │ (React Native)  │    │   (3rd Party)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │   API Gateway   │
                    │   (Fiber v2)    │
                    └─────────────────┘
                                 │
         ┌───────────────────────┼───────────────────────┐
         │                       │                       │
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Authentication │    │  Business Logic │    │   External      │
│   & Security    │    │    Services     │    │  Integrations   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │   Data Layer    │
                    │  (GORM + PG)    │
                    └─────────────────┘
                                 │
                    ┌─────────────────┐
                    │   PostgreSQL    │
                    │    Database     │
                    └─────────────────┘
```

### Technology Stack

- **Language**: Go 1.23.3
- **Web Framework**: Fiber v2 (Express-like HTTP framework)
- **Database**: PostgreSQL 15 with GORM ORM
- **Authentication**: JWT tokens with golang-jwt/jwt/v4
- **Password Security**: bcrypt hashing
- **API Documentation**: Swagger/OpenAPI with swaggo/swag
- **Containerization**: Docker & Docker Compose
- **Real-time Communication**: WebSocket support
- **File Storage**: Local/Cloud storage for documents and images
- **Caching**: Redis (optional for session management)

## Core Features & Components

### 1. User Management System

**Purpose**: Multi-role user system supporting carriers, shippers, and administrators.

**Key Components**:
- User registration and authentication
- Role-based access control (RBAC)
- Profile management with document verification
- Company information and business licenses
- Driver license management

**Models**:
- `User`: Core user entity with role-based permissions
- `NotificationPreferences`: User-specific notification settings
- `NotificationToken`: Device tokens for push notifications

**API Endpoints**:
- `POST /api/auth/register` - User registration
- `POST /api/auth/login` - User authentication
- `GET /api/users/profile` - Get user profile
- `PUT /api/users/profile` - Update user profile
- `POST /api/users/verify` - Document verification

### 2. Trip Management System

**Purpose**: Comprehensive trip planning, scheduling, and management for carriers.

**Key Components**:
- Trip creation with origin/destination routing
- Capacity management (weight and volume)
- Dynamic pricing configuration
- Trip status tracking and updates
- Real-time location tracking integration

**Models**:
- `Trip`: Core trip entity with routing and capacity information
- `TrackingRecord`: GPS location history
- `TrackingStatus`: Current trip status and ETA
- `TrackingEvent`: Trip milestone events

**API Endpoints**:
- `POST /api/trips` - Create new trip
- `GET /api/trips` - List trips with filtering
- `GET /api/trips/{id}` - Get trip details
- `PUT /api/trips/{id}` - Update trip information
- `POST /api/trips/{id}/tracking/location` - Update location
- `GET /api/trips/{id}/tracking/current` - Get current status

### 3. Load Booking System

**Purpose**: Shipper load booking and management with comprehensive cargo details.

**Key Components**:
- Load creation with detailed cargo specifications
- Pickup and delivery scheduling
- Special handling requirements (fragile, hazmat, refrigerated)
- Insurance and customs documentation
- Load status tracking throughout journey

**Models**:
- `Load`: Comprehensive load entity with cargo details
- `CustomsDocument`: International shipping documentation
- `Quote`: Pricing quotes from carriers

**API Endpoints**:
- `POST /api/loads` - Create load booking
- `GET /api/loads` - List loads with filtering
- `GET /api/loads/{id}` - Get load details
- `PUT /api/loads/{id}/status` - Update load status
- `POST /api/loads/{id}/documents` - Upload customs documents

### 4. Quote & Pricing System

**Purpose**: Dynamic pricing and quote management between carriers and shippers.

**Key Components**:
- Quote request and response workflow
- Dynamic pricing based on distance, weight, and market factors
- Quote comparison and selection
- Automated quote expiration
- Price negotiation support

**Models**:
- `Quote`: Quote entity with pricing and validity
- Integrated with `Load` and `Trip` models

**API Endpoints**:
- `POST /api/loads/{id}/quotes` - Submit quote for load
- `GET /api/quotes` - List quotes for user
- `PUT /api/quotes/{id}/accept` - Accept quote
- `PUT /api/quotes/{id}/reject` - Reject quote

### 5. Vehicle Management System

**Purpose**: Comprehensive fleet management for carriers.

**Key Components**:
- Vehicle registration with specifications
- Capacity and capability tracking
- Equipment features (liftgate, straps, refrigeration)
- Certification management (hazmat, food grade)
- Insurance and inspection tracking
- Vehicle image management

**Models**:
- `Vehicle`: Comprehensive vehicle entity with specifications

**API Endpoints**:
- `POST /api/vehicles` - Register new vehicle
- `GET /api/vehicles` - List user vehicles
- `PUT /api/vehicles/{id}` - Update vehicle information
- `POST /api/vehicles/{id}/images` - Upload vehicle images

### 6. Real-time Tracking System

**Purpose**: End-to-end shipment tracking with real-time location updates.

**Key Components**:
- GPS location tracking and history
- Automated status updates and notifications
- ETA calculations and delay detection
- Milestone tracking (departure, arrival, delays)
- Geofencing for pickup/delivery zones

**Models**:
- `TrackingRecord`: GPS location points with timestamps
- `TrackingStatus`: Current status and ETA information
- `TrackingEvent`: Milestone and event logging

**API Endpoints**:
- `POST /api/trips/{id}/tracking/location` - Update GPS location
- `GET /api/trips/{id}/tracking/history` - Get location history
- `GET /api/loads/{id}/tracking` - Get load tracking status
- `POST /api/tracking/events` - Log tracking events

### 7. Communication System

**Purpose**: Secure messaging between carriers and shippers.

**Key Components**:
- Direct messaging between users
- Load-specific communication threads
- File attachment support
- Message history and search
- Real-time message delivery

**Models**:
- `Message`: Message entity with sender/receiver information

**API Endpoints**:
- `POST /api/messages` - Send message
- `GET /api/messages` - Get message history
- `GET /api/messages/conversations` - List conversations

### 8. Review & Rating System

**Purpose**: Mutual rating system for building trust between users.

**Key Components**:
- Bidirectional reviews (carrier ↔ shipper)
- 5-star rating system with comments
- Aggregate rating calculations
- Review moderation and reporting
- Rating history and trends

**Models**:
- `Review`: Review entity with rating and comments

**API Endpoints**:
- `POST /api/reviews` - Submit review
- `GET /api/reviews/{userId}` - Get user reviews
- `GET /api/reviews/stats/{userId}` - Get rating statistics

### 9. Payment & Transaction System

**Purpose**: Secure payment processing with multiple gateway support.

**Key Components**:
- Multiple payment gateway integration (Stripe, PayPal)
- Platform fee calculation and collection
- Transaction history and reporting
- Refund and dispute management
- Automated payment processing

**Models**:
- `Transaction`: Transaction entity with payment details

**API Endpoints**:
- `POST /api/payments/process` - Process payment
- `GET /api/transactions` - Get transaction history
- `POST /api/payments/refund` - Process refund

### 10. Notification System

**Purpose**: Multi-channel notification delivery (push, email, SMS).

**Key Components**:
- Push notification delivery (iOS/Android)
- Email notification templates
- SMS notifications for critical updates
- User notification preferences
- Delivery tracking and retry logic

**Models**:
- `Notification`: Notification entity
- `NotificationToken`: Device tokens for push notifications
- `NotificationDelivery`: Delivery tracking
- `NotificationPreferences`: User preferences

**API Endpoints**:
- `GET /api/notifications` - Get user notifications
- `PUT /api/notifications/{id}/read` - Mark as read
- `PUT /api/notifications/preferences` - Update preferences

### 11. Document Management System

**Purpose**: Customs and shipping document management for international shipments.

**Key Components**:
- Document upload and storage
- Document type validation
- Expiration date tracking
- Digital signature support
- Document sharing and access control

**Models**:
- `CustomsDocument`: Document entity with metadata
- `Manifest`: Trip manifest generation

**API Endpoints**:
- `POST /api/documents/upload` - Upload document
- `GET /api/documents/{id}` - Get document
- `GET /api/trips/{id}/manifest` - Generate manifest

### 12. Analytics & Reporting System

**Purpose**: Business intelligence and performance analytics.

**Key Components**:
- Trip and load analytics
- Revenue and cost tracking
- Performance metrics and KPIs
- Custom report generation
- Data export capabilities

**API Endpoints**:
- `GET /api/analytics/trips` - Trip analytics
- `GET /api/analytics/revenue` - Revenue analytics
- `GET /api/analytics/performance` - Performance metrics

## Database Design

### Core Entities Relationships

```
User (1) ──── (N) Vehicle
User (1) ──── (N) Trip
User (1) ──── (N) Load
Trip (1) ──── (N) Load
Load (1) ──── (N) CustomsDocument
Load (1) ──── (N) Quote
Load (1) ──── (1) Transaction
Trip (1) ──── (N) TrackingRecord
Trip (1) ──── (1) TrackingStatus
Trip (1) ──── (N) TrackingEvent
User (1) ──── (N) Review (as reviewer)
User (1) ──── (N) Review (as reviewee)
User (1) ──── (N) Message (as sender)
User (1) ──── (N) Message (as receiver)
User (1) ──── (N) Notification
```

### Key Database Features

- **GORM ORM**: Provides type-safe database operations
- **Soft Deletes**: All entities support soft deletion
- **Timestamps**: Automatic created_at/updated_at tracking
- **Indexes**: Optimized queries with proper indexing
- **Constraints**: Foreign key relationships and data integrity
- **Migrations**: Version-controlled database schema changes

## Security Architecture

### Authentication & Authorization

- **JWT Tokens**: Stateless authentication with configurable expiration
- **Role-Based Access Control**: CARRIER, SHIPPER, ADMIN roles
- **Password Security**: bcrypt hashing with salt
- **API Rate Limiting**: Protection against abuse
- **CORS Configuration**: Cross-origin request security

### Data Security

- **Input Validation**: Comprehensive request validation
- **SQL Injection Prevention**: GORM ORM protection
- **File Upload Security**: Type validation and size limits
- **Sensitive Data Encryption**: Password and payment information
- **Audit Logging**: Security event tracking

## Performance & Scalability

### Optimization Strategies

- **Database Indexing**: Optimized query performance
- **Connection Pooling**: Efficient database connections
- **Caching Layer**: Redis for session and frequently accessed data
- **File Storage**: CDN integration for static assets
- **API Pagination**: Large dataset handling
- **Background Jobs**: Asynchronous processing for heavy operations

### Monitoring & Observability

- **Health Checks**: System health monitoring endpoints
- **Metrics Collection**: Performance and usage metrics
- **Error Tracking**: Comprehensive error logging
- **API Documentation**: Auto-generated Swagger documentation

## Deployment Architecture

### Containerization

```dockerfile
# Multi-stage Docker build
FROM golang:1.23.3-alpine AS builder
# Build application
FROM alpine:latest AS runtime
# Production runtime
```

### Environment Configuration

- **Development**: Local development with hot reload
- **Staging**: Pre-production testing environment
- **Production**: Scalable production deployment

### Database Management

```yaml
# docker-compose.yml
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: triplink
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
```

## API Design Principles

### RESTful Architecture

- **Resource-based URLs**: `/api/trips`, `/api/loads`
- **HTTP Methods**: GET, POST, PUT, DELETE
- **Status Codes**: Proper HTTP status code usage
- **JSON Responses**: Consistent response format

### Error Handling

```json
{
  "error": true,
  "message": "Validation failed",
  "details": {
    "field": "email",
    "issue": "Invalid email format"
  }
}
```

### Response Format

```json
{
  "success": true,
  "data": {...},
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 100
  }
}
```

## Integration Points

### External Services

- **Payment Gateways**: Stripe, PayPal integration
- **Mapping Services**: Google Maps API for routing
- **Push Notifications**: Firebase Cloud Messaging
- **Email Service**: SMTP/SendGrid integration
- **SMS Service**: Twilio integration
- **File Storage**: AWS S3 or local storage

### Webhook Support

- **Payment Webhooks**: Real-time payment status updates
- **Tracking Webhooks**: External tracking system integration
- **Notification Webhooks**: Third-party notification delivery

## Development Guidelines

### Code Organization

```
backend/
├── main.go                 # Application entry point
├── auth/                   # Authentication logic
├── database/               # Database connection
├── handlers/               # HTTP request handlers
├── models/                 # Data models
├── routes/                 # Route definitions
├── services/               # Business logic
├── middleware/             # HTTP middleware
├── config/                 # Configuration management
└── test/                   # Test files
```

### Testing Strategy

- **Unit Tests**: Individual function testing
- **Integration Tests**: API endpoint testing
- **Database Tests**: Data layer testing
- **Mock Services**: External service mocking

### Code Quality

- **Go Standards**: Following Go best practices
- **Code Reviews**: Mandatory peer reviews
- **Static Analysis**: Automated code quality checks
- **Documentation**: Comprehensive code documentation

This backend design provides a robust, scalable foundation for the TripLink logistics platform, supporting both web and mobile clients with comprehensive freight management capabilities.