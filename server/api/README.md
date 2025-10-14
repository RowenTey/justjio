# JustJio API Server 🎉

> Production-ready REST API server for JustJio - A social event planning platform

![server-landing](../../client/public/assets/JustJio-Server.gif)

## 📋 Table of Contents

- [Features](#-features)
- [Technology Stack](#-technology-stack)
- [Architecture](#-architecture)
- [Getting Started](#-getting-started)
- [Project Structure](#-project-structure)
- [Testing](#-testing)
- [API Documentation](#-api-documentation)
- [Database](#-database)
- [Environment Variables](#-environment-variables)

## ✨ Features

- 🔐 **JWT Authentication** - Secure user authentication with Google OAuth integration
- 👥 **User Management** - Friend system, user profiles, and social connections
- 🎪 **Event Management** - Create, join, and manage social events (rooms)
- 💰 **Bill Splitting** - Smart bill consolidation and payment tracking
- 📍 **Location Services** - Google Maps integration for venue search
- 🔔 **Push Notifications** - Web push notifications via VAPID
- 📊 **Monitoring** - Prometheus metrics and structured logging

## 🚀 Technology Stack

**Core Framework:**

- [Fiber v2](https://gofiber.io/) - Express-inspired web framework
- [GORM](https://gorm.io/) - ORM with PostgreSQL support
- Go 1.24.8

**Infrastructure:**

- PostgreSQL 15+ - Primary database
- Apache Kafka - Message queue for async processing
- Docker & Docker Compose - Containerization

**External Services:**

- Google OAuth 2.0 - Authentication
- Google Maps API - Venue search and location
- SMTP2GO - Email delivery
- Web Push (VAPID) - Browser notifications

## 🏗 Architecture

This project follows a **clean 3-layer architecture** for maintainability and testability:

```
┌─────────────┐
│   Handler   │  ← HTTP layer (Fiber routes, request/response)
└──────┬──────┘
       │
┌──────▼──────┐
│   Service   │  ← Business logic layer
└──────┬──────┘
       │
┌──────▼──────┐
│ Repository  │  ← Data access layer (GORM, SQL)
└──────┬──────┘
       │
┌──────▼──────┐
│  Database   │  ← PostgreSQL
└─────────────┘
```

**Layer Responsibilities:**

- **Handlers**: Request validation, HTTP status codes, response formatting
- **Services**: Business logic, transactions, external API calls
- **Repositories**: Database queries, CRUD operations, data mapping

## 🛠 Getting Started

### Prerequisites

- Go 1.24+ ([Download](https://go.dev/dl/))
- PostgreSQL 15+ ([Download](https://www.postgresql.org/download/))
- Docker & Docker Compose (optional, for containerized setup)
- [Air](https://github.com/cosmtrek/air) (optional, for hot reload)

### Quick Start

1. **Clone the repository**

   ```bash
   git clone https://github.com/RowenTey/JustJio.git
   cd JustJio/server/api
   ```

2. **Install dependencies**

   ```bash
   go mod download
   ```

3. **Set up environment variables**

   ```bash
   cp .env.example .env
   # Edit .env with your configuration (see Environment Variables section)
   ```

4. **Run the server**

   With Air (hot reload):

   ```bash
   air dev
   ```

   Or standard Go:

   ```bash
   go run main.go dev
   ```

The API will be available at [`http://localhost:8080`](http://localhost:8080)

## 📂 Project Structure

```
server/api/
├── config/              # Configuration management
├── database/            # Database connection and setup
├── docs/                # Swagger/OpenAPI documentation
├── dto/                 # Data Transfer Objects
│   ├── request/         # API request DTOs
│   └── response/        # API response DTOs
├── handlers/            # HTTP request handlers (Controller layer)
├── middleware/          # HTTP middleware
├── migrations/          # Database migrations
│   └── *.sql            # SQL migration files
├── model/               # Database models (GORM entities)
├── repository/          # Data access layer
├── router/              # API routing configuration
├── services/            # Business logic layer
├── tests/               # Test utilities and fixtures
├── utils/               # Utility functions
├── worker/              # Background workers
├── main.go              # Application entry point
├── Dockerfile           # Docker image definition
├── .air.toml            # Air configuration (hot reload)
├── .env.example         # Environment variables template
├── go.mod               # Go module definition
└── README.md            # This file
```

## 🧪 Testing

### Running Tests

**Run all tests:**

```bash
go test ./... -v
```

**Run tests with coverage:**

```bash
go test ./... -coverprofile=coverage.out -covermode=atomic
go tool cover -html=coverage.out
```

**Run tests in parallel:**

```bash
# Use all CPU cores
go test ./... -parallel $(nproc) -p 1 -v
```

**Run specific test suite:**

```bash
# Repository tests
go test ./repository/... -v

# Service tests
go test ./services/... -v

# Handler tests
go test ./handlers/... -v
```

**Run specific test:**

```bash
# Run a single test
go test ./handlers/... -run TestUserHandlerSuite/TestGetUser_Success -v

# Run all tests matching pattern
go test ./... -run "TestCreate" -v
```

**Generate test report:**

```bash
go test ./... -json > test-report.json
```

### Test Structure

Tests use the [testify/suite](https://pkg.go.dev/github.com/stretchr/testify/suite) framework:

```go
type UserServiceTestSuite struct {
    suite.Suite
    db          *gorm.DB
    userService *UserService
}

func (s *UserServiceTestSuite) SetupTest() {
    // Setup before each test
}

func (s *UserServiceTestSuite) TestCreateUser_Success() {
    // Test implementation
}
```

Integration tests use **Testcontainers** for real PostgreSQL and Kafka instances.

## 📚 API Documentation

### Swagger UI

Interactive API documentation is available at:

```
http://localhost:8080/docs
```

### OpenAPI Specification

The OpenAPI 3.0 specification is available at:

- JSON: `docs/swagger.json`
- Embedded: `/swagger/doc.json`

### Generate/Update Swagger Docs

```bash
# Install swag
go install github.com/swaggo/swag/cmd/swag@latest

# Generate docs
swag init

# Format swagger comments
swag fmt
```

### API Endpoints Overview

#### Authentication

- `POST /v1/auth/login` - User login
- `POST /v1/auth/signup` - User registration
- `POST /v1/auth/verify` - OTP verification
- `POST /v1/auth/send-otp` - Send OTP email
- `POST /v1/auth/reset-password` - Reset password

#### Users

- `GET /v1/users/:id` - Get user profile
- `PATCH /v1/users/:id/username` - Update username
- `GET /v1/users/:id/friends` - Get friend list
- `POST /v1/users/:id/friends` - Send friend request
- `DELETE /v1/users/:id/friends/:friendId` - Remove friend

#### Rooms (Events)

- `GET /v1/rooms` - List user's rooms
- `POST /v1/rooms` - Create new room
- `GET /v1/rooms/:id` - Get room details
- `PATCH /v1/rooms/:id` - Edit room
- `POST /v1/rooms/:id` - Invite users
- `PATCH /v1/rooms/:id/join` - Join room
- `PATCH /v1/rooms/:id/leave` - Leave room
- `PATCH /v1/rooms/:id/close` - Close room

#### Bills

- `POST /v1/bills` - Create bill
- `GET /v1/bills` - Get bills by room
- `POST /v1/bills/consolidate` - Consolidate bills

#### Messages

- `GET /v1/rooms/:roomId/messages` - Get messages
- `POST /v1/rooms/:roomId/messages` - Send message

## 💾 Database

### Schema Overview

**Core Tables:**

- `users` - User accounts and profiles
- `rooms` - Events/gatherings
- `room_users` - Room membership (many-to-many)
- `room_invites` - Room invitations
- `user_friends` - Friendship connections (many-to-many)
- `friend_requests` - Pending friend requests
- `bills` - Bill splitting records
- `transactions` - Payment tracking
- `messages` - Chat messages
- `notifications` - Push notifications
- `subscriptions` - Push notification subscriptions

**Views:**

- `user_non_friends` - Materialized view for friend search

### Migrations

Database migrations use [golang-migrate](https://github.com/golang-migrate/migrate):

## 🔐 Environment Variables

Create a `.env` file from `.env.example` and configure:

### Server Configuration

```bash
PORT=8080                        # API server port
```

### Database

```bash
POSTGRES_USER=postgres           # PostgreSQL username
POSTGRES_PASSWORD=yourpassword   # PostgreSQL password
POSTGRES_HOST=localhost          # PostgreSQL host
POSTGRES_PORT=5432               # PostgreSQL port
POSTGRES_DB=justjio              # Database name
```

### Authentication

```bash
JWT_SECRET=your-secret-key       # JWT signing secret (min 32 chars)
ADMIN_EMAIL=admin@justjio.com    # Admin email for system operations
```

### Google OAuth & Maps

```bash
GOOGLE_CLIENT_ID=your-client-id                    # Google OAuth Client ID
GOOGLE_CLIENT_SECRET=your-client-secret            # Google OAuth Client Secret
GOOGLE_REDIRECT_URL=http://localhost:3000/callback # OAuth redirect URL
GOOGLE_MAPS_API_KEY=your-maps-api-key             # Google Maps/Places API key
```

### Kafka (Message Queue)

```bash
KAFKA_HOST=localhost             # Kafka broker host
KAFKA_PORT=9092                  # Kafka broker port
KAFKA_TOPIC_PREFIX=justjio       # Topic prefix for namespacing
```

### Email (SMTP2GO)

```bash
SMTP2GO_API_KEY=your-api-key     # SMTP2GO API key for email delivery
```

### Web Push Notifications (VAPID)

```bash
VAPID_EMAIL=mailto:admin@justjio.com  # VAPID email
VAPID_PUBLIC_KEY=your-public-key       # VAPID public key
VAPID_PRIVATE_KEY=your-private-key     # VAPID private key
```

### CORS

```bash
ALLOWED_ORIGINS=http://localhost:3000  # Comma-separated allowed origins
```

### Generate VAPID Keys

```bash
# Install web-push CLI
npm install -g web-push

# Generate VAPID keys
web-push generate-vapid-keys
```
